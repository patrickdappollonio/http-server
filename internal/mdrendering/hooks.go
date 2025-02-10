package mdrendering

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
)

// Hooks is a function that can be passed to the gomarkdown parser to modify the AST.
func Hooks(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch elem := node.(type) {
	case *ast.Heading:
		return processHeading(w, elem, entering)

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

	link.Children = append(link.Children, &ast.HTMLBlock{
		Leaf: ast.Leaf{
			Literal: []byte("<i class=\"fas fa-link\"></i>"),
		},
	})

	node.Children = append(node.Children, link)

	return ast.GoToNext, false
}
