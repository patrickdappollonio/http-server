package mdrendering

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindAdmonition is the NodeKind for our custom admonition block.
var KindAdmonition = ast.NewNodeKind("Admonition")

// Admonition represents a GitHub-style admonition block.
type Admonition struct {
	ast.BaseBlock
	AdmonitionType string // e.g. "note", "warning", "tip", etc.
}

func (n *Admonition) Kind() ast.NodeKind {
	return KindAdmonition
}

func (n *Admonition) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"AdmonitionType": n.AdmonitionType}, nil)
}

// ────────────────────────────────────────────────────────────────────────────────
// Parser
// ────────────────────────────────────────────────────────────────────────────────
type admonitionParser struct{}

// NewAdmonitionParser returns a parser that recognizes GitHub-style admonitions.
func NewAdmonitionParser() parser.BlockParser {
	return &admonitionParser{}
}

func (b *admonitionParser) Trigger() []byte { return []byte{'>'} }

func (b *admonitionParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, seg := reader.PeekLine()
	trimmed := bytes.TrimSpace(line)
	// Must start with "> [!"
	if !bytes.HasPrefix(trimmed, []byte("> [!")) {
		return nil, parser.NoChildren
	}
	end := bytes.IndexByte(trimmed, ']')
	if end < 0 {
		return nil, parser.NoChildren
	}
	raw := trimmed[len("> [!"):end]
	node := &Admonition{AdmonitionType: strings.ToLower(string(raw))}
	reader.Advance(seg.Stop - seg.Start)
	return node, parser.HasChildren
}

func (b *admonitionParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	trimmed := bytes.TrimLeft(line, " \t")
	if len(trimmed) > 0 && trimmed[0] == '>' {
		return parser.Continue | parser.HasChildren
	}
	return parser.Close
}

func (b *admonitionParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}
func (b *admonitionParser) CanInterruptParagraph() bool                                { return true }
func (b *admonitionParser) CanAcceptIndentedLine() bool                                { return false }

// ────────────────────────────────────────────────────────────────────────────────
// AST Transformer
// ────────────────────────────────────────────────────────────────────────────────
type admonitionTransformer struct{}

// NewAdmonitionTransformer returns an ASTTransformer that unwraps nested blockquotes.
func NewAdmonitionTransformer() parser.ASTTransformer {
	return &admonitionTransformer{}
}

func (t *admonitionTransformer) Transform(root *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		adm, ok := n.(*Admonition)
		if !ok {
			return ast.WalkContinue, nil
		}
		// unwrap any blockquote children into the admonition
		for child := adm.FirstChild(); child != nil; {
			next := child.NextSibling()
			if bq, isBQ := child.(*ast.Blockquote); isBQ {
				for gc := bq.FirstChild(); gc != nil; {
					gnext := gc.NextSibling()
					bq.RemoveChild(bq, gc)
					adm.InsertBefore(adm, child, gc)
					gc = gnext
				}
				adm.RemoveChild(adm, child)
			}
			child = next
		}
		return ast.WalkContinue, nil
	})
}

// ────────────────────────────────────────────────────────────────────────────────
// Renderer
// ────────────────────────────────────────────────────────────────────────────────
type admonitionHTMLRenderer struct{ html.Config }

// NewAdmonitionHTMLRenderer returns a renderer for our admonition node.
func NewAdmonitionHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &admonitionHTMLRenderer{Config: html.NewConfig()}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// SVG icon path constants
const (
	noteIconPath      = `<path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path>`
	noteIconTip       = `<path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path>`
	noteIconImportant = `<path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path>`
	noteIconWarning   = `<path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path>`
	noteIconCaution   = `<path d="M4.47.22A.749.749 0 0 1 5 0h6c.199 0 .389.079.53.22l4.25 4.25c.141.14.22.331.22.53v6a.749.749 0 0 1-.22.53l-4.25 4.25A.749.749 0 0 1 11 16H5a.749.749 0 0 1-.53-.22L.22 11.53A.749.749 0 0 1 0 11V5c0-.199.079-.389.22-.53Zm.84 1.28L1.5 5.31v5.38l3.81 3.81h5.38l3.81-3.81V5.31L10.69 1.5ZM8 4a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5A.75.75 0 0 1 8 4Zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path>`
)

func (r *admonitionHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAdmonition, r.renderAdmonition)
}

func (r *admonitionHTMLRenderer) renderAdmonition(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*Admonition)
	if entering {
		// open admonition wrapper
		w.WriteString(`<div class="admonition admonition-` + n.AdmonitionType + `">`)
		// title + icon
		w.WriteString(`<p class="admonition-title">`)
		w.WriteString(`<svg class="admonition-icon" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">`)
		// insert appropriate icon constant
		switch n.AdmonitionType {
		case "note":
			w.WriteString(noteIconPath)
		case "tip":
			w.WriteString(noteIconTip)
		case "important":
			w.WriteString(noteIconImportant)
		case "warning":
			w.WriteString(noteIconWarning)
		case "caution":
			w.WriteString(noteIconCaution)
		default:
			w.WriteString(noteIconPath)
		}
		w.WriteString(`</svg>`)
		w.WriteString(`<strong>` + uppercaseFirstCharacter(n.AdmonitionType) + `</strong>`)
		w.WriteString(`</p>`)
	} else {
		// close admonition wrapper
		w.WriteString(`</div>`)
	}
	return ast.WalkContinue, nil
}

func uppercaseFirstCharacter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

// ────────────────────────────────────────────────────────────────────────────────
// Extension
// ────────────────────────────────────────────────────────────────────────────────
type AdmonitionExtension struct{}

func (e *AdmonitionExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewAdmonitionParser(), 100),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewAdmonitionTransformer(), 0),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewAdmonitionHTMLRenderer(), 500),
		),
	)
}
