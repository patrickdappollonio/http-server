package renderer

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestJSONRenderer_Format(t *testing.T) {
	renderer := NewJSONRenderer()
	if renderer.Format() != "json" {
		t.Errorf("Expected format to be 'json', got %q", renderer.Format())
	}
}

func TestJSONRenderer_Render(t *testing.T) {
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
	renderer := NewJSONRenderer()
	err := renderer.Render(config, w, files)
	// Check no error
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check content type
	resp := w.Result()
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %q, got %q", "application/json", contentType)
	}

	// Decode response and validate structure
	var response struct {
		CurrentPath string `json:"current_path"`
		ParentPath  string `json:"parent_path"`
		Files       []struct {
			Name        string `json:"name"`
			Size        int64  `json:"size"`
			IsDirectory bool   `json:"is_directory"`
			ModTime     string `json:"mod_time"`
			Path        string `json:"path"`
		} `json:"files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Error decoding response JSON: %v", err)
	}

	// Validate response fields
	if response.CurrentPath != "/path/to/dir" {
		t.Errorf("Expected CurrentPath to be %q, got %q", "/path/to/dir", response.CurrentPath)
	}

	if response.ParentPath != "/path/to" {
		t.Errorf("Expected ParentPath to be %q, got %q", "/path/to", response.ParentPath)
	}

	if len(response.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(response.Files))
	}

	// Validate first file (regular file)
	if response.Files[0].Name != "file1.txt" {
		t.Errorf("Expected file name to be %q, got %q", "file1.txt", response.Files[0].Name)
	}
	if response.Files[0].Size != 100 {
		t.Errorf("Expected file size to be %d, got %d", 100, response.Files[0].Size)
	}
	if response.Files[0].IsDirectory {
		t.Errorf("Expected IsDirectory to be false")
	}
	expectedTime := fixedTime.Format(time.RFC3339)
	if response.Files[0].ModTime != expectedTime {
		t.Errorf("Expected mod time to be %q, got %q", expectedTime, response.Files[0].ModTime)
	}
	// Check path
	expectedPath := "/path/to/dir/file1.txt"
	if response.Files[0].Path != expectedPath {
		t.Errorf("Expected path to be %q, got %q", expectedPath, response.Files[0].Path)
	}

	// Validate second file (directory)
	if response.Files[1].Name != "dir1" {
		t.Errorf("Expected file name to be %q, got %q", "dir1", response.Files[1].Name)
	}
	if response.Files[1].Size != 0 {
		t.Errorf("Expected file size to be %d, got %d", 0, response.Files[1].Size)
	}
	if !response.Files[1].IsDirectory {
		t.Errorf("Expected IsDirectory to be true")
	}
	if response.Files[1].ModTime != expectedTime {
		t.Errorf("Expected mod time to be %q, got %q", expectedTime, response.Files[1].ModTime)
	}
	// Check path with trailing slash for directory
	expectedPath = "/path/to/dir/dir1/"
	if response.Files[1].Path != expectedPath {
		t.Errorf("Expected path to be %q, got %q", expectedPath, response.Files[1].Path)
	}
}
