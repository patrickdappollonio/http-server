package mw

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

func applyStatusCode(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, strings.ToLower(http.StatusText(code))), code)
}

func DisableConfigAccess(files []string, filePrefixes []string, fileExt []string, statusCode int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the file's name
			fileName := filepath.Base(r.URL.Path)

			// Check if the path is a file
			for _, file := range files {
				if fileName == file {
					applyStatusCode(w, statusCode)
					return
				}
			}

			// Check if the path is a file with a prefix
			for _, prefix := range filePrefixes {
				if strings.HasPrefix(fileName, prefix) {
					applyStatusCode(w, statusCode)
					return
				}
			}

			// Check if the path is a file with a specific extension
			for _, ext := range fileExt {
				if strings.HasSuffix(fileName, ext) {
					applyStatusCode(w, statusCode)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
