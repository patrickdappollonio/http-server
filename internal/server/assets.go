package server

import (
	"embed"
	"net/http"
)

// Assets contains all the assets in the server/assets folder
//
//go:embed assets/*
var assets embed.FS

// serveAssets returns a handler that serves the assets from the code's "assets"
// folder as an embedded filesystem. The handler strips any potential prefix from
// the request path. Assets will be served after the server-wide prefix has been
// set.
func (s *Server) serveAssets(prefix string) func(http.ResponseWriter, *http.Request) {
	// Create a static handler that auto-removes the prefix from the request
	fs := http.StripPrefix(prefix, http.FileServer(http.FS(assets)))

	// Return the file server handler
	return func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}
}
