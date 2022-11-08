package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	mermaid "github.com/abhinav/goldmark-mermaid"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var allowedIndexFiles = []string{"README.md", "readme.md", "index.md"}

// generateMarkdown generates the markdown needed to render the content
// in the directory listing page
func (s *Server) generateMarkdown(pathLocation string, files []os.FileInfo, placeholder *bytes.Buffer) error {
	// Check if markdown is enabled or not, if not, don't bother running
	// the rest of the code
	if s.DisableMarkdown {
		return nil
	}

	// Find a file name among the available options that can be rendered
	var foundFilename string
	for _, f := range files {
		for _, allowed := range allowedIndexFiles {
			if f.Name() == allowed {
				foundFilename = allowed
				break
			}
		}
	}

	// If we couldn't find one, we exit
	if foundFilename == "" {
		return nil
	}

	// Generate a full path then open the file
	fullpath := path.Join(pathLocation, foundFilename)
	f, err := os.Open(fullpath)
	if err != nil {
		return fmt.Errorf("unable to open markdown file %q: %w", fullpath, err)
	}

	// Close the file when we're done
	defer f.Close()

	// Copy the file contents to an intermediate buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return fmt.Errorf("unable to read markdown file %q: %w", fullpath, err)
	}

	// Configure goldmark
	md := goldmark.New(
		goldmark.WithExtensions(
			emoji.Emoji,
			extension.GFM,
			&mermaid.Extender{
				RenderMode: mermaid.RenderModeClient,
				MermaidJS:  s.assetpath("mermaid-9.2.0.js"),
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(),
	)

	// Render the markdown
	if err := md.Convert(buf.Bytes(), placeholder); err != nil {
		return fmt.Errorf("unable to render markdown file %q: %w", fullpath, err)
	}

	return nil
}
