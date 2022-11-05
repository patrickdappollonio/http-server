package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickdappollonio/http-server/internal/mw"
)

const (
	logFormat  = `{http_method} "{url}" -- {proto} {status_code} {status_text} (served in {duration}; {bytes_written} bytes)`
	warnPrefix = "(WARN)"
)

func (s *Server) handler() http.Handler {
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
	r.Get(s.PathPrefix+"*", s.showOrRender)

	// If the path prefix is not the root of the server, then we
	// can preemptively redirect users to the appropriate destination
	// so they don't see a not found error
	if s.PathPrefix != "/" {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, s.PathPrefix, http.StatusFound)
		})
	}

	return r
}

func (s *Server) showOrRender(w http.ResponseWriter, r *http.Request) {
	relpath := filepath.Join(s.Path, r.URL.Path)

	// Generate an absolute path off a relative one
	currentPath, err := filepath.Abs(relpath)
	if err != nil {
		fmt.Fprintln(s.LogOutput, "error generating absolute path:", err)
		httpError(http.StatusInternalServerError, w, "internal error generating full paths -- see application logs for details")
		return
	}

	// Stat the current path
	info, err := os.Stat(currentPath)
	if err != nil {
		// If the path doesn't exist, return the 404 error but also print in the log
		// of the app the full path to the given location
		if os.IsNotExist(err) {
			fmt.Fprintln(s.LogOutput, warnPrefix, "attempted to acess non-existent path:", currentPath)
			httpError(http.StatusNotFound, w, "file or folder not found")
			return
		}

		// If it's any other kind of error, return the 500 error and log the actual error
		// to the app log
		fmt.Fprintf(s.LogOutput, warnPrefix+" unable to stat directory %q: %s\n", currentPath, err)
		httpError(http.StatusInternalServerError, w, "unable to stat directory -- see application logs for more information")
		return
	}

	// Check if the current path is a directory
	if info.IsDir() {
		// Check if the path doesn't ends in a slash, and redirect accordingly
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}

		s.walk(currentPath, w, r)
		return
	}

	// If the path is not a directory, then it's a file, so we can render it
	http.ServeFile(w, r, currentPath)
}

func (s *Server) walk(requestedPath string, w http.ResponseWriter, r *http.Request) {
	// Append index.html or index.htm to the path and see if the index
	// file exists, if so, return it instead
	for _, index := range []string{"index.html", "index.htm"} {
		indexPath := filepath.Join(requestedPath, index)
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
	}

	// Open the directory path and read all files
	dir, err := os.Open(requestedPath)
	if err != nil {
		// If the directory doesn't exist, render an appropriate message
		if os.IsNotExist(err) {
			fmt.Fprintln(s.LogOutput, warnPrefix, "attempted to acess non-existent path:", requestedPath)
			httpError(http.StatusNotFound, w, "file or folder not found")
			return
		}

		// Otherwise handle it generically speaking
		fmt.Fprintf(s.LogOutput, warnPrefix+" unable to open directory %q: %s\n", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to open directory -- see application logs for more information")
		return
	}

	// Read all files in the directory then close the directory
	list, err := dir.ReadDir(-1)
	dir.Close()

	// Handle error on readdir call
	if err != nil {
		fmt.Fprintf(s.LogOutput, warnPrefix+" unable to read directory %q: %s\n", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to read directory -- see application logs for more information")
		return
	}

	// Render the directory listing
	sort.Sort(foldersFirst(list))

	// // Generate a link to the parent folder, for the breadcrumb
	// parentWebFolder := ""
	// if relPath := path.Join(s.Path, r.URL.Path); relPath != "/" && relPath != s.Path {
	// 	parentWebFolder = path.Dir(strings.TrimSuffix(relPath, "/"))

	// 	if !strings.HasSuffix(parentWebFolder, "/") {
	// 		parentWebFolder += "/"
	// 	}

	// 	if parentWebFolder == "/" && s.Path != "/" {
	// 		parentWebFolder = ""
	// 	}
	// }

	// Render the directory listing
	content := map[string]any{
		"Path": r.URL.Path,
		// "ParentWebFolder": parentWebFolder,
		"Files":         fmt.Sprintf("%#v", list),
		"RequestedPath": requestedPath,
	}

	if err := json.NewEncoder(w).Encode(content); err != nil {
		fmt.Fprintf(s.LogOutput, warnPrefix+" unable to render directory %q: %s\n", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to render directory -- see application logs for more information")
		return
	}

	// fmt.Fprintln(w, "directory listing not implemented yet, getting:", requestedPath)
}

func httpError(statusCode int, w http.ResponseWriter, format string, args ...any) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, format, args...)
}
