package main

import (
	"fmt"
	"sync"
)

func A() {
	var mutex sync.Mutex

	mutex.Lock()
	mutex.Unlock()
} // OK

func B() {
	var mutex sync.Mutex

	mutex.Lock()
} // want "missing unlock"

func C() {
	var mutex sync.RWMutex

	mutex.RLock()
} // want "missing unlock"

type S struct {
	mu sync.Mutex
}

func (s *S) ReceiverTest() {
	s.mu.Lock()
	s.mu.Unlock()
}

func (s *S) D() {
	s.mu.Lock()
} // want "missing unlock"

func E() {
	fmt.Println("hello")
} // OK

func F(b bool) {
	var mutex sync.Mutex

	mutex.Lock()

	if b {
		return // want "missing unlock"
	}
	mutex.Unlock()
	fmt.Println("here")
}

func G() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	fmt.Println("aaaa")
} // OK

func H(b bool) {
	var mu sync.Mutex

	mu.Lock()

	if b {
		mu.Unlock()
	}
	fmt.Println("aaaa")
} // want "missing unlock"

func I(a, b bool) {
	var mu1 sync.Mutex
	mu1.Lock()
	if b {
		mu1.Unlock()
		return
	}
	if a {
		return // want "missing unlock"
	}
} // want "missing unlock"

func TwoMutexLockAndOneUnlock() {
	var mu1 sync.Mutex
	var mu2 sync.Mutex
	mu1.Lock()
	mu2.Lock()
	fmt.Println("hello")
	mu1.Unlock()
} // want "missing unlock"
