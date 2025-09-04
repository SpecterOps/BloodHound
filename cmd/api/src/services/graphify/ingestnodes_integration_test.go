package graphify_test

import (
	"context"
	"path"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

// 1. changelog disabled , ingest a node. created
// 2. changelog disabled , ingest a node. update.
// invalid case, 0 kinds
// 4 . changelog enabled. creating a node still works
// 5. changelog enabled. dedupe works.
// 6. changelog enabled. update works.

// new setup helper that does exactly what IngestNode needs
func TestIngestNode(t *testing.T) {
	t.Parallel()

	var (
		ctx         = context.Background()
		fixturePath = path.Join("fixtures", "IngestNode")
		testSuite   = setupIntegrationTestSuite(t, fixturePath)
		// todo: put the cl on the test suite struct
		cl = changelog.NewChangelog(testSuite.GraphDB, testSuite.BHDatabase, changelog.DefaultOptions())
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	t.Run("ingested node gets created", func(t *testing.T) {
		// open up a batchOp
		testSuite.GraphDB.BatchOperation(ctx,
			func(batch graph.Batch) error {
				ingestCtx := graphify.NewIngestContext(testSuite.Context, batch)

				// function under test
				err := graphify.IngestNode(ingestCtx, graph.StringKind("hellobase"), ein.IngestibleNode{
					ObjectID: "ABC123",
					Labels:   []graph.Kind{graph.StringKind("labelA"), graph.StringKind("labelB")},
					PropertyMap: map[string]any{
						"hello": "world",
						"1":     2,
					},
				})

				require.NoError(t, err)

				return nil
			})

		testSuite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {

			})
		})
	})

}
