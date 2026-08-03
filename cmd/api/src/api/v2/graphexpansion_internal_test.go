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
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphExpansionQuery(t *testing.T) {
	t.Parallel()

	t.Run("outbound", func(t *testing.T) {
		t.Parallel()

		query, err := buildGraphExpansionQuery(42, graphExpansionDirectionOutbound, []string{"CustomTraversable"}, 501)

		require.NoError(t, err)
		require.Equal(t, `MATCH (source)
WHERE ID(source) = 42
MATCH (source)-[r:ALL_ATTACK_PATHS|CustomTraversable]->(target)
RETURN source, r, target
LIMIT 501`, query)
	})

	t.Run("inbound", func(t *testing.T) {
		t.Parallel()

		query, err := buildGraphExpansionQuery(42, graphExpansionDirectionInbound, nil, 501)

		require.NoError(t, err)
		require.Contains(t, query, "MATCH (target)-[r:ALL_ATTACK_PATHS]->(source)")
	})

	t.Run("deduplicates relationship kinds", func(t *testing.T) {
		t.Parallel()

		query, err := buildGraphExpansionQuery(42, graphExpansionDirectionOutbound, []string{"Custom", "Custom"}, 501)

		require.NoError(t, err)
		require.Equal(t, 1, strings.Count(query, "Custom"))
	})

	t.Run("rejects invalid relationship kinds", func(t *testing.T) {
		t.Parallel()

		_, err := buildGraphExpansionQuery(42, graphExpansionDirectionOutbound, []string{"Unsafe-Kind"}, 501)

		require.EqualError(t, err, "invalid relationship kind: Unsafe-Kind")
	})

	t.Run("rejects invalid direction", func(t *testing.T) {
		t.Parallel()

		_, err := buildGraphExpansionQuery(42, "sideways", nil, 501)

		require.EqualError(t, err, "direction must be either inbound or outbound")
	})
}

func TestGraphExpansionLimit(t *testing.T) {
	t.Parallel()

	limit, err := graphExpansionLimit(0)
	require.NoError(t, err)
	require.Equal(t, defaultGraphExpansionLimit, limit)

	limit, err = graphExpansionLimit(25)
	require.NoError(t, err)
	require.Equal(t, 25, limit)

	_, err = graphExpansionLimit(-1)
	require.EqualError(t, err, "limit must be greater than 0")

	_, err = graphExpansionLimit(maxGraphExpansionLimit + 1)
	require.EqualError(t, err, "limit must be less than or equal to 1000")
}

func TestPruneGraphExpansionResponse(t *testing.T) {
	t.Parallel()

	graphResponse := model.UnifiedGraph{
		Nodes: map[string]model.UnifiedNode{
			"1": {Label: "one"},
			"2": {Label: "two"},
			"3": {Label: "three"},
		},
		Edges: []model.UnifiedEdge{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
		},
	}

	prunedGraph, truncated := pruneGraphExpansionResponse(graphResponse, 1)

	require.True(t, truncated)
	require.Len(t, prunedGraph.Edges, 1)
	require.Contains(t, prunedGraph.Nodes, "1")
	require.Contains(t, prunedGraph.Nodes, "2")
	require.NotContains(t, prunedGraph.Nodes, "3")
}
