package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSigningKey = "a_super_secret_test_key_that_is_long_enough"

// Helper function to create a signed HS256 token for testing
func createTestToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSigningKey))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	return tokenString
}

// Helper function to create a token with a specific signing method
// Used here to create a token claiming a non-HS256 method but signed with HS256 key
// to trigger the method validation error in the middleware.
func createTokenWithMethod(t *testing.T, claims jwt.MapClaims, method jwt.SigningMethod, key interface{}) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("Failed to sign token with method %s: %v", method.Alg(), err)
	}
	return tokenString
}

// Mocks for the functions and handlers

// mockNextHandler is a simple handler that records if it was called.
type mockNextHandler struct {
	called bool
}

func (m *mockNextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	w.WriteHeader(http.StatusOK) // Simulate the next handler doing something
}

// mockWarnFunction records the last message format and args it was called with.
type mockWarnFunction struct {
	msgFormat string
	args      []interface{}
	called    bool
}

func (m *mockWarnFunction) Call(msg string, args ...interface{}) {
	m.msgFormat = msg
	m.args = args
	m.called = true
}

// mockLoggedInFunction records the last message it was called with.
type mockLoggedInFunction struct {
	lastMsg string
	called  bool
}

func (m *mockLoggedInFunction) Call(msg string) {
	m.lastMsg = msg
	m.called = true
}

// --- Helper function for test setup and execution ---

type testMocks struct {
	next     *mockNextHandler
	warn     *mockWarnFunction
	loggedIn *mockLoggedInFunction
}

// setupAndRunMiddleware creates mocks, recorder, middleware, and runs the handler.
func setupAndRunMiddleware(t *testing.T, req *http.Request, jwtKey string, validateTimed bool) (*httptest.ResponseRecorder, *testMocks) {
	t.Helper() // Mark this as a helper function

	mockNext := &mockNextHandler{}
	mockWarn := &mockWarnFunction{}
	mockLogged := &mockLoggedInFunction{}
	rr := httptest.NewRecorder()

	middleware := ValidateJWTHS256(mockWarn.Call, mockLogged.Call, jwtKey, validateTimed)
	handler := middleware(mockNext)

	handler.ServeHTTP(rr, req)

	return rr, &testMocks{next: mockNext, warn: mockWarn, loggedIn: mockLogged}
}

// --- Assertion Helpers ---

// assertSuccess checks for a successful response and correct mock calls.
// expectedUser and expectedIssuer are the values expected in the loggedInFunction message.
func assertSuccess(t *testing.T, rr *httptest.ResponseRecorder, mocks *testMocks, reqPath string, expectedUser, expectedIssuer string) {
	t.Helper() // Mark this as a helper function

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK (%d), got %d", http.StatusOK, rr.Code)
	}
	if !mocks.next.called {
		t.Error("Expected next handler to be called, but it wasn't")
	}
	if mocks.warn.called {
		t.Errorf("Expected warn function not to be called, but it was with message: %q", mocks.warn.msgFormat)
	}

	if expectedUser != "" {
		if !mocks.loggedIn.called {
			t.Error("Expected loggedIn function to be called, but it wasn't")
		}
		// Construct the *expected* message based on how the middleware formats it
		expectedMsg := fmt.Sprintf(`JWT auth passed for url "%s": user: "%s"`, reqPath, expectedUser)
		if expectedIssuer != "" {
			expectedMsg += fmt.Sprintf(` (issuer: "%s")`, expectedIssuer)
		}

		if mocks.loggedIn.lastMsg != expectedMsg {
			t.Errorf("Expected loggedIn msg %q, got %q", expectedMsg, mocks.loggedIn.lastMsg)
		}
	} else {
		if mocks.loggedIn.called {
			t.Error("Expected loggedIn function not to be called (no 'sub' claim), but it was")
		}
	}
}

// assertUnauthorized checks for an unauthorized response and correct mock calls,
// including the expected warning message format and arguments.
// Pass expectedMsgFormat and expectedArgs only if a warn call is expected.
func assertUnauthorized(t *testing.T, rr *httptest.ResponseRecorder, mocks *testMocks, expectedMsgFormat string, expectedArgs ...interface{}) {
	t.Helper() // Mark this as a helper function

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized (%d), got %d", http.StatusUnauthorized, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "unauthorized" {
		t.Errorf("Expected body %q, got %q", "unauthorized", strings.TrimSpace(rr.Body.String()))
	}
	if mocks.next.called {
		t.Error("Expected next handler not to be called, but it was")
	}
	if mocks.loggedIn.called {
		t.Error("Expected loggedIn function not to be called, but it was")
	}

	// Check warnFunction based on whether a call is expected
	if expectedMsgFormat != "" {
		if !mocks.warn.called {
			t.Error("Expected warn function to be called, but it wasn't")
		}
		if mocks.warn.msgFormat != expectedMsgFormat {
			t.Errorf("Expected warn message format %q, got %q", expectedMsgFormat, mocks.warn.msgFormat)
		}
		// Use reflect.DeepEqual for slice and content comparison
		if !reflect.DeepEqual(mocks.warn.args, expectedArgs) {
			t.Errorf("Expected warn arguments %v, got %v", expectedArgs, mocks.warn.args)
		}
	} else {
		if mocks.warn.called {
			// This covers the "no token" case where warn is not called
			t.Errorf("Expected warn function not to be called, but it was with message format: %q and args %v", mocks.warn.msgFormat, mocks.warn.args)
		}
	}
}

