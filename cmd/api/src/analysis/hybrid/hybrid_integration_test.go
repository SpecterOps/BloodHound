// Copyright 2024 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/hybrid"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
)

func TestHybridAttackPaths(t *testing.T) {
	t.Run("SyncedEdgesCreatedAndLinkExistingNodes", func(t *testing.T) {
		// ADUser.ObjectID matches AZUser.OnPremID, AZUser.OnPremSyncEnabled is true
		// SyncedToEntraUser and SyncedToADUser edges should be created and link the two nodes
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := adUserObjectID
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, true)
			},
		)
	})

	t.Run("SyncedEdgesNotCreated", func(t *testing.T) {

		// ADUser.ObjectID do NOT match as AZUser.OnPremID is null, AZUser.OnPremSyncEnabled is false
		// SyncedToEntraUser and SyncedToADUser edges should NOT be created
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := ""
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, false, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("OnPremSyncEnabled False", func(t *testing.T) {
		// ADUser.ObjectID matches AZUser.OnPremID, AZUser.OnPremSyncEnabled is false
		// SyncedToEntraUser and SyncedToADUser edges should NOT be created
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := adUserObjectID
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, false, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("SyncedEdgesNotCreatedWithoutMatchingADUser", func(t *testing.T) {
		// ADUser does not exist. AZUser has OnPremID and OnPremSyncEnabled=true
		// No ADUser node should be created. SyncedToADUser and SyncedToEntraUser edges should not be created.
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := ""
				azUserOnPremID := integration.RandomObjectID(t)
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, false, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("SyncedEdgesNotCreatedForUnknownADEntity", func(t *testing.T) {
		// ADUser does not exist, but the objectid from a selected AZUser exists in the graph. Selected AZUser has OnPremID and
		// OnPremSyncEnabled=true
		// Hybrid post-processing only links existing AD user nodes, so no synced edges should be created.
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := ""
				azUserOnPremID := integration.RandomObjectID(t)
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, false, true)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})

	t.Run("SyncedEdgesNotCreatedWhenADObjectIDDoesNotMatchOnPremID", func(t *testing.T) {
		// ADUser.ObjectID does NOT match AZUser.OnPremID, AZUser.OnPremSyncEnabled is true
		// SyncedToEntraUser and SyncedToADUser edges should not be created.
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(
			func(harness *integration.HarnessDetails) error {
				adUserObjectID := integration.RandomObjectID(t)
				azUserOnPremID := integration.RandomObjectID(t)
				harness.HybridAttackPaths.Setup(testContext, adUserObjectID, azUserOnPremID, true, true, false)
				return nil
			},
			func(harness integration.HarnessDetails, db graph.Database) {
				operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

				if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
					t.Fatalf("failed post processing for hybrid attack paths: %v", err)
				}
				operation.Done()

				verifyHybridPaths(t, db, harness, false)
			},
		)
	})
}

func verifyHybridPaths(t *testing.T, db graph.Database, harness integration.HarnessDetails, shouldHaveEdges bool) {
	expectedEdgeCount := 1
	if !shouldHaveEdges {
		expectedEdgeCount = 0
	}

	// Verify the SyncedToADUser edge
	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		// Pull the edges
		syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.SyncedToADUser)
		}))
		assert.Nil(t, err)
		assert.Len(t, syncedToADUserEdges, expectedEdgeCount)

		for _, edge := range syncedToADUserEdges {
			// Retrieve the nodes connected to the edge
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			// Get ObjectID and OnPremID from the AZUser node
			startObjectProp := start.Properties.Get(common.ObjectID.String())
			startObjectID, err := startObjectProp.String()
			assert.Nil(t, err)

			// Get the ObjectID from the ADUser node
			endObjectProp := end.Properties.Get(common.ObjectID.String())
			endObjectID, err := endObjectProp.String()
			assert.Nil(t, err)

			// Ensure we got the correct node types
			assert.True(t, end.Kinds.ContainsOneOf(ad.User))
			assert.True(t, start.Kinds.ContainsOneOf(azure.User))

			// Verify the AZUser is the first node
			assert.Equal(t, harness.HybridAttackPaths.AZUserObjectID, startObjectID)
			assert.Equal(t, harness.HybridAttackPaths.ADUserObjectID, endObjectID)
		}

		return nil
	})

	// Verify the SyncedToEntraUser edge
	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		// Pull the edges
		syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), azure.SyncedToEntraUser)
		}))
		assert.Nil(t, err)
		assert.Len(t, syncedToADUserEdges, expectedEdgeCount)

		for _, edge := range syncedToADUserEdges {
			// Retrieve the nodes connected to the edge
			start, end, err := ops.FetchRelationshipNodes(tx, edge)
			assert.Nil(t, err)

			// Get the ObjectID from the ADUser node
			startObjectProp := start.Properties.Get(common.ObjectID.String())
			startObjectID, err := startObjectProp.String()
			assert.Nil(t, err)

			// Get ObjectID and OnPremID from the AZUser node
			endObjectProp := end.Properties.Get(common.ObjectID.String())
			endObjectID, err := endObjectProp.String()
			assert.Nil(t, err)

			// Ensure we got the correct node types
			assert.True(t, start.Kinds.ContainsOneOf(ad.User))
			assert.True(t, end.Kinds.ContainsOneOf(azure.User))

			assert.Equal(t, harness.HybridAttackPaths.ADUserObjectID, startObjectID)
			assert.Equal(t, harness.HybridAttackPaths.AZUserObjectID, endObjectID)
		}

		return nil
	})
}
