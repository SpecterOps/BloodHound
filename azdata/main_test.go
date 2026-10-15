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

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/packages/go/chow/ingest"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func ignoredDefaults() map[string]bool {
	return map[string]bool{"firstseen": true, "lastseen": true, "lastcollected": true, "primarykind": true}
}

// This pair was verified by ingesting each format into the real database.
// Also compare with the literal database export, so a bug shared by both
// in-memory ingest reductions cannot make the benchmark pass unnoticed.
func TestKnownGoodAzureOpenGraph(t *testing.T) {
	var (
		azureContent, azureError         = os.ReadFile("azurehound_example.json")
		openGraphContent, openGraphError = os.ReadFile("baseline.json")
		snapshot                         struct {
			Metadata ein.GenericMetadata `json:"metadata"`
			Graph    struct {
				Nodes []ein.GenericNode `json:"nodes"`
				Edges []ein.GenericEdge `json:"edges"`
			} `json:"graph"`
		}
		oracle = newMemoryGraph()
	)
	require.NoError(t, azureError)
	require.NoError(t, openGraphError)
	var (
		azureState, azureStateError         = readState(azureContent, ingest.DataTypeAzure, false)
		openGraphState, openGraphStateError = readState(openGraphContent, ingest.DataTypeOpenGraph, false)
	)
	require.NoError(t, azureStateError)
	require.NoError(t, openGraphStateError)
	require.Len(t, azureState.NodesByID, 13526)
	require.Len(t, azureState.EdgesByID, 27388)
	var result, err = compareStates(azureState, openGraphState, ignoredDefaults())
	require.NoError(t, err)
	require.True(t, result.Equivalent, "%+v", result.Differences)

	require.NoError(t, json.Unmarshal(openGraphContent, &snapshot))
	for _, node := range snapshot.Graph.Nodes {
		node.Properties["objectid"] = node.ID
		oracle.NodesByID[nodeKey{Source: snapshot.Metadata.SourceKind, ID: node.ID}] = graph.PrepareNode(graph.AsProperties(node.Properties), graph.StringsToKinds(node.Kinds)...)
	}
	for _, edge := range snapshot.Graph.Edges {
		oracle.EdgesByID[edgeKey{Start: nodeKey{Source: snapshot.Metadata.SourceKind, ID: edge.Start.Value}, Kind: edge.Kind, End: nodeKey{Source: snapshot.Metadata.SourceKind, ID: edge.End.Value}}] = graph.AsProperties(edge.Properties)
	}
	result, err = compareStates(azureState, oracle, ignoredDefaults())
	require.NoError(t, err)
	require.True(t, result.Equivalent, "%+v", result.Differences)
}

func testGraph() *memoryGraph {
	var state = newMemoryGraph()
	state.NodesByID[nodeKey{Source: "AZBase", ID: "user"}] = graph.PrepareNode(graph.AsProperties(map[string]any{"objectid": "user", "enabled": false, "values": []any{"a", "b"}, "lastseen": "earlier"}), azure.Entity, azure.User)
	state.NodesByID[nodeKey{Source: "AZBase", ID: "tenant"}] = graph.PrepareNode(graph.AsProperties(map[string]any{"objectid": "tenant"}), azure.Entity, azure.Tenant)
	state.EdgesByID[edgeKey{Start: nodeKey{Source: "AZBase", ID: "tenant"}, Kind: "AZContains", End: nodeKey{Source: "AZBase", ID: "user"}}] = graph.AsProperties(map[string]any{"enabled": false})
	return state
}

func TestMeaningfulDifferences(t *testing.T) {
	var cases = []struct {
		name   string
		mutate func(*memoryGraph)
	}{
		{"missing node", func(state *memoryGraph) { delete(state.NodesByID, nodeKey{Source: "AZBase", ID: "user"}) }},
		{"extra node", func(state *memoryGraph) {
			state.NodesByID[nodeKey{Source: "AZBase", ID: "extra"}] = graph.PrepareNode(graph.NewProperties(), azure.User)
		}},
		{"kind", func(state *memoryGraph) {
			state.NodesByID[nodeKey{Source: "AZBase", ID: "user"}].Kinds = graph.Kinds{azure.Entity, azure.Group}
		}},
		{"false versus missing", func(state *memoryGraph) {
			delete(state.NodesByID[nodeKey{Source: "AZBase", ID: "user"}].Properties.Map, "enabled")
		}},
		{"false versus zero", func(state *memoryGraph) {
			state.NodesByID[nodeKey{Source: "AZBase", ID: "user"}].Properties.Map["enabled"] = 0
		}},
		{"array order", func(state *memoryGraph) {
			state.NodesByID[nodeKey{Source: "AZBase", ID: "user"}].Properties.Map["values"] = []any{"b", "a"}
		}},
		{"edge property", func(state *memoryGraph) {
			for _, properties := range state.EdgesByID {
				properties.Map["enabled"] = true
			}
		}},
		{"edge direction", func(state *memoryGraph) {
			for key, properties := range state.EdgesByID {
				delete(state.EdgesByID, key)
				state.EdgesByID[edgeKey{Start: key.End, Kind: key.Kind, End: key.Start}] = properties
				break
			}
		}},
		{"source scope", func(state *memoryGraph) {
			var key = nodeKey{Source: "AZBase", ID: "user"}
			state.NodesByID[nodeKey{Source: "OtherBase", ID: "user"}] = state.NodesByID[key]
			delete(state.NodesByID, key)
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var left, right = testGraph(), testGraph()
			testCase.mutate(right)
			var result, err = compareStates(left, right, ignoredDefaults())
			require.NoError(t, err)
			require.False(t, result.Equivalent)
			require.NotEmpty(t, result.Differences)
		})
	}
}

