// Copyright 2025 Specter Ops, Inc.
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
//go:build slow_integration

package graphify_test

import (
	"context"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func setupTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	//#region Setup for dbs
	pool, err := pg.NewPool(connConf.URL())
	require.NoError(t, err)

	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	err = migrations.NewGraphMigrator(graphDB).Migrate(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	err = db.Migrate(ctx)
	require.NoError(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	ingestSchema, err := upload.LoadIngestSchema()
	require.NoError(t, err)

	//#endregion

	cfg := config.Configuration{}

	cl := changelog.NewChangelog(graphDB, db, changelog.DefaultOptions())
	cl.InitCacheForTest(ctx)
	cl.Start(ctx)

	return IntegrationTestSuite{
		Context:         ctx,
		GraphifyService: graphify.NewGraphifyService(ctx, db, graphDB, cfg, ingestSchema, cl),
		GraphDB:         graphDB,
		BHDatabase:      db,
		Changelog:       cl,
	}
}

// new setup helper that does exactly what IngestNode needs
func TestIngestNode(t *testing.T) {
	t.Parallel()

	var (
		ctx       = context.Background()
		testSuite = setupTestSuite(t)
		// when counting and querying the nodes table, we want to filter out this default node that always gets created on graph startup
		filter = query.Not(query.Kind(query.Node(), common.MigrationData))
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	t.Run("ingested node gets created", func(t *testing.T) {

		var (
			ingestedPropertyBag = graph.NewProperties().SetAll(map[string]any{
				"hello": "world",
				"1":     2,
			})
			// clone here because ingestNode decorates the incoming propety bag with additional properties like lastseen
			expectedPropertyBag = ingestedPropertyBag.Clone()
			sourceKind          = graph.StringKind("hellobase")
			ingestedKinds       = graph.Kinds{graph.StringKind("labelA"), graph.StringKind("labelB")}
		)

		// open up a batchOp
		err := testSuite.GraphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			ingestCtx := graphify.NewIngestContext(ctx, graphify.WithBatchUpdater(batch))

			// function under test
			err := graphify.IngestNode(ingestCtx, sourceKind,
				ein.IngestibleNode{
					ObjectID:    "ABC123",
					Labels:      ingestedKinds,
					PropertyMap: ingestedPropertyBag.Map,
				})

			require.NoError(t, err)

			return nil
		})

		require.NoError(t, err)

		err = testSuite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			count, err := tx.Nodes().Filter(filter).Count()
			require.NoError(t, err)
			require.Equal(t, int64(1), count)

			tx.Nodes().Filter(filter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for node := range cursor.Chan() {
					objectid, _ := node.Properties.Get(common.ObjectID.String()).String()
					require.Equal(t, "ABC123", objectid)

					lastseen, _ := node.Properties.Get(common.LastSeen.String()).String()
					require.NotEmpty(t, lastseen)

					// kinds assert
					expectedKinds := ingestedKinds.Add(sourceKind)
					actualKinds := node.Kinds
					for _, actualKind := range actualKinds {
						require.Contains(t, expectedKinds, actualKind)
					}

					actualPropertyBag := node.Properties.Map
					for k, v := range expectedPropertyBag.Map {
						require.EqualValues(t, v, actualPropertyBag[k])
					}

				}
				return nil
			})
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("ingested node gets created with changelog on", func(t *testing.T) {

		var (
			ingestedPropertyBag = graph.NewProperties().SetAll(map[string]any{
				"hello": "world",
				"1":     2,
			})
			// clone here because ingestNode decorates the incoming propety bag with additional properties like lastseen
			expectedPropertyBag = ingestedPropertyBag.Clone()
			sourceKind          = graph.StringKind("hellobase")
			ingestedKinds       = graph.Kinds{graph.StringKind("labelA"), graph.StringKind("labelB")}
		)

		// open up a batchOp
		err := testSuite.GraphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			// inject changelog here to simulate it being toggled on
			ingestCtx := graphify.NewIngestContext(
				testSuite.Context,
				graphify.WithChangeManager(testSuite.Changelog),
				graphify.WithBatchUpdater(batch))

			// function under test
			err := graphify.IngestNode(ingestCtx, sourceKind,
				ein.IngestibleNode{
					ObjectID:    "ABC123",
					Labels:      ingestedKinds,
					PropertyMap: ingestedPropertyBag.Map,
				})

			require.NoError(t, err)

			return nil
		})

		require.NoError(t, err)

		err = testSuite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			count, err := tx.Nodes().Filter(filter).Count()
			require.NoError(t, err)
			require.Equal(t, int64(1), count)

			tx.Nodes().Filter(filter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for node := range cursor.Chan() {
					objectid, _ := node.Properties.Get(common.ObjectID.String()).String()
					require.Equal(t, "ABC123", objectid)

					lastseen, _ := node.Properties.Get(common.LastSeen.String()).String()
					require.NotEmpty(t, lastseen)

					// kinds assert
					expectedKinds := ingestedKinds.Add(sourceKind)
					actualKinds := node.Kinds
					for _, actualKind := range actualKinds {
						require.Contains(t, expectedKinds, actualKind)
					}

					actualPropertyBag := node.Properties.Map
					for k, v := range expectedPropertyBag.Map {
						require.EqualValues(t, v, actualPropertyBag[k])
					}

				}
				return nil
			})
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("ingested node gets deduped with changelog on", func(t *testing.T) {
		var (
			ingestedPropertyBag = graph.NewProperties().SetAll(map[string]any{
				"hello": "world",
				"1":     2,
			})
			// clone here because ingestNode decorates the incoming propety bag with additional properties like lastseen
			expectedPropertyBag = ingestedPropertyBag.Clone()
			sourceKind          = graph.StringKind("hellobase")
			ingestedKinds       = graph.Kinds{graph.StringKind("labelA"), graph.StringKind("labelB")}
		)

		// open up a batchOp
		err := testSuite.GraphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			// inject changelog here to simulate it being toggled on
			ingestCtx := graphify.NewIngestContext(
				testSuite.Context,
				graphify.WithBatchUpdater(batch),
				graphify.WithChangeManager(testSuite.Changelog))

			// function under test. ingest same node twice
			node := ein.IngestibleNode{
				ObjectID:    "ABC123",
				Labels:      ingestedKinds,
				PropertyMap: ingestedPropertyBag.Map,
			}
			err := graphify.IngestNode(ingestCtx, sourceKind, node)
			require.NoError(t, err)
			err = graphify.IngestNode(ingestCtx, sourceKind, node)
			require.NoError(t, err)

			return nil
		})

		require.NoError(t, err)

		err = testSuite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			count, err := tx.Nodes().Filter(filter).Count()
			require.NoError(t, err)
			require.Equal(t, int64(1), count)

			tx.Nodes().Filter(filter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for node := range cursor.Chan() {
					objectid, _ := node.Properties.Get(common.ObjectID.String()).String()
					require.Equal(t, "ABC123", objectid)

					lastseen, _ := node.Properties.Get(common.LastSeen.String()).String()
					require.NotEmpty(t, lastseen)

					// kinds assert
					expectedKinds := ingestedKinds.Add(sourceKind)
					actualKinds := node.Kinds
					for _, actualKind := range actualKinds {
						require.Contains(t, expectedKinds, actualKind)
					}

					actualPropertyBag := node.Properties.Map
					for k, v := range expectedPropertyBag.Map {
						require.EqualValues(t, v, actualPropertyBag[k])
					}

				}
				return nil
			})
			return nil
		})

		require.NoError(t, err)
	})
}
