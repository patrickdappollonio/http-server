package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signMethod(signingKey []byte) func(*jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		hm, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || hm.Hash != jwt.SigningMethodHS256.Hash {
			// This error will be caught by the err check below
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return signingKey, nil
	}
}

// ValidateJWTHS256 validates a JWT token using the HS256 algorithm,
// the token can be passed in the "Authorization" header or in the
// "token" query parameter.
func ValidateJWTHS256(warnFunctionf func(string, ...interface{}), loggedInFunction func(string), jwtSigningKey string, validateTimedJWT bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var claims jwt.MapClaims

			// Get the token from the "Authorization" header or from the "token" query parameter
			token := firstNonEmpty(
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "),
				r.URL.Query().Get("token"),
			)

			// Check if the token is not empty
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check for errors during parsing (includes signature validation errors and method mismatch errors)
			tkn, err := jwt.ParseWithClaims(token, &claims, signMethod([]byte(jwtSigningKey)))
			if err != nil {
				warnFunctionf("error parsing token for URL %q: %s", r.URL.Path, err.Error())
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Basic token validity check (mainly signature and structure if no options passed)
			if !tkn.Valid {
				// This case should ideally not be hit if err is nil above, unless custom validation
				// options were passed which they are not. Including defensively.
				warnFunctionf("JWT token basic validation failed: invalid token for url: %s", r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if validateTimedJWT {
				now := time.Now()

				// Check for missing 'exp' claim first
				if _, ok := claims["exp"]; !ok {
					warnFunctionf("JWT token validation failed: missing 'exp' claim for url: %s", r.URL.Path)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// Claim exists, now try to get it as a NumericDate
				exp, err := claims.GetExpirationTime()
				if err != nil {
					warnFunctionf("JWT token validation failed: invalid 'exp' claim value for url: %s: %s", r.URL.Path, err.Error())
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// Now compare the time directly
				if now.After(exp.Time) {
					warnFunctionf("JWT token validation failed: token expired for url: %s", r.URL.Path)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// We check for 'iat' existence first
				if _, ok := claims["iat"]; !ok {
					warnFunctionf("JWT token validation failed: missing 'iat' claim for url: %s", r.URL.Path)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// Claim exists, now try to get it as a NumericDate
				iss, err := claims.GetIssuedAt()
				if err != nil {
					warnFunctionf("JWT token validation failed: invalid 'iat' claim value for url: %s: %s", r.URL.Path, err.Error())
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// Now compare the time directly
				if now.Before(iss.Time) {
					warnFunctionf("JWT token validation failed: token issued in the future for url: %s", r.URL.Path)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}

			// Logging successful authentication
			if user := claims["sub"]; user != nil {
				s := fmt.Sprintf("JWT auth passed for url %q: user: %q", r.URL.Path, user)

				if issuer := claims["iss"]; issuer != nil {
					s += fmt.Sprintf(" (issuer: %q)", issuer)
				}

				loggedInFunction(s)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		v = strings.TrimSpace(v)

		if v != "" {
			return v
		}
	}

	return ""
}
