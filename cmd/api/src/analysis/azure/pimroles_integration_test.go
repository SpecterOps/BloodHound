//go:build integration

package azure_test

import (
	"fmt"
	"testing"

	azureAnalysis "github.com/specterops/bloodhound/analysis/azure"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzurePIMRolesAZRoleApprover(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZPIMRolesHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		stats, err := azureAnalysis.CreateAZRoleApproverEdge(testContext.Context(), db)
		require.NoError(t, err)

		require.NotNil(t, stats)
		require.NotEmpty(t, stats.RelationshipsCreated)
		assert.Equal(t, int32(10), *stats.RelationshipsCreated[azure.AZRoleApprover])

		db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
			results, err := ops.FetchRelationships(tx.Relationships().Filter(query.Kind(query.Relationship(), azure.AZRoleApprover)))
			require.NoError(t, err)

			assert.Equal(t, len(results), 10)

			for _, result := range results {
				startNode, err := ops.FetchNode(tx, result.StartID)
				require.NoError(t, err)

				endNode, err := ops.FetchNode(tx, result.EndID)
				require.NoError(t, err)

				startName, err := startNode.Properties.Get(common.Name.String()).String()
				require.NoError(t, err)

				endName, err := endNode.Properties.Get(common.Name.String()).String()
				require.NoError(t, err)

				switch {
				case startName == "AZBase n5":
					assert.Equal(t, "PIMTestRole", endName)
				case startName == "AZBase n7":
					assert.True(t, endName == "PIMTestRole" || endName == "PIMTestRole3", fmt.Sprintf("expected: %s", endName))
				case startName == "AZBase n10":
					assert.Equal(t, "PIMTestRole3", endName)
				case startName == "AZBase n12":
					assert.Equal(t, "PIMTestRole2", endName)
				case startName == "AZBase n13":
					assert.Equal(t, "PIMTestRole2", endName)
				case startName == "AZBase n15":
					assert.Equal(t, "PIMTestRole", endName)
				case startName == "AZBase n16":
					assert.Equal(t, "PIMTestRole", endName)
				case startName == "Privileged Role Administrator":
					assert.Equal(t, "PIMTestRole4", endName)
				case startName == "Global Administrator":
					assert.Equal(t, "PIMTestRole4", endName)
				}
			}

			return nil
		})
	})
}
