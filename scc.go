package unlockcheck

import (
	"golang.org/x/tools/go/ssa"
)

var visited = map[Node]struct{}{}

type Block = *ssa.BasicBlock

func SCC(root Block) (bridges []Edge, attributes map[Node]int, lowlinks map[int][]Node) {
	visited := make(map[Node]struct{})
	id := 0
	attributes = make(map[Node]int)
	lowlinks = make(map[int][]Node)
	labels := make(map[Node]int)
	bridges = []Edge{}

	var dfs func(node Block)
	dfs = func(node Block) {
		index := Node(node.Index)
		visited[index] = struct{}{}
		attributes[index] = id
		labels[index] = id
		id++

		for _, succNode := range node.Succs {
			succ := Node(succNode.Index)
			if _, ok := visited[succ]; !ok {
				dfs(succNode)
				attributes[index] = min(attributes[index], attributes[succ])
				if labels[index] < attributes[succ] {
					bridges = append(bridges, Edge{
						From: int(index),
						To:   int(succ),
					})
				}
				continue
			}
			attributes[index] = min(attributes[index], attributes[succ])
		}
	}
	dfs(root)

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
