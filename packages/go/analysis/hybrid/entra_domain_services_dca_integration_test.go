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

//go:build integration

package hybrid

import (
	"context"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHybridEntraDSDCA(t *testing.T) {
	var (
		testContext                         = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		controlledAZUser, controlledAZGroup *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			azUser, _, azGroup, _ := setupEntraDSGroupMemberHarness(t, testContext, azure.AddMembers, true, true)
			controlledAZUser = azUser
			controlledAZGroup = azGroup
			setupManageEntraDSSyncHarness(t, testContext, validManageEntraDSSyncOptions())
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			kinds := graph.Kinds{
				azure.SyncedToEntraDSUser,
				azure.SyncedToEntraDSGroup,
				azure.AddEntraDSGroupMember,
				azure.EntraDSFor,
				azure.ManageEntraDSSync,
				azure.ManageEntraDSSyncFilter,
			}

			firstRunStats, err := PostHybrid(context.Background(), db, true)
			require.NoError(t, err)
			firstRunEdges := fetchRelationshipsByKinds(t, db, kinds)
			firstRunIDs := make(map[string]graph.ID)
			firstRunFirstSeen := make(map[string]time.Time)
			for _, kind := range kinds {
				require.NotEmpty(t, firstRunEdges[kind])
				require.NotNil(t, firstRunStats.RelationshipsCreated[kind])
				for _, edge := range firstRunEdges[kind] {
					firstSeen, err := edge.Properties.Get(common.FirstSeen.String()).Time()
					require.NoError(t, err)
					edgeIdentity := relationshipIdentity(edge)
					firstRunIDs[edgeIdentity] = edge.ID
					firstRunFirstSeen[edgeIdentity] = firstSeen
				}
			}

			secondRunStats, err := PostHybrid(context.Background(), db, true)
			require.NoError(t, err)
			secondRunEdges := fetchRelationshipsByKinds(t, db, kinds)
			for _, kind := range kinds {
				assert.NotContains(t, secondRunStats.RelationshipsCreated, kind)
				require.Len(t, secondRunEdges[kind], len(firstRunEdges[kind]))
				for _, edge := range secondRunEdges[kind] {
					edgeIdentity := relationshipIdentity(edge)
					assert.Equal(t, firstRunIDs[edgeIdentity], edge.ID)
					firstSeen, err := edge.Properties.Get(common.FirstSeen.String()).Time()
					require.NoError(t, err)
					assert.Equal(t, firstRunFirstSeen[edgeIdentity], firstSeen)
				}
			}

			err = db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
				return tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.InIDs(query.StartID(), controlledAZUser.ID),
						query.InIDs(query.EndID(), controlledAZGroup.ID),
						query.Kind(query.Relationship(), azure.AddMembers),
					)
				}).Delete()
			})
			require.NoError(t, err)

			_, err = PostHybrid(context.Background(), db, true)
			require.NoError(t, err)
			evidenceRemovedEdges := fetchRelationshipsByKinds(t, db, kinds)
			assert.Empty(t, evidenceRemovedEdges[azure.AddEntraDSGroupMember])
			for _, kind := range kinds {
				if kind == azure.AddEntraDSGroupMember {
					continue
				}

				require.Len(t, evidenceRemovedEdges[kind], len(firstRunEdges[kind]))
				for _, edge := range evidenceRemovedEdges[kind] {
					assert.Equal(t, firstRunIDs[relationshipIdentity(edge)], edge.ID)
				}
			}

			_, err = PostHybrid(context.Background(), db, false)
			require.NoError(t, err)
			disabledEdges := fetchRelationshipsByKinds(t, db, kinds)
			for _, kind := range kinds {
				assert.Empty(t, disabledEdges[kind])
			}
		},
	)
}

func TestPostHybridDisabledRetainsLegacyHybridEdges(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			adUserObjectID := integration.RandomObjectID(t)
			harness.HybridAttackPaths.Setup(testContext, adUserObjectID, adUserObjectID, true, true, false)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, err := PostHybrid(context.Background(), db, false)
			require.NoError(t, err)
			verifyHybridPaths(t, db, harness, true)
		},
	)
}

func fetchRelationshipsByKinds(t *testing.T, db graph.Database, kinds graph.Kinds) map[graph.Kind][]*graph.Relationship {
	t.Helper()

	relationshipsByKind := make(map[graph.Kind][]*graph.Relationship, len(kinds))
	err := db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		for _, kind := range kinds {
			relationships, err := ops.FetchRelationships(tx.Relationships().Filter(query.Kind(query.Relationship(), kind)))
			if err != nil {
				return err
			}
			relationshipsByKind[kind] = relationships
		}

		return nil
	})
	require.NoError(t, err)

	return relationshipsByKind
}

func relationshipIdentity(relationship *graph.Relationship) string {
	return relationship.Kind.String() + "|" + relationship.StartID.String() + "|" + relationship.EndID.String()
}
