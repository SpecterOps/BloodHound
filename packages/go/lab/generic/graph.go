// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"testing"

	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var excludedProperties = []string{
	"lastseen",
	"lastcollected",
}

type Node struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

type Terminal struct {
	MatchBy string `json:"match_by"`
	Value   string `json:"value"`
}

type Edge struct {
	Start      Terminal       `json:"start"`
	End        Terminal       `json:"end"`
	Kind       string         `json:"kind"`
	Properties map[string]any `json:"properties"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type GenericObject struct {
	Graph Graph `json:"graph"`
}

func WriteGraphToDatabase(db graph.Database, g *Graph) error {
	var nodeMap = make(map[string]graph.ID)
	if err := db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {

		//#region Write nodes
		for _, node := range g.Nodes {

			props := graph.AsProperties(node.Properties)
			props.Set(common.ObjectID.String(), node.ID)

			if dbNode, err := tx.CreateNode(props, graph.StringsToKinds(node.Kinds)...); err != nil {
				return fmt.Errorf("could not create node `%s`: %w", node.ID, err)
			} else {
				nodeMap[node.ID] = dbNode.ID
			}
		}
		//#endregion

		//#region Write edges
		for _, edge := range g.Edges {
			if startId, ok := nodeMap[edge.Start.Value]; !ok {
				return fmt.Errorf("could not find start node %s", edge.Start.Value)
			} else if endId, ok := nodeMap[edge.End.Value]; !ok {
				return fmt.Errorf("could not find end node %s", edge.End.Value)
			} else if _, err := tx.CreateRelationshipByIDs(startId, endId, graph.StringKind(edge.Kind), graph.AsProperties(edge.Properties)); err != nil {
				return fmt.Errorf("could not create relationship `%s` from `%s` to `%s`: %w", edge.Kind, edge.Start.Value, edge.End.Value, err)
			}
		}
		//#endregion

		return nil
	}); err != nil {
		return fmt.Errorf("error writing graph data: %w", err)
	}
	return nil
}

func LoadGraphFromFile(fSys fs.FS, path string) (Graph, error) {
	var graphFixture GenericObject
	fh, err := fSys.Open(path)
	if err != nil {
		return graphFixture.Graph, fmt.Errorf("could not open graph data file: %w", err)
	}
	defer fh.Close()

	decoder := json.NewDecoder(fh)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&graphFixture); err != nil {
		return graphFixture.Graph, fmt.Errorf("could not parse graph data file: %w", err)
	} else {
		return graphFixture.Graph, nil
	}
}

func AssertDatabaseGraph(t *testing.T, ctx context.Context, db graph.Database, expected *Graph) {
	actualNodes := make(map[string]*graph.Node, 100)
	nodeIDToObjectID := make(map[graph.ID]string, 100)
	actualEdges := make(map[string]*graph.Relationship, 100)

	//#region Reading from DB
	_ = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		err := tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				objectId, err := node.Properties.Get(common.ObjectID.String()).String()

				if err == nil {
					actualNodes[objectId] = node
					nodeIDToObjectID[node.ID] = objectId
				}
			}

			return cursor.Error()
		})

		require.NoError(t, err)

		err = tx.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for edge := range cursor.Chan() {

				fingerprint := nodeIDToObjectID[edge.StartID] + nodeIDToObjectID[edge.EndID] + edge.Kind.String()
				actualEdges[fingerprint] = edge
			}

			return cursor.Error()
		})

		require.NoError(t, err)

		return nil
	})
	//#endregion

	//#region Node Assertions
	for _, expectedNode := range expected.Nodes {
		t.Run(fmt.Sprintf("AssertNode_%s", expectedNode.ID), func(t *testing.T) {
			t.Parallel()

			// assert existence
			node, ok := actualNodes[expectedNode.ID]
			require.True(t, ok)

			// assert kinds
			kindMap := make(map[string]struct{})
			for _, kind := range node.Kinds.Strings() {
				kindMap[kind] = struct{}{}
			}
			for _, kind := range expectedNode.Kinds {
				assert.Contains(t, kindMap, kind)
			}

			// assert properties
			for expectedProperty, expectedValue := range expectedNode.Properties {
				value, ok := node.Properties.Map[expectedProperty]
				assert.Truef(t, ok, "could not find expected property `%s` on node `%s`", expectedProperty, node.ID)

				if ok && !slices.Contains(excludedProperties, expectedProperty) {
					assert.Equalf(t, expectedValue, value, "mismatched property `%s` on node `%s`: have %v, want %v", expectedProperty, node.ID, value, expectedValue)
				}
			}
		})
	}
	//#endregion

	//#region Edge Assertions
	for _, expectedEdge := range expected.Edges {
		t.Run(fmt.Sprintf("AssertEdge_%s-%s-%s", expectedEdge.Start.Value, expectedEdge.Kind, expectedEdge.End.Value), func(t *testing.T) {
			t.Parallel()

			// assert existence
			fingerprint := expectedEdge.Start.Value + expectedEdge.End.Value + expectedEdge.Kind
			edge, ok := actualEdges[fingerprint]

			require.True(t, ok, "expected edge `(%s)-[%s]->(%s)` is missing", expectedEdge.Start.Value, expectedEdge.Kind, expectedEdge.End.Value)

			// assert properties
			for expectedProperty, expectedValue := range expectedEdge.Properties {
				value, ok := edge.Properties.Map[expectedProperty]
				assert.Truef(t, ok, "could not find expected property `%s` on edge (%s)->(%s)", expectedProperty, expectedEdge.Start.Value, expectedEdge.End.Value)

				if ok && !slices.Contains(excludedProperties, expectedProperty) {
					assert.Equalf(t, expectedValue, value, "mismatched property `%s` on edge (%s)->(%s): have %v, want %v", expectedProperty, expectedEdge.Start.Value, expectedEdge.End.Value, value, expectedValue)
				}
			}
		})
	}
	//#endregion
}
