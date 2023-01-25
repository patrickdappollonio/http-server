package server

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type customizedRenderer struct {
	html.Config
}

// RegisterFuncs implements goldmark.Renderer.
func (r *customizedRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindImage, r.renderImageAlign)
}

func (r *customizedRenderer) renderImageAlign(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString("<img src=\"")

	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		u := util.URLEscape(n.Destination, true)
		if bytes.HasSuffix(n.Destination, []byte(`#align-right`)) {
			_, _ = w.Write(util.EscapeHTML(bytes.TrimSuffix(u, []byte(`#align-right`))))
			_, _ = w.WriteString(`" align="right`)
		} else if bytes.HasSuffix(n.Destination, []byte(`#align-center`)) {
			_, _ = w.Write(util.EscapeHTML(bytes.TrimSuffix(u, []byte(`#align-center`))))
			_, _ = w.WriteString(`" align="center`)
		} else if bytes.HasSuffix(n.Destination, []byte(`#align-left`)) {
			_, _ = w.Write(util.EscapeHTML(bytes.TrimSuffix(u, []byte(`#align-left`))))
			_, _ = w.WriteString(`" align="left`)
		} else {
			_, _ = w.Write(util.EscapeHTML(u))
		}
	}

	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(util.EscapeHTML(n.Text(source)))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}
	if r.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func (r *customizedRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	hn := node.(*ast.Heading)

	slug, _ := node.AttributeString("id")

	if entering {
		node.SetAttribute([]byte("id"), slug)
		_, _ = w.WriteString(fmt.Sprintf(`<h%d class="content-header" id="%s">`, hn.Level, slug))
		return ast.WalkContinue, nil
	}

	_, _ = w.WriteString(fmt.Sprintf(`<a href="#%s" tabindex="-1"><i class="fas fa-link"></i></a></h%d>`, slug, hn.Level))
	return ast.WalkContinue, nil
}
