package format

import (
	"encoding/json"
	"fmt"
	"graph-positioner/pkg/graph"
	"graph-positioner/pkg/vec"
	"os"
)

type InputNode struct {
	ParentIDs []int64                `json:"parentIDs"`
	Label     string                 `json:"label"`
	Children  []*graph.ChildNodeData `json:"children"`
}

type InputFile struct {
	Font  graph.FontConfig    `json:"font"`
	Nodes map[int64]InputNode `json:"nodes"`
}

func ReadJsonInput(filePath string) (*InputFile, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data InputFile
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	data.Font = graph.FontConfigWithDefault(data.Font)

	return &data, nil
}

func ConvertInputNodes(inNodes map[int64]InputNode, bounds vec.Vec) map[int64]*graph.Node {
	nodes := make(map[int64]*graph.Node, len(inNodes))

	// Initialize Node
	for id, inNode := range inNodes {
		nodes[id] = &graph.Node{
			ParentIDs: inNode.ParentIDs,
			Label:     inNode.Label,
			Children:  inNode.Children,
			Pos:       vec.Random([]byte(inNode.Label), vec.Vec{}, bounds),
		}
	}

	// Remove children with non-existing IDs
	for id, ni := range nodes {
		validChildren := ni.Children[:0]
		for _, c := range ni.Children {
			if _, ok := nodes[c.ID]; ok {
				validChildren = append(validChildren, c)
			} else {
				fmt.Printf("Warning: Node %d has child with non-existing ID %d, skipping\n", id, c.ID)
			}
		}
		ni.Children = validChildren
	}

	// Remove parent IDs that do not exist
	for id, ni := range nodes {
		validParents := ni.ParentIDs[:0]
		for _, pid := range ni.ParentIDs {
			if _, ok := nodes[pid]; ok {
				validParents = append(validParents, pid)
			} else {
				fmt.Printf("Warning: Node %d has non-existing parent ID %d, skipping\n", id, pid)
			}
		}
		ni.ParentIDs = validParents
	}

	// Detect cycles using DFS
	visited := make(map[int64]bool)
	recStack := make(map[int64]bool)
	var dfs func(id int64)
	dfs = func(id int64) {
		if recStack[id] {
			panic(fmt.Sprintf("Cycle detected at node %d", id))
		}
		if visited[id] {
			return
		}
		visited[id] = true
		recStack[id] = true
		for _, c := range nodes[id].Children {
			dfs(c.ID)
		}
		recStack[id] = false
	}

	for id := range nodes {
		if !visited[id] {
			dfs(id)
		}
	}

	return nodes
}
