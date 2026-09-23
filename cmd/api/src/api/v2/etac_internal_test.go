// Copyright 2026 Specter Ops, Inc.
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

package v2

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathSetToBloodHoundGraphETAC(t *testing.T) {
	var (
		hiddenNode = bloodhoundgraph.BloodHoundGraphNode{
			BloodHoundGraphItem: &bloodhoundgraph.BloodHoundGraphItem{
				Color: "#FFFFFF",
				Data: map[string]any{
					"nodetype": "HIDDEN",
					"objectid": "HIDDEN",
					"kinds":    []string{},
				},
			},
			FontIcon: &bloodhoundgraph.BloodHoundGraphFontIcon{Text: "fas fa-eye-slash"},
			Label:    &bloodhoundgraph.BloodHoundGraphNodeLabel{Text: "** Hidden Object **"},
			Size:     1,
		}
		cases = []struct {
			Name            string
			AllowList       []string
			EmptyPaths      bool
			MissingTenantID bool
			SharedPaths     bool
			HiddenNodeIDs   []graph.ID
			HiddenEdgeIDs   []graph.ID
		}{
			{
				Name:       "empty paths return an empty graph",
				AllowList:  []string{"tenant1"},
				EmptyPaths: true,
			},
			{
				Name:      "nil allowlist preserves all nodes and edges",
				AllowList: nil,
			},
			{
				Name:          "empty allowlist redacts all nodes and edges",
				AllowList:     []string{},
				HiddenNodeIDs: []graph.ID{1, 2, 3},
				HiddenEdgeIDs: []graph.ID{1, 2},
			},
			{
				Name:      "allowed tenants preserve all nodes and edges",
				AllowList: []string{"tenant1", "tenant2", "tenant3"},
			},
			{
				Name:          "hidden source node redacts its edge and preserves the accessible edge",
				AllowList:     []string{"tenant2", "tenant3"},
				HiddenNodeIDs: []graph.ID{1},
				HiddenEdgeIDs: []graph.ID{1},
			},
			{
				Name:          "hidden target node redacts its edge and preserves the accessible edge",
				AllowList:     []string{"tenant1", "tenant2"},
				HiddenNodeIDs: []graph.ID{3},
				HiddenEdgeIDs: []graph.ID{2},
			},
			{
				Name:          "two hidden endpoints redact their connecting edge",
				AllowList:     []string{"tenant3"},
				HiddenNodeIDs: []graph.ID{1, 2},
				HiddenEdgeIDs: []graph.ID{1, 2},
			},
			{
				Name:            "missing tenant ID redacts the node and both incident edges",
				AllowList:       []string{"tenant1", "tenant2", "tenant3"},
				MissingTenantID: true,
				HiddenNodeIDs:   []graph.ID{2},
				HiddenEdgeIDs:   []graph.ID{1, 2},
			},
			{
				Name:        "shared visible nodes and edges appear once",
				AllowList:   []string{"tenant1", "tenant2", "tenant3"},
				SharedPaths: true,
			},
			{
				Name:          "shared hidden nodes and edges appear once and remain redacted",
				AllowList:     []string{"tenant2", "tenant3"},
				SharedPaths:   true,
				HiddenNodeIDs: []graph.ID{1},
				HiddenEdgeIDs: []graph.ID{1},
			},
		}
	)

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var (
				nodes    = make([]*graph.Node, 0, 3)
				edges    = make([]*graph.Relationship, 0, 2)
				paths    = graph.NewPathSet()
				expected = make(map[string]any)
			)

			if !testCase.EmptyPaths {
				for index, tenantID := range []string{"tenant1", "tenant2", "tenant3"} {
					var node = &graph.Node{
						ID:    graph.ID(index + 1),
						Kinds: graph.Kinds{azure.Entity, azure.User},
						Properties: &graph.Properties{Map: map[string]any{
							common.Name.String():        fmt.Sprintf("user%d", index+1),
							common.ObjectID.String():    fmt.Sprintf("object%d", index+1),
							common.Description.String(): "private user details",
							azure.TenantID.String():     tenantID,
						}},
					}

					if testCase.MissingTenantID && node.ID == 2 {
						delete(node.Properties.Map, azure.TenantID.String())
					}
					nodes = append(nodes, node)

					if slices.Contains(testCase.HiddenNodeIDs, node.ID) {
						expected[node.ID.String()] = hiddenNode
					} else {
						var expectedNode = *node
						expectedNode.Properties = &graph.Properties{Map: maps.Clone(node.Properties.Map)}
						expected[node.ID.String()] = bloodhoundgraph.NodeToBloodHoundGraph(graphschema.ValidKinds, &expectedNode)
					}
				}

				for index := range 2 {
					var (
						edge = &graph.Relationship{
							ID:      graph.ID(index + 1),
							StartID: nodes[index].ID,
							EndID:   nodes[index+1].ID,
							Kind:    azure.Owns,
							Properties: &graph.Properties{Map: map[string]any{
								common.Description.String(): "private edge details",
							}},
						}
						edgeKey = "rel_" + edge.ID.String()
					)

					if slices.Contains(testCase.HiddenEdgeIDs, edge.ID) {
						expected[edgeKey] = bloodhoundgraph.BloodHoundGraphLink{
							ID:    edge.ID,
							ID1:   edge.StartID.String(),
							ID2:   edge.EndID.String(),
							End2:  &bloodhoundgraph.BloodHoundGraphLinkEnd{Arrow: true},
							Label: &bloodhoundgraph.BloodHoundGraphLinkLabel{Text: "** Hidden Edge **"},
						}
					} else {
						expected[edgeKey] = bloodhoundgraph.RelationshipToBloodHoundGraph(edge)
					}
					edges = append(edges, edge)
				}

				paths = graph.NewPathSet(graph.Path{Nodes: nodes, Edges: edges})
				if testCase.SharedPaths {
					paths.AddPath(graph.Path{Nodes: nodes[:2], Edges: edges[:1]})
				}
			}

			actual := pathSetToBloodHoundGraphETAC(
				graphschema.ValidKinds,
				paths,
				testCase.AllowList,
			)

			for _, node := range nodes {
				actualNode, isBloodHoundGraphNode := actual[node.ID.String()].(bloodhoundgraph.BloodHoundGraphNode)
				require.True(t, isBloodHoundGraphNode)
				require.NotNil(t, actualNode.BloodHoundGraphItem)
				require.NotNil(t, actualNode.Label)

				if slices.Contains(testCase.HiddenNodeIDs, node.ID) {
					assert.Equal(t, "HIDDEN", actualNode.Data["objectid"])
					assert.Equal(t, []string{}, actualNode.Data["kinds"])
					assert.Equal(t, "** Hidden Object **", actualNode.Label.Text)

					redactedNodeJSON, err := json.Marshal(actualNode)
					require.NoError(t, err)

					for propertyName, propertyValue := range node.Properties.Map {
						if propertyString, isString := propertyValue.(string); isString {
							assert.NotContains(t, string(redactedNodeJSON), propertyString, "redacted node must not contain original %s", propertyName)
						}
					}
				} else {
					assert.Equal(t, node.Properties.Map[common.Name.String()], actualNode.Data[common.Name.String()])
					assert.Equal(t, node.Properties.Map[common.ObjectID.String()], actualNode.Data[common.ObjectID.String()])
					assert.Equal(t, node.Properties.Map[common.Name.String()], actualNode.Label.Text)
				}
			}

			for _, edge := range edges {
				var edgeKey = "rel_" + edge.ID.String()

				actualEdge, isBloodHoundGraphLink := actual[edgeKey].(bloodhoundgraph.BloodHoundGraphLink)
				require.True(t, isBloodHoundGraphLink)

				if slices.Contains(testCase.HiddenEdgeIDs, edge.ID) {
					assert.Equal(t, edge.ID, actualEdge.ID)
					assert.Equal(t, edge.StartID.String(), actualEdge.ID1)
					assert.Equal(t, edge.EndID.String(), actualEdge.ID2)
					assert.Nil(t, actualEdge.BloodHoundGraphItem)
					require.NotNil(t, actualEdge.Label)
					assert.Equal(t, "** Hidden Edge **", actualEdge.Label.Text)

					redactedEdgeJSON, err := json.Marshal(actualEdge)
					require.NoError(t, err)
					assert.NotContains(t, string(redactedEdgeJSON), edge.Kind.String())
				} else {
					assert.Equal(t, bloodhoundgraph.RelationshipToBloodHoundGraph(edge), actualEdge)
				}
			}

			assert.Equal(t, expected, actual)
		})
	}
}

