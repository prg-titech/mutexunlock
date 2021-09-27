package main

import (
	"fmt"
	"sync"
)

func A() { // Ok
	var mutex sync.Mutex

	mutex.Lock()
	mutex.Unlock()
}

func B() { // Not OK
	var mutex sync.Mutex

	mutex.Lock()
}

func C() { // Not OK
	var mutex sync.RWMutex

	mutex.RLock()
}

type S struct {
	mu sync.Mutex
}

func (s *S) D() { // Not OK
	s.mu.Lock()
}

func E() { // OK
	fmt.Println("hello")
}

func F() { // Not OK
	var mutex sync.Mutex

	mutex.Lock()

	if true {
		return
	}
	mutex.Unlock()
}

func G() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	fmt.Println("aaaa")
}
