package unlockcheck

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func formatNode(pass *analysis.Pass, node ast.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, pass.Fset, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ms *MuStates) Report(pass *analysis.Pass, pos token.Pos, toEnd bool) {
	t, _ := formatNode(pass, ms.Peek().node)
	pass.Report(analysis.Diagnostic{
		Pos:            pos,
		Message:        fmt.Sprintf("missing unlock: No unlock for %s", string(t)),
		SuggestedFixes: ms.Suggest(pass, pos, toEnd),
	})
}

func (ms *MuStates) Suggest(pass *analysis.Pass, pos token.Pos, toEnd bool) []analysis.SuggestedFix {
	var fix ast.Node
	if ms.Peek().RLocked() {
		fix = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ms.Peek().node,
				Sel: ast.NewIdent("RUnlock"),
			},
		}
	}
	if ms.Peek().Locked() {
		fix = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ms.Peek().node,
				Sel: ast.NewIdent("Unlock"),
			},
		}
	}
	b, err := formatNode(pass, fix)
	if err != nil {
		panic(err)
	}
	if !toEnd {
		b = []byte(fmt.Sprintf("%s\n", string(b)))
	} else {
		b = []byte(fmt.Sprintf("\n%s", string(b)))
	}

	ret := []analysis.SuggestedFix{
		{
			Message: "Missing",
			TextEdits: []analysis.TextEdit{
				{
					Pos:     pos,
					End:     pos,
					NewText: b,
				},
			},
		},
	}
	return ret
}
