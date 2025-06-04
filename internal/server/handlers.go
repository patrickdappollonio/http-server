package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/patrickdappollonio/http-server/internal/ctype"
	"github.com/patrickdappollonio/http-server/internal/fileutil"
	"github.com/patrickdappollonio/http-server/internal/renderer"
	isort "github.com/patrickdappollonio/http-server/internal/sort"
	"github.com/saintfish/chardet"
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
			s.printWarningf("attempted to access non-existent path: %s", currentPath)

			// Overwrite custom page if one was set
			if s.CustomNotFoundPage != "" {
				// Check if the status code was generated
				statusCode := s.CustomNotFoundStatusCode
				if statusCode == 0 {
					statusCode = http.StatusNotFound
				}

				s.serveFile(statusCode, s.CustomNotFoundPage, w, r)
				return
			}
			httpError(http.StatusNotFound, w, "404 not found")
			return
		}

		// If it's any other kind of error, return the 500 error and log the actual error
		// to the app log
		s.printWarningf("unable to stat directory %q: %s", currentPath, err)
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

	// If the path is not a directory, then it's a file, so we can render it,
	// let's check first if it's a markdown file
	if ext := strings.ToLower(filepath.Ext(currentPath)); ext == ".md" || ext == ".markdown" {
		// Check FullMarkdownRender first to avoid unnecessary filename extraction
		if s.FullMarkdownRender {
			s.serveMarkdown(currentPath, w, r)
			return
		}

		// Not rendering all markdown, check if this is an index-like file
		filename := filepath.Base(currentPath)
		if slices.Contains(allowedIndexFiles, filename) {
			s.serveMarkdown(currentPath, w, r)
			return
		}

		// If not an index file and FullMarkdownRender is disabled, serve as plain text
		s.serveFile(0, currentPath, w, r)
		return
	}

	s.serveFile(0, currentPath, w, r)
}

func (s *Server) serveMarkdown(requestedPath string, w http.ResponseWriter, r *http.Request) {
	// Find if among the files there's a markdown readme
	var markdownContent bytes.Buffer
	if err := s.renderMarkdownFile(requestedPath, &markdownContent); err != nil {
		s.printWarningf("unable to generate markdown: %s", err)
		httpError(http.StatusInternalServerError, w, "unable to generate markdown for current directory -- see application logs for more information")
		return
	}

	// Define the parent directory
	parent := getParentURL(s.PathPrefix, r.URL.Path)

	// Render the directory listing
	content := map[string]any{
		"DirectoryRootPath":      s.PathPrefix,
		"PageTitle":              s.PageTitle,
		"CurrentPath":            r.URL.Path,
		"CacheBuster":            s.cacheBuster,
		"RequestedPath":          requestedPath,
		"IsRoot":                 s.PathPrefix == r.URL.Path,
		"UpDirectory":            parent,
		"Files":                  nil,
		"ShouldRenderFiles":      false, // we don't want file rendering for non index pages
		"IsSingleMarkdownRender": true,  // we want to differentiate single markdown render pages
		"HideLinks":              s.HideLinks,
		"MarkdownContent":        markdownContent.String(),
		"MarkdownBeforeDir":      s.MarkdownBeforeDir,
		"CustomCSS":              s.getCustomCSSURL(),
	}

	if err := s.templates.ExecuteTemplate(w, "app.tmpl", content); err != nil {
		s.printWarningf("unable to render directory listing: %s", err)
		httpError(http.StatusInternalServerError, w, "unable to render directory listing -- see application logs for more information")
		return
	}
}

// FileInfo represents a file's metadata for JSON output
type FileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	IsDirectory bool   `json:"is_directory"`
	ModTime     string `json:"mod_time"`
	Path        string `json:"path"`
}

