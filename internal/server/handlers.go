package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/patrickdappollonio/http-server/internal/ctype"
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
		s.serveMarkdown(currentPath, w, r)
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
	// Get the output format from query parameters
	outputFormat := r.URL.Query().Get("output")

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

	// Open the directory path and read all files
	dir, err := os.Open(requestedPath)
	if err != nil {
		// If the directory doesn't exist, render an appropriate message
		if os.IsNotExist(err) {
			s.printWarningf("attempted to access non-existent path: %s", requestedPath)
			httpError(http.StatusNotFound, w, "404 not found")
			return
		}

		// Otherwise handle it generically speaking
		s.printWarningf("unable to open directory %q: %s", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to open directory -- see application logs for more information")
		return
	}

	// Read all files in the directory then close the directory
	list, err := dir.ReadDir(-1)
	dir.Close()

	// Handle error on readdir call
	if err != nil {
		s.printWarningf("unable to read directory %q: %s", requestedPath, err)
		httpError(http.StatusInternalServerError, w, "unable to read directory -- see application logs for more information")
		return
	}

	// Render the directory listing
	sort.Sort(isort.FoldersFirst(list))

	// Generate a list of FileInfo objects
	files := make([]os.FileInfo, 0, len(list))
	for _, f := range list {
		fi, err := f.Info()
		if err != nil {
			s.printWarningf("unable to stat file %q: %s", f.Name(), err)
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
	if outputFormat != "" {
		switch strings.ToLower(outputFormat) {
		case "json":
			s.renderJSONListing(files, r.URL.Path, w)
			return
		case "terminal":
			s.renderTerminalListing(files, r.URL.Path, w)
			return
		case "plain-list":
			s.renderPlainListListing(files, w)
			return
		default:
			// Return an error for unsupported output formats
			s.printWarningf("unsupported output format requested: %s", outputFormat)
			httpError(http.StatusBadRequest, w, "unsupported output format: %s (supported formats: json, terminal, plain-list)", outputFormat)
			return
		}
	}
	// If no output format is specified, continue with HTML rendering

	// Find if among the files there's a markdown readme
	var markdownContent bytes.Buffer
	if err := s.findAndGenerateMarkdown(requestedPath, files, &markdownContent); err != nil {
		s.printWarningf("unable to generate markdown: %s", err)
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

// renderJSONListing renders a directory listing in JSON format
func (s *Server) renderJSONListing(files []os.FileInfo, currentPath string, w http.ResponseWriter) {
	parent := getParentURL(s.PathPrefix, currentPath)

	fileList := make([]FileInfo, 0, len(files))

	for _, file := range files {
		filePath := filepath.Join(currentPath, file.Name())
		if !strings.HasSuffix(filePath, "/") && file.IsDir() {
			filePath += "/"
		}

		fileList = append(fileList, FileInfo{
			Name:        file.Name(),
			Size:        file.Size(),
			IsDirectory: file.IsDir(),
			ModTime:     file.ModTime().Format("2006-01-02 15:04:05"),
			Path:        filePath,
		})
	}

	response := map[string]interface{}{
		"current_path": currentPath,
		"parent_path":  parent,
		"files":        fileList,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.printWarningf("unable to render JSON directory listing: %s", err)
		httpError(http.StatusInternalServerError, w, "unable to render JSON directory listing -- see application logs for more information")
	}
}

// renderTerminalListing renders a directory listing in terminal-friendly format
func (s *Server) renderTerminalListing(files []os.FileInfo, currentPath string, w http.ResponseWriter) {
	parent := getParentURL(s.PathPrefix, currentPath)

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tw, "Current directory: %s\n", currentPath)
	if parent != "" {
		fmt.Fprintf(tw, "Parent directory: %s\n", parent)
	}
	fmt.Fprintf(tw, "\n")

	// Write headers
	fmt.Fprintf(tw, "Type\tName\tSize\tModified\n")
	fmt.Fprintf(tw, "----\t----\t----\t--------\n")

	// Write file data
	for _, file := range files {
		fileType := "FILE"
		if file.IsDir() {
			fileType = "DIR"
		}

		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			fileType,
			file.Name(),
			file.Size(),
			file.ModTime().Format("2006-01-02 15:04:05"),
		)
	}

	tw.Flush()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(buf.Bytes())
}

// renderPlainListListing renders a simple list of filenames with directories having trailing slashes
func (s *Server) renderPlainListListing(files []os.FileInfo, w http.ResponseWriter) {
	var output strings.Builder

	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			name += "/"
		}
		output.WriteString(name + "\n")
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(output.String()))
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
	f, err := os.Open(location)
	if err != nil {
		if os.IsNotExist(err) {
			httpError(http.StatusNotFound, w, "404 not found")
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var contentType string
	if local := ctype.GetContentTypeForFilename(filepath.Base(location)); local != "" {
		contentType = local
	}

	var data [512]byte
	if _, err := f.Read(data[:]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if contentType == "" {
		if local := http.DetectContentType(data[:]); local != "application/octet-stream" {
			contentType = local
		}
	}

	charset := ""
	if utf8.Valid(data[:]) {
		charset = "utf-8"
	}

	if charset == "" {
		res, err := chardet.NewTextDetector().DetectBest(data[:])
		if err == nil && res.Confidence > 50 && res.Charset != "" {
			charset = res.Charset
		}
	}

	if contentType != "" && contentType != "application/octet-stream" {
		if charset != "" {
			contentType += "; charset=" + charset
		}

		w.Header().Set("Content-Type", contentType)
	}

	// Check if we should force download this file based on its extension
	if s.ShouldForceDownload(location) {
		filename := filepath.Base(location)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
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

func httpError(statusCode int, w http.ResponseWriter, format string, args ...any) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, format, args...)
}

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