func TestExclusionsAreExplicit(t *testing.T) {
	var left, right = testGraph(), testGraph()
	right.NodesByID[nodeKey{Source: "AZBase", ID: "user"}].Properties.Map["lastseen"] = "later"
	right.NodesByID[nodeKey{Source: "AZBase", ID: "user"}].Properties.Map["primarykind"] = "AZUser"
	var result, err = compareStates(left, right, ignoredDefaults())
	require.NoError(t, err)
	require.True(t, result.Equivalent)
	result, err = compareStates(left, right, nil)
	require.NoError(t, err)
	require.False(t, result.Equivalent)
}

func TestProductionIngestMergesAndCreatesEndpoints(t *testing.T) {
	var (
		content    = []byte(`{"metadata":{"source_kind":"AZBase"},"graph":{"nodes":[{"id":"user","kinds":["AZUser"],"properties":{"name":"First","enabled":true}},{"id":"user","kinds":["AZUser"],"properties":{"name":"Second"}}],"edges":[{"start":{"value":"tenant"},"end":{"value":"user","kind":"AZUser"},"kind":"AZContains","properties":{"a":true}},{"start":{"value":"tenant"},"end":{"value":"user","kind":"AZUser"},"kind":"AZContains","properties":{"b":true}}]}}`)
		state, err = readState(content, ingest.DataTypeOpenGraph, false)
	)
	require.NoError(t, err)
	require.Len(t, state.NodesByID, 2)
	require.Len(t, state.EdgesByID, 1)
	require.Equal(t, "SECOND", state.NodesByID[nodeKey{Source: "AZBase", ID: "USER"}].Properties.Map["name"])
	require.Equal(t, true, state.NodesByID[nodeKey{Source: "AZBase", ID: "USER"}].Properties.Map["enabled"])
	for _, properties := range state.EdgesByID {
		require.Equal(t, true, properties.Map["a"])
		require.Equal(t, true, properties.Map["b"])
	}
}

func TestRawIDsAndUnsupportedLookup(t *testing.T) {
	var (
		content    = []byte(`{"metadata":{"source_kind":"AZBase"},"graph":{"nodes":[{"id":"mixedCase","kinds":["AZUser"]}]}}`)
		state, err = readState(content, ingest.DataTypeOpenGraph, true)
	)
	require.NoError(t, err)
	require.Contains(t, state.NodesByID, nodeKey{Source: "AZBase", ID: "mixedCase"})
	_, err = readState([]byte(`{"metadata":{"source_kind":"AZBase"},"graph":{"edges":[{"start":{"match_by":"name","value":"user"},"end":{"value":"tenant"},"kind":"AZContains"}]}}`), ingest.DataTypeOpenGraph, false)
	require.ErrorContains(t, err, "only ID-based endpoints")
}

func TestCommandVerdicts(t *testing.T) {
	var cases = []struct {
		name      string
		azure     string
		opengraph string
		verdict   string
	}{
		{"different", `{"meta":{"type":"azure","version":5},"data":[]}`, `{"metadata":{"source_kind":"AZBase"},"graph":{"nodes":[{"id":"extra","kinds":["AZUser"]}]}}`, "different"},
		{"logged conversion error", `{"meta":{"type":"azure","version":5},"data":[{"kind":"AZUser","data":"invalid user"}]}`, `{"metadata":{"source_kind":"AZBase"},"graph":{"nodes":[]}}`, "inconclusive"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				directory     = t.TempDir()
				azurePath     = filepath.Join(directory, "azure.json")
				openGraphPath = filepath.Join(directory, "opengraph.json")
				reportPath    = filepath.Join(directory, "report.json")
				result        report
			)
			require.NoError(t, os.WriteFile(azurePath, []byte(testCase.azure), 0o600))
			require.NoError(t, os.WriteFile(openGraphPath, []byte(testCase.opengraph), 0o600))
			require.Error(t, run(azurePath, openGraphPath, reportPath, "", false))
			var content, err = os.ReadFile(reportPath)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(content, &result))
			require.Equal(t, testCase.verdict, result.Verdict)
			require.False(t, result.Equivalent)
		})
	}
}
