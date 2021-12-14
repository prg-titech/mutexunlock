package unlockcheck

import (
	"fmt"
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

	// Succsが0 = 帯域脱出，一番優先スべき
	// なんけど，本来forの中で閉じるべきものがforを出たあとのreturnの前でのUnlockになるので，これではだめ
	// 1. forの中からのreturn
	// 2. forの中のbreakの前
	// 3. for.body の一番最後
	// 4. 通常return

	// loopの中なのにbridgeというわけではなくSuccsが0 = loopの中からのreturn
	// → loopの中のifは別物

	// for _, succ := range block.Succs {
	// 	for _, bridge := range rc.bridges {
	// 		if (Edge{From: block.Index, To: succ.Index} == bridge) {
	// 			fmt.Println("ここはbridgeです", block, succ)
	// 			continue
	// 		}
	// 		if len(block.Succs) == 0 {
	// 			fmt.Println("ここはloopからのreturnです")
	// 		}
	// 	}
	// }

	for _, ms := range ls.Map() {
		if !(ms.Peek().Locked() || ms.Peek().RLocked()) {
			continue
		}

		pos := token.NoPos

		// the last of for loop
		for _, bridge := range rc.bridges {
			if len(rc.lowlinks[rc.attributes[Node(ms.Peek().block.Index)]]) > 1 { // 自分はloopの中にいる？
				if bridge.From == ms.Peek().block.Index {
					target := rc.blocks[bridge.To].Nodes
					if len(target) > 0 { // breakのときはこれが0になってしまう……
						pos = target[len(target)-1].Pos()
						ms.Report(rc.pass, pos, false)
					} else {
						pos = rc.lastpos
						if pos != token.NoPos {
							ms.Report(rc.pass, pos, true)
						}
					}
				}
			}

			// break & return
			for _, succ := range ms.Peek().block.Succs {
				if len(rc.lowlinks[rc.attributes[Node(succ.Index)]]) > 1 { // 1つ飛んだ先がloopかどうか？
					if bridge.From == succ.Index { // bridge.From かつ どっかのbridgeのToになっているところ？ → 違う， それはfor.done ここで引っかかってる
						for _, bridge2 := range rc.bridges {
							if bridge2.To == succ.Index { // ここでsuccがrange.loopであることがわかる
								target := ms.Peek().block.Nodes
								if len(target) > 0 {
									pos = target[len(target)-1].End()
									ms.Report(rc.pass, pos, true)
								} else {
									pos = rc.lastpos
									if pos != token.NoPos {
										ms.Report(rc.pass, pos, true)
									}
								}
							}
						}
					}
				}
			}
		}
		if pos != token.NoPos { // 上ですでにreport済み
			continue
		} else if block.Return() != nil {
			if len(block.Succs) == 0 {
				pos = block.Return().Pos()
				ms.Report(rc.pass, pos, false)
			}
			// } else if len(block.Nodes) > 0 {
			// 	if len(block.Succs) == 0 {
			// 		pos = block.Nodes[len(block.Nodes)-1].Pos()
			// 		ms.Report(rc.pass, pos, false)
			// 	}
		}
		rc.lastpos = pos
		// ms.Report(rc.pass, pos)
		// ls.Update(block, ms.Peek().node, MutexOpUnlock)
	}
	// }
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

	fmt.Println(pass.Fset.Position(nodeFuncDecl.Pos()), nodeFuncDecl.Name)

	funcDecls := cfgs.FuncDecl(nodeFuncDecl)

	// ==================== DEBUG ========================
	fmt.Println("=============")
	for _, v := range funcDecls.Blocks {
		fmt.Println(v)
		for _, node := range v.Nodes {
			fmt.Println("  +- ", node)
		}
		for _, succ := range v.Succs {
			fmt.Println("|- ", succ)
		}
	}
	// ==================== DEBUG ========================

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
