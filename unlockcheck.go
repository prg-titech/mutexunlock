package unlockcheck

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/cfg"
)

var Analyzer = &analysis.Analyzer{
	Name: "unlockcheck",
	Doc:  "hoge",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		ctrlflow.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	filter := []ast.Node{
		(*ast.FuncDecl)(nil),
		// (*ast.FuncLit)(nil),
	}
	inspect.Preorder(filter, func(node ast.Node) {
		switch node := node.(type) {
		case *ast.FuncDecl:
			checkFuncDecl(pass, node)
			// case *ast.FuncLit:
		}
	})

	return nil, nil
}

// Inside FuncDecl
func checkFuncDecl(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
	cfgcheck(pass, nodeFuncDecl)
}

type Edge struct {
	To   int32
	From int32
}

func cfgcheck(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
	visited := make(map[Edge]struct{})
	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)

	var f func(block *cfg.Block, locked bool)
	f = func(block *cfg.Block, locked bool) {
		var mutexObj MutexObj
		var mutexOp MutexOp
		for _, node := range block.Nodes {
			obj, op, found, _ := NodeToMutexOp(pass, node)
			if !found {
				continue
			}
			switch op {
			case MutexOpLock:
				mutexObj = obj
				mutexOp = op
				locked = true
			case MutexOpRLock:
				mutexObj = obj
				mutexOp = op
				locked = true
			case MutexOpUnlock:
				mutexObj = obj
				mutexOp = op
				locked = false
			case MutexOpRUnlock:
				mutexObj = obj
				mutexOp = op
				locked = false
			}
		}

		// function exit point
		if len(block.Succs) == 0 && locked {
			pass.Report(analysis.Diagnostic{
				Pos:     block.Return().Pos(),
				Message: fmt.Sprintf("No corresponding %s.%s() call", string(mutexObj), string(mutexOp)),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Maybe missing Unlock",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     block.Return().Pos(),
								End:     block.Return().Pos() + 1, // FIXME mutex varuable is not obtainable
								NewText: []byte{},
							},
						},
					},
				},
			})
		}

		for _, succ := range block.Succs {
			e := Edge{block.Index, succ.Index}
			if _, ok := visited[e]; ok {
				continue
			}
			visited[e] = struct{}{}
			f(succ, locked)
		}
	}

	// Blocks[0] is entry point
	f(cfgs.FuncDecl(nodeFuncDecl).Blocks[0], false)
}
