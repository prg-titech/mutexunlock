package unlockcheck

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/Qs-F/unlockcheck/internal/cfg"
	"github.com/Qs-F/unlockcheck/internal/ctrlflow"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "unlockcheck",
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
		}
	})

	return nil, nil
}

// Inside FuncDecl
func checkFuncDecl(pass *analysis.Pass, nodeFuncDecl *ast.FuncDecl) {
	cfgcheck(pass, nodeFuncDecl)
}

type RetCheck struct {
	pass       *analysis.Pass
	blocks     []*cfg.Block
	bridges    []Edge
	attributes map[Node]int
	lowlinks   map[int][]Node
	lastpos    token.Pos
}

func (rc *RetCheck) Check(block *cfg.Block, ls *LockState, walks func()) {
	defer walks()

	for _, node := range block.Nodes {
		// fmt.Println("|- ", block, node, rc.pass.Fset.Position(node.Pos()))
		_, op, found, x := NodeToMutexOp(rc.pass, node)
		if !found {
			continue
		}
		ls.Update(block, x, op)
	}

	for _, ms := range ls.Map() {
		if !(ms.Peek().Locked() || ms.Peek().RLocked()) {
			continue
		}

		pos := token.NoPos

	BRIDGE:
		for _, bridge := range rc.bridges {

			// the last of for loop
			for _, succ := range block.Succs { // succはrange.loop または for.loopであることを確認したい
				// fmt.Println("見てるblock", block, "succ", succ, "bridge", bridge)
				if rc.attributes[Node(succ.Index)] == rc.attributes[Node(block.Index)] { // succとblockが同じloopに属している
					if bridge.From == succ.Index { // succはbridgeのfromである (entryとかから)
						for _, bridge2 := range rc.bridges {
							if bridge2.To == succ.Index { // かつsuccはbridgeのtoである (range.done または for.done)
								if len(block.Nodes) > 0 {
									pos = block.Nodes[len(block.Nodes)-1].End()
								} else {
									pos = block.Pos
								}
								fmt.Println(rc.pass.Fset.Position(pos).Line)
								fmt.Println(block)
								ms.Report(rc.pass, pos, true)
								ls.Update(block, ms.Peek().node, MutexOpUnlock)
								break BRIDGE
							}
						}
					}
				}
			}

			// break & return
			if bridge.To == block.Index { // BlockはbridgeのToである
				if len(rc.lowlinks[rc.attributes[Node(bridge.From)]]) > 1 { // BridgeのFromはloopから伸びてる
					for _, pred := range block.Preds { // predがfor.loopでないことを確認したい (range.doneでないことを確認したい)
						for _, bridge2 := range rc.bridges {
							for _, bridge3 := range rc.bridges {
								if bridge2.From == pred.Index && bridge3.To == pred.Index { // predはどこかのbridgeのfromかつどこかのbridgeのtoではない (range.loopまたは for.loopではない)
									break BRIDGE
								}
							}
						}
					} // 上でbreakされなければpredはすべてfor.loopではない
					if block.Return() != nil {
						pos = block.Return().Pos()
						fmt.Println(rc.pass.Fset.Position(pos).Line)
						fmt.Println(block)
						ms.Report(rc.pass, pos, false)
						ls.Update(block, ms.Peek().node, MutexOpUnlock)
						break BRIDGE
					}
					if len(block.Nodes) > 0 {
						pos = block.Nodes[len(block.Nodes)-1].End()
					} else {
						pos = block.Pos
					}
					fmt.Println(rc.pass.Fset.Position(pos).Line)
					fmt.Println(block)
					ms.Report(rc.pass, pos, true)
					ls.Update(block, ms.Peek().node, MutexOpUnlock)
					break BRIDGE
				}
			}

		}

		if pos != token.NoPos { // 上ですでにreport済み
			continue
		} else if block.Return() != nil {
			fmt.Println(rc.pass.Fset.Position(pos).Line)
			fmt.Println(block)
			pos = block.Return().Pos()
			ms.Report(rc.pass, pos, false)
		}
	}
}

var _ WalkFunc = (&RetCheck{}).Check

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

	fmt.Println("=============")
	fmt.Println(pass.Fset.Position(nodeFuncDecl.Pos()), nodeFuncDecl.Name)

	funcDecls := cfgs.FuncDecl(nodeFuncDecl)
	// fmt.Println(funcDecls)

	// locks := &Locks{
	// 	pass: pass,
	// }
	// Walk(
	// 	funcDecls.Blocks[0],
	// 	NewLockState(),
	// 	locks.Check,
	// 	NewVisitedEdges(),
	// )

	bridges, attrs, lowlinks := SCC(funcDecls.Blocks[0])

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

	// RetCheck
	retCheck := &RetCheck{
		pass:       pass,
		bridges:    bridges,
		blocks:     funcDecls.Blocks,
		attributes: attrs,
		lowlinks:   lowlinks,
	}
	Walk(
		cfgs.FuncDecl(nodeFuncDecl).Blocks[0],
		NewLockState(),
		retCheck.Check,
		NewVisitedEdges(),
	)
}
