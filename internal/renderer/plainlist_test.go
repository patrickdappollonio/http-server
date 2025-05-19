package renderer

import (
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPlainListRenderer_Format(t *testing.T) {
	renderer := NewPlainListRenderer()
	if renderer.Format() != "plain-list" {
		t.Errorf("Expected format to be 'plain-list', got %q", renderer.Format())
	}
}

func TestPlainListRenderer_Render(t *testing.T) {
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
	renderer := NewPlainListRenderer()
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

	// Check for expected file entries
	expectedLines := []string{
		"file1.txt\n",
		"dir1/\n",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it did not", expected)
		}
	}

	// Verify the exact output (order matters)
	expectedOutput := "file1.txt\ndir1/\n"
	if output != expectedOutput {
		t.Errorf("Expected output:\n%q\nGot:\n%q", expectedOutput, output)
	}
}
