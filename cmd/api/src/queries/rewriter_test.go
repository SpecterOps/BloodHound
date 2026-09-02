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

package queries_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestNewRewriter(t *testing.T) {
	rewriter := queries.NewRewriter()

	require.NotNil(t, rewriter)
	require.NotNil(t, rewriter.Visitor)
	require.False(t, rewriter.HasMutation)
	require.False(t, rewriter.HasRelationshipTypeShortcut)
}

func TestRewriter_Enter_UpdatingClauseSetsHasMutation(t *testing.T) {
	rewriter := queries.NewRewriter()

	rewriter.Enter(&cypher.UpdatingClause{})

	require.True(t, rewriter.HasMutation)
	require.False(t, rewriter.HasRelationshipTypeShortcut)
}

func TestRewriter_Enter_ADAttackPathsShortcut_ReplacesKinds(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{graph.StringKind("AD_ATTACK_PATHS")},
		}
	)

	rewriter.Enter(relationshipPattern)

	require.True(t, rewriter.HasRelationshipTypeShortcut)
	require.False(t, rewriter.HasMutation)
	require.Equal(t, graph.Kinds(ad.PathfindingRelationships()), relationshipPattern.Kinds)
}

func TestRewriter_Enter_AzureAttackPathsShortcut_ReplacesKinds(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{graph.StringKind("AZ_ATTACK_PATHS")},
		}
	)

	rewriter.Enter(relationshipPattern)

	require.True(t, rewriter.HasRelationshipTypeShortcut)
	require.False(t, rewriter.HasMutation)
	require.Equal(t, graph.Kinds(azure.PathfindingRelationships()), relationshipPattern.Kinds)
}

func TestRewriter_Enter_CombinedAzureAndADShortcut_ConcatenatesKinds(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{
				graph.StringKind("AZ_ATTACK_PATHS"),
				graph.StringKind("AD_ATTACK_PATHS"),
			},
		}
		expectedKinds = graph.Kinds(append(azure.PathfindingRelationships(), ad.PathfindingRelationships()...))
	)

	rewriter.Enter(relationshipPattern)

	require.True(t, rewriter.HasRelationshipTypeShortcut)
	require.Equal(t, expectedKinds, relationshipPattern.Kinds)
}

func TestRewriter_Enter_CombinedADAndAzureShortcut_PreservesOrder(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{
				graph.StringKind("AD_ATTACK_PATHS"),
				graph.StringKind("AZ_ATTACK_PATHS"),
			},
		}
		expectedKinds = graph.Kinds(append(ad.PathfindingRelationships(), azure.PathfindingRelationships()...))
	)

	rewriter.Enter(relationshipPattern)

	require.True(t, rewriter.HasRelationshipTypeShortcut)
	require.Equal(t, expectedKinds, relationshipPattern.Kinds)
}

func TestRewriter_Enter_AllAttackPathsShortcut_ReplacesKinds(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{graph.StringKind("ALL_ATTACK_PATHS")},
		}
		expectedKinds = graph.Kinds(append(azure.PathfindingRelationships(), ad.PathfindingRelationships()...))
	)

	rewriter.Enter(relationshipPattern)

	require.True(t, rewriter.HasRelationshipTypeShortcut)
	require.Equal(t, expectedKinds, relationshipPattern.Kinds)
}

func TestRewriter_Enter_AllAttackPathsShortcut_TakesPrecedenceOverOtherShortcuts(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{
				graph.StringKind("ALL_ATTACK_PATHS"),
				graph.StringKind("AZ_ATTACK_PATHS"),
				graph.StringKind("AD_ATTACK_PATHS"),
			},
		}
		expectedKinds = graph.Kinds(append(azure.PathfindingRelationships(), ad.PathfindingRelationships()...))
	)

	rewriter.Enter(relationshipPattern)

	require.True(t, rewriter.HasRelationshipTypeShortcut)
	require.Equal(t, expectedKinds, relationshipPattern.Kinds)
}

func TestRewriter_Enter_NoShortcut_PreservesKinds(t *testing.T) {
	var (
		rewriter            = queries.NewRewriter()
		originalKind        = graph.StringKind("MemberOf")
		relationshipPattern = &cypher.RelationshipPattern{
			Kinds: graph.Kinds{originalKind},
		}
	)

	rewriter.Enter(relationshipPattern)

	require.False(t, rewriter.HasRelationshipTypeShortcut)
	require.False(t, rewriter.HasMutation)
	require.Equal(t, graph.Kinds{originalKind}, relationshipPattern.Kinds)
}
