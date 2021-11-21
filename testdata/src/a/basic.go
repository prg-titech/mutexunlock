package main

import (
	"fmt"
	"sync"
)

func _() {
	var mutex sync.Mutex

	mutex.Lock()
	mutex.Unlock()
} // OK

func _() {
	var mutex sync.Mutex

	mutex.Lock()
} // want "missing unlock"

func _() {
	var mutex sync.RWMutex

	mutex.RLock()
} // want "missing unlock"

type S struct {
	mu sync.Mutex
}

func (s *S) _() {
	s.mu.Lock()
	s.mu.Unlock()
} // OK

func (s *S) _() {
	s.mu.Lock()
} // want "missing unlock"

func _() {
	fmt.Println("hello")
} // OK

func _(b bool) {
	var mutex sync.Mutex

	mutex.Lock()

	if b {
		return // want "missing unlock"
	}
	mutex.Unlock()
	fmt.Println("here")
} // OK

func _() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	fmt.Println("aaaa")
} // OK

func _(b bool) {
	var mu sync.Mutex

	mu.Lock()

	if b {
		mu.Unlock()
	}
	fmt.Println("aaaa")
} // want "missing unlock"

func _(a, b bool) {
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

func _() {
	var mu1 sync.Mutex
	var mu2 sync.Mutex
	mu1.Lock()
	mu2.Lock()
	fmt.Println("hello")
	mu1.Unlock()
} // want "missing unlock"
