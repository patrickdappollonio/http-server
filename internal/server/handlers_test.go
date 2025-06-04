package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"embed"         // Ensure embed is imported
	"html/template" // Ensure html/template is imported for template.New
)

//go:embed templates/*
var testHandlerTemplates embed.FS

// newTestServer creates a new server instance for testing, with a temporary root directory.
// It also creates the RootCtx for the server using os.OpenRoot.
func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()

	// Create a temporary root directory for the server
	tempRoot := t.TempDir()

	// Create the root context for the server
	rootCtx, err := os.OpenRoot(tempRoot)
	if err != nil {
		t.Fatalf("Failed to create root context for temp root directory %q: %v", tempRoot, err)
	}

	// Initialize templates (minimal setup to avoid nil pointer in handlers)
	// In a real scenario, you might need a more complete template setup if testing rendering.
	// For sandboxing tests, we mainly care that the handlers don't panic before file access.
	tpl := template.New("test") // Provides a non-nil, empty template set.
	// If actual template parsing is needed in the future for these tests:
	// tpl, err := template.ParseFS(testHandlerTemplates, "templates/*.tmpl") // Corrected path for embed.FS
	// if err != nil {
	// t.Fatalf("Failed to parse test templates: %v", err)
	// }

	s := &Server{
		Path:       tempRoot,
		RootCtx:    rootCtx, // Use RootCtx now
		PathPrefix: "/",     // Default path prefix for simplicity
		LogOutput:  io.Discard,
		templates:  tpl, // Assign parsed templates
		// Add other essential fields if handlers require them to be non-nil/non-zero
		// For example, if markdown processing or other features are implicitly triggered.
		// For now, keeping it minimal for file access tests.
		DisableDirectoryList: false, // Ensure walk is tested
	}

	return s, tempRoot
}

func TestServeFile_OpenRoot_Sandboxing(t *testing.T) {
	s, tempRoot := newTestServer(t)

	// --- Setup Files ---
	// File inside the root
	publicFilePath := filepath.Join(tempRoot, "public.txt")
	if err := os.WriteFile(publicFilePath, []byte("public content"), 0644); err != nil {
		t.Fatalf("Failed to write public.txt: %v", err)
	}

	// File outside the root (in the parent of tempRoot)
	// To get the parent of tempRoot, we go up one level from tempRoot.
	parentOfTempRoot := filepath.Dir(tempRoot)
	sensitiveFilePath := filepath.Join(parentOfTempRoot, "sensitive.txt")
	if err := os.WriteFile(sensitiveFilePath, []byte("sensitive content"), 0644); err != nil {
		t.Fatalf("Failed to write sensitive.txt: %v", err)
	}
	// Cleanup sensitive file manually as it's outside t.TempDir()
	t.Cleanup(func() { os.Remove(sensitiveFilePath) })

	// File inside a subdirectory of the root
	subDirPath := filepath.Join(tempRoot, "subdir")
	if err := os.Mkdir(subDirPath, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	subFileContent := "subdir content"
	subFilePath := filepath.Join(subDirPath, "subfile.txt")
	if err := os.WriteFile(subFilePath, []byte(subFileContent), 0644); err != nil {
		t.Fatalf("Failed to write subfile.txt: %v", err)
	}

	// --- Test Cases ---
	tests := []struct {
		name           string
		urlPath        string
		expectedStatus int
		expectedBody   string // Empty if not checking body or if expecting error
	}{
		{
			name:           "access public file in root",
			urlPath:        "/public.txt",
			expectedStatus: http.StatusOK,
			expectedBody:   "public content",
		},
		{
			name:           "access file in subdirectory",
			urlPath:        "/subdir/subfile.txt",
			expectedStatus: http.StatusOK,
			expectedBody:   subFileContent,
		},
		{
			name:           "attempt path traversal to sensitive file",
			urlPath:        "/../sensitive.txt", // This path will be joined with s.Path by the handler logic
			                                     // but filepath.Rel(s.Path, joinedPath) will be tricky.
			                                     // The key is how os.OpenRoot(s.RootFD, relativePathToOpen) behaves.
			                                     // If relativePathToOpen starts with "../", os.OpenRoot should prevent it.
			expectedStatus: http.StatusNotFound, // Or potentially Bad Request, depending on how path cleaning occurs before OpenRoot
		},
		{
			name:           "attempt path traversal from subdir to sensitive file",
			urlPath:        "/subdir/../../sensitive.txt",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "access non-existent file in root",
			urlPath:        "/nonexistent.txt",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "access non-existent file in subdir",
			urlPath:        "/subdir/nonexistent.txt",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The URL path for NewRequest should be exactly what the user would type.
			// s.showOrRender will internally join s.Path + r.URL.Path, then calculate relative path for OpenRoot.
			req := httptest.NewRequest("GET", tc.urlPath, nil)
			rr := httptest.NewRecorder()

			s.showOrRender(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			if tc.expectedBody != "" {
				body := strings.TrimSpace(rr.Body.String())
				if body != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, body)
				}
			} else {
				// For error cases like 404, ensure sensitive content wasn't leaked
				if strings.Contains(rr.Body.String(), "sensitive content") {
					t.Errorf("Sensitive content leaked in response body for %s: %s", tc.urlPath, rr.Body.String())
				}
			}
		})
	}
}

