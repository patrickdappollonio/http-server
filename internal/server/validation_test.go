package server

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateTLS_BothCertAndKey(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  &bytes.Buffer{},
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if s.ActiveTLSMode() != TLSModeBYO {
		t.Errorf("expected TLSModeBYO, got %s", s.ActiveTLSMode())
	}
}

func TestValidateTLS_OnlyCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		LogOutput:  &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when only cert is provided")
	}
}

func TestValidateTLS_OnlyKey(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSKey:     keyPath,
		LogOutput:  &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when only key is provided")
	}
}

func TestValidateTLS_PortConflict(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:        8080, // non-default port conflicts with TLS
		Path:        dir,
		ETagMaxSize: "5M",
		TLSCert:     certPath,
		TLSKey:      keyPath,
		Hostname:    "localhost",
		HTTPPort:    80,
		HTTPSPort:   443,
		LogOutput:   &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when --port is set with TLS")
	}
}

func TestValidateTLS_HostnameRequired(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		LogOutput:  &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when hostname is missing with TLS")
	}
}

func TestValidateTLS_SameHTTPAndHTTPSPort(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8443,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when HTTP and HTTPS ports are the same")
	}
}

func TestValidateTLS_ExpiredCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateExpiredTestCert(t, certPath, keyPath)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error for expired certificate")
	}
}

func TestValidateTLS_ExpiringSoonCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 15*24*time.Hour) // expires in 15 days

	buf := &bytes.Buffer{}
	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  buf,
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error for expiring-soon cert, got: %v", err)
	}

	// Should contain a warning
	if !bytes.Contains(buf.Bytes(), []byte("expires in less than 30 days")) {
		t.Error("expected expiry warning in output")
	}
}

func TestValidateTLS_MismatchedCertKey(t *testing.T) {
	dir := t.TempDir()
	certPath1 := filepath.Join(dir, "cert1.pem")
	keyPath1 := filepath.Join(dir, "key1.pem")
	generateTestCert(t, certPath1, keyPath1, 365*24*time.Hour)

	certPath2 := filepath.Join(dir, "cert2.pem")
	keyPath2 := filepath.Join(dir, "key2.pem")
	generateTestCert(t, certPath2, keyPath2, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath1,
		TLSKey:     keyPath2,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error for mismatched cert and key")
	}
}

func TestValidateTLS_CertFileHiding(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	buf := &bytes.Buffer{}
	s := &Server{
		Port:       5000,
		Path:       dir, // cert is inside the served directory
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  buf,
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(s.forbiddenAbsPaths) != 2 {
		t.Errorf("expected 2 forbidden paths, got %d", len(s.forbiddenAbsPaths))
	}

	// Should contain a warning about files being inside served dir
	if !bytes.Contains(buf.Bytes(), []byte("inside the served directory")) {
		t.Error("expected warning about cert files inside served directory")
	}
}

func TestValidateTLS_CertOutsideServedDir(t *testing.T) {
	servedDir := t.TempDir()
	certDir := t.TempDir()
	certPath := filepath.Join(certDir, "cert.pem")
	keyPath := filepath.Join(certDir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	// Create a file in servedDir so it's a valid directory
	os.WriteFile(filepath.Join(servedDir, "index.html"), []byte("hello"), 0644)

	s := &Server{
		Port:       5000,
		Path:       servedDir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   8080,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  &bytes.Buffer{},
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(s.forbiddenAbsPaths) != 0 {
		t.Errorf("expected 0 forbidden paths when cert is outside served dir, got %d", len(s.forbiddenAbsPaths))
	}
}

func TestValidateTLS_NoTLSFlags(t *testing.T) {
	dir := t.TempDir()

	s := &Server{
		Port:        5000,
		Path:        dir,
		ETagMaxSize: "5M",
		HTTPPort:    80,
		HTTPSPort:   443,
		LogOutput:   &bytes.Buffer{},
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error without TLS flags, got: %v", err)
	}

	if s.IsTLSEnabled() {
		t.Error("expected TLS to be disabled when no flags set")
	}
}

func TestValidateTLS_HTTPPortZero(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:       5000,
		Path:       dir,
		ETagMaxSize: "5M",
		TLSCert:    certPath,
		TLSKey:     keyPath,
		HTTPPort:   0,
		HTTPSPort:  8443,
		Hostname:   "localhost",
		LogOutput:  &bytes.Buffer{},
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error with --http-port 0, got: %v", err)
	}
}

func TestValidateTLS_HostnameAloneTriggersAutoMode(t *testing.T) {
	dir := t.TempDir()

	s := &Server{
		Port:        5000,
		Path:        dir,
		ETagMaxSize: "5M",
		Hostname:    "example.com",
		HTTPPort:    80,
		HTTPSPort:   443,
		LogOutput:   &bytes.Buffer{},
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if s.ActiveTLSMode() != TLSModeAuto {
		t.Errorf("expected TLSModeAuto, got %s", s.ActiveTLSMode())
	}
}

func TestValidateTLS_HostnameWithCertKeyTriggersBYO(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	generateTestCert(t, certPath, keyPath, 365*24*time.Hour)

	s := &Server{
		Port:        5000,
		Path:        dir,
		ETagMaxSize: "5M",
		TLSCert:     certPath,
		TLSKey:      keyPath,
		Hostname:    "localhost",
		HTTPPort:    8080,
		HTTPSPort:   8443,
		LogOutput:   &bytes.Buffer{},
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if s.ActiveTLSMode() != TLSModeBYO {
		t.Errorf("expected TLSModeBYO, got %s", s.ActiveTLSMode())
	}
}

func TestValidateTLS_AutoModePortConflict(t *testing.T) {
	dir := t.TempDir()

	s := &Server{
		Port:        8080, // non-default port conflicts with auto TLS
		Path:        dir,
		ETagMaxSize: "5M",
		Hostname:    "example.com",
		HTTPPort:    80,
		HTTPSPort:   443,
		LogOutput:   &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when --port is set with auto TLS")
	}
}

func TestValidateTLS_TLSEmailWithoutTLS(t *testing.T) {
	dir := t.TempDir()

	s := &Server{
		Port:        5000,
		Path:        dir,
		ETagMaxSize: "5M",
		HTTPPort:    80,
		HTTPSPort:   443,
		TLSEmail:    "test@example.com",
		LogOutput:   &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when --tls-email is set without TLS")
	}
}

func TestValidateTLS_AutoModeHTTPPortZeroRejected(t *testing.T) {
	dir := t.TempDir()

	s := &Server{
		Port:        5000,
		Path:        dir,
		ETagMaxSize: "5M",
		Hostname:    "example.com",
		HTTPPort:    0,
		HTTPSPort:   443,
		LogOutput:   &bytes.Buffer{},
	}

	err := s.Validate()
	if err == nil {
		t.Fatal("expected error when --http-port 0 in auto TLS mode")
	}
}

func TestValidateTLS_AutoModeNonStandardHTTPPort(t *testing.T) {
	dir := t.TempDir()
	buf := &bytes.Buffer{}

	s := &Server{
		Port:        5000,
		Path:        dir,
		ETagMaxSize: "5M",
		Hostname:    "example.com",
		HTTPPort:    8080,
		HTTPSPort:   443,
		LogOutput:   buf,
	}

	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("ACME HTTP-01 challenges require port 80")) {
		t.Error("expected warning about non-standard HTTP port for ACME challenges")
	}
}
