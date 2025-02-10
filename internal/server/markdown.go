package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/patrickdappollonio/http-server/internal/mdrendering"
)

// allowedIndexFiles is a list of filenames that are allowed to be rendered
// in the directory listing page.
var allowedIndexFiles = []string{"README.md", "README.markdown", "readme.md", "readme.markdown", "index.md", "index.markdown"}

// supportedFullExtensions is the list of markdown extensions that are supported
// by the markdown parser.
var supportedFullExtensions = parser.CommonExtensions |
	parser.AutoHeadingIDs |
	parser.NoEmptyLineBeforeBlock

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

	// Render the markdown
	if err := mdToHTML(buf.Bytes(), placeholder, 0); err != nil {
		return fmt.Errorf("unable to render markdown file %q: %w", fullpath, err)
	}

	return nil
}

// mdToHTML converts markdown to HTML
func mdToHTML(md []byte, dest *bytes.Buffer, htmlFlags html.Flags) error {
	doc := parser.NewWithExtensions(supportedFullExtensions).Parse(md)

	if htmlFlags == 0 {
		htmlFlags = html.CommonFlags
	}

	r := html.NewRenderer(html.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: mdrendering.Hooks,
	})

	b := markdown.Render(doc, r)
	if _, err := dest.Write(b); err != nil {
		return fmt.Errorf("unable to write rendered markdown: %w", err)
	}

	return nil
}

// generateBannerMarkdown generates the markdown needed to render the banner
// in the directory listing page.
func (s *Server) generateBannerMarkdown() (string, error) {
	if s.cachedBannerMarkdown != "" {
		return s.cachedBannerMarkdown, nil
	}

	if s.BannerMarkdown == "" {
		return "", nil
	}

	s.BannerMarkdown = strings.ReplaceAll(s.BannerMarkdown, "\n", "")

	opts := html.CommonFlags |
		html.SkipHTML |
		html.SkipImages |
		html.NofollowLinks |
		html.NoreferrerLinks |
		html.NoopenerLinks

	var buf bytes.Buffer
	if err := mdToHTML([]byte(s.BannerMarkdown), &buf, opts); err != nil {
		return "", fmt.Errorf("unable to render banner markdown: %w", err)
	}

	s.cachedBannerMarkdown = buf.String()
	return s.cachedBannerMarkdown, nil
}
