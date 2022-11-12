package mw

import (
	"net/http"
	"strings"
)

var indexes = [...]string{"index.html", "index.htm"}

// RedirectIndexes is a middleware that redirects requests for a directory
// if the URL ends in a known index file back to the root of it, avoiding the
// need for longer urls.

func RedirectIndexes(statusCode int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, index := range indexes {
				if strings.HasSuffix(r.URL.Path, index) {
					http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, index), statusCode)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
