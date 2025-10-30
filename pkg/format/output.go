package format

import (
	"encoding/json"
	"graph-positioner/pkg/graph"
	"graph-positioner/pkg/vec"
	"os"
)

type OutputNode struct {
	ID  int64   `json:"id"`
	Pos vec.Vec `json:"pos"`
}

func WriteJsonOut(positions []OutputNode, fileName string) error {
	data, err := json.Marshal(positions)
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, data, 0o644)
}

func OutputFromNodes(inNodes map[int64]*graph.Node) (out []OutputNode) {
	out = make([]OutputNode, 0, len(inNodes))
	for id, node := range inNodes {
		out = append(out, OutputNode{
			ID:  id,
			Pos: node.Pos,
		})
	}
	return out
}
