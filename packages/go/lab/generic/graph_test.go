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

package generic_test

import (
	"testing"
	"testing/fstest"

	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/stretchr/testify/require"
)

func TestLoadGraphFromFilePreservesNumbers(t *testing.T) {
	var (
		fixture = fstest.MapFS{"graph.json": {Data: []byte(`{
			"graph": {
				"nodes": [{"id": "node", "kinds": ["Test"], "properties": {
					"large": 9007199254740993, "fraction": 1.25,
					"nested": [1, {"negative": -9007199254740993, "exponent": 1e3}],
					"text": "1", "flag": true, "empty": null
				}}],
				"edges": [{"start": {"value": "node"}, "end": {"value": "node"},
					"kind": "TestEdge", "properties": {"port": 1433, "weight": 0.5}}]
			}
		}`)}}
	)

	loadedGraph, err := generic.LoadGraphFromFile(fixture, "graph.json")
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"large":    int64(9007199254740993),
		"fraction": 1.25,
		"nested":   []any{int64(1), map[string]any{"negative": int64(-9007199254740993), "exponent": float64(1000)}},
		"text":     "1", "flag": true, "empty": nil,
	}, loadedGraph.Nodes[0].Properties)
	require.Equal(t, map[string]any{"port": int64(1433), "weight": 0.5}, loadedGraph.Edges[0].Properties)
}

func TestLoadGraphFromFileRejectsOutOfRangeNumbers(t *testing.T) {
	for _, properties := range []string{
		`"nodes": [{"properties": {"integer": 9223372036854775808}}]`,
		`"nodes": [{"properties": {"nested": [1e400]}}]`,
		`"edges": [{"properties": {"integer": -9223372036854775809}}]`,
	} {
		t.Run(properties, func(t *testing.T) {
			var fixture = fstest.MapFS{"graph.json": {Data: []byte(`{"graph": {` + properties + `}}`)}}
			_, err := generic.LoadGraphFromFile(fixture, "graph.json")
			require.Error(t, err)
		})
	}
}
