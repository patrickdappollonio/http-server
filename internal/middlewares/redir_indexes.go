package middlewares

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
					target := strings.TrimSuffix(r.URL.Path, index)

					// Collapse duplicate leading slashes so the target can't be
					// interpreted by browsers as a scheme-relative URL ("//evil.com/").
					if strings.HasPrefix(target, "//") {
						target = "/" + strings.TrimLeft(target, "/")
					}

					http.Redirect(w, r, target, statusCode) //nolint:gosec // target is derived from the request path with leading slashes collapsed, so it is always same-origin
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
