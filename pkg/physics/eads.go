package physics

import (
	"graph-positioner/pkg/graph"
	"graph-positioner/pkg/vec"
	"math"
)

const minDist = 1e-6
const maxForce = 1e6

func CoolingFunc(initCooling float64, currentIter float64) (cooling float64) {
	return initCooling * (1 - currentIter*currentIter)
}

// RepulsiveForce calculates the repulsive force between two positions
func RepulsiveForceEads(from, to vec.Vec, force float64) vec.Vec {
	diff := vec.Sub(from, to)
	dist := vec.Mag(diff)

	if dist < minDist {
		dist = minDist
	}

	magnitude := force / (dist * dist)
	if magnitude > maxForce {
		magnitude = maxForce
	}

	return vec.Mul(vec.Unit(diff), magnitude)
}

// AttractiveForce calculates the attractive force between two positions
func AttractiveForceEads(from, to vec.Vec, fRep vec.Vec, force, springLen float64) vec.Vec {
	diff := vec.Sub(to, from)
	dist := vec.Mag(diff)

	if springLen <= 0 {
		springLen = minDist // avoid div-by-zero
	}
	if dist < minDist {
		dist = minDist
	}

	return vec.Sub(
		vec.Mul(vec.Unit(diff), force*math.Log(dist/springLen)), fRep)
}

// c is constant (some small value, smaller than 1, bigger than 0)
func CentralForce(nodePos vec.Vec, c float64) vec.Vec {
	return vec.Mul(nodePos, c*-1.0)
}

func AntiCollisionForce(from, to vec.Vec, collisionRad float64) vec.Vec {
	diff := vec.Sub(from, to)
	dist := vec.Mag(diff)

	if dist >= collisionRad || dist == 0 {
		return vec.Vec{}
	}
	magnitude := collisionRad - dist
	return vec.Mul(vec.Unit(diff), magnitude)
}

// ComputeForces recursively calculates forces in the hierarchy
func RecursiveComputeForcesEads(
	id int64,
	nodes map[int64]*graph.Node,
	seenMap map[int64]struct{},
	forceMap map[int64]vec.Vec,
	centerForce,
	attractiveForce,
	repulsiveForce,
	springLen,
	collisionRad float64,
) {
	if _, seen := seenMap[id]; seen {
		return
	}
	seenMap[id] = struct{}{}
	node := nodes[id]

	// compute repulsive
	for otherID, other := range nodes {
		forceMap[id] = vec.Add(forceMap[id], AntiCollisionForce(node.Pos, other.Pos, collisionRad))
		_, isParent := node.ParentIDsMap[otherID]
		_, isChild := node.ChildrenIDsMap[otherID]
		if otherID == id || isParent || isChild {
			continue
		}
		forceMap[id] = vec.Add(forceMap[id], RepulsiveForceEads(node.Pos, other.Pos, repulsiveForce))
	}
	forceMap[id] = vec.Add(forceMap[id], CentralForce(node.Pos, centerForce))

	for _, parentID := range node.ParentIDs {
		parent := nodes[parentID]
		fRep := RepulsiveForceEads(node.Pos, parent.Pos, repulsiveForce)
		forceMap[id] = vec.Add(
			forceMap[id],
			fRep,
			AttractiveForceEads(node.Pos, parent.Pos, fRep, attractiveForce, springLen),
		)
	}
	for _, childData := range node.Children {
		child := nodes[childData.ID]
		fRep := RepulsiveForceEads(node.Pos, child.Pos, repulsiveForce)
		forceMap[id] = vec.Add(
			forceMap[id],
			fRep,
			AttractiveForceEads(node.Pos, child.Pos, fRep, attractiveForce, springLen),
		)
		RecursiveComputeForcesEads(
			childData.ID,
			nodes,
			seenMap,
			forceMap,
			centerForce,
			attractiveForce,
			repulsiveForce,
			springLen,
			collisionRad,
		)
	}
}