func (s *Server) walk(requestedPath string, w http.ResponseWriter, r *http.Request) {
	// Append index.html or index.htm to the path and see if the index
	// file exists, if so, return it instead
	for _, index := range []string{"index.html", "index.htm"} {
		indexPath := filepath.Join(requestedPath, index)
		if _, err := os.Stat(indexPath); err == nil {
			s.serveFile(0, indexPath, w, r)
			return
		}
	}

	// Check if directory listing is disabled, if so,
	// return here with a 404 error
	if s.DisableDirectoryList {
		httpError(http.StatusNotFound, w, "404 not found")
		return
	}

	// Calculate the relative path
	relativePath, err := filepath.Rel(s.Path, requestedPath)
	if err != nil {
		s.printWarningf("unable to calculate relative path for %q: %s", requestedPath, err)
		// User sees the requestedPath, which is more direct than the calculated relativePath here.
		httpError(http.StatusInternalServerError, w, "internal error calculating relative path for %q -- see application logs for details", requestedPath)
		return
	}

	// Open the directory path using the root context and relative path.
	// s.RootCtx is an *os.Root context for s.Path.
	dir, err := s.RootCtx.Open(relativePath)
	if err != nil {
		// If the directory doesn't exist, render an appropriate message
		if os.IsNotExist(err) {
			s.printWarningf("attempted to access non-existent path %q (relative: %q) using sandboxed context: %s", requestedPath, relativePath, err)
			// For user-facing "not found", the specific relative path is good.
			httpError(http.StatusNotFound, w, "directory %q not found", relativePath)
			return
		}

		// Otherwise handle it generically speaking
		s.printWarningf("unable to open directory %q (relative: %q) using sandboxed context: %s", requestedPath, relativePath, err)
		httpError(http.StatusInternalServerError, w, "unable to open directory %q -- see application logs for more information", relativePath)
		return
	}
	defer dir.Close() // Ensure directory is closed

	// Read all files in the directory. os.File.ReadDir is used.
	list, err := dir.ReadDir(-1)

	// Handle error on readdir call
	if err != nil {
		s.printWarningf("unable to read directory %q (relative: %s): %s", requestedPath, relativePath, err)
		httpError(http.StatusInternalServerError, w, "unable to read directory %q -- see application logs for more information", relativePath)
		return
	}

	// Render the directory listing
	sort.Sort(isort.FoldersFirst(list))

	// Generate a list of FileInfo objects
	files := make([]os.FileInfo, 0, len(list))
	for _, f := range list {
		fi, err := f.Info()
		if err != nil {
			// When using fs.DirEntry, f.Name() is the base name. For full path for logging, we might need to reconstruct or log relative path.
			s.printWarningf("unable to stat file %q in %q: %s", f.Name(), relativePath, err)
			httpError(http.StatusInternalServerError, w, "unable to stat file %q -- see application logs for more information", f.Name())
			return
		}

		// Check if file starts with config prefix
		if s.isFiltered(fi.Name()) {
			continue
		}

		files = append(files, fi)
	}

	// Handle different output formats
	if outputFormat := r.URL.Query().Get("output"); outputFormat != "" {
		// Get parent directory URL
		parent := getParentURL(s.PathPrefix, r.URL.Path)

		// Create renderer configuration
		config := renderer.Config{
			CurrentPath: r.URL.Path,
			ParentPath:  parent,
			Logger:      s.LogOutput,
		}

		// Render the directory listing
		if err := renderer.Render(outputFormat, config, w, files); err != nil {
			if errors.Is(err, renderer.UnsupportedFormatError{}) {
				s.printWarningf("unsupported output format: %s", err)
				httpError(http.StatusBadRequest, w, "unsupported output format: %q (supported formats: %s)",
					outputFormat, renderer.GetSupportedFormatsString())
				return
			}

			s.printWarningf("error rendering directory listing: %s", err)
			httpError(http.StatusInternalServerError, w, "error rendering directory listing -- see application logs for more information")
			return
		}
		return
	}

	// Find if among the files there's a markdown readme
	var markdownContent bytes.Buffer
	// Pass s.Path and relativePath to findAndGenerateMarkdown if it needs to construct full paths.
	// For now, assuming findAndGenerateMarkdown works with requestedPath or can be adapted.
	// If it directly uses file paths, it might need s.Path and the relative paths of entries.
	// However, `files` contains os.FileInfo, which doesn't inherently know its full path if obtained from fs.FS.
	// This might be a concern for findAndGenerateMarkdown if it relies on absolute paths from os.FileInfo.
	// For now, we'll assume it works or will be adapted in a future step.
	// The `requestedPath` is still the absolute path to the directory being walked.
	if err := s.findAndGenerateMarkdown(requestedPath, files, &markdownContent); err != nil {
		s.printWarningf("unable to generate markdown for %q: %s", requestedPath, err)
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
		"RequestedPath":     requestedPath,
		"IsRoot":            s.PathPrefix == r.URL.Path,
		"UpDirectory":       parent,
		"Files":             files,
		"ShouldRenderFiles": !s.HideFilesInMarkdown,
		"HideLinks":         s.HideLinks,
		"MarkdownContent":   markdownContent.String(),
		"MarkdownBeforeDir": s.MarkdownBeforeDir,
		"CustomCSS":         s.getCustomCSSURL(),
	}

	if err := s.templates.ExecuteTemplate(w, "app.tmpl", content); err != nil {
		s.printWarningf("unable to render directory listing: %s", err)
		httpError(http.StatusInternalServerError, w, "unable to render directory listing -- see application logs for more information")
		return
	}
}

// statusCodeHijacker is a response writer that captures the status code
// and the body of the response that would have been sent to the client.
type statusCodeHijacker struct {
	http.ResponseWriter
	givenStatusCode int
}

