package main

import (
	"graph-positioner/pkg/format"
	"graph-positioner/pkg/graph"
	"graph-positioner/pkg/physics"
	"graph-positioner/pkg/vec"
	"math"
)

func main() {
	const (
		maxIterations  = int64(1_000_000)
		forceThreshold = 0.001
		initCooling    = 0.9
	)

	input, err := format.ReadJsonInput("input.json")
	if err != nil {
		panic(err)
	}

	boundVec := vec.Boundaries(len(input.Nodes), 10)
	nodes := format.ConvertInputNodes(input.Nodes, boundVec)
	classified := graph.GetClassifiedNodes(nodes)
	graph.SetupSpringLengths(nodes, classified, input.Font)

	forceMap := make(map[int64]vec.Vec, len(nodes))
	for id := range nodes {
		forceMap[id] = vec.Vec{}
	}

	var (
		iter        int64
		maxForceMag float64
	)

	for {
		iterNormalized := float64(iter) / float64(maxIterations)
		cooling := physics.CoolingFunc(initCooling, iterNormalized)
		maxForceMag = 0

		// Compute forces for root -> children and global repulsion
		for _, rootID := range classified.Root {
			parent := nodes[rootID]

			fSum := vec.Vec{}
			for toID, toNode := range nodes {
				if toID == rootID {
					continue
				}
				fSum = vec.Add(fSum, physics.RepulsiveForce(parent.Pos, toNode.Pos, parent.SpringLen))
			}
			forceMap[rootID] = vec.Add(forceMap[rootID], fSum)

			for _, childData := range parent.Children {
				child := nodes[childData.ID]

				fAttr := physics.AttractiveForce(child.Pos, parent.Pos, parent.SpringLen)
				fRep := physics.RepulsiveForce(child.Pos, parent.Pos, parent.SpringLen)

				// Net parent-child force
				fSum := vec.Add(fAttr, fRep)

				// Repulsion with all other nodes
				for toID, toNode := range nodes {
					if toID == rootID || toID == childData.ID {
						continue
					}
					fSum = vec.Add(fSum, physics.RepulsiveForce(child.Pos, toNode.Pos, child.SpringLen))
				}

				forceMap[childData.ID] = vec.Add(forceMap[childData.ID], fSum)
			}
		}

		// Handle isolated (free) nodes
		for _, id := range classified.Free {
			fromNode := nodes[id]
			fSum := vec.Vec{}
			for toID, toNode := range nodes {
				if toID == id {
					continue
				}
				fSum = vec.Add(fSum, physics.RepulsiveForce(fromNode.Pos, toNode.Pos, toNode.SpringLen))
			}
			forceMap[id] = vec.Add(forceMap[id], fSum)
		}

		// Apply forces and update positions
		for id, force := range forceMap {
			force = vec.Mul(force, cooling)
			nodes[id].Pos = vec.Add(nodes[id].Pos, force)
			maxForceMag = math.Max(maxForceMag, vec.Mag(force))
			forceMap[id] = vec.Vec{} // reset
		}

		iter++
		if maxForceMag < forceThreshold || iter >= maxIterations {
			break
		}
	}

	if err := format.WriteJsonOut(format.OutputFromNodes(nodes), "out.json"); err != nil {
		panic(err)
	}
}
