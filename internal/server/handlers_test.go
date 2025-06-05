package server

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed templates/*
var testHandlerTemplates embed.FS

// newTestServer creates a new server instance for testing, with a temporary root directory.
// It also creates the RootCtx for the server using os.OpenRoot.
func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()

	tempRoot := t.TempDir()

	rootCtx, err := os.OpenRoot(tempRoot)
	if err != nil {
		t.Fatalf("Failed to create root context for temp root directory %q: %v", tempRoot, err)
	}

	// Initialize templates by parsing the actual template files.
	tpl, err := template.ParseFS(testHandlerTemplates, "*.tmpl")
	if err != nil {
		t.Fatalf("Failed to parse test templates: %v", err)
	}

	s := &Server{
		Path:                 tempRoot,
		RootCtx:              rootCtx,
		PathPrefix:           "/",
		LogOutput:            io.Discard,
		templates:            tpl,
		DisableDirectoryList: false,
	}

	return s, tempRoot
}

func TestServeFile_OpenRoot_Sandboxing(t *testing.T) {
	s, tempRoot := newTestServer(t)

	publicFilePath := filepath.Join(tempRoot, "public.txt")
	if err := os.WriteFile(publicFilePath, []byte("public content"), 0644); err != nil {
		t.Fatalf("Failed to write public.txt: %v", err)
	}

	parentOfTempRoot := filepath.Dir(tempRoot)
	sensitiveFilePath := filepath.Join(parentOfTempRoot, "sensitive.txt")
	if err := os.WriteFile(sensitiveFilePath, []byte("sensitive content"), 0644); err != nil {
		t.Fatalf("Failed to write sensitive.txt: %v", err)
	}
	t.Cleanup(func() { os.Remove(sensitiveFilePath) })

	subDirPath := filepath.Join(tempRoot, "subdir")
	if err := os.Mkdir(subDirPath, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	subFileContent := "subdir content"
	subFilePath := filepath.Join(subDirPath, "subfile.txt")
	if err := os.WriteFile(subFilePath, []byte(subFileContent), 0644); err != nil {
		t.Fatalf("Failed to write subfile.txt: %v", err)
	}

	tests := []struct {
		name           string
		urlPath        string
		expectedStatus int
		expectedBody   string
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
			urlPath:        "/../sensitive.txt",
			expectedStatus: http.StatusNotFound,
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
				if strings.Contains(rr.Body.String(), "sensitive content") {
					t.Errorf("Sensitive content leaked in response body for %s: %s", tc.urlPath, rr.Body.String())
				}
			}
		})
	}
}

func TestWalk_OpenRoot_Sandboxing(t *testing.T) {
	s, tempRoot := newTestServer(t)

	if err := os.WriteFile(filepath.Join(tempRoot, "file_in_root.txt"), []byte("root file"), 0644); err != nil {
		t.Fatalf("Failed to write file_in_root.txt: %v", err)
	}
	subDirPath := filepath.Join(tempRoot, "sub")
	if err := os.Mkdir(subDirPath, 0755); err != nil {
		t.Fatalf("Failed to create subdir 'sub': %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDirPath, "file_in_sub.txt"), []byte("sub file"), 0644); err != nil {
		t.Fatalf("Failed to write file_in_sub.txt: %v", err)
	}

	parentOfTempRoot := filepath.Dir(tempRoot)
	externalDirPath := filepath.Join(parentOfTempRoot, "external_dir")
	if err := os.Mkdir(externalDirPath, 0755); err != nil {
		if !os.IsExist(err) {
			t.Fatalf("Failed to create external_dir: %v", err)
		}
	} else {
		t.Cleanup(func() { os.RemoveAll(externalDirPath) })
	}
	if err := os.WriteFile(filepath.Join(externalDirPath, "external_file.txt"), []byte("external content"), 0644); err != nil {
		t.Fatalf("Failed to write external_file.txt: %v", err)
	}

	tests := []struct {
		name           string
		urlPath        string
		expectedStatus int
		mustContain    []string
		mustNotContain []string
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
			urlPath:        "/../",
			expectedStatus: http.StatusNotFound,
			mustNotContain: []string{filepath.Base(tempRoot), "external_dir"},
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
