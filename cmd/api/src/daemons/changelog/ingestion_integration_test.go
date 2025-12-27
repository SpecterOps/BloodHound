// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package changelog

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context      context.Context
	GraphDB      graph.Database
	BloodhoundDB *database.BloodhoundDB
}

// enable changelog by getting existing flag and setting it
func (s *IntegrationTestSuite) enableChangelog(t *testing.T) {
	flag, err := s.BloodhoundDB.GetFlagByKey(s.Context, appcfg.FeatureChangelog)
	require.NoError(t, err)
	flag.Enabled = true
	require.NoError(t, s.BloodhoundDB.SetFlag(s.Context, flag))
}

// disable changelog by getting existing flag and setting it
func (s *IntegrationTestSuite) disableChangelog(t *testing.T) {
	flag, err := s.BloodhoundDB.GetFlagByKey(s.Context, appcfg.FeatureChangelog)
	require.NoError(t, err)
	flag.Enabled = false
	require.NoError(t, s.BloodhoundDB.SetFlag(s.Context, flag))
}

var (
	nodeKinds = graph.Kinds{
		graph.StringKind("NK1"),
		graph.StringKind("NK2"),
		graph.StringKind("NK3")}
	edgeKinds = graph.Kinds{graph.StringKind("EK2")}
)

func schema() graph.Schema {
	defaultGraph := graph.Graph{
		Name:  "default",
		Nodes: nodeKinds,
		Edges: edgeKinds,
		NodeConstraints: []graph.Constraint{{
			Field: "objectid",
			Type:  graph.BTreeIndex,
		}},
	}

	return graph.Schema{
		Graphs:       []graph.Graph{defaultGraph},
		DefaultGraph: defaultGraph,
	}
}

// setupIntegrationTest creates a real PostgreSQL database for testing
func setupIntegrationTest(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	// Create connection pool
	pool, err := pg.NewPool(connConf.URL())
	require.NoError(t, err)

	// Open graph database
	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	// Run migrations
	err = graphDB.AssertSchema(ctx, schema())
	require.NoError(t, err)

	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())
	require.NoError(t, db.Migrate(ctx))
	require.NoError(t, db.PopulateExtensionData(ctx))

	return IntegrationTestSuite{
		Context:      ctx,
		GraphDB:      graphDB,
		BloodhoundDB: db,
	}
}

// getPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	config, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(config.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func teardownIntegrationTest(t *testing.T, suite *IntegrationTestSuite) {
	t.Helper()

	if suite.GraphDB != nil {
		if err := suite.GraphDB.Close(suite.Context); err != nil {
			t.Logf("Failed to close GraphDB: %v", err)
		}
	}

	if suite.BloodhoundDB != nil {
		suite.BloodhoundDB.Close(suite.Context)
	}
}

