package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/specterops/bloodhound/dawgs/graph"
)

// GraphFixture is the JSON representation of the graph we are importing.
type GraphFixture struct {
	Nodes         []Node `json:"nodes"`
	Relationships []Edge `json:"relationships"`
}

// Node is the JSON representation of a graph node.
type Node struct {
	// ID is the local identifier for the Node within the file.
	// This ID is not preserved when imported into the database.
	ID string `json:"id"`

	// Labels are the node types. This is equivalent to what we call
	// Kinds in BloodHound.
	Labels []string `json:"labels"`

	// Properties is the key:value map used for storing extra information
	// on the Node. We currently do not validate that Nodes have an
	// `object_id` property, but it is best practice to include one as
	// `object_id` is the main identifier we use in BloodHound.
	Properties map[string]any `json:"properties"`
}

// Edge is the JSON representation of a graph edge.
type Edge struct {
	// FromID is the local Node identifier for the start of an Edge.
	FromID string `json:"fromId"`

	// ToID is the local Node identifier for the end of an Edge.
	ToID string `json:"toId"`

	// Type is the 'label' we apply to the Edge. This is synonymous to
	// the edge Kind in BloodHound.
	Type string `json:"type"`

	// Properties is the key:value map used for storing extra information
	// on the Edge.
	Properties map[string]any `json:"properties"`
}

// ParseGraphFixtureJsonFile takes in a fs.File interface and parses its contents into a
// GraphFixture struct.
func ParseGraphFixtureJsonFile(fh fs.File) (GraphFixture, error) {
	var graphFixture GraphFixture
	if err := json.NewDecoder(fh).Decode(&graphFixture); err != nil {
		return graphFixture, fmt.Errorf("could not unmarshal graph data: %w", err)
	} else {
		return graphFixture, nil
	}
}

// LoadGraphFixture requires a graph.Database interface and a GraphFixture struct.
// It will import the nodes and edges from the GraphFixture and inserts them into
// the graph database. It uses the `id` property of the nodes as the local
// identifier and then maps them to database IDs, meaning that the `id` given in
// the file will not be preserved.
func LoadGraphFixture(db graph.Database, g *GraphFixture) error {
	var nodeMap = make(map[string]graph.ID)
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

// LoadGraphFixtureFile takes a graph.Database interface, a fs.FS interface,
// and a path string. It will attempt to read the given path from the FS and
// parse then file into a GraphFixture, and finally import the GraphFixture
// Nodes and Edges into the graph.Database.
func LoadGraphFixtureFile(db graph.Database, fSys fs.FS, path string) error {
	if fh, err := fSys.Open(path); err != nil {
		return fmt.Errorf("could not open graph data file: %w", err)
	} else {
		defer fh.Close()
		if graphFixture, err := ParseGraphFixtureJsonFile(fh); err != nil {
			return fmt.Errorf("could not parse graph data file: %w", err)
		} else if err := LoadGraphFixture(db, &graphFixture); err != nil {
			return fmt.Errorf("could not load graph data: %w", err)
		}
	}
	return nil
}
