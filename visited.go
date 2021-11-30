package unlockcheck

import "errors"

type VisitedItem interface {
	Is(v VisitedItem) (bool, error)
}

type VisitedMap interface {
	Done(v VisitedItem) error
	Visited(v VisitedItem) (bool, error)
}

// Edge

type Edge struct {
	To   int32 // *cfg.Block.Index
	From int32 // *cfg.Block.Index
}

func (edge Edge) Is(v VisitedItem) (bool, error) {
	// if _, ok := e.(Edge); !ok {
	// 	return false, errors.New("VisitedItem is not a valid type of VisitedMap")
	// }
	return edge == v, nil
}

type VisitedEdges map[Edge]struct{}

func NewVisitedEdges() VisitedEdges {
	return make(VisitedEdges)
}

func (vs VisitedEdges) Done(v VisitedItem) error {
	e, ok := v.(Edge)
	if !ok {
		return errors.New("VisitedItem is not a valid type of VisitedMap")
	}
	vs[e] = struct{}{}
	return nil
}

func (vs VisitedEdges) Visited(v VisitedItem) (bool, error) {
	e, ok := v.(Edge)
	if !ok {
		return false, errors.New("VisitedItem is not a valid type of VisitedMap")
	}
	_, ok = vs[e]
	return ok, nil
}

var _ VisitedMap = NewVisitedEdges()

// Node

type Node int32

func (node Node) Is(v VisitedItem) (bool, error) {
	return node == v, nil
}

type VisitedNodes map[Node]struct{}

func NewVisitedNodes() VisitedNodes {
	return make(VisitedNodes)
}

func (vs VisitedNodes) Done(v VisitedItem) error {
	e, ok := v.(Node)
	if !ok {
		return errors.New("VisitedItem is not a valid type of VisitedMap")
	}
	vs[e] = struct{}{}
	return nil
}

func (vs VisitedNodes) Visited(v VisitedItem) (bool, error) {
	e, ok := v.(Node)
	if !ok {
		return false, errors.New("VisitedItem is not a valid type of VisitedMap")
	}
	_, ok = vs[e]
	return ok, nil
}

var _ VisitedMap = NewVisitedNodes()
