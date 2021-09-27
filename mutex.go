package unlockcheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type MutexObj string

const (
	MutexObjInvalid     MutexObj = ""
	MutexObjSyncMutex   MutexObj = "sync.Mutex"
	MutexObjSyncRWMutex MutexObj = "sync.RWMutex"
)

var mutexObjs = []MutexObj{
	MutexObjSyncMutex,
	MutexObjSyncRWMutex,
}

func UnderlyingMutex(ty types.TypeAndValue) (MutexObj, bool) {
	for _, mu := range mutexObjs {
		if ty.Type.String() == string(mu) {
			return mu, true
		}
	}
	return MutexObjInvalid, false
}

type MutexOp string

const (
	MutexOpInvalid MutexOp = ""
	MutexOpLock    MutexOp = "Lock"
	MutexOpRLock   MutexOp = "RLock"
	MutexOpUnlock  MutexOp = "Unlock"
	MutexOpRUnlock MutexOp = "RUnlock"
)

var mutexOps = []MutexOp{
	MutexOpLock,
	MutexOpRLock,
	MutexOpUnlock,
	MutexOpRUnlock,
}

func NodeToMutexOp(pass *analysis.Pass, node ast.Node) (mutexObj MutexObj, mutexOp MutexOp, found bool, target ast.Expr) {
	mutexObj = MutexObjInvalid
	mutexOp = MutexOpInvalid
	found = false

	ast.Inspect(node, func(node ast.Node) bool {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ty, ok := pass.TypesInfo.Types[selectorExpr.X]
		if !ok {
			return true
		}
		obj, ok := UnderlyingMutex(ty)
		if !ok {
			return true
		}

		target = selectorExpr.X
		for _, op := range mutexOps {
			if selectorExpr.Sel.Name == string(op) {
				mutexObj = obj
				mutexOp = op
				found = true
				return false
			}
		}
		return true
	})
	return mutexObj, mutexOp, found, target
}