// --- Refactored Test Cases ---

func TestValidateJWTHS256_Success_Header(t *testing.T) {
	claims := jwt.MapClaims{"sub": "testuser", "iss": "testissuer"}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	assertSuccess(t, rr, mocks, reqPath, "testuser", "testissuer")
}

func TestValidateJWTHS256_Success_QueryParam(t *testing.T) {
	claims := jwt.MapClaims{"sub": "queryuser"}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath+"?token="+token, nil)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	assertSuccess(t, rr, mocks, reqPath, "queryuser", "")
}

func TestValidateJWTHS256_Fail_NoToken(t *testing.T) {
	reqPath := "/protected"
	req := httptest.NewRequest("GET", reqPath, nil)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	// No warn message is expected for this case in the current code
	assertUnauthorized(t, rr, mocks, "") // ExpectedMsgFormat is empty
}

func TestValidateJWTHS256_Fail_InvalidTokenFormat(t *testing.T) {
	invalidToken := "this-is-not-a-valid-jwt-token"
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	// Expect error logged by warnFunction about parsing
	expectedMsgFormat := "error parsing token for URL %q: %s"
	// The second arg is the actual error string from jwt.ParseWithClaims, which varies.
	// We can check its type and that it's not empty, or check for a common substring.
	// Checking for the substring "token contains an invalid number of segments" is common.
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, reqPath, testing.Anypoint{}) // Use Anypoint as we can't predict the exact error string
	// Note: Anypoint requires Go 1.22 or later. For older Go, you'd check len(args) and type/substring of the second arg manually.
}

func TestValidateJWTHS256_Fail_WrongSigningKey(t *testing.T) {
	claims := jwt.MapClaims{"sub": "testuser"}
	token := createTestToken(t, claims) // Signed with testSigningKey
	reqPath := "/protected"
	wrongKey := "another_different_secret_key"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Middleware is configured with the *wrong* key. Signature validation fails during ParseWithClaims.
	rr, mocks := setupAndRunMiddleware(t, req, wrongKey, false)

	// Expect error logged by warnFunction about parsing (due to signature mismatch)
	expectedMsgFormat := "error parsing token for URL %q: %s"
	// The second arg is the actual error string from jwt.ParseWithClaims (signature validation error).
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, reqPath, testing.Anypoint{}) // Use Anypoint
}

