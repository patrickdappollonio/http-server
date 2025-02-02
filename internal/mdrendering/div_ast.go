package mdrendering

import "github.com/yuin/goldmark/ast"

// DivKind is a unique NodeKind for  <div>.
var DivKind = ast.NewNodeKind("Div")

// Div is a custom block node that behaves like a container for child nodes.
type Div struct {
	ast.BaseBlock
}

// Kind implements the ast.Node interface.
func (d *Div) Kind() ast.NodeKind {
	return DivKind
}

// Dump implements the ast.Node interface.
func (d *Div) Dump(source []byte, level int) {
	ast.DumpHelper(d, source, level, nil, nil)
}