func TestIngestionCoordinator(t *testing.T) {
	t.Parallel()

	t.Run("real NodeChange operations work end-to-end", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)

		var (
			coordinator   = newIngestionCoordinator(suite.GraphDB)
			ctx           = suite.Context
			batchSize     = 3
			flushInterval = 100 * time.Millisecond
			lastseen      = time.Now().UTC()
		)

		// Start the coordinator
		coordinator.start(ctx, batchSize, flushInterval)
		defer coordinator.stop(context.Background())

		// Create real NodeChange instances with required properties
		changes := []*NodeChange{
			NewNodeChange("node1", graph.Kinds{graph.StringKind("NK1")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "node1",
					"lastseen": lastseen,
				})),
			NewNodeChange("node2", graph.Kinds{graph.StringKind("NK2")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "node2",
					"lastseen": lastseen,
				})),
			NewNodeChange("node3", graph.Kinds{graph.StringKind("NK3")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "node3",
					"lastseen": lastseen,
				})),
		}

		// Submit changes
		for _, change := range changes {
			require.True(t, coordinator.submit(ctx, change))
		}

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Verify nodes were created in the database by querying for them
		var nodeCount int

		err := suite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {

			return tx.Nodes().
				Filter(
					query.Or(
						query.Equals(query.Property(query.Node(), "objectid"), "node1"),
						query.Equals(query.Property(query.Node(), "objectid"), "node2"),
						query.Equals(query.Property(query.Node(), "objectid"), "node3"),
					),
				).
				Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					for range cursor.Chan() {
						nodeCount++
					}

					return cursor.Error()
				})
		})
		require.NoError(t, err)
		require.Equal(t, len(changes), nodeCount)
	})

	t.Run("size-based batching creates nodes in database", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)
		var (
			coordinator   = newIngestionCoordinator(suite.GraphDB)
			ctx           = suite.Context
			batchSize     = 2
			flushInterval = 1 * time.Second // Long interval to test size-based flushing
		)

		coordinator.start(ctx, batchSize, flushInterval)
		defer coordinator.stop(context.Background())

		// Submit exactly batchSize changes
		changes := []*NodeChange{
			NewNodeChange("batch1", graph.Kinds{graph.StringKind("NK1")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "batch1",
					"lastseen": time.Now(),
					"batch":    "size-test",
				})),
			NewNodeChange("batch2", graph.Kinds{graph.StringKind("NK1")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "batch2",
					"lastseen": time.Now(),
					"batch":    "size-test",
				})),
		}

		for _, change := range changes {
			require.True(t, coordinator.submit(ctx, change))
		}

		// Wait for size-based flush
		time.Sleep(200 * time.Millisecond)

		// Verify nodes were created
		var nodeCount int
		err := suite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			return tx.Nodes().
				Filter(
					query.Or(query.Equals(query.NodeProperty("objectid"), "batch1"),
						query.Equals(query.NodeProperty("objectid"), "batch2"))).
				Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					for node := range cursor.Chan() {
						fmt.Println(node)
						nodeCount++
					}
					return cursor.Error()
				})
		})
		require.NoError(t, err)
		require.Equal(t, 2, nodeCount)
	})

	t.Run("idle flush creates nodes in database", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)

		var (
			coordinator   = newIngestionCoordinator(suite.GraphDB)
			ctx           = suite.Context
			batchSize     = 10 // Large batch size
			flushInterval = 50 * time.Millisecond
		)

		coordinator.start(ctx, batchSize, flushInterval)
		defer coordinator.stop(context.Background())

		// Submit single change (less than batch size)
		change := NewNodeChange("idle1", graph.Kinds{graph.StringKind("NK1")},
			graph.NewProperties().SetAll(map[string]any{
				"objectid": "idle1",
				"lastseen": time.Now(),
				"batch":    "idle-test",
			}))

		require.True(t, coordinator.submit(ctx, change))

		// Wait for idle flush
		time.Sleep(150 * time.Millisecond)

		// Verify node was created
		var nodeCount int
		err := suite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			return tx.Nodes().Filter(query.Equals(query.NodeProperty("objectid"), "idle1")).
				Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					for range cursor.Chan() {
						nodeCount++
					}
					return cursor.Error()
				})
		})
		require.NoError(t, err)
		require.Equal(t, 1, nodeCount)
	})

	t.Run("final flush on shutdown creates remaining nodes", func(t *testing.T) {
		suite := setupIntegrationTest(t)
		defer teardownIntegrationTest(t, &suite)

		var (
			coordinator   = newIngestionCoordinator(suite.GraphDB)
			ctx           = suite.Context
			batchSize     = 10              // Large batch size
			flushInterval = 1 * time.Second // Long interval
		)

		coordinator.start(ctx, batchSize, flushInterval)

		// Submit changes that won't trigger size-based flush
		changes := []*NodeChange{
			NewNodeChange("final1", graph.Kinds{graph.StringKind("NK1")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "final1",
					"lastseen": time.Now(),
					"batch":    "final-test",
				})),
			NewNodeChange("final2", graph.Kinds{graph.StringKind("NK1")},
				graph.NewProperties().SetAll(map[string]any{
					"objectid": "final2",
					"lastseen": time.Now(),
					"batch":    "final-test",
				})),
		}

		for _, change := range changes {
			require.True(t, coordinator.submit(ctx, change))
		}

		// Give changes time to be received
		time.Sleep(10 * time.Millisecond)

		// Stop should trigger final flush
		require.NoError(t, coordinator.stop(context.Background()))

		// Verify nodes were created by final flush
		var nodeCount int
		err := suite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			return tx.Nodes().Filter(query.Or(
				query.Equals(query.NodeProperty("objectid"), "final1"),
				query.Equals(query.NodeProperty("objectid"), "final2"),
			)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for range cursor.Chan() {
					nodeCount++
				}
				return cursor.Error()
			})
		})
		require.NoError(t, err)
		require.Equal(t, 2, nodeCount)
	})
}
