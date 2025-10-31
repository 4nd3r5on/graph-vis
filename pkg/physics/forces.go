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

// c is constant (some small value, smaller than 1, bigger than 0)
func CentralForce(nodePos vec.Vec, c float64) vec.Vec {
	return vec.Mul(nodePos, c*-1.0)
}

// ComputeForces recursively calculates forces in the hierarchy
func RecursiveComputeForces(id int64, nodes map[int64]*graph.Node, seenMap map[int64]struct{}, forceMap map[int64]vec.Vec, centerForce float64) {
	if _, seen := seenMap[id]; seen {
		return
	}
	seenMap[id] = struct{}{}
	node := nodes[id]

	// compute repulsive
	for otherID, other := range nodes {
		if otherID == id {
			continue
		}
		forceMap[id] = vec.Add(forceMap[id], RepulsiveForce(node.Pos, other.Pos, node.SpringLen))
	}
	forceMap[id] = vec.Add(forceMap[id], CentralForce(node.Pos, centerForce))

	for _, parentID := range node.ParentIDs {
		parent := nodes[parentID]
		forceMap[id] = vec.Add(
			forceMap[id],
			AttractiveForce(node.Pos, parent.Pos, node.SpringLen),
		)
	}
	for _, childData := range node.Children {
		child := nodes[childData.ID]
		forceMap[id] = vec.Add(
			forceMap[id],
			AttractiveForce(node.Pos, child.Pos, node.SpringLen),
		)
		RecursiveComputeForces(childData.ID, nodes, seenMap, forceMap, centerForce)
	}
}
