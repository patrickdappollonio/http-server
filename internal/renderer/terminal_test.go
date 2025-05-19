package renderer

import (
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTerminalRenderer_Format(t *testing.T) {
	renderer := NewTerminalRenderer()
	if renderer.Format() != "terminal" {
		t.Errorf("Expected format to be 'terminal', got %q", renderer.Format())
	}
}

func TestTerminalRenderer_Render(t *testing.T) {
	// Create test files
	fixedTime := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)
	files := []os.FileInfo{
		mockFileInfo{
			name:    "file1.txt",
			size:    100,
			mode:    0o644,
			modTime: fixedTime,
			isDir:   false,
		},
		mockFileInfo{
			name:    "dir1",
			size:    0,
			mode:    0o755 | os.ModeDir,
			modTime: fixedTime,
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

	// Create renderer and render
	renderer := NewTerminalRenderer()
	err := renderer.Render(config, w, files)
	// Check no error
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check content type
	resp := w.Result()
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Expected Content-Type %q, got %q", "text/plain; charset=utf-8", contentType)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	output := string(body)

	// Check that the output contains the expected information
	expectedStrings := []string{
		"Current directory: /path/to/dir",
		"Parent directory: /path/to",
		"Type",
		"Name",
		"Size",
		"Modified",
		"FILE", // For file1.txt
		"file1.txt",
		"100", // Size of file1.txt
		"DIR", // For dir1
		"dir1",
		"0", // Size of dir1
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it did not", expected)
		}
	}

	// Validate formatting structure (headers then data)
	if !strings.Contains(output, "Type") || !strings.Contains(output, "Name") ||
		!strings.Contains(output, "Size") || !strings.Contains(output, "Modified") {
		t.Errorf("Expected output to contain header columns")
	}

	// Check the timestamp format
	expectedTimeStr := fixedTime.Format(time.RFC3339)
	if !strings.Contains(output, expectedTimeStr) {
		t.Errorf("Expected output to contain timestamp %q, but it did not", expectedTimeStr)
	}
}
