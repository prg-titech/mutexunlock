package mutexunlock

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"time"

	"github.com/Qs-F/mutexunlock/internal/cfg"
	"github.com/Qs-F/mutexunlock/internal/ctrlflow"
	"golang.org/x/tools/go/analysis"
)

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
						ms.Report(rc.pass, pos, true)
						ls.Update(block, ms.Peek().node, ms.Peek().Op.Reverse())
						continue MS
					}
				}
			}
		}

		// break
		for _, bridge := range rc.bridges {
			if bridge.To == Node(block.Index) { // BlockはbridgeのToである
				if len(rc.lowlinks[rc.attributes[Node(bridge.From)]]) > 1 { // BridgeのFromはloopから伸びてる
					if rc.attributes[Node(ms.Peek().block.Index)] == rc.attributes[Node(bridge.From)] {
						ok := false
					SUCCS:
						for _, succ := range block.Succs {
							for _, pred := range succ.Preds { // blockのすべてのsuccのすべてのpredが，
								if len(succ.Preds) == 1 {
									ok = true
									break SUCCS
								}
								if pred.Index == block.Index { // (自分自身は無視)
									continue
								} // じぶんと異なるpredのsuccについて，
								if rc.attributes[Node(pred.Index)] == rc.attributes[Node(bridge.From)] { // bridgeのFromと同一のSCCである？ == Loopから伸びている？
									ok = true
									break SUCCS
								}
							}
						}
						if len(block.Nodes) > 0 {
							pos = block.Nodes[len(block.Nodes)-1].End()
						} else {
							pos = block.Pos
						}
						if ok {
							ms.Report(rc.pass, pos, true)
							ls.Update(block, ms.Peek().node, ms.Peek().Op.Reverse())
							continue MS
						}
					}
				}
			}
		}

		// return
		if block.Return() != nil {
			pos = block.Return().Pos()
			ms.Report(rc.pass, pos, false)
		}
	}
}

var _ WalkFunc = (&Check{}).Walk

func check(pass *analysis.Pass, blocks []*cfg.Block) (int, int) {
	start := time.Now()
	bridges, attrs, lowlinks := NewSCC(blocks[0])
	check := &Check{
		pass:       pass,
		bridges:    bridges,
		blocks:     blocks,
		attributes: attrs,
		lowlinks:   lowlinks,
	}
	Walk(
		blocks[0],
		NewLockState(),
		check.Walk,
		NewVisitedEdges(),
	)

	t := time.Now().Sub(start)
	vlevel := os.Getenv("VERBOSE_LEVEL")
	if vlevel == "1" || vlevel == "2" {
		fmt.Println("----------------")
		fmt.Println("pos", "\t", pass.Fset.Position(blocks[0].Pos))
		fmt.Println("time", "\t", t.Nanoseconds())
		fmt.Println("N blocks", "\t", len(blocks))
		if vlevel == "2" {
			fmt.Println("succs:")
			for _, block := range blocks {
				if len(block.Succs) == 0 {
					fmt.Println("  From: ", block, "To: exit")
					continue
				}
				for _, succ := range block.Succs {
					fmt.Println("  From: ", block, "To: ", succ)
				}
			}
			fmt.Println("bridges:")
			for _, bridge := range bridges {
				fmt.Println("  From: ", blocks[bridge.From], "To: ", blocks[bridge.To])
			}
			fmt.Println("lowlinks: ", lowlinks)
		}
	}
	succs := 0
	for _, b := range blocks {
		succs += len(b.Succs)
	}
	return len(blocks), succs
}

func declCheck(pass *analysis.Pass, cfgs *ctrlflow.CFGs, nodeFuncDecl *ast.FuncDecl) (int, int) {
	funcDecl := cfgs.FuncDecl(nodeFuncDecl)
	if funcDecl == nil || len(funcDecl.Blocks) < 1 {
		return 0, 0
	}
	return check(pass, funcDecl.Blocks)
}

func litCheck(pass *analysis.Pass, cfgs *ctrlflow.CFGs, nodeFuncLit *ast.FuncLit) (int, int) {
	funcDecl := cfgs.FuncLit(nodeFuncLit)
	if funcDecl == nil || len(funcDecl.Blocks) < 1 {
		return 0, 0
	}
	return check(pass, funcDecl.Blocks)
}
