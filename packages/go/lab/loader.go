// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

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

	// Caption will be used as the `name` property of a Node.
	Caption string `json:"caption"`

	// Labels are the node types. This is equivalent to what we call
	// Kinds in BloodHound.
	Labels []string `json:"labels"`

	// Properties is the key:value map used for storing extra information
	// on the Node. For timestamp properties, you can use a timestamp function
	// like `NOW()` or `NOW()-3600` where `-3600` modifies the timestamp to
	// be 3600 seconds in the past.
	Properties map[string]string `json:"properties"`
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
	Properties map[string]string `json:"properties"`
}

// WriteGraphFixture requires a graph.Database interface and a GraphFixture struct.
// It will import the nodes and edges from the GraphFixture and inserts them into
// the graph database. It uses the `ID` property of the nodes as the local
// identifier and then maps them to database IDs, meaning that the `ID` given in
// the GraphFixture will not be preserved.
func WriteGraphFixture(db graph.Database, g *GraphFixture) error {
	var nodeMap = make(map[string]graph.ID)
	if err := db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		for _, node := range g.Nodes {

			props, err := processProperties(node.Properties)
			if err != nil {
				return fmt.Errorf("failed to process node properties: %w", err)
			}
			props.Set("name", node.Caption)

			if dbNode, err := tx.CreateNode(props, graph.StringsToKinds(node.Labels)...); err != nil {
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
			} else if props, err := processProperties(edge.Properties); err != nil {
				return fmt.Errorf("failed to process edge properties: %w", err)
			} else if testEdge, err := props.Get("testedge").Bool(); err == nil && testEdge {
				// It's a test edge for a harness test - skip creating it
				continue
			} else if _, err := tx.CreateRelationshipByIDs(startId, endId, graph.StringKind(edge.Type), props); err != nil {
				return fmt.Errorf("could not create relationship `%s` from `%s` to `%s`: %w", edge.Type, edge.FromID, edge.ToID, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error writing graph data: %w", err)
	}
	return nil
}

// LoadGraphFixtureFromFile a fs.FS interface, and a path string. It will
// attempt to read the given path from the FS and parse the file into a
// GraphFixture and return it.
func LoadGraphFixtureFromFile(fSys fs.FS, path string) (GraphFixture, error) {
	var graphFixture GraphFixture
	fh, err := fSys.Open(path)
	if err != nil {
		return graphFixture, fmt.Errorf("could not open graph data file: %w", err)
	}
	defer fh.Close()
	if err := json.NewDecoder(fh).Decode(&graphFixture); err != nil {
		return graphFixture, fmt.Errorf("could not parse graph data file: %w", err)
	} else {
		return graphFixture, nil
	}
}

func processProperties(props map[string]string) (*graph.Properties, error) {
	var out = graph.NewProperties()
	for k, v := range props {
		switch {
		case strings.HasPrefix(v, "NOW()"):
			if ts, err := processTimeFunctionProperty(v); err != nil {
				return nil, fmt.Errorf("could not process time function `%s`: %w", v, err)
			} else {
				out.Set(k, ts)
			}
		case strings.HasPrefix(v, "BOOL:"):
			_, val, found := strings.Cut(v, "BOOL:")
			if !found {
				return nil, fmt.Errorf("could not process bool value `%s`", v)
			} else if boolVal, err := strconv.ParseBool(val); err != nil {
				return nil, fmt.Errorf("could not process bool value `%s`: %w", v, err)
			} else {
				out.Set(k, boolVal)
			}
		default:
			out.Set(k, v)
		}
	}
	return out, nil
}

func processTimeFunctionProperty(prop string) (time.Time, error) {
	ts := time.Now().UTC()
	mod := strings.TrimPrefix(prop, "NOW()")
	if mod != "" {
		if modi, err := strconv.Atoi(mod); err != nil {
			return ts, fmt.Errorf("could not parse %d to int", modi)
		} else {
			ts = ts.Add(time.Duration(modi) * time.Second)
		}
	}
	return ts, nil
}
