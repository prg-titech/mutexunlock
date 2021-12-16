package unlockcheck

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/Qs-F/unlockcheck/internal/cfg"
	"golang.org/x/tools/go/analysis"
)

type Block struct {
	Nodes []ast.Node
	Succs []*Block
	Preds []*Block
	Pos   token.Pos
	End   token.Pos
	Index int32
	Live  bool
	Raw   *cfg.Block
}

type CFG struct {
	Blocks []*Block
}

func NewCFG(pass *analysis.Pass, node ast.Node, ctrlflow *cfg.CFG, pos token.Pos /*, end token.Pos */) *CFG {
	ret := &CFG{
		Blocks: make([]*Block, len(ctrlflow.Blocks)),
	}
	if len(ctrlflow.Blocks) < 1 {
		return ret
	}
	// To support unreachable node
	for i := range ret.Blocks {
		b := ctrlflow.Blocks[i]
		ret.Blocks[i] = &Block{
			Nodes: b.Nodes,
			Index: b.Index,
			Live:  b.Live,
			Raw:   b,
		}
	}
	for i := range ret.Blocks {
		b := ctrlflow.Blocks[i]
		for _, succ := range b.Succs {
			ret.Blocks[b.Index].Succs = append(ret.Blocks[b.Index].Succs, ret.Blocks[succ.Index])
			ret.Blocks[succ.Index].Preds = append(ret.Blocks[succ.Index].Preds, ret.Blocks[b.Index])
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if n != nil {
			fmt.Println(n.Pos(), n.End(), n)
		}
		return true
	})

	visited := make(map[int32]struct{})
	var f func(block *cfg.Block, pos token.Pos)
	f = func(block *cfg.Block, pos token.Pos) {
		if len(block.Nodes) > 0 {
			pos = block.Nodes[0].Pos()
			// ret.Blocks[block.Index].End = block.Nodes[len(block.Nodes)-1].End()
		} else {
			// TODO
			// Concept: search next { or } or break or continue or fallthrouw
		}
		ret.Blocks[block.Index].Pos = pos
		for _, succ := range block.Succs {
			if _, ok := visited[succ.Index]; ok {
				continue
			}
			visited[succ.Index] = struct{}{}
			f(succ, pos)
		}
	}
	f(ctrlflow.Blocks[0], pos)

	return ret
}

func (block *Block) String() string {
	return fmt.Sprint(block.Index, block.Raw)
}

func (cfg *CFG) String() string {
	ret := ""
	for _, block := range cfg.Blocks {
		ret += fmt.Sprint(block, "@", block.Pos, "-", block.End)
		if len(block.Nodes) > 0 {
			ret += fmt.Sprint(" ", block.Nodes)
		}
		ret += fmt.Sprintln("")
		for _, succ := range block.Succs {
			ret += fmt.Sprintf("|-> %s\n", succ.String())
		}
	}
	return ret
}
