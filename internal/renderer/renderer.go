// Package renderer provides directory listing rendering in various formats.
package renderer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Renderer defines the interface for directory listing renderers.
type Renderer interface {
	// Format returns the format identifier for this renderer.
	Format() string
	// Render renders the directory listing to the response writer.
	Render(config Config, w http.ResponseWriter, files []os.FileInfo) error
}

// Config holds the common configuration for renderers.
type Config struct {
	// CurrentPath is the current directory path.
	CurrentPath string
	// ParentPath is the parent directory path.
	ParentPath string
	// Logger is used for logging errors.
	Logger io.Writer
}

// renderers is a slice of all available renderers
var renderers = []Renderer{
	NewJSONRenderer(),
	NewTerminalRenderer(),
	NewPlainListRenderer(),
}

// Render renders directory listing using the specified format.
func Render(format string, config Config, w http.ResponseWriter, files []os.FileInfo) error {
	if format == "html" {
		return ErrHTMLHandledSeparately
	}

	for _, r := range renderers {
		if r.Format() == format {
			err := r.Render(config, w, files)
			if err != nil {
				return fmt.Errorf("rendering with format %s: %w", format, err)
			}
			return nil
		}
	}

	return UnsupportedFormatError{Format: format}
}

// GetSupportedFormats returns a list of supported formats.
func GetSupportedFormats() []string {
	formats := make([]string, 0, len(renderers))
	for _, r := range renderers {
		if r.Format() != "html" {
			formats = append(formats, r.Format())
		}
	}
	return formats
}

// GetSupportedFormatsString returns a comma-separated string of supported formats.
func GetSupportedFormatsString() string {
	return strings.Join(GetSupportedFormats(), ", ")
}

// ParseFormat converts a string to a Format, returning
// an error if the format is not supported.
func ParseFormat(format string) (string, error) {
	if format == "" {
		return "html", nil
	}

	for _, r := range renderers {
		if r.Format() == format {
			return format, nil
		}
	}

	return "", UnsupportedFormatError{Format: format}
}

// UnsupportedFormatError is returned when an unsupported format is requested.
type UnsupportedFormatError struct {
	Format string
}

func (e UnsupportedFormatError) Error() string {
	return "unsupported output format: " + e.Format
}

func (e UnsupportedFormatError) Is(target error) bool {
	_, ok := target.(UnsupportedFormatError)
	return ok
}

// ErrHTMLHandledSeparately is returned when attempting to create an HTML renderer.
var ErrHTMLHandledSeparately = errors.New("HTML format is handled separately")
