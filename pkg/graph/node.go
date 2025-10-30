package graph

import "graph-positioner/pkg/vec"

type ChildNodeData struct {
	ID        int64  `json:"id"`
	ConnLabel string `json:"connLabel"`
}

type Node struct {
	ParentIDs []int64          `json:"parentIDs"`
	Label     string           `json:"label"`
	Children  []*ChildNodeData `json:"children"`
	Pos       vec.Vec

	SpringLen float64 `json:"springLen"` // for physics simulations to not recalculate it
}
