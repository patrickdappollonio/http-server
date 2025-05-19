package renderer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() any           { return nil }

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()

	// Check that we have the expected formats
	expectedFormats := []string{"json", "terminal", "plain-list"}

	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}

	// Check that each expected format is in the list
	for _, expected := range expectedFormats {
		found := false
		for _, actual := range formats {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %q not found in supported formats", expected)
		}
	}
}

func TestGetSupportedFormatsString(t *testing.T) {
	formatString := GetSupportedFormatsString()

	// This might need to be updated if the supported formats change
	expectedFormats := []string{"json", "terminal", "plain-list"}

	for _, expected := range expectedFormats {
		if !strings.Contains(formatString, expected) {
			t.Errorf("Expected format %q not found in format string %q", expected, formatString)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "empty string defaults to html",
			input:       "",
			expected:    "html",
			expectError: false,
		},
		{
			name:        "json format",
			input:       "json",
			expected:    "json",
			expectError: false,
		},
		{
			name:        "terminal format",
			input:       "terminal",
			expected:    "terminal",
			expectError: false,
		},
		{
			name:        "plain-list format",
			input:       "plain-list",
			expected:    "plain-list",
			expectError: false,
		},
		{
			name:        "unsupported format",
			input:       "invalid",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := ParseFormat(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, got nil", tt.input)
				}
				var formatErr UnsupportedFormatError
				if !errors.As(err, &formatErr) {
					t.Errorf("Expected UnsupportedFormatError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if format != tt.expected {
					t.Errorf("Expected format %q, got %q", tt.expected, format)
				}
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		expectError    bool
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "json format",
			format:         "json",
			expectError:    false,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "terminal format",
			format:         "terminal",
			expectError:    false,
			expectedStatus: http.StatusOK,
			expectedType:   "text/plain; charset=utf-8",
		},
		{
			name:           "plain-list format",
			format:         "plain-list",
			expectError:    false,
			expectedStatus: http.StatusOK,
			expectedType:   "text/plain; charset=utf-8",
		},
		{
			name:        "html format (handled separately)",
			format:      "html",
			expectError: true,
		},
		{
			name:        "invalid format",
			format:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test files
			files := []os.FileInfo{
				mockFileInfo{
					name:    "file1.txt",
					size:    100,
					mode:    0o644,
					modTime: time.Now(),
					isDir:   false,
				},
				mockFileInfo{
					name:    "dir1",
					size:    0,
					mode:    0o755 | os.ModeDir,
					modTime: time.Now(),
					isDir:   true,
				},
			}

			// Create config
			config := Config{
				CurrentPath: "/path/to/dir",
				ParentPath:  "/path/to",
				Logger:      io.Discard,
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call Render
			err := Render(tt.format, config, w, files)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for format %q, got nil", tt.format)
				}
				if tt.format == "html" && err != ErrHTMLHandledSeparately {
					t.Errorf("Expected ErrHTMLHandledSeparately, got %v", err)
				}
				if tt.format != "html" {
					var formatErr UnsupportedFormatError
					if !errors.As(err, &formatErr) {
						t.Errorf("Expected UnsupportedFormatError, got %T", err)
					}
				}
				return
			}

			// Check no error
			if err != nil {
				t.Errorf("Unexpected error for format %q: %v", tt.format, err)
				return
			}

			// Check response status
			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if contentType != tt.expectedType {
				t.Errorf("Expected Content-Type %q, got %q", tt.expectedType, contentType)
			}
		})
	}
}

func TestUnsupportedFormatError(t *testing.T) {
	err := UnsupportedFormatError{Format: "test"}

	// Test Error method
	if err.Error() != "unsupported output format: test" {
		t.Errorf("Expected error message %q, got %q", "unsupported output format: test", err.Error())
	}

	// Test Is method
	if !errors.Is(err, UnsupportedFormatError{}) {
		t.Errorf("Expected err to match UnsupportedFormatError")
	}

	// Test with different error
	if errors.Is(err, errors.New("some other error")) {
		t.Errorf("Expected err not to match 'some other error'")
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return s == substr ||
		strings.HasPrefix(s, substr+",") ||
		strings.HasSuffix(s, ","+substr) ||
		strings.Contains(s, ","+substr+",")
}

// errorWriter is a test helper that fails on Write
type errorWriter struct {
	http.ResponseWriter
}

func (w *errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("forced write error")
}

func (w *errorWriter) Header() http.Header {
	return http.Header{}
}

func TestRenderErrors(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "json format",
			format: "json",
		},
		{
			name:   "terminal format",
			format: "terminal",
		},
		{
			name:   "plain-list format",
			format: "plain-list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal test files
			files := []os.FileInfo{
				mockFileInfo{
					name:    "file1.txt",
					size:    100,
					modTime: time.Now(),
					isDir:   false,
				},
			}

			// Create config
			config := Config{
				CurrentPath: "/path",
				ParentPath:  "/",
				Logger:      io.Discard,
			}

			// Create error response writer
			w := httptest.NewRecorder()
			ew := &errorWriter{ResponseWriter: w}

			// Call Render and expect error
			err := Render(tt.format, config, ew, files)
			if err == nil {
				t.Errorf("Expected error for format %q, got nil", tt.format)
			}
		})
	}
}
