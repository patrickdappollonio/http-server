package renderer

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

// terminalRenderer implements the Renderer interface for terminal-friendly output.
type terminalRenderer struct{}

// NewTerminalRenderer creates a new terminal renderer.
func NewTerminalRenderer() *terminalRenderer {
	return &terminalRenderer{}
}

// Format returns the format identifier for this renderer.
func (r *terminalRenderer) Format() string {
	return "terminal"
}

// Render renders a directory listing in terminal-friendly format.
func (r *terminalRenderer) Render(config Config, w http.ResponseWriter, files []os.FileInfo) error {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tw, "Current directory: %s\n", config.CurrentPath)
	if config.ParentPath != "" {
		fmt.Fprintf(tw, "Parent directory: %s\n", config.ParentPath)
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
			file.ModTime().Format(time.RFC3339),
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flushing tabwriter: %w", err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("writing terminal output: %w", err)
	}
	return nil
}
