package mutexunlock

import (
	"errors"
	"fmt"
	"go/ast"
	"os"
	"regexp"
	"strings"

	"github.com/Qs-F/mutexunlock/internal/ctrlflow"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "mutexunlock",
	Doc:  "If lcoked mutex is not unlocked before exitting from function or locked block, then report and fix when called with -fix option.",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		ctrlflow.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfgs, ok := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
	if !ok {
		return nil, errors.New("ctrlflow Analyzer is not ready")
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	filter := []ast.Node{
		(*ast.File)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}
	n := 0
	bn := 0
	inspect.Nodes(filter, func(node ast.Node, push bool) bool {
		if !push {
			return false
		}

		switch node := node.(type) {
		case *ast.File:
			f := pass.Fset.File(node.Pos())
			return !(strings.HasSuffix(f.Name(), "_test.go") || generated(node))
		case *ast.FuncDecl:
			n++
			bn += declCheck(pass, cfgs, node)
		case *ast.FuncLit:
			n++
			bn += litCheck(pass, cfgs, node)
		}
		return false
	})

	vlevel := os.Getenv("VERBOSE_LEVEL")
	if vlevel == "1" || vlevel == "2" {
		fmt.Println("================")
		fmt.Println("pass", "\t", pass)
		fmt.Println("N funcs", "\t", n)
		fmt.Println("N blocks", "\t", bn)
	}

	return nil, nil
}

// Belows are copied from https://github.com/ichiban/prodinspect/blob/master/inspector.go

// https://github.com/golang/go/issues/13560#issuecomment-288457920
var pattern = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func generated(f *ast.File) bool {
	for _, c := range f.Comments {
		for _, l := range c.List {
			if pattern.MatchString(l.Text) {
				return true
			}
		}
	}
	return false
}
