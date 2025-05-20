package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// sendUnauthorized sends a 401 response with a warning message.
func sendUnauthorized(w http.ResponseWriter, warnFunctionf func(string, ...any), format string, args ...any) {
	warnFunctionf(format, args...)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

// signMethod returns a key function that validates the signing method is HS256.
func signMethod(signingKey []byte) func(*jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		// Validate the signing method
		hm, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || hm.Hash != jwt.SigningMethodHS256.Hash {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return signingKey, nil
	}
}

// ValidateJWTHS256 validates a JWT token using the HS256 algorithm,
// the token can be passed in the "Authorization" header or in the
// "token" query parameter.
func ValidateJWTHS256(warnFunctionf func(string, ...any), loggedInFunction func(string), jwtSigningKey string, validateTimedJWT bool) func(http.Handler) http.Handler {
	// Cache the signing key bytes
	signingKeyBytes := []byte(jwtSigningKey)
	keyfunc := signMethod(signingKeyBytes)

	// Configure parser options based on validation requirements
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"HS256"}),
	}

	if validateTimedJWT {
		// Add time validation options
		options = append(options,
			jwt.WithLeeway(time.Second*5), // Small leeway for clock skew
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
		)
	}

	// Create a parser with the specified options
	parser := jwt.NewParser(options...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path // Cache path to avoid repeated access
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

			// Parse token with claims using the configured parser
			tkn, err := parser.ParseWithClaims(token, &claims, keyfunc)

			// Handle parsing errors with more specific messages based on error type
			if err != nil {
				// Check specific JWT error types for better error messages
				switch {
				case errors.Is(err, jwt.ErrTokenExpired):
					sendUnauthorized(w, warnFunctionf, "JWT token validation failed: token expired for URL: %s", path)
				case errors.Is(err, jwt.ErrTokenUsedBeforeIssued), errors.Is(err, jwt.ErrTokenNotValidYet):
					sendUnauthorized(w, warnFunctionf, "JWT token validation failed: token not valid yet for URL: %s", path)
				case errors.Is(err, jwt.ErrTokenMalformed):
					sendUnauthorized(w, warnFunctionf, "JWT token validation failed: malformed token for URL: %s", path)
				case errors.Is(err, jwt.ErrTokenSignatureInvalid):
					sendUnauthorized(w, warnFunctionf, "JWT token validation failed: invalid signature for URL: %s", path)
				case errors.Is(err, jwt.ErrTokenRequiredClaimMissing):
					sendUnauthorized(w, warnFunctionf, "JWT token validation failed: required claim missing for URL: %s", path)
				default:
					sendUnauthorized(w, warnFunctionf, "Error parsing token for URL %q: %s", path, err.Error())
				}
				return
			}

			// Basic token validity check
			if !tkn.Valid {
				sendUnauthorized(w, warnFunctionf, "JWT token basic validation failed: invalid token for URL: %s", path)
				return
			}

			// Log successful authentication
			if user := claims["sub"]; user != nil {
				log := fmt.Sprintf("JWT auth passed for URL %q: user: %q", path, user)

				if issuer := claims["iss"]; issuer != nil {
					log += fmt.Sprintf(" (issuer: %q)", issuer)
				}

				loggedInFunction(log)
			}

			// Continue to the next handler
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
