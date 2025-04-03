package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type FileSet map[fs.FS][]string

type GraphFixture struct {
	Nodes         []Node `json:"nodes"`
	Relationships []Edge `json:"relationships"`
}

type Node struct {
	ID         string         `json:"id"`
	Caption    string         `json:"caption"`
	Labels     []string       `json:"labels"`
	Properties map[string]any `json:"properties"`
}

type Edge struct {
	ID         string         `json:"id"`
	FromID     string         `json:"fromId"`
	ToID       string         `json:"toId"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
}

func ParseGraphFixtureJsonFile(fh fs.File) (GraphFixture, error) {
	var graphFixture GraphFixture
	if err := json.NewDecoder(fh).Decode(&graphFixture); err != nil {
		return graphFixture, fmt.Errorf("could not unmarshal graph data: %w", err)
	} else {
		return graphFixture, nil
	}
}

func LoadGraphFixture(db graph.Database, g GraphFixture) error {
	var nodeMap map[string]graph.ID
	if err := db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		for _, node := range g.Nodes {
			if dbNode, err := tx.CreateNode(graph.AsProperties(node.Properties), graph.StringsToKinds(node.Labels)...); err != nil {
				return fmt.Errorf("could not create node `%s`: %w", node.ID, err)
			} else {
				nodeMap[node.ID] = dbNode.ID
			}
		}
		for _, edge := range g.Relationships {
			if startId, ok := nodeMap[edge.FromID]; !ok {
				return fmt.Errorf("could not find start node %s", edge.FromID)
			} else if endId, ok := nodeMap[edge.ToID]; !ok {
				return fmt.Errorf("could not find end node %s", edge.ToID)
			} else if _, err := tx.CreateRelationshipByIDs(startId, endId, graph.StringKind(edge.Type), graph.AsProperties(edge.Properties)); err != nil {
				return fmt.Errorf("could not create relationship `%s` from `%s` to `%s`: %w", edge.Type, edge.FromID, edge.ToID, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error writing graph data: %w", err)
	}
	return nil
}

func LoadGraphFixtureFile(db graph.Database, fSys fs.FS, path string) error {
	if fh, err := fSys.Open(path); err != nil {
		return fmt.Errorf("could not open graph data file: %w", err)
	} else {
		defer fh.Close()
		if graphFixture, err := ParseGraphFixtureJsonFile(fh); err != nil {
			return fmt.Errorf("could not parse graph data file: %w", err)
		} else if err := LoadGraphFixture(db, graphFixture); err != nil {
			return fmt.Errorf("could not load graph data: %w", err)
		}
	}
	return nil
}

func LoadGraphFixtureFiles(db graph.Database, fileSet map[fs.FS][]string) error {
	for fSys, files := range fileSet {
		for _, file := range files {
			if err := LoadGraphFixtureFile(db, fSys, file); err != nil {
				return fmt.Errorf("could not load graph fixture file %s: %w", file, err)
			}
		}
	}
	return nil
}
