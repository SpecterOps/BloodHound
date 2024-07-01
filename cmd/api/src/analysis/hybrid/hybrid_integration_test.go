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

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.HybridAttackPaths.Setup(testContext, false)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

			if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
				t.Fatalf("failed post processing for hybrid attack paths: %v", err)
			}

			operation.Done()

			// Verify the SyncedToADUser edge
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), azure.SyncedToADUser)
				}))
				assert.Nil(t, err)
				assert.Len(t, syncedToADUserEdges, 1)

				for _, edge := range syncedToADUserEdges {
					start, end, err := ops.FetchRelationshipNodes(tx, edge)
					assert.Nil(t, err)

					startObjectProp := start.Properties.Get(common.ObjectID.String())
					startObjectID, err := startObjectProp.String()
					assert.Nil(t, err)

					endObjectProp := end.Properties.Get(common.ObjectID.String())
					endObjectID, err := endObjectProp.String()
					assert.Nil(t, err)

					assert.Equal(t, azure.SyncedToADUser, edge.Kind)
					assert.Equal(t, harness.HybridAttackPaths.AZUserID, startObjectID)
					assert.Equal(t, harness.HybridAttackPaths.ADUserID, endObjectID)
				}

				return nil
			})

			// Verify the SyncedToEntraUser edge
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				syncedToEntraUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.SyncedToEntraUser)
				}))
				assert.Nil(t, err)
				assert.Len(t, syncedToEntraUserEdges, 1)

				for _, edge := range syncedToEntraUserEdges {
					start, end, err := ops.FetchRelationshipNodes(tx, edge)
					assert.Nil(t, err)

					startObjectProp := start.Properties.Get(common.ObjectID.String())
					startObjectID, err := startObjectProp.String()
					assert.Nil(t, err)

					endObjectProp := end.Properties.Get(common.ObjectID.String())
					endObjectID, err := endObjectProp.String()
					assert.Nil(t, err)

					assert.Equal(t, ad.SyncedToEntraUser, edge.Kind)
					assert.Equal(t, harness.HybridAttackPaths.ADUserID, startObjectID)
					assert.Equal(t, harness.HybridAttackPaths.AZUserID, endObjectID)
				}

				return nil
			})
		},
	)
}

func TestHybridAttackPathsWithNoADUserNode(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.HybridAttackPaths.Setup(testContext, true)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "Hybrid Attack Path Post Process Test")

			if _, err := hybrid.PostHybrid(context.Background(), db); err != nil {
				t.Fatalf("failed post processing for hybrid attack paths: %v", err)
			}

			operation.Done()

			// Verify the SyncedToADUser edge
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), azure.SyncedToADUser)
				}))
				assert.Nil(t, err)
				assert.Len(t, syncedToADUserEdges, 1)

				for _, edge := range syncedToADUserEdges {
					start, end, err := ops.FetchRelationshipNodes(tx, edge)
					assert.Nil(t, err)

					startObjectProp := start.Properties.Get(common.ObjectID.String())
					startObjectID, err := startObjectProp.String()
					assert.Nil(t, err)

					endObjectProp := end.Properties.Get(common.ObjectID.String())
					endObjectID, err := endObjectProp.String()
					assert.Nil(t, err)

					assert.Equal(t, azure.SyncedToADUser, edge.Kind)

					assert.True(t, end.Kinds.ContainsOneOf(ad.User))
					assert.Equal(t, harness.HybridAttackPaths.AZUserID, startObjectID)
					assert.Equal(t, harness.HybridAttackPaths.ADUserID, endObjectID)
				}

				return nil
			})

			// Verify the SyncedToEntraUser edge
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				syncedToADUserEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.SyncedToEntraUser)
				}))
				assert.Nil(t, err)
				assert.Len(t, syncedToADUserEdges, 1)

				for _, edge := range syncedToADUserEdges {
					start, end, err := ops.FetchRelationshipNodes(tx, edge)
					assert.Nil(t, err)

					startObjectProp := start.Properties.Get(common.ObjectID.String())
					startObjectID, err := startObjectProp.String()
					assert.Nil(t, err)

					endObjectProp := end.Properties.Get(common.ObjectID.String())
					endObjectID, err := endObjectProp.String()
					assert.Nil(t, err)

					assert.Equal(t, ad.SyncedToEntraUser, edge.Kind)

					assert.True(t, end.Kinds.ContainsOneOf(azure.User))
					assert.Equal(t, harness.HybridAttackPaths.ADUserID, startObjectID)
					assert.Equal(t, harness.HybridAttackPaths.AZUserID, endObjectID)
				}

				return nil
			})
		},
	)
}
