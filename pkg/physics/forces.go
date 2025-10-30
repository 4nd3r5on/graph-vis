package physics

import (
	"graph-positioner/pkg/graph"
	"graph-positioner/pkg/vec"
)

const minDist = 1e-6 // prevent div-by-zero
const maxForce = 1e6

func CoolingFunc(initCooling float64, currentIter float64) (cooling float64) {
	return initCooling * (1 - currentIter*currentIter)
}

// RepulsiveForce calculates the repulsive force between two positions
// f_rep = ((springLen^2) / dist) * unitVec
func RepulsiveForce(from, to vec.Vec, springLen float64) vec.Vec {
	diff := vec.Sub(from, to)
	dist := vec.Mag(diff)

	if dist < minDist {
		dist = minDist
	}

	magnitude := (springLen * springLen) / dist
	if magnitude > maxForce {
		magnitude = maxForce
	}

	return vec.Mul(vec.Unit(diff), magnitude)
}

// AttractiveForce calculates the attractive force between two positions
// f_attr = (dist^2 / springLen) * unitVec
func AttractiveForce(from, to vec.Vec, springLen float64) vec.Vec {
	diff := vec.Sub(to, from)
	dist := vec.Mag(diff)

	if springLen <= 0 {
		springLen = minDist // avoid div-by-zero
	}
	if dist < minDist {
		dist = minDist
	}

	magnitude := (dist * dist) / springLen
	if magnitude > maxForce {
		magnitude = maxForce
	}

	return vec.Mul(vec.Unit(diff), magnitude)
}

// ComputeForces recursively calculates forces in the hierarchy
func RecursiveComputeForces(id int64, nodes map[int64]*graph.Node, forceMap map[int64]vec.Vec) {
	node := nodes[id]

	// compute force for root from every non-parent
	fSum := vec.Vec{}
	for otherID, other := range nodes {
		_, isParent := node.ParentIDsMap[otherID]
		_, isChild := node.ChildrenIDsMap[otherID]
		if otherID == id || isParent || isChild {
			continue
		}
		fSum = vec.Add(
			fSum,
			RepulsiveForce(node.Pos, other.Pos, node.SpringLen*2),
			AttractiveForce(node.Pos, other.Pos, node.SpringLen*2),
		)
	}
	forceMap[id] = vec.Add(forceMap[id], fSum)

	// compute forces for every children
	for _, childData := range node.Children {
		child := nodes[childData.ID]
		forceMap[childData.ID] = vec.Add(
			forceMap[childData.ID],
			AttractiveForce(child.Pos, node.Pos, node.SpringLen),
			vec.Div(RepulsiveForce(child.Pos, node.Pos, node.SpringLen), float64(len(child.ParentIDs))),
		)
		RecursiveComputeForces(childData.ID, nodes, forceMap)
	}
}

// HandleFreeNodes applies only repulsive forces between isolated nodes
func HandleFreeNodes(classified graph.ClassifiedNodes, nodes map[int64]*graph.Node, forceMap map[int64]vec.Vec) {
	free := classified.Free
	for i := range free {
		for j := i + 1; j < len(free); j++ {
			from := nodes[free[i]]
			to := nodes[free[j]]
			f := RepulsiveForce(from.Pos, to.Pos, from.SpringLen)
			forceMap[free[i]] = vec.Add(forceMap[free[i]], f)
			forceMap[free[j]] = vec.Add(forceMap[free[j]], vec.Mul(f, -1))
		}
	}
}
