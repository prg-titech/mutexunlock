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
	To   int32 // *cfg.Block.Index
	From int32 // *cfg.Block.Index
}

type Lock struct {
	O string
}

func cfgcheck(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
	visited := make(map[Edge]struct{})
	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)

	var f func(block *cfg.Block, ls *LockState)
	f = func(block *cfg.Block, ls *LockState) {
		for _, node := range block.Nodes {
			_, op, found, x := NodeToMutexOp(pass, node)
			if !found {
				continue
			}
			ls.Update(x, op)
		}

		// function exit point
		if len(block.Succs) == 0 {
			for _, ms := range ls.Map() {
				if ms.Peek().Locked() || ms.Peek().RLocked() {
					t, _ := formatNode(pass, ms.Peek().node)
					pass.Report(analysis.Diagnostic{
						Pos:            block.Return().Pos(),
						Message:        fmt.Sprintf("missing unlock: No unlock for %s", string(t)),
						SuggestedFixes: ms.Suggest(pass, block.Return().Pos()),
					})
				}
			}
		}

		for _, succ := range block.Succs {
			e := Edge{block.Index, succ.Index}
			if _, ok := visited[e]; ok {
				continue
			}
			visited[e] = struct{}{}
			// recursive point
			f(succ, ls.Copy())
		}
	}

	// Blocks[0] is entry point
	f(cfgs.FuncDecl(nodeFuncDecl).Blocks[0], NewLockState())
}
