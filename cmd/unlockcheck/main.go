package main

import (
	"github.com/Qs-F/unlockcheck"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(unlockcheck.Analyzer)
}
