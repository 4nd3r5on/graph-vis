package main

import (
	"fmt"
	"sort"
)

// TopoSort returns topological ordering (parents before children). Assumes DAG; detects cycles.
func TopoSort(nodes map[int64]*NodeInfo) ([]int64, error) {
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

// OrderLayerByBarycenter reorders nodes in layer l based on barycenter of their adjacent nodes in the given direction.
// direction=true means use parents (upwards) to compute barycenter; false => use children.
func OrderLayerByBarycenter(layers map[int][]int64, nodes map[int64]*NodeInfo, l int, useParents bool) {
	list := layers[l]
	type pair struct {
		id   int64
		bary float64
	}
	var arr []pair
	for _, id := range list {
		n := nodes[id]
		var neigh []int64
		if useParents {
			neigh = n.ParentIDs
		} else {
			for _, c := range n.Children {
				neigh = append(neigh, c.ID)
			}
		}
		if len(neigh) == 0 {
			arr = append(arr, pair{id, float64(n.Order)})
			continue
		}
		// average order of neighbors
		sum := 0.0
		cnt := 0.0
		for _, nid := range neigh {
			if nb, ok := nodes[nid]; ok {
				sum += float64(nb.Order)
				cnt += 1.0
			}
		}
		if cnt == 0 {
			arr = append(arr, pair{id, float64(n.Order)})
		} else {
			arr = append(arr, pair{id, sum / cnt})
		}
	}
	// stable sort by barycenter then by id
	sort.SliceStable(arr, func(i, j int) bool {
		if arr[i].bary == arr[j].bary {
			return arr[i].id < arr[j].id
		}
		return arr[i].bary < arr[j].bary
	})
	newList := make([]int64, len(arr))
	for i, p := range arr {
		newList[i] = p.id
	}
	layers[l] = newList
}
