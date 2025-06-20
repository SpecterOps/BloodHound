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

package arrows

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

// Graph is the JSON representation of the graph we are importing.
type Graph struct {
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

// WriteGraphToDatabase will import the nodes and edges from the arrows.app
// Graph and insert them into the graph database. It uses the `ID` property
// of the nodes as the local identifier and then maps them to database IDs,
// meaning that the `ID` given in the Graph will not be preserved.
func WriteGraphToDatabase(db graph.Database, g *Graph) error {
	var nodeMap = make(map[string]graph.ID)
	if err := db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {

		//#region Write nodes
		for _, node := range g.Nodes {

			props, err := processProperties(node.Properties)
			if err != nil {
				return fmt.Errorf("failed to process node properties: %w", err)
			}
			props.Set("name", node.Caption)

			// Determine node kinds and add Base and AZBase
			nodeKinds := graph.StringsToKinds(node.Labels)
			isAD := slices.ContainsFunc(nodeKinds, func(k graph.Kind) bool { return slices.Contains(ad.NodeKinds(), k) })
			isAzure := slices.ContainsFunc(nodeKinds, func(k graph.Kind) bool { return slices.Contains(azure.NodeKinds(), k) })
			if isAD {
				nodeKinds = append(nodeKinds, ad.Entity)
			}
			if isAzure {
				nodeKinds = append(nodeKinds, azure.Entity)
			}

			if dbNode, err := tx.CreateNode(props, nodeKinds...); err != nil {
				return fmt.Errorf("could not create node `%s`: %w", node.ID, err)
			} else {
				nodeMap[node.ID] = dbNode.ID
			}
		}
		//#endregion

		//#region Write edges
		for _, edge := range g.Relationships {
			if startId, ok := nodeMap[edge.FromID]; !ok {
				return fmt.Errorf("could not find start node %s", edge.FromID)
			} else if endId, ok := nodeMap[edge.ToID]; !ok {
				return fmt.Errorf("could not find end node %s", edge.ToID)
			} else if props, err := processProperties(edge.Properties); err != nil {
				return fmt.Errorf("failed to process edge properties: %w", err)
			} else if _, err := tx.CreateRelationshipByIDs(startId, endId, graph.StringKind(edge.Type), props); err != nil {
				return fmt.Errorf("could not create relationship `%s` from `%s` to `%s`: %w", edge.Type, edge.FromID, edge.ToID, err)
			}
		}
		//#endregion

		return nil
	}); err != nil {
		return fmt.Errorf("error writing graph data: %w", err)
	}
	return nil
}

// LoadGraphFromFile will attempt to read the given path from the
// FS and parse the file into an arrows.app Graph.
func LoadGraphFromFile(fSys fs.FS, path string) (Graph, error) {
	var graphFixture Graph
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
		kLowercase := strings.ToLower(k)

		switch {
		case strings.HasPrefix(v, "NOW()"):
			if ts, err := processTimeFunctionProperty(v); err != nil {
				return nil, fmt.Errorf("could not process time function `%s`: %w", v, err)
			} else {
				out.Set(kLowercase, ts)
			}
		case strings.HasPrefix(v, "BOOL:"):
			_, val, found := strings.Cut(v, "BOOL:")
			if !found {
				return nil, fmt.Errorf("could not process bool value `%s`", v)
			} else if boolVal, err := strconv.ParseBool(val); err != nil {
				return nil, fmt.Errorf("could not process bool value `%s`: %w", v, err)
			} else {
				out.Set(kLowercase, boolVal)
			}
		case strings.HasPrefix(v, "STRLIST:"):
			_, val, found := strings.Cut(v, "STRLIST:")
			if !found {
				return nil, fmt.Errorf("could not process list value `%s`", v)
			}
			listVal := strings.Split(val, ",")
			for i := range listVal {
				listVal[i] = strings.TrimSpace(listVal[i])
			}
			out.Set(k, listVal)
		case strings.HasPrefix(v, "DECIMAL:"):
			_, val, found := strings.Cut(v, "DECIMAL:")
			if !found {
				return nil, fmt.Errorf("could not process decimal value `%s`", v)
			}
			if floatVal, err := strconv.ParseFloat(val, 64); err != nil {
				return nil, fmt.Errorf("could not process decimal value `%s`: %w", v, err)
			} else {
				out.Set(k, floatVal)
			}
		default:
			out.Set(kLowercase, v)
		}
	}
	return out, nil
}

func processTimeFunctionProperty(prop string) (time.Time, error) {
	ts := time.Now().UTC()
	mod := strings.TrimPrefix(prop, "NOW()")
	if mod != "" {
		if modi, err := strconv.Atoi(mod); err != nil {
			return ts, fmt.Errorf("could not parse %s to integer value", mod)
		} else {
			ts = ts.Add(time.Duration(modi) * time.Second)
		}
	}
	return ts, nil
}
