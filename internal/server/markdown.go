package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/patrickdappollonio/http-server/internal/mdrendering"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	mermaid "go.abhg.dev/goldmark/mermaid"
)

// allowedIndexFiles is a list of filenames that are allowed to be rendered
// in the directory listing page.
var allowedIndexFiles = []string{"README.md", "README.markdown", "readme.md", "readme.markdown", "index.md", "index.markdown"}

// renderMarkdownFile renders a markdown file from a given location
func (s *Server) renderMarkdownFile(location string, v *bytes.Buffer) error {
	// Generate a full path then open the file
	f, err := os.Open(location)
	if err != nil {
		return fmt.Errorf("unable to open markdown file %q: %w", location, err)
	}

	// Close the file when we're done
	defer f.Close()

	// Copy the file contents to an intermediate buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return fmt.Errorf("unable to read markdown file %q: %w", location, err)
	}

	// Configure goldmark
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // enables Table, Strikethrough, Linkify and TaskList
			extension.Footnote,
			extension.DefinitionList,
			&mermaid.Extender{
				RenderMode: mermaid.RenderModeClient,
				MermaidURL: s.assetpath("mermaid-11.6.0.min.js"),
			},
			&mdrendering.AdmonitionExtension{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(&mdrendering.HTTPServerRendering{}, 500),
			),
			html.WithUnsafe(),
		),
	)

	// Render the markdown
	if err := md.Convert(buf.Bytes(), v); err != nil {
		return fmt.Errorf("unable to render markdown file %q: %w", location, err)
	}

	return nil
}

// generateMarkdown generates the markdown needed to render the content
// in the directory listing page
func (s *Server) findAndGenerateMarkdown(pathLocation string, files []os.FileInfo, placeholder *bytes.Buffer) error {
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

	// Generate the full path of the found file
	fullpath := path.Join(pathLocation, foundFilename)
	return s.renderMarkdownFile(fullpath, placeholder)
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

	srvParser := parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(parser.NewParagraphParser(), 500),
		),
		parser.WithInlineParsers(
			util.Prioritized(parser.NewEmphasisParser(), 500),
			util.Prioritized(parser.NewLinkParser(), 501),
			util.Prioritized(parser.NewAutoLinkParser(), 502),
			util.Prioritized(parser.NewCodeSpanParser(), 503),
		),
	)

	md := goldmark.New(goldmark.WithParser(srvParser))

	var buf bytes.Buffer
	if err := md.Convert([]byte(s.BannerMarkdown), &buf); err != nil {
		return "", fmt.Errorf("unable to render banner markdown: %w", err)
	}

	s.cachedBannerMarkdown = buf.String()
	return s.cachedBannerMarkdown, nil
}
