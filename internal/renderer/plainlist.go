package renderer

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// PlainListRenderer implements the Renderer interface for plain list output.
type PlainListRenderer struct{}

// NewPlainListRenderer creates a new plainlist renderer.
func NewPlainListRenderer() *PlainListRenderer {
	return &PlainListRenderer{}
}

// Format returns the format identifier for this renderer.
func (r *PlainListRenderer) Format() string {
	return "plain-list"
}

// Render renders a directory listing as a simple list of filenames.
func (r *PlainListRenderer) Render(_ Config, w http.ResponseWriter, files []os.FileInfo) error {
	var output strings.Builder

	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			name += "/"
		}
		output.WriteString(name + "\n")
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := w.Write([]byte(output.String()))
	if err != nil {
		return fmt.Errorf("writing plain list output: %w", err)
	}
	return nil
}
