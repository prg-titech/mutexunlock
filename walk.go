package unlockcheck

import (
	"golang.org/x/tools/go/cfg"
)

type WalkFunc func(block *cfg.Block, ls *LockState)

func walk(root *cfg.Block, ls *LockState, f WalkFunc, visited VisitedMap) {
	f(root, ls)

	// SCC
	if len(root.Succs) == 0 {
	}

	for _, succ := range root.Succs {
		e := Edge{
			From: root.Index,
			To:   succ.Index,
		}
		ok, err := visited.Visited(e)
		if err != nil {
			// TODO: Error handling
		}
		if ok {
			continue
		}
		if err := visited.Done(e); err != nil {
			// TODO: Error handling
		}
		walk(succ, ls.Copy(), f, visited)
	}
}

func Walk(root *cfg.Block, ls *LockState, f WalkFunc, visited VisitedMap) {
	walk(root, ls, f, visited)
}
