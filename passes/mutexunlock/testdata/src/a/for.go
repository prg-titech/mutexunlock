package main

import (
	"fmt"
	"sync"
)

func _(a bool) {
	var mu1 sync.Mutex

	for a {
		mu1.Lock()
		fmt.Println("hello") // want "missing unlock"
	}
}

func _(l []int) int {
	var mu1 sync.Mutex

	for k, v := range l {
		mu1.Lock()
		if v == 0 {
			return k // want "missing unlock"
		} // want "missing unlock"
	}
	return -1 // OK
}

func _(l []int) int {
	var mu1 sync.Mutex

	for _, v := range l {
		mu1.Lock()
		if v == 0 {
			fmt.Println("To Avoid Nodes is empty") // want "missing unlock"
			break
		} // want "missing unlock"
	}

	return -1 // OK
}

func _(l []int) int {
	var mu1 sync.Mutex

	for _, v := range l {
		mu1.Lock()
		if v == 0 { // want "missing unlock"
			break
		} // want "missing unlock"
	}

	return -1 // OK
}

func _(l []int) {
	var mu1 sync.Mutex

	mu1.Lock()
	for _, v := range l {
		fmt.Println("Hello", v)
	}
	mu1.Unlock()
}

// func _(l []int) {
// 	var mu1 sync.Mutex
//
// 	for _, v := range l {
// 		mu1.Lock()
// 		if v == 0 {
// 			if v != 2 {
// 				fmt.Println(v)
// 			}
// 			mu1.Unlock()
// 			return
// 		}
// 		mu1.Unlock()
// 	}
// }

// func _(l []int) {
// 	var mu1 sync.Mutex
//
// 	for _, v := range l {
// 		mu1.Lock()
// 		if v == 0 {
// 			if v != 2 {
// 				fmt.Println(v)
// 			}
// 			mu1.Unlock()
// 			break
// 		}
// 		mu1.Unlock()
// 	}
// }

func _(l []int) int {
	var mu1 sync.Mutex

	for {
		mu1.Lock()
		if l[0] == 0 { // want "missing unlock"
			break
		}
		mu1.Unlock()
	}

	return -1 // OK
}

func _(l []int) int {
	var mu1 sync.Mutex

	for {
		mu1.Lock()
		if l[0] == 0 {
			return -1 // want "missing unlock"
		}
		mu1.Unlock()
	}

	return -1 // OK
}
