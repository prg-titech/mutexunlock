package mutexunlock

import (
	"go/ast"
	"go/types"

	"github.com/Qs-F/mutexunlock/internal/cfg"
	"golang.org/x/tools/go/analysis"
)

type MutexObj string

const (
	MutexObjInvalid            MutexObj = ""
	MutexObjSyncMutex          MutexObj = "sync.Mutex"
	MutexObjSyncRWMutex        MutexObj = "sync.RWMutex"
	MutexPointerObjSyncMutex   MutexObj = "*sync.Mutex"
	MutexPointerObjSyncRWMutex MutexObj = "*sync.RWMutex"
)

var mutexObjs = []MutexObj{
	MutexObjSyncMutex,
	MutexObjSyncRWMutex,
	MutexPointerObjSyncMutex,
	MutexPointerObjSyncRWMutex,
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

func (op MutexOp) Reverse() MutexOp {
	switch op {
	case MutexOpLock:
		return MutexOpUnlock
	case MutexOpRLock:
		return MutexOpRUnlock
	case MutexOpUnlock:
		return MutexOpLock
	case MutexOpRUnlock:
		return MutexOpRLock
	}
	return MutexOpInvalid
}

func GetMuState(pass *analysis.Pass, block *cfg.Block, node ast.Node) *MuState {
	var ret *MuState
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
		if _, ok := UnderlyingMutex(ty); !ok {
			return true
		}

		target := selectorExpr.X
		for _, op := range mutexOps {
			if selectorExpr.Sel.Name == string(op) {
				ret = &MuState{
					Op:    op,
					block: block,
					node:  target,
				}
				return false
			}
		}
		return true
	})
	return ret
}
