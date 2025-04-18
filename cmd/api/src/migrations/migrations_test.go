package migrations_test

import (
	"context"
	"errors"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersion_730_Migration(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	t.Run("Migration_v730 Success", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version730_Migration.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_730_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				computers, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)))

				require.Nil(t, err)

				for _, computer := range computers {
					if computer.ID == harness.Version730_Migration.Computer1.ID {
						smbSigning, err := computer.Properties.Get(ad.SMBSigning.String()).Bool()
						require.Nil(t, err)
						require.True(t, smbSigning)
					} else {
						_, err := computer.Properties.Get(ad.SMBSigning.String()).Bool()
						require.Error(t, err)
						require.True(t, errors.Is(err, graph.ErrPropertyNotFound))
					}
				}

				return nil
			})
		})
	})
}
