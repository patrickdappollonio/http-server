package mw

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// ValidateJWTHS256 validates a JWT token using the HS256 algorithm,
// the token can be passed in the "Authorization" header or in the
// "token" query parameter.
func ValidateJWTHS256(warnFunction func(string, ...interface{}), loggedInFunction func(string), jwtSigningKey string, validateTimedJWT bool) func(http.Handler) http.Handler {
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

			// Parse token and validate signing algorithm
			tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
				hm, ok := token.Method.(*jwt.SigningMethodHMAC)
				if !ok || hm.Hash != jwt.SigningMethodHS256.Hash {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(jwtSigningKey), nil
			})

			// Check for errors during parsing
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if the token is valid
			if !tkn.Valid {
				warnFunction("JWT token validation failed: invalid token for url: %s", r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if validateTimedJWT {
				t := time.Now().Unix()

				if !claims.VerifyExpiresAt(t, true) {
					warnFunction("JWT token validation failed: token expired for url: %s", r.URL.Path)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				if !claims.VerifyIssuedAt(t, true) {
					warnFunction("JWT token validation failed: token issued in the future for url: %s", r.URL.Path)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}

			if user := claims["user"]; user != nil {
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
