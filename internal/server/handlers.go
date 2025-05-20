package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patrickdappollonio/http-server/internal/ctype"
	"github.com/patrickdappollonio/http-server/internal/fileutil"
	"github.com/patrickdappollonio/http-server/internal/httputil"
	"github.com/patrickdappollonio/http-server/internal/pathutil"
	"github.com/patrickdappollonio/http-server/internal/renderer"
	isort "github.com/patrickdappollonio/http-server/internal/sort"
)

const (
	logFormat   = `{http_method} "{url}" -- {proto} {status_code} {status_text} (served in {duration}; {bytes_written} bytes)`
	specialPath = "_"
)

// Standard index files to check when serving a directory
var indexFiles = []string{"index.html", "index.htm"}

// FileInfo represents a file's metadata for JSON output
type FileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	IsDirectory bool   `json:"is_directory"`
	ModTime     string `json:"mod_time"`
	Path        string `json:"path"`
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

// showOrRender is the main handler for the server. It will either render the
// content requested or show a directory listing.
func (s *Server) showOrRender(w http.ResponseWriter, r *http.Request) {
	// Create absolute path from URL path
	currentPath, err := s.makeAbsPath(r.URL.Path)
	if err != nil {
		fmt.Fprintln(s.LogOutput, "error generating absolute path:", err)
		httputil.Error(http.StatusInternalServerError, w, "internal error generating full paths -- see application logs for details")
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
			httputil.Error(http.StatusNotFound, w, "404 not found")
			return
		}

		// If it's any other kind of error, return the 500 error and log the actual error
		// to the app log
		s.printWarningf("unable to stat directory %q: %s", currentPath, err)
		httputil.Error(http.StatusInternalServerError, w, "unable to stat directory -- see application logs for more information")
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
	if fileutil.IsMarkdownFile(currentPath) {
		// Serve as markdown if it's an index file or if rendering all markdown is enabled
		if s.FullMarkdownRender || fileutil.IsIndexFile(filepath.Base(currentPath), allowedIndexFiles) {
			s.serveMarkdown(currentPath, w, r)
			return
		}

		// Otherwise serve as plain text
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
		httputil.Error(http.StatusInternalServerError, w, "unable to generate markdown for current directory -- see application logs for more information")
		return
	}

	// Prepare template data - using nil for files in single markdown view
	content := s.prepareTemplateData(requestedPath, r, nil, &markdownContent, true)

	if err := s.templates.ExecuteTemplate(w, "app.tmpl", content); err != nil {
		s.printWarningf("unable to render directory listing: %s", err)
		httputil.Error(http.StatusInternalServerError, w, "unable to render directory listing -- see application logs for more information")
		return
	}
}

func (s *Server) walk(requestedPath string, w http.ResponseWriter, r *http.Request) {
	// Check for index files in the directory
	for _, index := range indexFiles {
		indexPath := filepath.Join(requestedPath, index)
		if fileutil.FileExists(indexPath) {
			s.serveFile(0, indexPath, w, r)
			return
		}
	}

	// Check if directory listing is disabled, if so,
	// return here with a 404 error
	if s.DisableDirectoryList {
		httputil.Error(http.StatusNotFound, w, "404 not found")
		return
	}

	// Read directory entries
	files, err := s.readDirEntries(requestedPath)
	if err != nil {
		// Handle directory access errors
		if os.IsNotExist(err) {
			s.handleError(w, http.StatusNotFound, fmt.Sprintf("attempted to access non-existent path: %s", requestedPath), "404 not found", nil)
			return
		}

		// Handle other errors
		logMsg := fmt.Sprintf("unable to read directory %q", requestedPath)
		userMsg := "unable to read directory -- see application logs for more information"
		s.handleError(w, http.StatusInternalServerError, logMsg, userMsg, err)
		return
	}

	// Handle different output formats
	if outputFormat := r.URL.Query().Get("output"); outputFormat != "" {
		s.handleCustomOutputFormat(w, r, outputFormat, files)
		return
	}

	// Find if among the files there's a markdown readme
	var markdownContent bytes.Buffer
	if err := s.findAndGenerateMarkdown(requestedPath, files, &markdownContent); err != nil {
		s.printWarningf("unable to generate markdown: %s", err)
		httputil.Error(http.StatusInternalServerError, w, "unable to generate markdown for current directory -- see application logs for more information")
		return
	}

	// Prepare template data - using files for directory listing
	content := s.prepareTemplateData(requestedPath, r, files, &markdownContent, false)

	if err := s.templates.ExecuteTemplate(w, "app.tmpl", content); err != nil {
		s.printWarningf("unable to render directory listing: %s", err)
		httputil.Error(http.StatusInternalServerError, w, "unable to render directory listing -- see application logs for more information")
		return
	}
}

