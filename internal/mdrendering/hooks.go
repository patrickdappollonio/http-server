package mdrendering

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

// Hooks is a function that can be passed to the gomarkdown parser to modify the AST.
func Hooks(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch n := node.(type) {
	case *ast.Heading:
		return processHeading(w, n, entering)
	case *ast.Image:
		return processImages(w, n, entering)
	default:
		return ast.GoToNext, false
	}
}

// processHeading is a hook that adds a link to each heading.
func processHeading(w io.Writer, node *ast.Heading, entering bool) (ast.WalkStatus, bool) {
	if node.Attribute == nil {
		node.Attribute = &ast.Attribute{}
	}

	node.Classes = append(node.Classes, []byte("content-header"))
	link := &ast.Link{
		Destination:          []byte("#" + string(node.HeadingID)),
		AdditionalAttributes: []string{"tabindex=\"-1\""},
	}

	// Append an icon inside the link.
	link.Children = append(link.Children, &ast.HTMLBlock{
		Leaf: ast.Leaf{
			Literal: []byte("<i class=\"fas fa-link\"></i>"),
		},
	})

	node.Children = append(node.Children, link)

	return ast.GoToNext, false
}

// processImages is a hook that processes images for alignment markers
// It checks if an image's Destination has a suffix like "#align-right", "#align-center" or "#align-left".
// It then removes that marker and stores the alignment value in the Title field.
func processImages(w io.Writer, node *ast.Image, entering bool) (ast.WalkStatus, bool) {
	// Only modify on entering.
	if !entering {
		return ast.GoToNext, false
	}

	dest := node.Destination
	var align string
	if bytes.HasSuffix(dest, []byte("#align-right")) {
		align = "right"
		dest = bytes.TrimSuffix(dest, []byte("#align-right"))
	} else if bytes.HasSuffix(dest, []byte("#align-center")) {
		align = "center"
		dest = bytes.TrimSuffix(dest, []byte("#align-center"))
	} else if bytes.HasSuffix(dest, []byte("#align-left")) {
		align = "left"
		dest = bytes.TrimSuffix(dest, []byte("#align-left"))
	}

	// Update the image destination without the alignment marker.
	node.Destination = dest

	// If an alignment was detected, add it
	if align != "" {
		if node.Attribute == nil {
			node.Attribute = &ast.Attribute{}
		}

		if node.Attribute.Attrs == nil {
			node.Attribute.Attrs = map[string][]byte{}
		}

		node.Attribute.Attrs["align"] = []byte(align)
	}
	return ast.GoToNext, false
}
