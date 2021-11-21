package unlockcheck_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Qs-F/unlockcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestFuncs(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, unlockcheck.Analyzer, "a")
}

func TestPlayground(t *testing.T) {
	testdata := analysistest.TestData()
	_, err := os.Open(filepath.Join("testdata", "src", "_playground"))
	if err != nil {
		return
	}
	analysistest.Run(t, testdata, unlockcheck.Analyzer, "_playground")
}
