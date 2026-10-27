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

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/extensions/internal/appdb"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
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

// setupStore spins up an isolated postgres database via pgtestdb, applies the
// relational migrations and populates the built-in extension data, then returns a
// Store backed by the resulting pgx pool.
func setupStore(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bhDB := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bhDB.Migrate(ctx))
	require.NoError(t, bhDB.PopulateExtensionData(ctx))
	t.Cleanup(func() { bhDB.Close(ctx) })

	return appdb.NewStore(dbPool), dbPool
}

func TestStore_GetNodeKind_Integration(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		wantErr error
		assert  func(t *testing.T, nodeKind services.NodeKind, data testSetupData)
	}

	tests := []testCase{
		{
			name:  "success_-_returns_node_kind_fields",
			setup: seedNodeKind,
			assert: func(t *testing.T, nodeKind services.NodeKind, data testSetupData) {
				assert.Equal(t, data.nodeKindID, nodeKind.ID)
				assert.Equal(t, data.kindID, nodeKind.KindID)
				assert.Equal(t, "TestNodeKind", nodeKind.Name)
				assert.Equal(t, "Test Node Kind", nodeKind.DisplayName)
				assert.Equal(t, "a test node kind", nodeKind.Description)
				assert.True(t, nodeKind.IsDisplayKind)
				assert.Equal(t, "user", nodeKind.Icon)
				assert.Equal(t, "#fff", nodeKind.Color)
			},
		},
		{
			name: "error_-_returns_ErrNodeKindNotFound",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				return testSetupData{nodeKindID: int32(999999999)}
			},
			wantErr: services.ErrNodeKindNotFound,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				ctx         = context.Background()
				data        = testCase.setup(t, ctx, pool)
			)

			nodeKind, err := store.GetNodeKind(ctx, data.nodeKindID)
			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}
			require.NoError(t, err)
			testCase.assert(t, nodeKind, data)
		})
	}
}

func TestStore_CreateGraphSchemaNodeKind_Integration(t *testing.T) {
	type testSetupData struct {
		extensionID int32
		input       services.NodeInput
		wantErr     error
	}

	type testCase struct {
		name     string
		setup    func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		assert   func(t *testing.T, data testSetupData, created services.NodeKind, err error)
		teardown func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData)
	}

	var (
		ctx   = context.Background()
		input = services.NodeInput{
			Name:          "Test_Kind_1",
			DisplayName:   "Test_Kind_1",
			Description:   "A test kind",
			IsDisplayKind: false,
			Icon:          "test_icon",
			IconColor:     "blue",
		}
		teardown = func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData) {
			t.Helper()
			_, err := pool.Exec(ctx, `DELETE FROM kind WHERE name = $1`, data.input.Name)
			require.NoError(t, err)
		}
	)

	tests := []testCase{
		{
			name: "success_-_creates_node_kind",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				t.Helper()
				seeded := seedNodeKind(t, ctx, pool)
				return testSetupData{extensionID: seeded.extensionID, input: input}
			},
			assert: func(t *testing.T, data testSetupData, created services.NodeKind, err error) {
				t.Helper()
				assert.NoError(t, err)
				assert.Positive(t, created.ID)
				assert.Positive(t, created.KindID)
				assert.Equal(t, data.extensionID, created.SchemaExtensionID)
				assert.Equal(t, data.input.Name, created.Name)
				assert.Equal(t, data.input.DisplayName, created.DisplayName)
				assert.Equal(t, data.input.Description, created.Description)
				assert.Equal(t, data.input.IsDisplayKind, created.IsDisplayKind)
				assert.Equal(t, data.input.Icon, created.Icon)
				assert.Equal(t, data.input.IconColor, created.Color)
				assert.False(t, created.CreatedAt.IsZero())
				assert.False(t, created.UpdatedAt.IsZero())
				assert.Nil(t, created.DeletedAt)
			},
			teardown: teardown,
		},
		{
			name: "error_-_duplicate_node_kind_name",
			setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
				t.Helper()
				seeded := seedNodeKind(t, ctx, pool)
				duplicateInput := input
				duplicateInput.Name = "TestNodeKind"
				return testSetupData{
					extensionID: seeded.extensionID,
					input:       duplicateInput,
					wantErr:     services.ErrDuplicateSchemaNodeKindName,
				}
			},
			assert: func(t *testing.T, data testSetupData, created services.NodeKind, err error) {
				t.Helper()
				assert.ErrorIs(t, err, data.wantErr)
				assert.Equal(t, services.NodeKind{}, created)
			},
			teardown: teardown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, pool)
			)

			t.Cleanup(func() {
				test.teardown(t, ctx, pool, data)
			})

			created, err := store.CreateGraphSchemaNodeKind(ctx, data.extensionID, data.input)
			test.assert(t, data, created, err)
		})
	}
}

