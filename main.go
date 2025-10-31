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
		maxIterations   = int64(100_000)
		forceThreshold  = 0.01
		initCooling     = 0.5
		centerForce     = 0.05
		attractiveForce = 1
		repulsiveForce  = 50
		springLen       = 250
	)

	input, err := format.ReadJsonInput("input.json")
	if err != nil {
		panic(err)
	}

	boundVec := vec.Boundaries(len(input.Nodes), 10)
	nodes := format.ConvertInputNodes(input.Nodes, boundVec)
	classified := graph.GetClassifiedNodes(nodes)
	graph.SetupNodes(nodes, classified, input.Font)

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
		seenMap := make(map[int64]struct{}, 0)
		maxForceMag = 0

		// Compute forces for root -> children and global repulsion
		for nodeID := range nodes {
			if len(nodes[nodeID].ParentIDs) == 0 {
				physics.RecursiveComputeForcesEads(nodeID, nodes, seenMap, forceMap, centerForce, attractiveForce, repulsiveForce, springLen)
			}
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
