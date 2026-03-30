package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/caddyserver/certmagic"
)

// TLSMode represents the server's TLS operating mode.
type TLSMode string

const (
	TLSModeOff  TLSMode = "off"
	TLSModeBYO  TLSMode = "byo"
	TLSModeAuto TLSMode = "auto"
)

// IsTLSEnabled returns true if TLS is active in any mode.
func (s *Server) IsTLSEnabled() bool {
	return s.activeTLSMode != TLSModeOff && s.activeTLSMode != ""
}

// ActiveTLSMode returns the current TLS mode.
func (s *Server) ActiveTLSMode() TLSMode {
	return s.activeTLSMode
}

// loadAndStoreCert loads the certificate and key from disk, validates
// expiry, and stores the result in the atomic pointer.
func (s *Server) loadAndStoreCert() error {
	cert, err := tls.LoadX509KeyPair(s.TLSCert, s.TLSKey)
	if err != nil {
		return fmt.Errorf("unable to load TLS certificate and key: %w", err)
	}

	// Parse the leaf certificate for metadata and expiry checking
	if cert.Leaf == nil {
		parsed, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return fmt.Errorf("unable to parse TLS certificate: %w", err)
		}
		cert.Leaf = parsed
	}

	// Check if the certificate is not yet valid or has expired
	now := time.Now()
	if now.Before(cert.Leaf.NotBefore) {
		return fmt.Errorf("TLS certificate is not yet valid (valid from %s)", cert.Leaf.NotBefore.Format(time.RFC3339))
	}
	if now.After(cert.Leaf.NotAfter) {
		return fmt.Errorf("TLS certificate expired on %s", cert.Leaf.NotAfter.Format(time.RFC3339))
	}

	// Warn if the certificate expires within 30 days
	if time.Until(cert.Leaf.NotAfter) < 30*24*time.Hour {
		s.printWarningf("TLS certificate expires in less than 30 days (expires: %s)", cert.Leaf.NotAfter.Format(time.RFC3339))
	}

	s.certPointer.Store(&cert)
	return nil
}

// reloadCert reloads the certificate. In BYO mode, it reloads from disk.
// In auto mode, it triggers a certmagic renewal check.
func (s *Server) reloadCert() error { //nolint:contextcheck // reload is triggered by signals and HTTP handlers which don't have a meaningful context
	if s.activeTLSMode == TLSModeAuto && s.certmagicConfig != nil {
		//nolint:wrapcheck // certmagic errors are already descriptive
		return s.certmagicConfig.ManageSync(context.Background(), []string{s.Hostname})
	}
	return s.loadAndStoreCert()
}

// getCertificate is the tls.Config.GetCertificate callback that returns
// the current certificate. In BYO mode, reads from the atomic pointer.
// In auto mode, delegates to certmagic.
func (s *Server) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if s.activeTLSMode == TLSModeAuto && s.certmagicConfig != nil {
		//nolint:wrapcheck // certmagic errors are already descriptive
		return s.certmagicConfig.GetCertificate(hello)
	}
	cert := s.certPointer.Load()
	if cert == nil {
		return nil, errors.New("no TLS certificate loaded")
	}
	return cert, nil
}

// setupAutoTLS configures certmagic for automatic certificate provisioning
// via Let's Encrypt. It synchronously obtains the certificate at startup.
func (s *Server) setupAutoTLS(ctx context.Context) error {
	certmagic.DefaultACME.Email = s.TLSEmail
	certmagic.DefaultACME.Agreed = true

	// Default storage to .certmagic/ inside the served directory,
	// so certs persist alongside the content and work in containers
	// with mounted volumes. Override with --tls-cache-dir.
	cacheDir := s.TLSCacheDir
	if cacheDir == "" {
		cacheDir = filepath.Join(s.Path, ".certmagic")
	}
	certmagic.Default.Storage = &certmagic.FileStorage{Path: cacheDir}

	magic := certmagic.NewDefault()

	fmt.Fprintf(s.LogOutput, "Provisioning TLS certificate for %q via Let's Encrypt...\n", s.Hostname)

	if err := magic.ManageSync(ctx, []string{s.Hostname}); err != nil {
		return fmt.Errorf("unable to provision TLS certificate for %q: %w", s.Hostname, err)
	}

	s.certmagicConfig = magic
	return nil
}

// certMetadata returns metadata about the currently loaded certificate.
func (s *Server) certMetadata() map[string]any {
	var leaf *x509.Certificate

	if s.activeTLSMode == TLSModeAuto && s.certmagicConfig != nil {
		// In auto mode, get cert from certmagic
		cert, err := s.certmagicConfig.GetCertificate(&tls.ClientHelloInfo{ServerName: s.Hostname})
		if err == nil && cert != nil && cert.Leaf != nil {
			leaf = cert.Leaf
		}
	} else {
		// In BYO mode, get cert from atomic pointer
		cert := s.certPointer.Load()
		if cert != nil && cert.Leaf != nil {
			leaf = cert.Leaf
		}
	}

	if leaf == nil {
		return nil
	}

	sans := make([]string, 0, len(leaf.DNSNames)+len(leaf.IPAddresses))
	sans = append(sans, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		sans = append(sans, ip.String())
	}

	return map[string]any{
		"tls_mode":            string(s.activeTLSMode),
		"tls_cert_subject":    leaf.Subject.CommonName,
		"tls_cert_sans":       sans,
		"tls_cert_issuer":     leaf.Issuer.CommonName,
		"tls_cert_not_after":  leaf.NotAfter.Format(time.RFC3339),
		"tls_cert_not_before": leaf.NotBefore.Format(time.RFC3339),
	}
}

// tlsInfoHandler handles GET requests to /_/tls, returning certificate
// metadata as JSON.
func (s *Server) tlsInfoHandler(w http.ResponseWriter, _ *http.Request) {
	meta := s.certMetadata()
	if meta == nil {
		http.Error(w, "no certificate loaded", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//nolint:errchkjson // best-effort JSON response to HTTP client
	json.NewEncoder(w).Encode(meta)
}

// tlsReloadHandler handles POST requests to /_/tls/reload, triggering
// a certificate reload from disk.
func (s *Server) tlsReloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.reloadCert(); err != nil { //nolint:contextcheck // HTTP handler reload has no propagatable context
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		//nolint:errchkjson // best-effort JSON error response
		json.NewEncoder(w).Encode(map[string]any{
			"reloaded": false,
			"error":    err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//nolint:errchkjson // best-effort JSON response
	json.NewEncoder(w).Encode(map[string]any{
		"reloaded": true,
	})
}
