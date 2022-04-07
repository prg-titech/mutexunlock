package main

import (
	"github.com/Qs-F/mutexunlock/passes/mutexunlock"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(mutexunlock.Analyzer)
}