func TestValidateJWTHS256_Fail_WrongSigningMethod(t *testing.T) {
	claims := jwt.MapClaims{"sub": "testuser"}
	reqPath := "/protected"
	// Create a token claiming RS256 but sign it with the HS256 key bytes.
	// This triggers the method check mismatch in the middleware's ParseWithClaims callback.
	tokenString := createTokenWithMethod(t, claims, jwt.SigningMethodRS256, []byte(testSigningKey))

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Middleware is configured for HS256
	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	// Expect error logged by warnFunction about unexpected signing method
	expectedMsgFormat := "error parsing token for URL %q: %s"
	// The second arg is the error string "unexpected signing method: RS256"
	expectedArgs := []interface{}{reqPath, fmt.Errorf("unexpected signing method: RS256").Error()}
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

func TestValidateJWTHS256_Success_TimedValid(t *testing.T) {
	now := time.Now()
	// Use slightly larger margins for timed tests to avoid flakiness
	claims := jwt.MapClaims{
		"sub": "timeduser",
		"exp": jwt.NewNumericDate(now.Add(5 * time.Minute)),  // Expires in 5 minutes (future)
		"iat": jwt.NewNumericDate(now.Add(-5 * time.Minute)), // Issued 5 minutes ago (past)
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	assertSuccess(t, rr, mocks, reqPath, "timeduser", "")
}

func TestValidateJWTHS256_Fail_TimedExpired(t *testing.T) {
	now := time.Now()
	// Use slightly larger margins for timed tests to avoid flakiness
	claims := jwt.MapClaims{
		"sub": "expireduser",
		"exp": jwt.NewNumericDate(now.Add(-5 * time.Minute)),  // Expires 5 minutes ago (past)
		"iat": jwt.NewNumericDate(now.Add(-10 * time.Minute)), // Issued 10 minutes ago (past)
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	expectedMsgFormat := "JWT token validation failed: token expired for url: %s"
	expectedArgs := []interface{}{reqPath}
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

func TestValidateJWTHS256_Fail_TimedIssuedInFuture(t *testing.T) {
	now := time.Now()
	// Use slightly larger margins for timed tests to avoid flakiness
	claims := jwt.MapClaims{
		"sub": "futureuser",
		"exp": jwt.NewNumericDate(now.Add(10 * time.Minute)), // Expires in 10 minutes (future)
		"iat": jwt.NewNumericDate(now.Add(5 * time.Minute)),  // Issued 5 minutes in the future
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	expectedMsgFormat := "JWT token validation failed: token issued in the future for url: %s"
	expectedArgs := []interface{}{reqPath}
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

func TestValidateJWTHS256_Fail_TimedMissingExpClaim(t *testing.T) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": "noexpuser",
		// "exp" is missing
		"iat": jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Issued 1 hour ago (past)
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	expectedMsgFormat := "JWT token validation failed: missing 'exp' claim for url: %s"
	expectedArgs := []interface{}{reqPath}
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

func TestValidateJWTHS256_Fail_TimedMissingIatClaim(t *testing.T) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": "noiatuser",
		"exp": jwt.NewNumericDate(now.Add(1 * time.Hour)), // Expires in 1 hour (future)
		// "iat" is missing
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	expectedMsgFormat := "JWT token validation failed: missing 'iat' claim for url: %s"
	expectedArgs := []interface{}{reqPath}
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

// Test case: exp claim exists but is wrong type (e.g., a string)
func TestValidateJWTHS256_Fail_TimedInvalidExpClaimType(t *testing.T) {
	invalidExpValue := "not a number"
	claims := jwt.MapClaims{
		"sub": "invalidexpuser",
		"exp": invalidExpValue, // Invalid type for exp
		"iat": jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	expectedMsgFormat := "JWT token validation failed: invalid 'exp' claim value (%v) for url: %s: %s"
	// The third arg is the actual error string from GetExpirationTime, which varies.
	// We can check its type or use Anypoint.
	expectedArgs := []interface{}{invalidExpValue, reqPath, testing.Anypoint{}} // Use Anypoint for error string
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

// Test case: iat claim exists but is wrong type (e.g., a boolean)
func TestValidateJWTHS256_Fail_TimedInvalidIatClaimType(t *testing.T) {
	invalidIatValue := true
	claims := jwt.MapClaims{
		"sub": "invalidiatuser",
		"exp": jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		"iat": invalidIatValue, // Invalid type for iat
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, true) // Enable timed validation

	expectedMsgFormat := "JWT token validation failed: invalid 'iat' claim value (%v) for url: %s: %s"
	// The third arg is the actual error string from GetIssuedAt, which varies.
	expectedArgs := []interface{}{invalidIatValue, reqPath, testing.Anypoint{}} // Use Anypoint for error string
	assertUnauthorized(t, rr, mocks, expectedMsgFormat, expectedArgs...)
}

func TestValidateJWTHS256_TimedDisabled_ExpiredTokenIgnored(t *testing.T) {
	now := time.Now()
	// Use slightly larger margins for timed tests to avoid flakiness
	claims := jwt.MapClaims{
		"sub": "ignoreduser",
		"exp": jwt.NewNumericDate(now.Add(-5 * time.Minute)),  // Expires 5 minutes ago (past)
		"iat": jwt.NewNumericDate(now.Add(-10 * time.Minute)), // Issued 10 minutes ago (past)
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false) // Disable timed validation

	// Should succeed because timed validation is off
	assertSuccess(t, rr, mocks, reqPath, "ignoreduser", "")
}

func TestValidateJWTHS256_HeaderPrecedence(t *testing.T) {
	headerClaims := jwt.MapClaims{"sub": "headeruser"}
	headerToken := createTestToken(t, headerClaims)

	queryClaims := jwt.MapClaims{"sub": "queryuser"}
	queryToken := createTestToken(t, queryClaims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath+"?token="+queryToken, nil)
	req.Header.Set("Authorization", "Bearer "+headerToken) // Header token should be used

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	// Assert that the user from the header token was logged
	assertSuccess(t, rr, mocks, reqPath, "headeruser", "") // Assuming issuer is not expected if not in claims
}

func TestValidateJWTHS256_Success_NoSubClaim(t *testing.T) {
	claims := jwt.MapClaims{
		"iss": "testissuer", // No "sub" claim
	}
	token := createTestToken(t, claims)
	reqPath := "/protected"

	req := httptest.NewRequest("GET", reqPath, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr, mocks := setupAndRunMiddleware(t, req, testSigningKey, false)

	// loggedInFunction should NOT be called because there's no 'sub' claim
	assertSuccess(t, rr, mocks, reqPath, "", "") // ExpectedUser and ExpectedIssuer are empty
}

// --- Test for the firstNonEmpty helper ---

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"First is non-empty", []string{"hello", "world"}, "hello"},
		{"Second is non-empty", []string{"", "world"}, "world"},
		{"Third is non-empty after spaces", []string{"", "  ", " world "}, "world"},
		{"All empty", []string{"", " ", "  "}, ""},
		{"Empty slice", []string{}, ""},
		{"First is non-empty with spaces", []string{" hello ", "world"}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			result := firstNonEmpty(tt.input...)
			if result != tt.expected {
				t.Errorf("Input: %v, Expected: %q, Got: %q", tt.input, tt.expected, result)
			}
		})
	}
}
