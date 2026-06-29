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

package appdb_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getPostgresConfig reads the integration test connection details from the
// configured environment and returns a pgtestdb.Config suitable for spinning
// up isolated databases. Supports both TCP and unix-socket host values.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for entry := range strings.FieldsSeq(cfg.Database.Connection) {
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
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
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

func setupStoreWithGraph(t *testing.T) (*appdb.Store, *pgxpool.Pool, graph.Database) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	dawgsPool, err := dbpool.NewDawgsPool(cfg.Database)
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bhDB := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bhDB.Migrate(ctx))
	require.NoError(t, bhDB.PopulateExtensionData(ctx))

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  dawgsPool,
	})
	require.NoError(t, err)

	require.NoError(t, migrations.NewGraphMigrator(graphDB).Migrate(ctx))
	require.NoError(t, graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema()))

	t.Cleanup(func() {
		graphDB.Close(ctx)
		bhDB.Close(ctx)
	})

	return appdb.NewStore(graphDB, dbPool), dbPool, graphDB
}

func TestStore_GetNode_Integration(t *testing.T) {
	type testCase struct {
		name       string
		setup      func(t *testing.T, graphDB graph.Database) graph.ID
		wantErr    error
		assertNode func(t *testing.T, node services.Node)
	}

	tests := []testCase{
		{
			name: "success_-_returns_node_with_kinds_and_properties",
			setup: func(t *testing.T, graphDB graph.Database) graph.ID {
				t.Helper()
				ctx := context.Background()

				var nodeID graph.ID
				err := graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
					props := graph.NewProperties()
					props.Set("name", "TestUser")
					props.Set("enabled", true)
					props.Set("objectid", "S-1-5-21-test")

					node, err := tx.CreateNode(
						props,
						graph.StringKind("User"),
						graph.StringKind("Base"),
					)
					if err != nil {
						return err
					}
					nodeID = node.ID
					return nil
				})
				require.NoError(t, err)
				return nodeID
			},
			assertNode: func(t *testing.T, node services.Node) {
				t.Helper()
				assert.Len(t, node.Kinds, 2)
				assert.Equal(t, "TestUser", node.Properties["name"])
				assert.Equal(t, true, node.Properties["enabled"])
				assert.Equal(t, "S-1-5-21-test", node.Properties["objectid"])
			},
		},
		{
			name: "error_-_returns_ErrNodeNotFound_for_nonexistent_node",
			setup: func(t *testing.T, graphDB graph.Database) graph.ID {
				t.Helper()
				return graph.ID(999999999)
			},
			wantErr: services.ErrNodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				store, _, graphDB = setupStoreWithGraph(t)
				ctx               = context.Background()
				nodeID            = tt.setup(t, graphDB)
			)

			node, err := store.GetNode(ctx, int64(nodeID))

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, int64(nodeID), node.ID)
				if tt.assertNode != nil {
					tt.assertNode(t, node)
				}
			}
		})
	}
}

