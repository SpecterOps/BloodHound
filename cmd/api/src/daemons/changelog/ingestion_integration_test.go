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
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
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

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	// Create connection pool
	pool, err := dbpool.NewPool(cfg.Database)
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

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
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

	type args struct {
		batchSize     int
		flushInterval time.Duration
		objectIDs     []string
	}

	tests := []struct {
		name   string
		args   args
		setup  func(t *testing.T, coordinator *ingestionCoordinator, ctx context.Context, objectIDs []string)
		assert func(t *testing.T, suite IntegrationTestSuite, objectIDs []string)
	}{
		{
			name: "submits changes and flushes nodes to database",
			args: args{
				batchSize:     3,
				flushInterval: 100 * time.Millisecond,
				objectIDs:     []string{"node1", "node2", "node3"},
			},
			setup: func(t *testing.T, coordinator *ingestionCoordinator, ctx context.Context, objectIDs []string) {
				t.Helper()
				for _, objectID := range objectIDs {
					change := NewNodeChange(objectID, graph.Kinds{graph.StringKind("NK1")},
						graph.NewProperties().SetAll(map[string]any{
							"objectid": objectID,
							"lastseen": time.Now().UTC(),
						}))
					require.True(t, coordinator.submit(ctx, change))
				}
				time.Sleep(200 * time.Millisecond)
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, objectIDs []string) {
				t.Helper()
				assertNodesExist(t, suite, objectIDs)
			},
		},
		{
			name: "flushes when buffer reaches batch size",
			args: args{
				batchSize:     2,
				flushInterval: 1 * time.Second,
				objectIDs:     []string{"batch1", "batch2"},
			},
			setup: func(t *testing.T, coordinator *ingestionCoordinator, ctx context.Context, objectIDs []string) {
				t.Helper()
				for _, objectID := range objectIDs {
					change := NewNodeChange(objectID, graph.Kinds{graph.StringKind("NK1")},
						graph.NewProperties().SetAll(map[string]any{
							"objectid": objectID,
							"lastseen": time.Now().UTC(),
						}))
					require.True(t, coordinator.submit(ctx, change))
				}
				time.Sleep(200 * time.Millisecond)
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, objectIDs []string) {
				t.Helper()
				assertNodesExist(t, suite, objectIDs)
			},
		},
		{
			name: "periodic timer flushes buffer below batch size",
			args: args{
				batchSize:     10,
				flushInterval: 50 * time.Millisecond,
				objectIDs:     []string{"idle1"},
			},
			setup: func(t *testing.T, coordinator *ingestionCoordinator, ctx context.Context, objectIDs []string) {
				t.Helper()
				change := NewNodeChange(objectIDs[0], graph.Kinds{graph.StringKind("NK1")},
					graph.NewProperties().SetAll(map[string]any{
						"objectid": objectIDs[0],
						"lastseen": time.Now().UTC(),
					}))
				require.True(t, coordinator.submit(ctx, change))
				time.Sleep(150 * time.Millisecond)
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, objectIDs []string) {
				t.Helper()
				assertNodesExist(t, suite, objectIDs)
			},
		},
		{
			name: "shutdown drains remaining buffer before exiting",
			args: args{
				batchSize:     10,
				flushInterval: 1 * time.Second,
				objectIDs:     []string{"final1", "final2"},
			},
			setup: func(t *testing.T, coordinator *ingestionCoordinator, ctx context.Context, objectIDs []string) {
				t.Helper()
				for _, objectID := range objectIDs {
					change := NewNodeChange(objectID, graph.Kinds{graph.StringKind("NK1")},
						graph.NewProperties().SetAll(map[string]any{
							"objectid": objectID,
							"lastseen": time.Now().UTC(),
						}))
					require.True(t, coordinator.submit(ctx, change))
				}
				time.Sleep(10 * time.Millisecond)
				require.NoError(t, coordinator.stop(context.Background()))
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, objectIDs []string) {
				t.Helper()
				assertNodesExist(t, suite, objectIDs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupIntegrationTest(t)
			defer teardownIntegrationTest(t, &suite)

			coordinator := newIngestionCoordinator(suite.GraphDB, DefaultRetryConfig())
			coordinator.start(suite.Context, tt.args.batchSize, tt.args.flushInterval)
			defer coordinator.stop(context.Background())

			tt.setup(t, coordinator, suite.Context, tt.args.objectIDs)
			tt.assert(t, suite, tt.args.objectIDs)
		})
	}
}

// assertNodesExist verifies that nodes with the given objectIDs exist in the graph database.
func assertNodesExist(t *testing.T, suite IntegrationTestSuite, objectIDs []string) {
	t.Helper()

	var nodeCount int
	filters := make([]graph.Criteria, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		filters = append(filters, query.Equals(query.NodeProperty("objectid"), objectID))
	}

	err := suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		return tx.Nodes().
			Filter(query.Or(filters...)).
			Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for range cursor.Chan() {
					nodeCount++
				}
				return cursor.Error()
			})
	})
	require.NoError(t, err)
	require.Equal(t, len(objectIDs), nodeCount)
}
