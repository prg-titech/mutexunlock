package main

import (
	"sync"
)

func _(l []int) int {
	var mu1 sync.Mutex

	for k, v := range l {
		mu1.Lock()
		if v == 0 {
			return k // want "missing unlock"
		} // want "missng unlock"
	}

	mu1.Unlock()
	return -1 // OK
}