func TestStore_GetGraphSchemaNodeKinds_Integration(t *testing.T) {
	type testSetupData struct {
		filters      params.Filters
		sort         params.SortItems
		skip         int
		limit        int
		wantNodeKind services.NodeKind
		wantCount    int
	}

	type testCase struct {
		name   string
		setup  func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		assert func(t *testing.T, data testSetupData, nodeKinds []services.NodeKind, count int, err error)
	}

	var (
		ctx   = context.Background()
		tests = []testCase{
			{
				name: "success_-_filters_sorts_and_paginates_node_kinds",
				setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seeded := seedNodeKind(t, ctx, pool)
					_, err := pool.Exec(ctx, `
						WITH inserted_kind AS (INSERT INTO kind (name) VALUES ('TestNodeKind_Z') RETURNING id)
						INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
						SELECT $1, id, 'Test Node Kind Z', 'a test node kind Z', true, 'user', '#fff' FROM inserted_kind`, seeded.extensionID)
					require.NoError(t, err)
					t.Cleanup(func() {
						_, err := pool.Exec(ctx, `DELETE FROM kind WHERE name = $1`, "TestNodeKind_Z")
						require.NoError(t, err)
					})
					return testSetupData{
						filters: params.Filters{
							"schema_extension_id": {
								{Operator: params.Equals, Value: fmt.Sprint(seeded.extensionID)},
							},
						},
						sort: params.SortItems{
							{Field: "name", Direction: params.Descending},
						},
						skip:         1,
						limit:        1,
						wantNodeKind: services.NodeKind{ID: seeded.nodeKindID, Name: "TestNodeKind"},
						wantCount:    2,
					}
				},
				assert: func(t *testing.T, data testSetupData, nodeKinds []services.NodeKind, count int, err error) {
					t.Helper()
					assert.NoError(t, err)
					require.Len(t, nodeKinds, 1)
					assert.Equal(t, data.wantNodeKind.ID, nodeKinds[0].ID)
					assert.Equal(t, data.wantNodeKind.Name, nodeKinds[0].Name)
					assert.Equal(t, data.wantCount, count)
				},
			},
			{
				name: "error_-_invalid_node_kind_id",
				setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					return testSetupData{
						filters: params.Filters{
							"id": {
								{Operator: params.Equals, Value: "invalid"},
							},
						},
					}
				},
				assert: func(t *testing.T, data testSetupData, nodeKinds []services.NodeKind, count int, err error) {
					var postgresError *pgconn.PgError

					t.Helper()
					if assert.ErrorAs(t, err, &postgresError) {
						assert.Equal(t, "22P02", postgresError.Code)
					}
					assert.Nil(t, nodeKinds)
					assert.Zero(t, count)
				},
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, pool)
			)

			nodeKinds, count, err := store.GetGraphSchemaNodeKinds(
				ctx, data.filters, data.sort, data.skip, data.limit,
			)
			test.assert(t, data, nodeKinds, count, err)
		})
	}
}

