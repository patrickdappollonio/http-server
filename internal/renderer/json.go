package renderer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// fileInfo represents a file's metadata for JSON output.
type fileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	IsDirectory bool   `json:"is_directory"`
	ModTime     string `json:"mod_time"`
	Path        string `json:"path"`
}

// jsonResponse represents the JSON response structure.
type jsonResponse struct {
	CurrentPath string     `json:"current_path"`
	ParentPath  string     `json:"parent_path"`
	Files       []fileInfo `json:"files"`
}

// JSONRenderer implements the Renderer interface for JSON output.
type JSONRenderer struct{}

// NewJSONRenderer creates a new JSON renderer.
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

// Format returns the format identifier for this renderer.
func (r *JSONRenderer) Format() string {
	return "json"
}

// Render renders a directory listing in JSON format.
func (r *JSONRenderer) Render(config Config, w http.ResponseWriter, files []os.FileInfo) error {
	fileList := make([]fileInfo, 0, len(files))

	for _, file := range files {
		filePath := filepath.Join(config.CurrentPath, file.Name())
		if !strings.HasSuffix(filePath, "/") && file.IsDir() {
			filePath += "/"
		}

		fileList = append(fileList, fileInfo{
			Name:        file.Name(),
			Size:        file.Size(),
			IsDirectory: file.IsDir(),
			ModTime:     file.ModTime().Format(time.RFC3339),
			Path:        filePath,
		})
	}

	response := jsonResponse{
		CurrentPath: config.CurrentPath,
		ParentPath:  config.ParentPath,
		Files:       fileList,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return fmt.Errorf("encoding JSON response: %w", err)
	}
	return nil
}
