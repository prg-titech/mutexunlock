package mutexunlock

import (
	"go/ast"
	"go/token"

	"github.com/Qs-F/mutexunlock/internal/cfg"
	"github.com/Qs-F/mutexunlock/internal/ctrlflow"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "mutexunlock",
	Doc:  "If lcoked mutex is not unlocked before exitting from function or locked block, then report and fix when called with -fix option.",
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
			// 	checkFuncLit(pass, node)
		}
	})

	return nil, nil
}

func checkFuncDecl(pass *analysis.Pass, node *ast.FuncDecl) {
	cfgcheck(pass, node)
}

// func checkFuncLit(pass *analysis.Pass, node *ast.FuncLit) {
// 	cfgcheck(pass, node)
// }

type Check struct {
	pass       *analysis.Pass
	blocks     []*cfg.Block
	bridges    Bridges
	attributes map[Node]int
	lowlinks   map[int][]Node
}

func (rc *Check) Walk(block *cfg.Block, ls *LockState) {
	for _, node := range block.Nodes {
		mu := GetMuState(rc.pass, block, node)
		ls.Push(mu)
	}

MS:
	for _, ms := range ls.Map() {
		locked := ms.Peek().Locked()
		rlocked := ms.Peek().RLocked()
		if !(locked || rlocked) {
			continue
		}

		pos := token.NoPos

		// the last of for loop
		if rc.attributes[Node(ms.Peek().block.Index)] == rc.attributes[Node(block.Index)] {
			for _, succ := range block.Succs { // succはrange.loop または for.loopであることを確認したい
				if rc.attributes[Node(succ.Index)] == rc.attributes[Node(block.Index)] { // succとblockが同じloopに属している
					if rc.bridges.IsFrom(Node(succ.Index)) && rc.bridges.IsTo(Node(succ.Index)) { // succはbridgeのfromである (entryとかから)かつsuccはbridgeのtoである (range.done または for.done)

						if len(block.Nodes) > 0 {
							pos = block.Nodes[len(block.Nodes)-1].End()
						} else {
							pos = block.Pos
						}
						// fmt.Println(rc.pass.Fset.Position(pos).Line)
						// fmt.Println(block)
						ms.Report(rc.pass, pos, true)
						ls.Update(block, ms.Peek().node, ms.Peek().Op.Reverse())
						continue MS
					}
				}
			}
		}

		// break & return
		for _, bridge := range rc.bridges {
			if bridge.To == Node(block.Index) { // BlockはbridgeのToである
				if len(rc.lowlinks[rc.attributes[Node(bridge.From)]]) > 1 { // BridgeのFromはloopから伸びてる
					if rc.attributes[Node(ms.Peek().block.Index)] == rc.attributes[Node(bridge.From)] {
						for _, pred := range block.Preds { // predがfor.loopでないことを確認したい (range.doneでないことを確認したい)
							if rc.bridges.IsFrom(Node(pred.Index)) && rc.bridges.IsTo(Node(pred.Index)) { // predはどこかのbridgeのfromかつどこかのbridgeのtoではない (range.loopまたは for.loopではない)
								continue MS
							}
						}
						// 上でbreakされなければpredはすべてfor.loopではない
						if block.Return() != nil {
							pos = block.Return().Pos()
							// fmt.Println(rc.pass.Fset.Position(pos).Line)
							// fmt.Println(block)
							ms.Report(rc.pass, pos, false)
							ls.Update(block, ms.Peek().node, ms.Peek().Op.Reverse())
							continue MS
						}
						if len(block.Nodes) > 0 {
							pos = block.Nodes[len(block.Nodes)-1].End()
						} else {
							pos = block.Pos
						}
						// fmt.Println(rc.pass.Fset.Position(pos).Line)
						// fmt.Println(block)
						ms.Report(rc.pass, pos, true)
						ls.Update(block, ms.Peek().node, ms.Peek().Op.Reverse())
						continue MS
					}
				}
			}
		}

		// }

		if block.Return() != nil {
			// fmt.Println(rc.pass.Fset.Position(pos).Line)
			// fmt.Println(block)
			pos = block.Return().Pos()
			ms.Report(rc.pass, pos, false)
		}
	}
}

var _ WalkFunc = (&Check{}).Walk

func cfgcheck(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
	cfgs, ok := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
	if !ok {
		return
	}
	funcDecl := cfgs.FuncDecl(nodeFuncDecl)
	if funcDecl == nil {
		return
	}
	if len(funcDecl.Blocks) < 1 {
		return
	}

	// fmt.Println("=============")
	// fmt.Println(pass.Fset.Position(nodeFuncDecl.Pos()), nodeFuncDecl.Name)

	bridges, attrs, lowlinks := NewSCC(funcDecl.Blocks[0])

	// ==================== DEBUG ========================
	// for _, bridge := range bridges {
	// 	fmt.Println(
	// 		"From: ",
	// 		funcDecls.Blocks[bridge.From],
	// 		"To: ",
	// 		funcDecls.Blocks[bridge.To],
	// 	)
	// }
	// fmt.Println("lowlinks: ", lowlinks)
	// ==================== DEBUG ========================

	check := &Check{
		pass:       pass,
		bridges:    bridges,
		blocks:     funcDecl.Blocks,
		attributes: attrs,
		lowlinks:   lowlinks,
	}

	Walk(
		funcDecl.Blocks[0],
		NewLockState(),
		check.Walk,
		NewVisitedEdges(),
	)
}