func TestStore_UpdateGraphSchemaNodeKind_Integration(t *testing.T) {
	type testSetupData struct {
		nodeKind     services.NodeKind
		wantNodeKind services.NodeKind
		wantErr      error
	}

	type testCase struct {
		name   string
		setup  func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData
		assert func(t *testing.T, data testSetupData, updated services.NodeKind, err error)
	}

	var (
		ctx   = context.Background()
		tests = []testCase{
			{
				name: "error_-_returns_node_kind_not_found",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					return testSetupData{
						nodeKind:     services.NodeKind{ID: -1},
						wantNodeKind: services.NodeKind{},
						wantErr:      services.ErrNodeKindNotFound,
					}
				},
				assert: func(t *testing.T, data testSetupData, updated services.NodeKind, err error) {
					t.Helper()
					assert.Equal(t, data.wantNodeKind, updated)
					assert.ErrorIs(t, err, data.wantErr)
				},
			},
			{
				name: "success_-_updates_node_kind_fields",
				setup: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seeded := seedNodeKind(t, ctx, pool)
					nodeKind, err := store.GetNodeKind(ctx, seeded.nodeKindID)
					require.NoError(t, err)
					nodeKind.DisplayName = "Updated Node Kind"
					nodeKind.Description = "updated description"
					nodeKind.IsDisplayKind = false
					nodeKind.Icon = "updated_icon"
					nodeKind.Color = "blue"
					return testSetupData{
						nodeKind:     nodeKind,
						wantNodeKind: nodeKind,
					}
				},
				assert: func(t *testing.T, data testSetupData, updated services.NodeKind, err error) {
					t.Helper()
					assert.NoError(t, err)
					assert.Equal(t, data.wantNodeKind.ID, updated.ID)
					assert.Equal(t, data.wantNodeKind.SchemaExtensionID, updated.SchemaExtensionID)
					assert.Equal(t, data.wantNodeKind.KindID, updated.KindID)
					assert.Equal(t, data.wantNodeKind.Name, updated.Name)
					assert.Equal(t, data.wantNodeKind.DisplayName, updated.DisplayName)
					assert.Equal(t, data.wantNodeKind.Description, updated.Description)
					assert.Equal(t, data.wantNodeKind.IsDisplayKind, updated.IsDisplayKind)
					assert.Equal(t, data.wantNodeKind.Icon, updated.Icon)
					assert.Equal(t, data.wantNodeKind.Color, updated.Color)
					assert.Equal(t, data.wantNodeKind.CreatedAt, updated.CreatedAt)
					assert.False(t, updated.UpdatedAt.IsZero())
					assert.Equal(t, data.wantNodeKind.DeletedAt, updated.DeletedAt)
				},
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, store, pool)
			)

			updated, err := store.UpdateGraphSchemaNodeKind(ctx, data.nodeKind)
			test.assert(t, data, updated, err)
		})
	}
}

func TestStore_DeleteGraphSchemaNodeKind_Integration(t *testing.T) {
	type testSetupData struct {
		nodeKindID       int32
		storedNodeKindID int32
		wantCount        int
		wantErr          error
	}

	type testCase struct {
		name   string
		setup  func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData
		assert func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, err error)
	}

	var (
		ctx   = context.Background()
		tests = []testCase{
			{
				name: "success_-_deletes_node_kind",
				setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seeded := seedNodeKind(t, ctx, pool)
					return testSetupData{
						nodeKindID:       seeded.nodeKindID,
						storedNodeKindID: seeded.nodeKindID,
					}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, err error) {
					var rowCount int

					t.Helper()
					assert.ErrorIs(t, err, data.wantErr)
					require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_node_kinds WHERE id = $1`, data.storedNodeKindID).Scan(&rowCount))
					assert.Equal(t, data.wantCount, rowCount)
				},
			},
			{
				name: "error_-_node_kind_not_found",
				setup: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
					t.Helper()
					seeded := seedNodeKind(t, ctx, pool)
					return testSetupData{
						nodeKindID:       -1,
						storedNodeKindID: seeded.nodeKindID,
						wantCount:        1,
						wantErr:          services.ErrNodeKindNotFound,
					}
				},
				assert: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool, data testSetupData, err error) {
					var rowCount int

					t.Helper()
					assert.ErrorIs(t, err, data.wantErr)
					require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_node_kinds WHERE id = $1`, data.storedNodeKindID).Scan(&rowCount))
					assert.Equal(t, data.wantCount, rowCount)
				},
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				store, pool = setupStore(t)
				data        = test.setup(t, ctx, pool)
			)

			err := store.DeleteGraphSchemaNodeKind(ctx, data.nodeKindID)
			test.assert(t, ctx, pool, data, err)
		})
	}
}
