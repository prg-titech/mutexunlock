package main

import (
	"sync"
)

var pointerOfMutex = &sync.Mutex{}

func PointerOfMutex() {
	pointerOfMutex.Lock()
} // want "missing unlock"

func MutexAsArg(mu *sync.Mutex) {
	mu.Lock()
} // want "missing unlock"
