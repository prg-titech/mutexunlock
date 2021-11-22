package unlockcheck

import (
	"golang.org/x/tools/go/cfg"
)

type Edge struct {
	To   int32 // *cfg.Block.Index
	From int32 // *cfg.Block.Index
}

type Visited map[Edge]struct{}

func NewVisited() Visited {
	return make(Visited)
}

type WalkFunc func(block *cfg.Block, ls *LockState)

func walk(root *cfg.Block, ls *LockState, f WalkFunc, visited Visited) {
	f(root, ls)
	for _, succ := range root.Succs {
		e := Edge{
			From: root.Index,
			To:   succ.Index,
		}
		if _, ok := visited[e]; ok {
			continue
		}
		visited[e] = struct{}{}
		walk(succ, ls.Copy(), f, visited)
	}
}

func Walk(root *cfg.Block, ls *LockState, f WalkFunc) {
	walk(root, ls, f, NewVisited())
}
