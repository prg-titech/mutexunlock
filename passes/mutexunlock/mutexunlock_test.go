package mutexunlock_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Qs-F/mutexunlock/passes/mutexunlock"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestFuncs(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, mutexunlock.Analyzer, "a")
}

func TestPlayground(t *testing.T) {
	testdata := analysistest.TestData()
	_, err := os.Open(filepath.Join("testdata", "src", "_playground"))
	if err != nil {
		return
	}
	analysistest.Run(t, testdata, mutexunlock.Analyzer, "_playground")
}
