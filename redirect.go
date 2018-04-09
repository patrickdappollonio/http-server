package main

import (
	"net/http"
)

func redirect(expectedPath, nextURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedPath != "*" && r.URL.Path != expectedPath {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Redirect(w, r, nextURL, http.StatusFound)
	})
}