func TestWalk_OpenRoot_Sandboxing(t *testing.T) {
	s, tempRoot := newTestServer(t)

	// --- Setup Files & Directories ---
	// File in root
	if err := os.WriteFile(filepath.Join(tempRoot, "file_in_root.txt"), []byte("root file"), 0644); err != nil {
		t.Fatalf("Failed to write file_in_root.txt: %v", err)
	}
	// Subdirectory in root
	subDirPath := filepath.Join(tempRoot, "sub")
	if err := os.Mkdir(subDirPath, 0755); err != nil {
		t.Fatalf("Failed to create subdir 'sub': %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDirPath, "file_in_sub.txt"), []byte("sub file"), 0644); err != nil {
		t.Fatalf("Failed to write file_in_sub.txt: %v", err)
	}

	// External directory (sibling to tempRoot, should not be accessible)
	parentOfTempRoot := filepath.Dir(tempRoot)
	externalDirPath := filepath.Join(parentOfTempRoot, "external_dir")
	if err := os.Mkdir(externalDirPath, 0755); err != nil {
		// If it already exists from a previous failed run, ignore. Otherwise, fail.
		if !os.IsExist(err) {
			t.Fatalf("Failed to create external_dir: %v", err)
		}
	} else {
		// Only schedule cleanup if we created it
		t.Cleanup(func() { os.RemoveAll(externalDirPath) })
	}
	if err := os.WriteFile(filepath.Join(externalDirPath, "external_file.txt"), []byte("external content"), 0644); err != nil {
		t.Fatalf("Failed to write external_file.txt: %v", err)
	}


	tests := []struct {
		name             string
		urlPath          string // Note: For walk, paths should end with "/"
		expectedStatus   int
		mustContain      []string // Substrings that must be in the response body for success
		mustNotContain   []string // Substrings that must NOT be in the response body
	}{
		{
			name:           "list root directory",
			urlPath:        "/",
			expectedStatus: http.StatusOK,
			mustContain:    []string{"file_in_root.txt", "sub/"},
		},
		{
			name:           "list subdirectory",
			urlPath:        "/sub/",
			expectedStatus: http.StatusOK,
			mustContain:    []string{"file_in_sub.txt"},
		},
		{
			name:           "attempt to list parent using traversal",
			urlPath:        "/../", // This should resolve relative to s.Path, then be used by OpenRoot.
			                         // Effectively asking OpenRoot to open "." on its existing FD if path becomes empty or ".".
			                         // Or, if the logic makes it "../" for OpenRoot, it should be denied.
			expectedStatus: http.StatusNotFound, // Or Bad Request. The crucial part is no listing of parent.
			mustNotContain: []string{filepath.Base(tempRoot), "external_dir"}, // tempRoot's name, external_dir name
		},
		{
			name:           "attempt to list external_dir using traversal",
			urlPath:        fmt.Sprintf("/../%s/", filepath.Base(externalDirPath)),
			expectedStatus: http.StatusNotFound,
			mustNotContain: []string{"external_file.txt"},
		},
		{
			name:           "list non-existent directory",
			urlPath:        "/nonexistent_dir/",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.urlPath, nil)
			rr := httptest.NewRecorder()

			s.showOrRender(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("URL: %s - Expected status %d, got %d. Body: %s", tc.urlPath, tc.expectedStatus, rr.Code, rr.Body.String())
			}

			for _, substr := range tc.mustContain {
				if !strings.Contains(rr.Body.String(), substr) {
					t.Errorf("URL: %s - Expected body to contain %q, but it didn't. Body: %s", tc.urlPath, substr, rr.Body.String())
				}
			}
			for _, substr := range tc.mustNotContain {
				if strings.Contains(rr.Body.String(), substr) {
					t.Errorf("URL: %s - Expected body NOT to contain %q, but it did. Body: %s", tc.urlPath, substr, rr.Body.String())
				}
			}
		})
	}
}

// Minimalistic template parsing logic, similar to what's in server/template.go
// but without all the functions if they are not strictly needed for handler execution
// up to the point of file access.
func parseTemplates(fsInstance assets.FsProvider) (*template.Template, error) {
	entries, err := fsInstance.ReadDir("internal/server/templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %w", err)
	}

	var templateFilePaths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			templateFilePaths = append(templateFilePaths, "internal/server/templates/"+entry.Name())
		}
	}

	if len(templateFilePaths) == 0 {
		return nil, fmt.Errorf("no template files found")
	}

	// Using a simple name for the template collection for testing purposes
