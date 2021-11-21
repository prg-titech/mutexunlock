package main

import (
	"sync"
)

func LockInFor(l []int) int {
	var mu1 sync.Mutex

	// LockはじまりでUnlockされないまま自分自身に戻ってきてしまうようなパスがあると検知したい
	// 自分自身もexit nodeと同じようなもの exitnode || 自分自身
	// Lockしているnodeを全部集めてきて，そこからグラフをもう一度辿る
	// Lockをみつるまでのvisitedと，Lockを見つけてからのvisited
	for k, v := range l {
		mu1.Lock()
		if v == 0 {
			// Here it is ok no unlock since mu1 is local, but to alleviate it is better to unlock
			return k // want "missing unlock"
		}
		// want "missng unlock"
	}

	mu1.Unlock()
	return -1 // OK
}
