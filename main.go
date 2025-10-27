package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	charWidth       = 11 // pixels per character
	charHeight      = 18 // line height
	nodeHPadding    = 8  // left+right padding in px
	nodeVPadding    = 6  // top+bottom padding in px
	connLabelPadBot = 10 // extra bottom padding to fit connection labels
)

type ChildNodeData struct {
	ID        int64  `json:"id"`
	ConnLabel string `json:"connLabel"`
}
type InputNode struct {
	ParentIDs []int64          `json:"parentIDs"`
	Label     string           `json:"label"`
	Children  []*ChildNodeData `json:"children"`
}
type InputFile struct {
	CharWidthPx     int                 `json:"charWidthPx"`
	CharHeightPx    int                 `json:"charHeightPx"`
	NodeHPadding    int                 `json:"nodeHPadding"`
	NodeVPadding    int                 `json:"nodeVPadding"`
	ConnLabelPadBot int                 `json:"connLabelPadBot"`
	Nodes           map[int64]InputNode `json:"nodes"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type OutputNode struct {
	ID  int64    `json:"id"`
	Pos Position `json:"pos"`
}

func useNotZero[T comparable](a, b T) T {
	var zero T
	if a == zero {
		return b
	}
	return a
}

func SetGlobalVarsFromFile(data InputFile) {
	charWidth = useNotZero(data.CharWidthPx, charWidth)
	charHeight = useNotZero(data.CharHeightPx, charHeight)
	nodeHPadding = useNotZero(data.NodeHPadding, nodeHPadding)
	nodeVPadding = useNotZero(data.NodeVPadding, nodeVPadding)
	connLabelPadBot = useNotZero(data.ConnLabelPadBot, connLabelPadBot)
}

func ReadJsonInput(filePath string) (map[int64]InputNode, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data InputFile
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	SetGlobalVarsFromFile(data)
	return data.Nodes, nil
}

func WriteJsonOut(positions []OutputNode, fileName string) error {
	data, err := json.Marshal(positions)
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, data, 0o644)
}

func InputToNodeInfo(inNodes map[int64]InputNode) map[int64]*NodeInfo {
	nodes := make(map[int64]*NodeInfo, len(inNodes))

	// Initialize NodeInfo
	for id, inNode := range inNodes {
		nodes[id] = &NodeInfo{InputNode: inNode}
	}

	// Remove children with non-existing IDs
	for id, ni := range nodes {
		validChildren := ni.InputNode.Children[:0]
		for _, c := range ni.InputNode.Children {
			if _, ok := nodes[c.ID]; ok {
				validChildren = append(validChildren, c)
			} else {
				fmt.Printf("Warning: Node %d has child with non-existing ID %d, skipping\n", id, c.ID)
			}
		}
		ni.InputNode.Children = validChildren
	}

	// Remove parent IDs that do not exist
	for id, ni := range nodes {
		validParents := ni.InputNode.ParentIDs[:0]
		for _, pid := range ni.InputNode.ParentIDs {
			if _, ok := nodes[pid]; ok {
				validParents = append(validParents, pid)
			} else {
				fmt.Printf("Warning: Node %d has non-existing parent ID %d, skipping\n", id, pid)
			}
		}
		ni.InputNode.ParentIDs = validParents
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
		for _, c := range nodes[id].InputNode.Children {
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

func main() {
	inNodes, err := ReadJsonInput("input.json")
	if err != nil {
		panic(err)
	}
	nodes := InputToNodeInfo(inNodes)
	ComputeNodesGeometry(nodes)
	layersData, err := AssignNodesLayer(nodes)
	if err != nil {
		panic(err)
	}
	ComputeNodeX(nodes, layersData)
	ComputeNodeY(nodes, layersData)

	out := make([]OutputNode, 0, len(nodes))
	for id, node := range nodes {
		out = append(out, OutputNode{
			ID:  id,
			Pos: node.Pos,
		})
	}
	err = WriteJsonOut(out, "out.json")
	if err != nil {
		panic(err)
	}
}
