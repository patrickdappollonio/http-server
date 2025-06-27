// Package httputil provides utility functions for HTTP operations.
package httputil

import (
	"fmt"
	"net/http"
)

// Error is a helper function to return HTTP errors
func Error(statusCode int, w http.ResponseWriter, message string) {
	w.WriteHeader(statusCode)
	fmt.Fprint(w, message)
}
