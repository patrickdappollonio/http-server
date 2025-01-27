package middlewares

import "net/http"

// VerbsAllowed is a middleware that allows only specific HTTP verbs to be
// processed. If the request verb is not in the list of allowed verbs, a
// 405 Method Not Allowed response is returned.
func VerbsAllowed(allowedVerbs ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, allowedVerb := range allowedVerbs {
				if r.Method == allowedVerb {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 method not allowed"))
		})
	}
}
