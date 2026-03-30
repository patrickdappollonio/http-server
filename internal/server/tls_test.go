package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestCert creates a self-signed certificate and key in PEM format,
// writing them to the specified paths. The certificate is valid for the
// given duration from now.
func generateTestCert(t *testing.T, certPath, keyPath string, validFor time.Duration) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certFile, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("create cert file: %v", err)
	}
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certFile.Close()

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	keyFile, err := os.Create(keyPath)
	if err != nil {
		t.Fatalf("create key file: %v", err)
	}
	pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyFile.Close()
}

// generateExpiredTestCert creates a certificate that expired in the past.
func generateExpiredTestCert(t *testing.T, certPath, keyPath string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "expired"},
		NotBefore:    time.Now().Add(-48 * time.Hour),
		NotAfter:     time.Now().Add(-24 * time.Hour),
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certFile, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("create cert file: %v", err)
	}
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certFile.Close()

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	keyFile, err := os.Create(keyPath)
	if err != nil {
		t.Fatalf("create key file: %v", err)
	}
	pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyFile.Close()
}

func TestTLSMode(t *testing.T) {
	t.Run("default is off", func(t *testing.T) {
		s := &Server{}
		if s.IsTLSEnabled() {
			t.Error("expected TLS to be disabled by default")
		}
		// Zero value of TLSMode is "", which is treated as off
		if mode := s.ActiveTLSMode(); mode != "" && mode != TLSModeOff {
			t.Errorf("expected empty or TLSModeOff, got %s", mode)
		}
	})

	t.Run("byo mode", func(t *testing.T) {
		s := &Server{activeTLSMode: TLSModeBYO}
		if !s.IsTLSEnabled() {
			t.Error("expected TLS to be enabled in BYO mode")
		}
		if s.ActiveTLSMode() != TLSModeBYO {
			t.Errorf("expected TLSModeBYO, got %s", s.ActiveTLSMode())
		}
	})
}

func TestLoadAndStoreCert(t *testing.T) {
	t.Run("valid cert", func(t *testing.T) {
		dir := t.TempDir()
		certPath := filepath.Join(dir, "cert.pem")
		keyPath := filepath.Join(dir, "key.pem")
		generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

		s := &Server{
			TLSCert: certPath,
			TLSKey:  keyPath,
		}

		if err := s.loadAndStoreCert(); err != nil {
			t.Fatalf("loadAndStoreCert() error: %v", err)
		}

		cert := s.certPointer.Load()
		if cert == nil {
			t.Fatal("cert pointer is nil after loading")
		}
		if cert.Leaf == nil {
			t.Fatal("cert leaf is nil after loading")
		}
		if cert.Leaf.Subject.CommonName != "localhost" {
			t.Errorf("expected CN=localhost, got %s", cert.Leaf.Subject.CommonName)
		}
	})

	t.Run("expired cert", func(t *testing.T) {
		dir := t.TempDir()
		certPath := filepath.Join(dir, "cert.pem")
		keyPath := filepath.Join(dir, "key.pem")
		generateExpiredTestCert(t, certPath, keyPath)

		s := &Server{
			TLSCert: certPath,
			TLSKey:  keyPath,
		}

		err := s.loadAndStoreCert()
		if err == nil {
			t.Fatal("expected error for expired cert")
		}
	})

	t.Run("nonexistent cert", func(t *testing.T) {
		s := &Server{
			TLSCert: "/nonexistent/cert.pem",
			TLSKey:  "/nonexistent/key.pem",
		}

		err := s.loadAndStoreCert()
		if err == nil {
			t.Fatal("expected error for nonexistent cert files")
		}
	})

	t.Run("mismatched cert and key", func(t *testing.T) {
		dir := t.TempDir()
		certPath1 := filepath.Join(dir, "cert1.pem")
		keyPath1 := filepath.Join(dir, "key1.pem")
		generateTestCert(t, certPath1, keyPath1, 365*24*time.Hour)

		certPath2 := filepath.Join(dir, "cert2.pem")
		keyPath2 := filepath.Join(dir, "key2.pem")
		generateTestCert(t, certPath2, keyPath2, 365*24*time.Hour)

		s := &Server{
			TLSCert: certPath1,
			TLSKey:  keyPath2, // mismatched key
		}

		err := s.loadAndStoreCert()
		if err == nil {
			t.Fatal("expected error for mismatched cert and key")
		}
	})
}

func TestGetCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		TLSCert: certPath,
		TLSKey:  keyPath,
	}

	if err := s.loadAndStoreCert(); err != nil {
		t.Fatalf("loadAndStoreCert() error: %v", err)
	}

	cert, err := s.getCertificate(&tls.ClientHelloInfo{})
	if err != nil {
		t.Fatalf("getCertificate() error: %v", err)
	}
	if cert == nil {
		t.Fatal("getCertificate() returned nil")
	}
}

func TestGetCertificate_NilPointer(t *testing.T) {
	s := &Server{}

	_, err := s.getCertificate(&tls.ClientHelloInfo{})
	if err == nil {
		t.Fatal("expected error when no cert is loaded")
	}
}

func TestReloadCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		TLSCert: certPath,
		TLSKey:  keyPath,
	}

	// Initial load
	if err := s.loadAndStoreCert(); err != nil {
		t.Fatalf("initial load: %v", err)
	}

	oldCert := s.certPointer.Load()

	// Generate a new cert at the same paths
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	// Reload
	if err := s.reloadCert(); err != nil {
		t.Fatalf("reloadCert() error: %v", err)
	}

	newCert := s.certPointer.Load()
	if newCert == oldCert {
		t.Error("cert pointer should have changed after reload")
	}
}

func TestReloadCert_PreservesOldOnFailure(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		TLSCert: certPath,
		TLSKey:  keyPath,
	}

	if err := s.loadAndStoreCert(); err != nil {
		t.Fatalf("initial load: %v", err)
	}

	oldCert := s.certPointer.Load()

	// Delete the cert file to cause reload failure
	os.Remove(certPath)

	err := s.reloadCert()
	if err == nil {
		t.Fatal("expected error when cert file is missing")
	}

	// Old cert should still be loaded
	currentCert := s.certPointer.Load()
	if currentCert != oldCert {
		t.Error("cert should be preserved after failed reload")
	}
}

func TestCertMetadata(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		TLSCert:       certPath,
		TLSKey:        keyPath,
		activeTLSMode: TLSModeBYO,
	}

	if err := s.loadAndStoreCert(); err != nil {
		t.Fatalf("loadAndStoreCert() error: %v", err)
	}

	meta := s.certMetadata()
	if meta == nil {
		t.Fatal("certMetadata() returned nil")
	}

	if meta["tls_mode"] != "byo" {
		t.Errorf("expected tls_mode=byo, got %v", meta["tls_mode"])
	}
	if meta["tls_cert_subject"] != "localhost" {
		t.Errorf("expected subject=localhost, got %v", meta["tls_cert_subject"])
	}
	sans, ok := meta["tls_cert_sans"].([]string)
	if !ok || len(sans) == 0 {
		t.Error("expected non-empty SANs")
	}
}

func TestTLSInfoHandler(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		TLSCert:       certPath,
		TLSKey:        keyPath,
		activeTLSMode: TLSModeBYO,
	}

	if err := s.loadAndStoreCert(); err != nil {
		t.Fatalf("loadAndStoreCert() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/_/tls", nil)
	rec := httptest.NewRecorder()
	s.tlsInfoHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if result["tls_mode"] != "byo" {
		t.Errorf("expected tls_mode=byo, got %v", result["tls_mode"])
	}
	if result["tls_cert_subject"] != "localhost" {
		t.Errorf("expected subject=localhost, got %v", result["tls_cert_subject"])
	}
}

func TestTLSInfoHandler_NoCert(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/_/tls", nil)
	rec := httptest.NewRecorder()
	s.tlsInfoHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestTLSReloadHandler(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		TLSCert:       certPath,
		TLSKey:        keyPath,
		activeTLSMode: TLSModeBYO,
	}

	if err := s.loadAndStoreCert(); err != nil {
		t.Fatalf("loadAndStoreCert() error: %v", err)
	}

	t.Run("POST reloads cert", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/_/tls/reload", nil)
		rec := httptest.NewRecorder()
		s.tlsReloadHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var result map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if result["reloaded"] != true {
			t.Error("expected reloaded=true")
		}
	})

	t.Run("GET returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/_/tls/reload", nil)
		rec := httptest.NewRecorder()
		s.tlsReloadHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})
}
