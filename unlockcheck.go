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

// type Lock struct {
// 	O string
// }
//
// type Start struct {
// 	Block  *cfg.Block
// 	backed bool
// }
//
// func NewStart(block *cfg.Block) *Start {
// 	return &Start{
// 		Block:  block,
// 		backed: false,
// 	}
// }
//
// func (s *Start) Get() (*cfg.Block, bool) {
// 	if s.backed {
// 		return s.Block, true
// 	}
// 	s.backed = true
// 	return s.Block, false
// }
//
// func cfgcheck(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
// 	// TODO: Remove this debug code
// 	fmt.Println("===", pass.Fset.Position(nodeFuncDecl.Pos()), "===")
//
// 	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
// 	lockBBs := []*cfg.Block{}
//
// 	var f func(cycle bool, start *Start, block *cfg.Block, ls *LockState, visited map[Edge]struct{})
// 	f = func(cycle bool, start *Start, block *cfg.Block, ls *LockState, visited map[Edge]struct{}) {
// 		for _, node := range block.Nodes {
// 			_, op, found, x := NodeToMutexOp(pass, node)
// 			if !found {
// 				continue
// 			}
// 			if (op == MutexOpLock || op == MutexOpRLock) && !cycle {
// 				lockBBs = append(lockBBs, block)
// 			}
// 			ls.Update(x, op)
// 		}
//
// 		pos := token.NoPos
// 		if !cycle && len(block.Succs) == 0 {
// 			pos = block.Return().Pos()
// 		}
// 		st, _ := start.Get()
// 		for _, succ := range block.Succs {
// 			if cycle && succ == st {
// 				pos = succ.Nodes[len(succ.Nodes)-1].Pos()
// 				break
// 			}
// 		}
// 		// function exit point
// 		if pos != token.NoPos {
// 			for _, ms := range ls.Map() {
// 				if ms.Peek().Locked() || ms.Peek().RLocked() {
// 					t, _ := formatNode(pass, ms.Peek().node)
// 					pass.Report(analysis.Diagnostic{
// 						Pos:            pos,
// 						Message:        fmt.Sprintf("missing unlock: No unlock for %s", string(t)),
// 						SuggestedFixes: ms.Suggest(pass, pos),
// 					})
// 				}
// 			}
// 		}
//
// 		for _, succ := range block.Succs {
// 			e := Edge{
// 				From: block.Index,
// 				To:   succ.Index,
// 			}
// 			if _, ok := visited[e]; ok {
// 				continue
// 			}
// 			visited[e] = struct{}{}
// 			// recursive point
// 			f(cycle, start, succ, ls.Copy(), visited)
// 		}
// 	}
//
// 	// Blocks[0] is entry point
// 	start := cfgs.FuncDecl(nodeFuncDecl).Blocks[0]
// 	visited := make(map[Edge]struct{})
// 	f(false, NewStart(start), start, NewLockState(), visited)
// 	for _, lockBB := range lockBBs {
// 		visited = make(map[Edge]struct{})
// 		f(true, NewStart(lockBB), lockBB, NewLockState(), visited)
// 	}

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
				ms.Report(rc.pass, block.Return().Pos())
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

	retCheck := &RetCheck{
		pass: pass,
	}
	Walk(
		cfgs.FuncDecl(nodeFuncDecl).Blocks[0],
		NewLockState(),
		retCheck.Check,
	)

	for _, lockBB := range retCheck.lockBBs {
		Walk(
			lockBB,
			NewLockState(),
			(&CyclicCheck{
				root: lockBB,
				pass: pass,
			}).Check,
		)
	}
}