func TestStore_GetNodeKindsByNames_Integration(t *testing.T) {
	type testSetupData struct {
		kindsCreated bool
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, pool *pgxpool.Pool) testSetupData
		assert   func(t *testing.T, store *appdb.Store, setupData testSetupData)
		teardown func(t *testing.T, pool *pgxpool.Pool, setupData testSetupData)
	}

	tests := []testCase{
		{
			name: "success_-_returns_registered_kinds_with_ids",
			setup: func(t *testing.T, pool *pgxpool.Pool) testSetupData {
				t.Helper()
				ctx := context.Background()

				// Create a test extension
				var extensionID int
				err := pool.QueryRow(ctx, `
					INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
					VALUES ('TestExtension', 'Test Extension', '1.0.0', false, 'TST')
					RETURNING id
				`).Scan(&extensionID)
				require.NoError(t, err)

				// Insert test kinds
				_, err = pool.Exec(ctx, `
					INSERT INTO kind (name)
					VALUES ('TestRegisteredKind1'), ('TestRegisteredKind2')
					ON CONFLICT (name) DO NOTHING
				`)
				require.NoError(t, err)

				// Register them in schema_node_kinds with the extension_id
				_, err = pool.Exec(ctx, `
					INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
					SELECT $1, id, name, '', false, '', ''
					FROM kind
					WHERE name IN ('TestRegisteredKind1', 'TestRegisteredKind2')
					ON CONFLICT (kind_id) DO NOTHING
				`, extensionID)
				require.NoError(t, err)

				return testSetupData{kindsCreated: true}
			},
			assert: func(t *testing.T, store *appdb.Store, setupData testSetupData) {
				t.Helper()
				ctx := context.Background()

				kinds, err := store.GetNodeKindsByNames(ctx, []string{"TestRegisteredKind1", "TestRegisteredKind2"})

				require.NoError(t, err)
				require.Len(t, kinds, 2)

				for _, kind := range kinds {
					require.NotNil(t, kind.ID, "kind %s should have non-nil ID", kind.Name)
					assert.NotEqual(t, int32(0), *kind.ID, "kind %s should have non-zero ID", kind.Name)
					assert.Contains(t, []string{"TestRegisteredKind1", "TestRegisteredKind2"}, kind.Name)
				}
			},
			teardown: func(t *testing.T, pool *pgxpool.Pool, setupData testSetupData) {
				t.Helper()
				if setupData.kindsCreated {
					ctx := context.Background()
					// Delete extension (cascades to schema_node_kinds)
					_, _ = pool.Exec(ctx, `DELETE FROM schema_extensions WHERE name = 'TestExtension'`)
					_, _ = pool.Exec(ctx, `DELETE FROM kind WHERE name IN ('TestRegisteredKind1', 'TestRegisteredKind2')`)
				}
			},
		},
		{
			name: "success_-_returns_unregistered_kinds_with_zero_id",
			setup: func(t *testing.T, pool *pgxpool.Pool) testSetupData {
				t.Helper()
				ctx := context.Background()

				// Create a test extension
				var extensionID int
				err := pool.QueryRow(ctx, `
					INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
					VALUES ('TestExtensionMixed', 'Test Extension Mixed', '1.0.0', false, 'TSM')
					RETURNING id
				`).Scan(&extensionID)
				require.NoError(t, err)

				// Insert one registered kind
				_, err = pool.Exec(ctx, `
					INSERT INTO kind (name) VALUES ('TestRegisteredKindMixed')
					ON CONFLICT (name) DO NOTHING
				`)
				require.NoError(t, err)

				// Register it in schema_node_kinds with extension_id
				_, err = pool.Exec(ctx, `
					INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
					SELECT $1, id, name, '', false, '', ''
					FROM kind
					WHERE name = 'TestRegisteredKindMixed'
					ON CONFLICT (kind_id) DO NOTHING
				`, extensionID)
				require.NoError(t, err)

				return testSetupData{kindsCreated: true}
			},
			assert: func(t *testing.T, store *appdb.Store, setupData testSetupData) {
				t.Helper()
				ctx := context.Background()

				// Request one registered and one unregistered kind
				kinds, err := store.GetNodeKindsByNames(ctx, []string{"TestRegisteredKindMixed", "TestUnregisteredKind"})

				require.NoError(t, err)
				require.Len(t, kinds, 2)

				var registeredKind, unregisteredKind services.Kind
				for _, kind := range kinds {
					if kind.Name == "TestRegisteredKindMixed" {
						registeredKind = kind
					} else if kind.Name == "TestUnregisteredKind" {
						unregisteredKind = kind
					}
				}

				require.NotNil(t, registeredKind.ID, "registered kind should have non-nil ID")
				assert.NotEqual(t, int32(0), *registeredKind.ID, "registered kind should have non-zero ID")
				assert.Equal(t, "TestRegisteredKindMixed", registeredKind.Name)
				assert.Nil(t, unregisteredKind.ID, "unregistered kind should have nil ID")
				assert.Equal(t, "TestUnregisteredKind", unregisteredKind.Name)
			},
			teardown: func(t *testing.T, pool *pgxpool.Pool, setupData testSetupData) {
				t.Helper()
				if setupData.kindsCreated {
					ctx := context.Background()
					// Delete extension (cascades to schema_node_kinds)
					_, _ = pool.Exec(ctx, `DELETE FROM schema_extensions WHERE name = 'TestExtensionMixed'`)
					_, _ = pool.Exec(ctx, `DELETE FROM kind WHERE name = 'TestRegisteredKindMixed'`)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			store, pool, _ := setupStoreWithGraph(t)

			setupData := testCase.setup(t, pool)
			defer testCase.teardown(t, pool, setupData)

			testCase.assert(t, store, setupData)
		})
	}
}
