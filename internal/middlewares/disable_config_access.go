package middlewares

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// DisableAccessToFile returns a middleware that disables access to files
// that match the given function.
func DisableAccessToFile(fn func(string) bool, statusCode int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the file's name
			fileName := filepath.Base(r.URL.Path)

			// Check if the function passes
			if fn(fileName) {
				http.Error(w, fmt.Sprintf("%d %s", statusCode, strings.ToLower(http.StatusText(statusCode))), statusCode)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
