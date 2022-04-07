package mutexunlock

import (
	"errors"
	"go/ast"
	"go/token"

	"github.com/Qs-F/mutexunlock/internal/cfg"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	ErrDoublyLocked     = errors.New("Lock after Lock")
	ErrInvalidOp        = errors.New("Invalid Operation to mutex object")
	ErrLockLocked       = errors.New("Attempt to Lock the Locked mutex object")
	ErrLockRLocked      = errors.New("Attempt to Lock the RLocked mutex object")
	ErrRLockLocked      = errors.New("Attempt to RLock the Locked mutex object")
	ErrRLockRLocked     = errors.New("Attempt to RLock the RLocked mutex object")
	ErrUnlockUnlocked   = errors.New("Attempt to Unlock the Unlocked mutex object")
	ErrUnlockRLocked    = errors.New("Attempt to Unlock the RLocked mutex object")
	ErrRUnlockLocked    = errors.New("Attempt to RUnlock the Locked mutex object")
	ErrRUnlockRUnlocked = errors.New("Attempt to RUnlock the RUnlocked mutex object")
)

type MuState struct {
	Op MutexOp

	node    ast.Expr
	block   *cfg.Block
	locked  bool
	rlocked bool
	err     error
}

func (mu *MuState) Error() error {
	return mu.err
}

func (mu *MuState) Locked() bool {
	return mu.locked
}

func (mu *MuState) RLocked() bool {
	return mu.rlocked
}

type MuStates []*MuState

func (ms *MuStates) Len() int {
	if ms == nil {
		return 0
	}
	return len(*ms)
}

func (ms *MuStates) Peek() *MuState {
	if ms.Len() < 1 {
		return &MuState{
			Op: MutexOpInvalid,
		}
	}
	return (*ms)[ms.Len()-1]
}

func (ms *MuStates) Push(state *MuState) {
	prev := ms.Peek()
	var e error
	locked := prev.Locked()
	rlocked := prev.RLocked()

	switch state.Op {
	case MutexOpLock:
		// Both rlocked and locked never be true at the same time
		if prev.Locked() {
			e = ErrLockLocked
			break
		}
		if prev.RLocked() {
			e = ErrLockRLocked
			break
		}
		locked = true
	case MutexOpRLock:
		// Both rlocked and locked never be true at the same time
		if prev.Locked() {
			e = ErrRLockLocked
			break
		}
		if prev.RLocked() {
			e = ErrRLockRLocked
			break
		}
		rlocked = true
	case MutexOpUnlock:
		if !prev.Locked() {
			e = ErrUnlockUnlocked
			break
		}
		if prev.RLocked() {
			e = ErrUnlockRLocked
			break
		}
		locked = false
	case MutexOpRUnlock:
		if !prev.RLocked() {
			e = ErrRUnlockRUnlocked
			break
		}
		if prev.Locked() {
			e = ErrRUnlockLocked
			break
		}
		rlocked = false
	}

	state.locked = locked
	state.rlocked = rlocked
	state.err = e

	*ms = append(*ms, state)
}

type LockState struct {
	ms map[ast.Expr]*MuStates
}

func NewLockState() *LockState {
	return &LockState{
		ms: make(map[ast.Expr]*MuStates),
	}
}

func (ls *LockState) Init(node ast.Expr) {
	ls.ms[node] = &MuStates{}
}

func (ls *LockState) Push(mu *MuState) {
	if mu == nil {
		return
	}
	key := mu.node
	for k := range ls.ms {
		if cmp.Equal(k, mu.node, cmpopts.IgnoreTypes(token.Pos(0))) {
			key = k
			break
		}
	}

	if _, ok := ls.ms[key]; !ok {
		ls.Init(key)
	}
	ls.ms[key].Push(mu)
}

func (ls *LockState) Update(block *cfg.Block, node ast.Expr, op MutexOp) {
	ls.Push(&MuState{
		Op:    op,
		block: block,
		node:  node,
	})
}

func (ls *LockState) Map() map[ast.Expr]*MuStates {
	return ls.ms
}

func (ls *LockState) Get(key ast.Expr) (*MuStates, bool) {
	node := key
	for k := range ls.ms {
		if cmp.Equal(k, key, cmpopts.IgnoreTypes(token.Pos(0))) {
			node = k
			break
		}
	}
	ret, ok := ls.ms[node]
	return ret, ok
}

func (ls *LockState) Copy() *LockState {
	if ls == nil {
		return nil
	}
	ret := NewLockState()
	for k, v := range ls.ms {
		ret.Init(k)
		*ret.ms[k] = append(MuStates{}, *v...)
	}
	return ret
}
