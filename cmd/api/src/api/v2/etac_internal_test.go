// Copyright 2026 Specter Ops, Inc.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivilegeZoneEnvironmentNodeETAC(t *testing.T) {
	t.Parallel()

	node := graph.PrepareNode(graph.AsProperties(map[string]any{
		graphschema.EnvironmentIDKey: "allowed-environment",
	}), graph.StringKind("PZ_PrivilegeZoneEnvironment"))

	assert.False(t, nodeGatedByETAC([]string{"allowed-environment"}, node))
	assert.True(t, nodeGatedByETAC([]string{"other-environment"}, node))
}

func TestFilterETACGraphPrivilegeZoneCanonicalVisibility(t *testing.T) {
	t.Parallel()

	graphResponse := model.UnifiedGraph{
		Nodes: map[string]model.UnifiedNode{
			"allowed": {
				Kinds:      []string{"PZ_PrivilegeZoneEnvironment"},
				Properties: map[string]any{graphschema.EnvironmentIDKey: "allowed-environment"},
			},
			"denied": {
				Kinds:      []string{"PZ_PrivilegeZoneEnvironment"},
				Properties: map[string]any{graphschema.EnvironmentIDKey: "denied-environment"},
			},
			"canonical": {Kinds: []string{"PZ_PrivilegeZone"}, Properties: map[string]any{}},
		},
		Edges: []model.UnifiedEdge{
			{Source: "allowed", Target: "canonical", Kind: "PZ_PartOfZone", Label: "Part Of Zone"},
			{Source: "denied", Target: "canonical", Kind: "PZ_PartOfZone", Label: "Part Of Zone"},
		},
	}
	user := model.User{EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "allowed-environment"}}}

	filtered, err := filterETACGraph(graphResponse, user)
	require.NoError(t, err)
	assert.False(t, filtered.Nodes["allowed"].Hidden)
	assert.True(t, filtered.Nodes["denied"].Hidden)
	assert.False(t, filtered.Nodes["canonical"].Hidden)
	assert.Equal(t, "PZ_PartOfZone", filtered.Edges[0].Kind)
	assert.Equal(t, "HIDDEN", filtered.Edges[1].Kind)
}

func TestFilterETACGraphHidesCanonicalZoneWithoutAuthorizedEnvironment(t *testing.T) {
	t.Parallel()

	graphResponse := model.UnifiedGraph{
		Nodes: map[string]model.UnifiedNode{
			"denied": {
				Kinds:      []string{"PZ_PrivilegeZoneEnvironment"},
				Properties: map[string]any{graphschema.EnvironmentIDKey: "denied-environment"},
			},
			"canonical": {Kinds: []string{"PZ_PrivilegeZone"}, Properties: map[string]any{}},
		},
		Edges: []model.UnifiedEdge{{Source: "denied", Target: "canonical", Kind: "PZ_PartOfZone"}},
	}
	user := model.User{EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "allowed-environment"}}}

	filtered, err := filterETACGraph(graphResponse, user)
	require.NoError(t, err)
	assert.True(t, filtered.Nodes["canonical"].Hidden)
}
