package unlockcheck

import (
	"golang.org/x/tools/go/cfg"
)

// 次に進むか進まないかを決められる関数になっている必要がある
// 1. callback関数がcontinueするか，stopするかをreturnで決められる
// 2. callback関数の中でwalkの再起関数を呼び出す
// traverseの順序を保証するには

type WalkFunc func(block *cfg.Block, ls *LockState, walks func())

func walk(root *cfg.Block, ls *LockState, f WalkFunc, visited VisitedMap) {
	walks := func() {
		for _, succ := range root.Succs {
			e := visited.New(root.Index, succ.Index)
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

	f(root, ls, walks)
}

func Walk(root *cfg.Block, ls *LockState, f WalkFunc, visited VisitedMap) {
	walk(root, ls, f, visited)
}
