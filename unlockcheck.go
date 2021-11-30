package unlockcheck

import (
	"go/ast"
	"go/token"

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

type RetCheck struct {
	pass    *analysis.Pass
	lockBBs []*cfg.Block
}

func (rc *RetCheck) Check(block *cfg.Block, ls *LockState) {
	for _, node := range block.Nodes {
		_, op, found, x := NodeToMutexOp(rc.pass, node)
		if !found {
			continue
		}
		if op == MutexOpLock || op == MutexOpRLock {
			rc.lockBBs = append(rc.lockBBs, block)
		}
		ls.Update(x, op)
	}
	if len(block.Succs) == 0 {
		for _, ms := range ls.Map() {
			if ms.Peek().Locked() || ms.Peek().RLocked() {
				pos := token.NoPos
				if block.Return() != nil {
					pos = block.Return().Pos()
				} else if len(block.Nodes) > 0 {
					pos = block.Nodes[len(block.Nodes)-1].Pos()
				}
				ms.Report(rc.pass, pos)
			}
		}
	}
}

type CyclicCheck struct {
	pass *analysis.Pass
	root *cfg.Block
}

func (cc *CyclicCheck) Check(block *cfg.Block, ls *LockState) {
	for _, node := range block.Nodes {
		_, op, found, x := NodeToMutexOp(cc.pass, node)
		if !found {
			continue
		}
		ls.Update(x, op)
	}

	pos := token.NoPos
	for _, succ := range block.Succs {
		if succ == cc.root {
			pos = succ.Nodes[len(succ.Nodes)-1].Pos()
			break
		}
	}
	if pos == token.NoPos {
		return
	}

	for _, ms := range ls.Map() {
		if ms.Peek().Locked() || ms.Peek().RLocked() {
			ms.Report(cc.pass, pos)
		}
	}
}

func cfgcheck(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
	cfgs, ok := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
	if !ok {
		return
	}
	if cfgs.FuncDecl(nodeFuncDecl) == nil {
		return
	}
	if len(cfgs.FuncDecl(nodeFuncDecl).Blocks) < 1 {
		return
	}

	nodes := NewVisitedNodes()
	Walk(
		cfgs.FuncDecl(nodeFuncDecl).Blocks[0],
		NewLockState(),
		retCheck.Check,
		NewVisitedNodes(),
	)

	retCheck := &RetCheck{
		pass: pass,
	}
	Walk(
		cfgs.FuncDecl(nodeFuncDecl).Blocks[0],
		NewLockState(),
		retCheck.Check,
		NewVisitedEdges(),
	)

	for _, lockBB := range retCheck.lockBBs {
		Walk(
			lockBB,
			NewLockState(),
			(&CyclicCheck{
				root: lockBB,
				pass: pass,
			}).Check,
			NewVisitedEdges(),
		)
	}
}
