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
// +build integration

package hybrid

import (
	"context"
	"testing"

	schema "github.com/specterops/bloodhound/graphschema"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/hybrid"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
)

func TestHybridAttackPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	// ADUser.ObjectID matches AZUser.OnPremID, AZUser.OnPremSyncEnabled is true
	// SyncedToEntraUser and SyncedToADUser edges should be created and link the two nodes
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

			verifyHybridPaths(t, db, harness, true, true)
		},
	)

	// ADUser.ObjectID do NOT match as AZUser.OnPremID is null, AZUser.OnPremSyncEnabled is false
	// SyncedToEntraUser and SyncedToADUser edges should NOT be created
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

			verifyHybridPaths(t, db, harness, false, true)
		},
	)

	// ADUser.ObjectID matches AZUser.OnPremID, AZUser.OnPremSyncEnabled is false
	// SyncedToEntraUser and SyncedToADUser edges should NOT be created
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

			verifyHybridPaths(t, db, harness, false, true)
		},
	)

	// ADUser does not exist. AZUser has OnPremID and OnPremSyncEnabled=true
	// A new ADUser node should be created. SyncedToADUser and SyncedToEntraUser edges should be created and linked to new ADUser node.
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

			verifyHybridPaths(t, db, harness, true, true)
		},
	)

	// ADUser does not exist, but the objectid from a selected AZUser exists in the graph. Selected AZUser has OnPremID and
	// OnPremSyncEnabled=true
	// The existing node should be used to create SyncedToADUser and SyncedToEntraUser edges.
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

			verifyHybridPaths(t, db, harness, true, false)
		},
	)

	// ADUser.ObjectID does NOT match AZUser.OnPremID, AZUser.OnPremSyncEnabled is true
	// SyncedToEntraUser and SyncedToADUser edges should be created, but a new ADUser node should be created with ObjectID that matches AZUser.OnPremID
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

			verifyHybridPaths(t, db, harness, true, true)
		},
	)
}

func verifyHybridPaths(t *testing.T, db graph.Database, harness integration.HarnessDetails, shouldHaveEdges bool, shouldHaveUserNode bool) {
	expectedEdgeCount := 1
	if !shouldHaveEdges {
		expectedEdgeCount = 0
	}

	// Verify the SyncedToADUser edge
	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		// Pull the edges
		syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), azure.SyncedToADUser)
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

			startObjectOnPremIdProp := start.Properties.Get(azure.OnPremID.String())
			startObjectOnPremId, err := startObjectOnPremIdProp.String()
			assert.Nil(t, err)

			// Get the ObjectID from the ADUser node
			endObjectProp := end.Properties.Get(common.ObjectID.String())
			endObjectID, err := endObjectProp.String()
			assert.Nil(t, err)

			// Ensure we got the correct node types
			if shouldHaveUserNode {
				assert.True(t, end.Kinds.ContainsOneOf(ad.User))
			} else {
				assert.True(t, end.Kinds.ContainsOneOf(ad.Entity))
			}
			assert.True(t, start.Kinds.ContainsOneOf(azure.User))

			// Verify the AZUser is the first node
			assert.Equal(t, harness.HybridAttackPaths.AZUserObjectID, startObjectID)

			// Verify the ADUser, but we have to handle the case where the ADUser node is created by the post-processing logic
			if harness.HybridAttackPaths.ADUserObjectID != startObjectOnPremId {
				// Node was created during post-processing. Pull AZUser.OnPremID from the node itself.
				assert.Equal(t, startObjectOnPremId, endObjectID)
			} else {
				// Node existed prior to post-processing. Use the node configured during setup.
				assert.Equal(t, harness.HybridAttackPaths.ADUserObjectID, endObjectID)
			}
		}

		return nil
	})

	// Verify the SyncedToEntraUser edge
	db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		// Pull the edges
		syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.SyncedToEntraUser)
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

			endObjectOnPremIdProp := end.Properties.Get(azure.OnPremID.String())
			endObjectOnPremId, err := endObjectOnPremIdProp.String()
			assert.Nil(t, err)

			// Ensure we got the correct node types
			if shouldHaveUserNode {
				assert.True(t, start.Kinds.ContainsOneOf(ad.User))
			} else {
				assert.True(t, start.Kinds.ContainsOneOf(ad.Entity))
			}
			assert.True(t, end.Kinds.ContainsOneOf(azure.User))

			// Verify the ADUser, but we have to handle the case where the ADUser node is created by the post-processing logic
			if harness.HybridAttackPaths.ADUserObjectID != endObjectOnPremId {
				// Node was created during post-processing. Pull AZUser.OnPremID from the node itself.
				assert.Equal(t, endObjectOnPremId, startObjectID)
			} else {
				// Node existed prior to post-processing. Use the node configured during setup.
				assert.Equal(t, harness.HybridAttackPaths.ADUserObjectID, startObjectID)
			}
			assert.Equal(t, harness.HybridAttackPaths.AZUserObjectID, endObjectID)
		}

		return nil
	})
}
