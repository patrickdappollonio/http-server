package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- firstNonEmpty tests ---
func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"   ", "\tfoo", "bar"}, "foo"},
		{[]string{"", "  ", ""}, ""},
		{[]string{"first", "second"}, "first"},
	}
	for _, tc := range tests {
		if got := firstNonEmpty(tc.input...); got != tc.want {
			t.Errorf("firstNonEmpty(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

// --- signMethod tests ---
func TestSignMethod_ValidHS256(t *testing.T) {
	key := []byte("secret")
	validator := signMethod(key)

	// Create a dummy token with HS256
	token := &jwt.Token{Method: jwt.SigningMethodHS256, Header: map[string]interface{}{"alg": "HS256"}}
	got, err := validator(token)
	if err != nil {
		t.Fatalf("expected no error for HS256, got %v", err)
	}
	if sk := got.([]byte); string(sk) != string(key) {
		t.Errorf("signMethod returned %v; want %v", sk, key)
	}
}

func TestSignMethod_InvalidMethod(t *testing.T) {
	key := []byte("secret")
	validator := signMethod(key)

	// Simulate RS256 token
	token := &jwt.Token{Method: jwt.SigningMethodRS256, Header: map[string]interface{}{"alg": "RS256"}}
	_, err := validator(token)
	if err == nil || !strings.Contains(err.Error(), "unexpected signing method") {
		t.Errorf("expected unexpected signing method error, got %v", err)
	}
}

// --- middleware tests ---
func TestValidateJWTHS256_NoToken(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)

	var warns []string
	var logs []string

	// Middleware with timing disabled
	mw := ValidateJWTHS256(
		func(fmtStr string, args ...interface{}) { warns = append(warns, fmt.Sprintf(fmtStr, args...)) },
		func(msg string) { logs = append(logs, msg) },
		"secret",
		false,
	)

	// next handler should never be called
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", rr.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Error("next handler was called on missing token")
	}
	if len(warns) != 0 || len(logs) != 0 {
		t.Error("no warnings or logs expected when token is missing")
	}
}

func TestValidateJWTHS256_BadSignature(t *testing.T) {
	// Create a token signed with the wrong key
	wrongKey := []byte("wrong")
	claims := jwt.MapClaims{"sub": "alice"}
	tkn, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(wrongKey)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+tkn)

	var warns []string
	var logs []string

	mw := ValidateJWTHS256(
		func(fmtStr string, args ...interface{}) { warns = append(warns, fmt.Sprintf(fmtStr, args...)) },
		func(msg string) { logs = append(logs, msg) },
		"secret", // correct key
		false,
	)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d on bad signature", rr.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Error("next handler was called on invalid signature")
	}
	if len(warns) == 0 {
		t.Error("expected a warning for parse error")
	}
}

func TestValidateJWTHS256_ValidToken_NoTiming(t *testing.T) {
	// Sign a valid token
	signingKey := []byte("secret")
	claims := jwt.MapClaims{"sub": "bob", "iss": "issuer"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknStr, err := token.SignedString(signingKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/home", nil)
	req.Header.Set("Authorization", "Bearer "+tknStr)

	var warns []string
	var logs []string

	mw := ValidateJWTHS256(
		func(fmtStr string, args ...interface{}) { warns = append(warns, fmt.Sprintf(fmtStr, args...)) },
		func(msg string) { logs = append(logs, msg) },
		string(signingKey),
		false, // timing disabled
	)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d; want %d on valid token", rr.Code, http.StatusOK)
	}
	if !nextCalled {
		t.Error("next handler was not called on valid token")
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "JWT auth passed for url") {
		t.Errorf("expected a login log, got %v", logs)
	}
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestValidateJWTHS256_ExpiredToken_WithTiming(t *testing.T) {
	// Sign a token that expired an hour ago
	signingKey := []byte("secret")
	claims := jwt.MapClaims{
		"sub": "carol",
		"exp": jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		"iat": jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknStr, err := token.SignedString(signingKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/timed", nil)
	req.Header.Set("Authorization", "Bearer "+tknStr)

	var warns []string
	mw := ValidateJWTHS256(
		func(fmtStr string, args ...interface{}) { warns = append(warns, fmt.Sprintf(fmtStr, args...)) },
		func(msg string) { /* ignore */ },
		string(signingKey),
		true, // timing enabled
	)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d on expired token", rr.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Error("next handler should not be called on expired token")
	}
	if len(warns) == 0 || !strings.Contains(warns[0], "token has invalid claims") {
		t.Errorf("expected a parse‐error warning about expired claims, got %v", warns)
	}
}
