package mdrendering

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type HTTPServerRendering struct {
	html.Config
}

// RegisterFuncs implements goldmark.Renderer.
func (r *HTTPServerRendering) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindImage, r.renderImageAlign)
}

func (r *HTTPServerRendering) renderImageAlign(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	w.WriteString("<img src=\"")

	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		u := util.URLEscape(n.Destination, true)
		if bytes.HasSuffix(n.Destination, []byte(`#align-right`)) {
			w.Write(util.EscapeHTML(bytes.TrimSuffix(u, []byte(`#align-right`))))
			w.WriteString(`" align="right`)
		} else if bytes.HasSuffix(n.Destination, []byte(`#align-center`)) {
			w.Write(util.EscapeHTML(bytes.TrimSuffix(u, []byte(`#align-center`))))
			w.WriteString(`" align="center`)
		} else if bytes.HasSuffix(n.Destination, []byte(`#align-left`)) {
			w.Write(util.EscapeHTML(bytes.TrimSuffix(u, []byte(`#align-left`))))
			w.WriteString(`" align="left`)
		} else {
			w.Write(util.EscapeHTML(u))
		}
	}

	w.WriteString(`" alt="`)
	w.Write(util.EscapeHTML(n.Text(source)))
	w.WriteString(`"`)
	if n.Title != nil {
		w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		w.WriteString(`"`)
	}
	if r.XHTML {
		w.WriteString(" />")
	} else {
		w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func (r *HTTPServerRendering) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	hn := node.(*ast.Heading)

	slug, _ := node.AttributeString("id")

	if entering {
		node.SetAttribute([]byte("id"), slug)
		w.WriteString(fmt.Sprintf(`<h%d class="content-header" id="%s">`, hn.Level, slug))
		return ast.WalkContinue, nil
	}

	w.WriteString(fmt.Sprintf(`<a href="#%s" tabindex="-1"><i class="fas fa-link"></i></a></h%d>`, slug, hn.Level))
	return ast.WalkContinue, nil
}
