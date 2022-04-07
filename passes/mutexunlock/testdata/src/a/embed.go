package main

import "sync"

type T struct {
	sync.Mutex
}

func (t *T) F() {
	t.Lock()
} // want "missing unlock"
