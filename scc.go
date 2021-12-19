package unlockcheck

import (
	"github.com/Qs-F/unlockcheck/internal/cfg"
)

type Bridges []Edge

func (bridges Bridges) IsFrom(node Node) bool {
	for _, bridge := range bridges {
		if bridge.From == node {
			return true
		}
	}
	return false
}

func (bridges Bridges) IsTo(node Node) bool {
	for _, bridge := range bridges {
		if bridge.To == node {
			return true
		}
	}
	return false
}

func NewSCC(root *cfg.Block) (bridges Bridges, attributes map[Node]int, lowlinks map[int][]Node) {
	visited := make(map[Node]struct{})
	stack := []Node{}
	id := 0
	attributes = make(map[Node]int)
	lowlinks = make(map[int][]Node)
	labels := make(map[Node]int)
	bridges = Bridges{}

	var f func(node *cfg.Block, pre *cfg.Block)
	f = func(node *cfg.Block, pre *cfg.Block) {
		index := Node(node.Index)
		visited[index] = struct{}{}
		stack = append(stack, index)
		attributes[index] = id
		labels[index] = id
		id++

		for _, succNode := range node.Succs {
			if node == pre {
				continue
			}
			succ := Node(succNode.Index)
			if _, ok := visited[succ]; !ok {
				f(succNode, node)
				attributes[index] = min(attributes[index], attributes[succ])
				if labels[index] < attributes[succ] {
					bridges = append(bridges, Edge{
						From: index,
						To:   succ,
					})
				}
			} else {
				attributes[index] = min(attributes[index], labels[succ])
			}
		}
		if attributes[index] == labels[index] {
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				delete(visited, w)
				if w == index {
					break
				}
			}
		}
	}

	f(root, nil)

	for node, attr := range attributes {
		lowlinks[attr] = append(lowlinks[attr], node)
	}

	return bridges, attributes, lowlinks
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
