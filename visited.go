package unlockcheck

import "errors"

type VisitedItem interface {
	Is(v VisitedItem) bool
}

type VisitedMap interface {
	New(from int32, to int32) VisitedItem
	Visit(v VisitedItem) error
	Visited(v VisitedItem) (bool, error)
}

// Edge

type Edge struct {
	To   int32 // *cfg.Block.Index
	From int32 // *cfg.Block.Index
}

func (edge Edge) Is(v VisitedItem) bool {
	return edge == v
}

type VisitedEdges map[Edge]struct{}

func NewVisitedEdges() VisitedEdges {
	return make(VisitedEdges)
}

func (vs VisitedEdges) Visit(v VisitedItem) error {
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

func (_ VisitedEdges) New(from int32, to int32) VisitedItem {
	return VisitedItem(Edge{
		From: from,
		To:   to,
	})
}

var _ VisitedMap = NewVisitedEdges()

// Node

type Node int32 // *cfg.Block.Index

func (node Node) Is(v VisitedItem) bool {
	return node == v
}

type VisitedNodes map[Node]struct{}

func NewVisitedNodes() VisitedNodes {
	return make(VisitedNodes)
}

func (vs VisitedNodes) Visit(v VisitedItem) error {
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

func (_ VisitedNodes) New(node int32, _ int32) VisitedItem {
	return VisitedItem(Node(node))
}

var _ VisitedMap = NewVisitedNodes()
