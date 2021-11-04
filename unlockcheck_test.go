package unlockcheck_test

import (
	"testing"

	"github.com/Qs-F/unlockcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, unlockcheck.Analyzer, "a")
}