// serveFile serves a file with the appropriate headers, including support
// for ETag and Last-Modified headers, as well as range requests.
// If the status code is not 0, the status code provided will be used
// when serving the file in the given path.
func (s *Server) serveFile(statusCode int, location string, w http.ResponseWriter, r *http.Request) {
	// Open the file
	f, err := os.Open(location)
	if err != nil {
		if os.IsNotExist(err) {
			httputil.Error(http.StatusNotFound, w, "404 not found")
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer f.Close()

	// Get file info
	fi, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle content detection and headers efficiently
	var data [512]byte
	n, err := io.ReadAtLeast(f, data[:], 1)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		// Only treat as error if it's not EOF (empty file) or unexpected EOF (small file)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Detect content type and set appropriate headers
	contentType, _ := ctype.DetectContentType(location, data[:n], n)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Set content disposition header for file download if needed
	s.setContentDispositionHeader(w, location)

	// Reset file position to beginning after reading first bytes for content detection
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// handleError logs an error with optional context and returns an HTTP error
func (s *Server) handleError(w http.ResponseWriter, statusCode int, logMsg, userMsg string, err error) {
	if err != nil {
		s.printWarningf("%s: %s", logMsg, err)
	} else {
		s.printWarningf("%s", logMsg)
	}
	// Use a const format string to satisfy the linter
	w.WriteHeader(statusCode)
	fmt.Fprint(w, userMsg)
}

// makeAbsPath creates an absolute path from a URL path
func (s *Server) makeAbsPath(urlPath string) (string, error) {
	relpath := filepath.Join(s.Path, strings.TrimPrefix(urlPath, s.PathPrefix))
	absPath, err := filepath.Abs(relpath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}

// setContentDispositionHeader sets the Content-Disposition header for file downloads
func (s *Server) setContentDispositionHeader(w http.ResponseWriter, filePath string) {
	if fileutil.ShouldForceDownload(filePath, s.ForceDownloadExtensions, s.SkipForceDownloadFiles) {
		filename := filepath.Base(filePath)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	}
}

// handleCustomOutputFormat handles rendering content in custom output formats
func (s *Server) handleCustomOutputFormat(w http.ResponseWriter, r *http.Request, format string, files []os.FileInfo) {
	// Get parent directory URL
	parent := pathutil.GetParentURL(s.PathPrefix, r.URL.Path)
	// Create renderer configuration
	config := renderer.Config{
		CurrentPath: r.URL.Path,
		ParentPath:  parent,
		Logger:      s.LogOutput,
	}

	// Render the directory listing
	err := renderer.Render(format, config, w, files)
	if err == nil {
		// Successfully rendered the output format
		return
	}

	// Handle any rendering errors
	if errors.Is(err, renderer.UnsupportedFormatError{}) {
		logMsg := fmt.Sprintf("unsupported output format: %s", format)
		userMsg := fmt.Sprintf("unsupported output format: %q (supported formats: %s)",
			format, renderer.GetSupportedFormatsString())
		s.handleError(w, http.StatusBadRequest, logMsg, userMsg, err)
	} else {
		logMsg := "error rendering directory listing"
		userMsg := "error rendering directory listing -- see application logs for more information"
		s.handleError(w, http.StatusInternalServerError, logMsg, userMsg, err)
	}
}

// readDirEntries reads directory entries and returns a sorted, filtered list of file info objects
func (s *Server) readDirEntries(dirPath string) ([]os.FileInfo, error) {
	// Try to open the file
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer f.Close()

	// Read directory entries
	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory entries: %w", err)
	}

	// Sort the directory listing
	sort.Sort(isort.FoldersFirst(entries))

	// Generate a list of FileInfo objects
	result := make([]os.FileInfo, 0, len(entries))
	for _, f := range entries {
		fi, err := f.Info()
		if err != nil {
			return nil, fmt.Errorf("unable to get info for file %q: %w", f.Name(), err)
		}

		// Skip filtered files
		if s.isFiltered(fi.Name()) {
			continue
		}

		result = append(result, fi)
	}

	return result, nil
}

// prepareTemplateData creates a map with common template data
func (s *Server) prepareTemplateData(requestedPath string, r *http.Request, files []os.FileInfo, markdownContent *bytes.Buffer, isSingleMarkdown bool) map[string]any {
	parent := pathutil.GetParentURL(s.PathPrefix, r.URL.Path)

	data := map[string]any{
		"DirectoryRootPath": s.PathPrefix,
		"PageTitle":         s.PageTitle,
		"CurrentPath":       r.URL.Path,
		"CacheBuster":       s.cacheBuster,
		"RequestedPath":     requestedPath,
		"IsRoot":            s.PathPrefix == r.URL.Path,
		"UpDirectory":       parent,
		"Files":             files,
		"HideLinks":         s.HideLinks,
		"MarkdownContent":   markdownContent.String(),
		"MarkdownBeforeDir": s.MarkdownBeforeDir,
		"CustomCSS":         s.getCustomCSSURL(),
	}

	if isSingleMarkdown {
		data["ShouldRenderFiles"] = false
		data["IsSingleMarkdownRender"] = true
	} else {
		data["ShouldRenderFiles"] = !s.HideFilesInMarkdown
	}

	return data
}
