package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// TLSMode represents the server's TLS operating mode.
type TLSMode string

const (
	TLSModeOff  TLSMode = "off"
	TLSModeBYO  TLSMode = "byo"
	TLSModeAuto TLSMode = "auto" // Phase 2: certmagic
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

// reloadCert reloads the certificate and key from disk.
func (s *Server) reloadCert() error {
	return s.loadAndStoreCert()
}

// getCertificate is the tls.Config.GetCertificate callback that returns
// the current certificate from the atomic pointer.
func (s *Server) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert := s.certPointer.Load()
	if cert == nil {
		return nil, errors.New("no TLS certificate loaded")
	}
	return cert, nil
}

// certMetadata returns metadata about the currently loaded certificate.
func (s *Server) certMetadata() map[string]any {
	cert := s.certPointer.Load()
	if cert == nil || cert.Leaf == nil {
		return nil
	}

	leaf := cert.Leaf

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

	if err := s.reloadCert(); err != nil {
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
