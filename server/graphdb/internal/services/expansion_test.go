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

package services_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/bloodhound/server/graphdb/internal/services/mocks"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestGraphExpansionLimit(t *testing.T) {
	t.Parallel()

	limit, err := services.GraphExpansionLimit(0)
	require.NoError(t, err)
	require.Equal(t, services.DefaultGraphExpansionLimit, limit)

	limit, err = services.GraphExpansionLimit(25)
	require.NoError(t, err)
	require.Equal(t, 25, limit)

	_, err = services.GraphExpansionLimit(-1)
	require.ErrorIs(t, err, services.ErrInvalidGraphExpansionLimit)
	require.Contains(t, err.Error(), "greater than 0")

	_, err = services.GraphExpansionLimit(services.MaxGraphExpansionLimit + 1)
	require.ErrorIs(t, err, services.ErrInvalidGraphExpansionLimit)
	require.Contains(t, err.Error(), "less than or equal to 1000")
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

	prunedGraph, truncated := services.PruneGraphExpansionResponse(graphResponse, 1)

	require.True(t, truncated)
	require.Len(t, prunedGraph.Edges, 1)
	require.Contains(t, prunedGraph.Nodes, "1")
	require.Contains(t, prunedGraph.Nodes, "2")
	require.NotContains(t, prunedGraph.Nodes, "3")
}

func TestService_ExpandGraphFiltersETAC(t *testing.T) {
	t.Parallel()

	var (
		ctx          = t.Context()
		databaseMock = mocks.NewMockDatabase(t)
		svc          = services.NewService(databaseMock, newAllowAllNodeAccessChecker(t), testAccessControl{shouldFilter: true})
		anchorNode   = services.Node{ID: 42}
	)

	databaseMock.EXPECT().GetNode(ctx, int64(42)).Return(anchorNode, nil)
	databaseMock.EXPECT().ExpandGraph(ctx, int64(42), services.GraphExpansionDirectionOutbound, 3).Return(model.UnifiedGraph{
		Nodes: map[string]model.UnifiedNode{
			"1": {
				Label:      "allowed",
				Kinds:      []string{"User"},
				Properties: map[string]any{ad.DomainSID.String(): "allowed-env", "name": "allowed"},
			},
			"2": {
				Label:      "hidden",
				Kinds:      []string{"Computer"},
				Properties: map[string]any{ad.DomainSID.String(): "hidden-env", "name": "hidden"},
			},
		},
		Edges: []model.UnifiedEdge{
			{Source: "1", Target: "2", Label: "edge", Kind: "MemberOf", Properties: map[string]any{"edge": "secret"}},
		},
		Literals: graph.Literals{{Value: "secret"}},
	}, nil)

	expansion, err := svc.ExpandGraph(ctx, services.GraphExpansionRequest{
		NodeID:            42,
		Direction:         services.GraphExpansionDirectionOutbound,
		Limit:             2,
		IncludeProperties: true,
		User: model.User{
			EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
				{EnvironmentID: "allowed-env"},
			},
		},
	})

	require.NoError(t, err)
	require.False(t, expansion.Truncated)
	require.False(t, expansion.Graph.Nodes["1"].Hidden)
	require.Equal(t, "allowed", expansion.Graph.Nodes["1"].Properties["name"])
	require.True(t, expansion.Graph.Nodes["2"].Hidden)
	require.Nil(t, expansion.Graph.Nodes["2"].Properties)
	require.Equal(t, "HIDDEN", expansion.Graph.Edges[0].Kind)
	require.Nil(t, expansion.Graph.Edges[0].Properties)
	require.Empty(t, expansion.Graph.Literals)
}

func TestService_ExpandGraphDeniesInaccessibleAnchorNode(t *testing.T) {
	t.Parallel()

	var (
		ctx          = t.Context()
		databaseMock = mocks.NewMockDatabase(t)
		anchorNode   = services.Node{ID: 42}
		svc          = services.NewService(databaseMock, newDenyAllNodeAccessChecker(t), testAccessControl{})
	)

	databaseMock.EXPECT().GetNode(ctx, int64(42)).Return(anchorNode, nil)

	_, err := svc.ExpandGraph(ctx, services.GraphExpansionRequest{
		NodeID:    42,
		Direction: services.GraphExpansionDirectionOutbound,
	})

	require.ErrorIs(t, err, services.ErrNodeAccessDenied)
}