func TestFilterNodeSetByETAC(t *testing.T) {
	var cases = []struct {
		Name            string
		AllowList       []string
		EmptyNodeSet    bool
		MissingTenantID bool
		ExpectedNodeIDs []graph.ID
	}{
		{
			Name:         "empty node set returns an empty node set",
			AllowList:    []string{"tenant1"},
			EmptyNodeSet: true,
		},
		{
			Name:            "nil allowlist preserves all nodes",
			AllowList:       nil,
			ExpectedNodeIDs: []graph.ID{1, 2, 3},
		},
		{
			Name:      "empty allowlist filters all nodes",
			AllowList: []string{},
		},
		{
			Name:            "allowed tenants preserve all nodes",
			AllowList:       []string{"tenant1", "tenant2", "tenant3"},
			ExpectedNodeIDs: []graph.ID{1, 2, 3},
		},
		{
			Name:            "tenants not in the user's allowlist are filtered",
			AllowList:       []string{"tenant1", "tenant3"},
			ExpectedNodeIDs: []graph.ID{1, 3},
		},
		{
			Name:            "missing tenant ID filters the node",
			AllowList:       []string{"tenant1", "tenant2", "tenant3"},
			MissingTenantID: true,
			ExpectedNodeIDs: []graph.ID{1, 3},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var (
				nodeSet  = make(graph.NodeSet)
				expected = make(graph.NodeSet)
			)

			if !testCase.EmptyNodeSet {
				for index, tenantID := range []string{"tenant1", "tenant2", "tenant3"} {
					var node = &graph.Node{
						ID:    graph.ID(index + 1),
						Kinds: graph.Kinds{azure.Entity, azure.User},
						Properties: &graph.Properties{Map: map[string]any{
							common.Name.String():    fmt.Sprintf("user%d", index+1),
							azure.TenantID.String(): tenantID,
						}},
					}

					if testCase.MissingTenantID && node.ID == 2 {
						delete(node.Properties.Map, azure.TenantID.String())
					}
					nodeSet[node.ID] = node

					if slices.Contains(testCase.ExpectedNodeIDs, node.ID) {
						expected[node.ID] = node
					}
					//assert.Len(t, nodeSet, 3, "filterNodeSetByETAC must not modify the input node set")
				}
			}

			inputNodeCount := len(nodeSet)
			actual := filterNodeSetByETAC(nodeSet, testCase.AllowList)
			assert.Len(t, nodeSet, inputNodeCount, "filterNodeSetByETAC must not modify the input node set")

			assert.Equal(t, expected, actual)
		})
	}
}
