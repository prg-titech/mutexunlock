package main

import (
	"fmt"
	"sync"
)

func For_A(a bool) {
	var mu1 sync.Mutex

	for a {
		mu1.Lock()
		fmt.Println("hello")
	} // want "missing unlock"
}

func For_B(l []int) int {
	var mu1 sync.Mutex

	for k, v := range l {
		mu1.Lock()
		if v == 0 {
			return k // want "missing unlock"
		}
	} // want "missing unlock"

	return -1 // OK
}

func For_C(l []int) int {
	var mu1 sync.Mutex

	for _, v := range l {
		mu1.Lock()
		if v == 0 {
			break // want "missing unlock"
		}
	} // want "missing unlock"

	return -1 // OK
}