// WriteHeader captures the status code that would have been sent to the client.
func (s *statusCodeHijacker) WriteHeader(code int) {
	s.givenStatusCode = code
}

// serveFile serves a file with the appropriate headers, including support
// for ETag and Last-Modified headers, as well as range requests.
// If the status code is not 0, the status code provided will be used
// when serving the file in the given path.
func (s *Server) serveFile(statusCode int, location string, w http.ResponseWriter, r *http.Request) {
	// Calculate the relative path from the server's root path
	relativePath, err := filepath.Rel(s.Path, location)
	if err != nil {
		s.printWarningf("unable to calculate relative path for %q (root: %q): %s", location, s.Path, err)
		// User sees the location they requested.
		httpError(http.StatusInternalServerError, w, "internal error calculating relative path for %q", location)
		return
	}

	// Open the file using the root context and relative path.
	// s.RootCtx is an *os.Root context for s.Path.
	f, err := s.RootCtx.Open(relativePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.printWarningf("file not found %q (relative: %q) using sandboxed context: %s", location, relativePath, err)
			httpError(http.StatusNotFound, w, "file %q not found", relativePath)
			return
		}
		s.printWarningf("error opening file %q (relative: %q) using sandboxed context: %s", location, relativePath, err)
		httpError(http.StatusInternalServerError, w, "error opening file %q", relativePath)
		return
	}
	defer f.Close()

	// Stat the file. os.File implements Stat().
	fi, err := f.Stat()
	if err != nil {
		s.printWarningf("error stating file %q (relative: %q): %s", location, relativePath, err)
		httpError(http.StatusInternalServerError, w, "error stating file %q", relativePath)
		return
	}

	var contentType string
	// filepath.Base(location) is fine here as `location` is the original full path.
	if local := ctype.GetContentTypeForFilename(filepath.Base(location)); local != "" {
		contentType = local
	}

	var data [512]byte
	// f.Read() is fine as fs.File implements io.Reader.
	n, err := f.Read(data[:])
	if err != nil && !errors.Is(err, io.EOF) {
		s.printWarningf("error reading file %q (relative: %q) for content type detection: %s", location, relativePath, err)
		httpError(http.StatusInternalServerError, w, "error reading file %q for content type detection", relativePath)
		return
	}

	// Only detect content type if we have content to examine
	if contentType == "" && n > 0 {
		if local := http.DetectContentType(data[:n]); local != "application/octet-stream" {
			contentType = local
		}
	}

	// Only attempt charset detection if we have content to examine
	charset := ""
	if n > 0 && contentType != "" && !strings.HasPrefix(contentType, "application/octet-stream") {
		if utf8.Valid(data[:n]) {
			charset = "utf-8"
		} else {
			res, err := chardet.NewTextDetector().DetectBest(data[:n])
			if err == nil && res.Confidence > 50 && res.Charset != "" {
				charset = res.Charset
			}
		}
	}

	if contentType != "" {
		// Add charset for text-based content types only
		if charset != "" && contentType != "application/octet-stream" {
			contentType += "; charset=" + charset
		}

		w.Header().Set("Content-Type", contentType)
	}

	// Check if we should force download this file based on its extension
	// `location` is still the original absolute path, which is fine for ShouldForceDownload.
	if fileutil.ShouldForceDownload(location, s.ForceDownloadExtensions, s.SkipForceDownloadFiles) {
		filename := filepath.Base(location)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	}

	// Reset file position to beginning after reading first bytes for content detection
	// f.Seek() is fine as fs.File implements io.Seeker.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		s.printWarningf("error seeking file %q (relative: %q): %s", location, relativePath, err)
		httpError(http.StatusInternalServerError, w, "error seeking in file %q", relativePath)
		return
	}

	// Check if the caller changed the status code, if not, simply call
	// the appropriate handler/
	if statusCode == 0 {
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
		return
	}

	// Write the status code sent by the caller.
	w.WriteHeader(statusCode)

	// Call serve content with the hijacked response writer, which won't
	// be able to overwrite the status code.
	http.ServeContent(&statusCodeHijacker{ResponseWriter: w}, r, fi.Name(), fi.ModTime(), f)
}

// healthCheck is a simple health check endpoint that returns 200 OK
func (s *Server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// httpError writes an error message to the response writer.
func httpError(statusCode int, w http.ResponseWriter, format string, args ...any) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, format, args...)
}

// getParentURL returns the parent URL for the given location.
func getParentURL(base, loc string) string {
	if loc == base {
		return ""
	}

	s := path.Dir(strings.TrimSuffix(loc, "/"))

	if strings.HasSuffix(s, "/") {
		return s
	}

	return s + "/"
}
