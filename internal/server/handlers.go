package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patrickdappollonio/http-server/internal/mw"
)

const (
	logFormat   = `{http_method} "{url}" -- {proto} {status_code} {status_text} (served in {duration}; {bytes_written} bytes)`
	specialPath = "_"
)

// showOrRender is the main handler for the server. It will either render the
// content requested or show a directory listing.
func (s *Server) showOrRender(w http.ResponseWriter, r *http.Request) {
	relpath := filepath.Join(s.Path, strings.TrimPrefix(r.URL.Path, s.PathPrefix))

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
			s.printWarning("attempted to access non-existent path: %s", currentPath)
			httpError(http.StatusNotFound, w, "404 not found")
			return
		}

		// If it's any other kind of error, return the 500 error and log the actual error
		// to the app log
		s.printWarning("unable to stat directory %q: %s", currentPath, err)
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
	s.serveFile(currentPath, w, r)
}

func (s *Server) walk(requestedPath string, w http.ResponseWriter, r *http.Request) {
	// Append index.html or index.htm to the path and see if the index
	// file exists, if so, return it instead
	for _, index := range []string{"index.html", "index.htm"} {
		indexPath := filepath.Join(requestedPath, index)
		if _, err := os.Stat(indexPath); err == nil {
			s.serveFile(indexPath, w, r)
			return
		}
	}

	// Open the directory path and read all files
	dir, err := os.Open(requestedPath)
	if err != nil {
		// If the directory doesn't exist, render an appropriate message
		if os.IsNotExist(err) {
			s.printWarning("attempted to access non-existent path: %s", requestedPath)
			httpError(http.StatusNotFound, w, "404 not found")
			return
		}

		// Otherwise handle it generically speaking
		s.printWarning("unable to open directory %q: %s", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to open directory -- see application logs for more information")
		return
	}

	// Read all files in the directory then close the directory
	list, err := dir.ReadDir(-1)
	dir.Close()

	// Handle error on readdir call
	if err != nil {
		s.printWarning("unable to read directory %q: %s", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to read directory -- see application logs for more information")
		return
	}

	// Render the directory listing
	sort.Sort(foldersFirst(list))

	// Generate a list of FileInfo objects
	files := make([]os.FileInfo, 0, len(list))
	for _, f := range list {
		fi, err := f.Info()
		if err != nil {
			s.printWarning("unable to stat file %q: %s", f.Name(), err)
			httpError(http.StatusInternalServerError, w, "unable to stat file %q -- see application logs for more information", f.Name())
			return
		}

		// Check if file starts with config prefix
		if strings.HasPrefix(fi.Name(), s.ConfigFilePrefix) {
			continue
		}

		files = append(files, fi)
	}

	// Find if among the files there's a markdown readme
	var markdownContent bytes.Buffer
	if err := s.generateMarkdown(requestedPath, files, &markdownContent); err != nil {
		s.printWarning("unable to generate markdown: %s", err)
		httpError(http.StatusInternalServerError, w, "unable to generate markdown for current directory -- see application logs for more information")
		return
	}

	// Define the parent directory
	parent := getParentURL(s.PathPrefix, r.URL.Path)

	// Render the directory listing
	content := map[string]any{
		"DirectoryRootPath": s.PathPrefix,
		"PageTitle":         s.PageTitle,
		"CurrentPath":       r.URL.Path,
		"CacheBuster":       s.cacheBuster,
		"Files":             files,
		"RequestedPath":     requestedPath,
		"IsRoot":            s.PathPrefix == r.URL.Path,
		"UpDirectory":       parent,
		"HideLinks":         s.HideLinks,
		"MarkdownContent":   markdownContent.String(),
		"MarkdownBeforeDir": s.MarkdownBeforeDir,
	}

	if err := s.templates.ExecuteTemplate(w, "app.tmpl", content); err != nil {
		s.printWarning("unable to render directory listing: %s", err)
		httpError(http.StatusInternalServerError, w, "unable to render directory listing -- see application logs for more information")
		return
	}
}

// serveFile serves a file with the appropriate headers, including support
// for ETag and Last-Modified headers, as well as range requests.
func (s *Server) serveFile(fp string, w http.ResponseWriter, r *http.Request) {
	mw.Etag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, fp)
	})).ServeHTTP(w, r)
}

// healthCheck is a simple health check endpoint that returns 200 OK
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func httpError(statusCode int, w http.ResponseWriter, format string, args ...any) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, format, args...)
}

func getParentURL(base string, loc string) string {
	if loc == base {
		return ""
	}

	s := path.Dir(strings.TrimSuffix(loc, "/"))

	if strings.HasSuffix(s, "/") {
		return s
	}

	return s + "/"
}
