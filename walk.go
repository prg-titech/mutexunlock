package unlockcheck

import (
	"github.com/Qs-F/unlockcheck/internal/cfg"
)

type WalkFunc func(block *cfg.Block, ls *LockState)

func walk(root *cfg.Block, ls *LockState, f WalkFunc, visited VisitedMap) {
	f(root, ls)
	for _, succ := range root.Succs {
		e := visited.New(Node(root.Index), Node(succ.Index))
		ok, err := visited.Visited(e)
		if err != nil {
			// TODO: Error handling
			panic(err)
		}
		if ok {
			continue
		}
		if err := visited.Visit(e); err != nil {
			// TODO: Error handling
			panic(err)
		}
		walk(succ, ls.Copy(), f, visited)
	}
}

func Walk(root *cfg.Block, ls *LockState, f WalkFunc, visited VisitedMap) {
	walk(root, ls, f, visited)
}
