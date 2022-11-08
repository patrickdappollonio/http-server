package server

import (
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickdappollonio/http-server/internal/mw"
)

func (s *Server) router() http.Handler {
	r := chi.NewRouter()

	// Allow logging all request to our custom logger
	r.Use(mw.LogRequest(s.LogOutput, logFormat))

	// Recover the request in case of a panic
	r.Use(middleware.Recoverer)

	// Only allow specific methods in all our requests
	r.Use(mw.VerbsAllowed("GET", "HEAD"))

	// Enable basic authentication if needed
	if s.IsAuthEnabled() {
		r.Use(middleware.BasicAuth("http-server", map[string]string{
			s.Username: s.Password,
		}))
	}

	// Check if the request is against a URL ending on a known
	// index file, and if so, redirect to the directory
	r.Use(mw.RedirectIndexes(http.StatusMovedPermanently))

	// Handle emptiness of path prefix
	if s.PathPrefix == "" {
		s.PathPrefix = "/"
	}

	// Create a route based on a path prefix, prevalidated that
	// the prefix is a valid prefix
	routePrefix := path.Join(s.PathPrefix, "*")
	r.Get(routePrefix, s.showOrRender)

	// Create a route for static assets, including
	// the cache buster randomized string so we can
	// force reload the assets on each execution
	assetsPrefix := path.Join(s.PathPrefix, specialPath, s.cacheBuster)
	r.Get(path.Join(assetsPrefix, "assets", "*"), s.serveAssets(assetsPrefix))

	// Handle special path prefix cases
	if s.PathPrefix != "/" {
		// If the path prefix is not the root of the server, then we
		// can preemptively redirect users to the appropriate destination
		// so they don't see a not found error
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, s.PathPrefix, http.StatusFound)
		})

		// Redirect path prefix without trailing slash to a canonical location
		r.Get(strings.TrimSuffix(s.PathPrefix, "/"), func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, s.PathPrefix, http.StatusMovedPermanently)
		})
	}

	return r
}
