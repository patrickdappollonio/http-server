package main

import "net/http"

func cors(enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			next.ServeHTTP(w, r)
		})
	}
}
