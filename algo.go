package main

import (
	"math"
	"sort"
)

var (
	minEdgeGap    = 12 // minimal extra horizontal gap beyond node half-widths
	layerVSpacing = 32 // vertical spacing added between layers (plus node height)
	relaxIters    = 40 // iterations of compacting relaxation
)

type NodeInfo struct {
	InputNode
	Width  float64
	Height float64
	Layer  int
	Order  int // Position inder inside layer
	Pos    Position
}

func ComputeNodesGeometry(nodes map[int64]*NodeInfo) {
	for id, node := range nodes {
		nodes[id].Width = float64(len([]rune(node.Label))*charWidth + 2*nodeHPadding)
		nodes[id].Height = float64(charHeight + 2*nodeVPadding)
	}
}

type LayersData struct {
	Layers   map[int][]int64
	MaxLayer int
}

func AssignNodesLayer(nodes map[int64]*NodeInfo) (data *LayersData, err error) {
	TopoSortedIDs, err := TopoSort(nodes)
	if err != nil {
		return nil, err
	}
	// TODO: Check if needed ! AI WARNING
	for _, id := range TopoSortedIDs {
		node := nodes[id]
		best := 0
		for _, p := range node.ParentIDs {
			if pl, ok := nodes[p]; ok {
				if pl.Layer+1 > best {
					best = pl.Layer + 1
				}
			}
		}
		node.Layer = best
	}

	layers := map[int][]int64{}
	maxLayer := 0
	for id, node := range nodes {
		layers[node.Layer] = append(layers[node.Layer], id)
		if node.Layer > maxLayer {
			maxLayer = node.Layer
		}
	}

	for pass := 0; pass < 6; pass++ {
		// downward pass
		for l := 1; l <= maxLayer; l++ {
			OrderLayerByBarycenter(layers, nodes, l, true)
		}
		// upward pass
		for l := maxLayer - 1; l >= 0; l-- {
			OrderLayerByBarycenter(layers, nodes, l, false)
		}
		// update orders
		for l := 0; l <= maxLayer; l++ {
			for idx, id := range layers[l] {
				nodes[id].Order = idx
			}
		}
	}

	for l := 0; l <= maxLayer; l++ {
		sort.SliceStable(layers[l], func(i, j int) bool { return layers[l][i] < layers[l][j] })
		for idx, id := range layers[l] {
			nodes[id].Order = idx
		}
	}

	return &LayersData{
		Layers:   layers,
		MaxLayer: maxLayer,
	}, nil
}

func ComputeNodeX(nodes map[int64]*NodeInfo, ld *LayersData) {
	// TODO: Check if needed ! AI WARNING

	// Inodetial X positions by order and widths with minodemal pair spacing
	for l := 0; l <= ld.MaxLayer; l++ {
		list := ld.Layers[l]
		x := 0.0
		for i, id := range list {
			node := nodes[id]
			// center x for the node
			if i == 0 {
				x = node.Width / 2.0
			} else {
				prev := nodes[list[i-1]]
				// minodemal center separation so nodes don't overlap
				minSep := (prev.Width+node.Width)/2.0 + float64(minEdgeGap)
				x = prev.Pos.X + minSep
			}
			node.Pos.X = x
		}
	}

	// Enforce minodemal horizontal separation constraints derived from edges (conn labels)
	// For each edge parent->child, minodemal distance between centers should also allow edge label width
	edgeMinSep := map[[2]int64]float64{} // (p,c) -> min separation
	for pid, pnode := range nodes {
		for _, ch := range pnode.Children {
			cnode := nodes[ch.ID]
			edgeLabelWidth := float64(len([]rune(ch.ConnLabel))*charWidth + 2) // small pad
			minSep := math.Max((pnode.Width+cnode.Width)/2.0+float64(minEdgeGap), edgeLabelWidth)
			edgeMinSep[[2]int64{pid, ch.ID}] = minSep
		}
	}

	// Compact each layer using relaxation while respecting pairwise node separation AND edge min sep between arbitrary pairs in same layer or across layers (only influences x)
	// We'll run a simple iterative push to satisfy constraints between adjacent nodes in layer and also across layers projectively.
	for it := 0; it < relaxIters; it++ {
		// within-layer neighbors
		for l := 0; l <= ld.MaxLayer; l++ {
			list := ld.Layers[l]
			if len(list) <= 1 {
				continue
			}
			// left to right ensure minodemal distances
			for i := 1; i < len(list); i++ {
				left := nodes[list[i-1]]
				right := nodes[list[i]]
				minSep := (left.Width+right.Width)/2.0 + float64(minEdgeGap)
				if right.Pos.X-left.Pos.X < minSep {
					shift := (minSep - (right.Pos.X - left.Pos.X)) / 2.0
					left.Pos.X -= shift
					right.Pos.X += shift
				}
			}
		}
		// edge-based constraints (parent-child may be in different layers)
		for pair, minSep := range edgeMinSep {
			p := nodes[pair[0]]
			c := nodes[pair[1]]
			if c.Pos.X-p.Pos.X < minSep {
				// push them apart symmetrically (but prefer moving child right if parent is root-ish)
				needed := (minSep - (c.Pos.X - p.Pos.X)) / 2.0
				p.Pos.X -= needed
				c.Pos.X += needed
			}
		}
	}

	// 6) final horizontal compacting: normalize so min x >= 0
	minX := math.Inf(1)
	for _, node := range nodes {
		if node.Pos.X-node.Width/2.0 < minX {
			minX = node.Pos.X - node.Width/2.0
		}
	}
	if minX < 0 {
		shift := -minX + 4 // small margin
		for _, node := range nodes {
			node.Pos.X += shift
		}
	}
}

func ComputeNodeY(nodes map[int64]*NodeInfo, ld *LayersData) {
	// TODO: Check if needed ! AI WARNING

	// compute Y per layer (vertical spacing must allow edge labels underneath)
	layerHeights := map[int]float64{}
	for l := 0; l <= ld.MaxLayer; l++ {
		// layer height = max node height in layer
		h := 0.0
		for _, id := range ld.Layers[l] {
			if nodes[id].Height > h {
				h = nodes[id].Height
			}
		}
		layerHeights[l] = h
	}

	y := 0.0
	for l := 0; l <= ld.MaxLayer; l++ {
		h := layerHeights[l]
		// center y for this layer
		centerY := y + h/2.0
		for _, id := range ld.Layers[l] {
			node := nodes[id]
			node.Pos.Y = centerY
		}
		// step y down by layer height + vertical spacing + room for connection labels below edges (use connLabelPadBot)
		y += h + float64(layerVSpacing) + float64(connLabelPadBot)
	}
}
