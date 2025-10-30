package graph

type ClassifiedNodes struct {
	Root []int64 // has no parents
	End  []int64 // has no children
	Free []int64 // has no parents, no children

	NonFree []int64 // has connections
}

func GetClassifiedNodes(nodes map[int64]*Node) ClassifiedNodes {
	out := ClassifiedNodes{
		Root: make([]int64, 0),
		End:  make([]int64, 0),
		Free: make([]int64, 0),
	}

	for id, n := range nodes {
		noParents := len(n.ParentIDs) == 0
		noChildren := len(n.Children) == 0

		switch {
		case noParents && noChildren:
			out.Free = append(out.Free, id)
			continue
		case noParents:
			out.Root = append(out.Root, id)
		case noChildren:
			out.End = append(out.End, id)
		}
		out.NonFree = append(out.NonFree, id)
	}

	return out
}

func SetupSpringLengths(nodes map[int64]*Node, classified ClassifiedNodes, font FontConfig) {
	const gapCoef = 1.3
	var avgSpringLen float64 = 0
	var maxSpringLen float64 = 0

	for _, nodeID := range classified.Root {
		var maxLenLocal float64 = 0
		for _, childNode := range nodes[nodeID].Children {
			maxLenLocal = max(float64(len(childNode.ConnLabel))*float64(font.CharWidthPx)*gapCoef, maxLenLocal)
		}
		nodes[nodeID].SpringLen = maxLenLocal
		avgSpringLen = (avgSpringLen + maxLenLocal) / 2
		maxSpringLen = max(maxLenLocal, maxSpringLen)
	}

	for _, nodeID := range classified.End {
		nodes[nodeID].SpringLen = avgSpringLen
	}

	// they have bigger spring length cuz we want to push them more on the outside from the system
	for _, nodeID := range classified.Free {
		nodes[nodeID].SpringLen = maxSpringLen
	}
}
