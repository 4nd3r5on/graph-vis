package main

import (
	"fmt"
	"graph-positioner/pkg/graph"
	"sort"
)

// TopoSort returns topological ordering (parents before children). Assumes DAG; detects cycles.
func TopoSort(nodes map[int64]*graph.Node) ([]int64, error) {
	inDeg := map[int64]int{}
	adj := map[int64][]int64{}
	for id := range nodes {
		inDeg[id] = 0
	}
	for id, node := range nodes {
		for _, c := range node.Children {
			adj[id] = append(adj[id], c.ID)
			inDeg[c.ID]++
		}
	}
	// Kahn
	var q []int64
	for id, d := range inDeg {
		if d == 0 {
			q = append(q, id)
		}
	}
	var out []int64
	for len(q) > 0 {
		// deterministic: pop smallest id
		sort.Slice(q, func(i, j int) bool { return q[i] < q[j] })
		n := q[0]
		q = q[1:]
		out = append(out, n)
		for _, to := range adj[n] {
			inDeg[to]--
			if inDeg[to] == 0 {
				q = append(q, to)
			}
		}
	}
	if len(out) != len(nodes) {
		return nil, fmt.Errorf("graph has cycle or disconnected nodes; topo sorted %d of %d", len(out), len(nodes))
	}
	return out, nil
}
