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
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupStoreAndPool spins up an isolated postgres database via pgtestdb, applies all
// migrations, and returns an identity Store backed by the resulting pgx pool along
// with the pool itself so callers can read seeded rows directly.
func setupStoreAndPool(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
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

	t.Cleanup(func() { bhDB.Close(ctx) })

	return appdb.NewStore(bhDB.Pool()), bhDB.Pool()
}

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

// seededPermission reads a single permission seeded by the migrations directly
// from the pool so tests have a known-good id to look up via the Store.
func seededPermission(t *testing.T, ctx context.Context, pool *pgxpool.Pool) services.Permission {
	t.Helper()

	var found services.Permission
	err := pool.QueryRow(ctx, "SELECT id, authority, name FROM permissions ORDER BY id LIMIT 1").
		Scan(&found.ID, &found.Authority, &found.Name)
	require.NoError(t, err)

	return found
}

// seededRole reads a single role seeded by the migrations directly from the pool.
func seededRole(t *testing.T, ctx context.Context, pool *pgxpool.Pool) services.Role {
	t.Helper()

	var found services.Role
	err := pool.QueryRow(ctx, "SELECT id, name FROM roles ORDER BY id LIMIT 1").
		Scan(&found.ID, &found.Name)
	require.NoError(t, err)

	return found
}

func TestStore_GetPermission_Integration(t *testing.T) {
	t.Run("returns the permission for a seeded id", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		expected := seededPermission(t, ctx, pool)

		retrieved, err := store.GetPermission(ctx, int(expected.ID))
		require.NoError(t, err)
		assert.Equal(t, expected.ID, retrieved.ID)
		assert.Equal(t, expected.Authority, retrieved.Authority)
		assert.Equal(t, expected.Name, retrieved.Name)
	})

	t.Run("returns ErrNoPermissionFound when the permission does not exist", func(t *testing.T) {
		var (
			ctx      = context.Background()
			store, _ = setupStoreAndPool(t)
		)

		_, err := store.GetPermission(ctx, 99999999)
		assert.ErrorIs(t, err, services.ErrNoPermissionFound)
	})
}

func TestStore_GetRole_Integration(t *testing.T) {
	t.Run("returns the role with its permissions for a seeded id", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		expected := seededRole(t, ctx, pool)

		retrieved, err := store.GetRole(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected.ID, retrieved.ID)
		assert.Equal(t, expected.Name, retrieved.Name)
		assert.NotEmpty(t, retrieved.Permissions, "expected the seeded role to have at least one permission")
	})

	t.Run("returns ErrNoRoleFound when the role does not exist", func(t *testing.T) {
		var (
			ctx      = context.Background()
			store, _ = setupStoreAndPool(t)
		)

		_, err := store.GetRole(ctx, 99999999)
		assert.ErrorIs(t, err, services.ErrNoRoleFound)
	})
}

// seededRoleCount reads the number of roles seeded by the migrations directly
// from the pool so tests can assert against the full set.
func seededRoleCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool) int {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, "SELECT count(*) FROM roles").Scan(&count)
	require.NoError(t, err)

	return count
}

func seededPermissionCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool) int {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, "SELECT count(*) FROM permissions").Scan(&count)
	require.NoError(t, err)

	return count
}

func TestStore_ListRoles_Integration(t *testing.T) {
	t.Run("returns every seeded role with its permissions", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		expectedCount := seededRoleCount(t, ctx, pool)
		require.NotZero(t, expectedCount, "expected migrations to seed at least one role")

		roles, err := store.ListRoles(ctx, params.Filters{}, params.SortItems{})
		require.NoError(t, err)
		assert.Len(t, roles, expectedCount)

		var withPermissions int
		for _, r := range roles {
			if len(r.Permissions) > 0 {
				withPermissions++
			}
		}
		assert.NotZero(t, withPermissions, "expected at least one role to preload its permissions")
	})

	t.Run("returns roles sorted by name ascending", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		expectedCount := seededRoleCount(t, ctx, pool)

		roles, err := store.ListRoles(ctx, params.Filters{}, params.SortItems{{Field: "name", Direction: params.Ascending}})
		require.NoError(t, err)
		require.Len(t, roles, expectedCount)

		for i := 1; i < len(roles); i++ {
			assert.LessOrEqual(t, roles[i-1].Name, roles[i].Name, "roles should be sorted by name ascending")
		}
	})

	t.Run("returns roles filtered by name", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		target := seededRole(t, ctx, pool)

		roles, err := store.ListRoles(ctx, params.Filters{
			"name": {{Field: "name", Operator: params.Equals, Value: target.Name, SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, target.Name, roles[0].Name)
	})

	t.Run("returns an empty slice when no role matches the filter", func(t *testing.T) {
		var (
			ctx      = context.Background()
			store, _ = setupStoreAndPool(t)
		)

		roles, err := store.ListRoles(ctx, params.Filters{
			"name": {{Field: "name", Operator: params.Equals, Value: "does-not-exist", SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("returns roles filtered by created_at date comparison", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
			epoch       = "2000-01-01T00:00:00Z"
		)

		expectedCount := seededRoleCount(t, ctx, pool)
		require.NotZero(t, expectedCount, "expected migrations to seed at least one role")

		afterEpoch, err := store.ListRoles(ctx, params.Filters{
			"created_at": {{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		assert.Len(t, afterEpoch, expectedCount, "every seeded role was created after the epoch")

		beforeEpoch, err := store.ListRoles(ctx, params.Filters{
			"created_at": {{Field: "created_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		assert.Empty(t, beforeEpoch, "no seeded role was created before the epoch")
	})

	t.Run("returns roles filtered by updated_at date comparison", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
			epoch       = "2000-01-01T00:00:00Z"
		)

		expectedCount := seededRoleCount(t, ctx, pool)
		require.NotZero(t, expectedCount, "expected migrations to seed at least one role")

		afterEpoch, err := store.ListRoles(ctx, params.Filters{
			"updated_at": {{Field: "updated_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		assert.Len(t, afterEpoch, expectedCount, "every seeded role was updated after the epoch")

		beforeEpoch, err := store.ListRoles(ctx, params.Filters{
			"updated_at": {{Field: "updated_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		assert.Empty(t, beforeEpoch, "no seeded role was updated before the epoch")
	})
}

func TestStore_ListPermissions_Integration(t *testing.T) {
	type expected struct {
		assertResponse func(t *testing.T, permissions []services.Permission, pool *pgxpool.Pool, ctx context.Context)
	}

	type testData struct {
		name       string
		buildQuery func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) (params.Filters, params.SortItems)
		expected   expected
	}

	const epoch = "2000-01-01T00:00:00Z"

	tests := []testData{
		{
			name: "Success: all seeded permissions are returned",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, pool *pgxpool.Pool, ctx context.Context) {
				expectedCount := seededPermissionCount(t, ctx, pool)
				require.NotZero(t, expectedCount, "expected migrations to seed at least one permission")
				assert.Len(t, permissions, expectedCount)
			}},
		},
		{
			name: "Success: permissions are sorted by name",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{}, params.SortItems{{Field: "name", Direction: params.Ascending}}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, pool *pgxpool.Pool, ctx context.Context) {
				expectedCount := seededPermissionCount(t, ctx, pool)
				require.Len(t, permissions, expectedCount)
				for i := 1; i < len(permissions); i++ {
					assert.LessOrEqual(t, permissions[i-1].Name, permissions[i].Name, "permissions should be sorted by name ascending")
				}
			}},
		},
		{
			name: "Success: permissions are filtered by authority",
			buildQuery: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) (params.Filters, params.SortItems) {
				target := seededPermission(t, ctx, pool)
				return params.Filters{"authority": {{Field: "authority", Operator: params.Equals, Value: target.Authority, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, pool *pgxpool.Pool, ctx context.Context) {
				target := seededPermission(t, ctx, pool)
				require.NotEmpty(t, permissions)
				for _, permission := range permissions {
					assert.Equal(t, target.Authority, permission.Authority)
				}
			}},
		},
		{
			name: "Success: no permissions match",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "does-not-exist", SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, _ *pgxpool.Pool, _ context.Context) {
				assert.Empty(t, permissions)
			}},
		},
		{
			name: "Success: created_at supports date comparisons after the epoch",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{"created_at": {{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, pool *pgxpool.Pool, ctx context.Context) {
				assert.Len(t, permissions, seededPermissionCount(t, ctx, pool))
			}},
		},
		{
			name: "Success: created_at supports date comparisons before the epoch",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{"created_at": {{Field: "created_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, _ *pgxpool.Pool, _ context.Context) {
				assert.Empty(t, permissions)
			}},
		},
		{
			name: "Success: updated_at supports date comparisons after the epoch",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{"updated_at": {{Field: "updated_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, pool *pgxpool.Pool, ctx context.Context) {
				assert.Len(t, permissions, seededPermissionCount(t, ctx, pool))
			}},
		},
		{
			name: "Success: updated_at supports date comparisons before the epoch",
			buildQuery: func(_ *testing.T, _ *pgxpool.Pool, _ context.Context) (params.Filters, params.SortItems) {
				return params.Filters{"updated_at": {{Field: "updated_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, permissions []services.Permission, _ *pgxpool.Pool, _ context.Context) {
				assert.Empty(t, permissions)
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
			)

			filters, sortItems := tt.buildQuery(t, pool, ctx)
			permissions, err := store.ListPermissions(ctx, filters, sortItems)
			require.NoError(t, err)
			tt.expected.assertResponse(t, permissions, pool, ctx)
		})
	}
}

// addDeprecatedDeletedAtColumn simulates the shape of an upgraded/GORM-era database by
// adding a deleted_at column that the current migrations do not create and that the
// Go structs do not model. Reading from such a table with SELECT * would return an
// extra column the strict pgx.RowToStructByName mapper cannot place, reproducing the
// original "struct doesn't have corresponding row field deleted_at" failure.
func addDeprecatedDeletedAtColumn(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()

	_, err := pool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN deleted_at timestamp with time zone", table))
	require.NoError(t, err)
}

// TestStore_SchemaDrift_DeletedAt_Integration is a regression test for the case where
// the physical roles/permissions tables carry a stray deleted_at column (as on
// databases upgraded from the GORM era) that the Go structs do not model. The reads
// must remain resilient to that schema drift by selecting explicit columns rather than
// SELECT *. This test fails against the old SELECT * queries and passes against the
// explicit-column queries.
func TestStore_SchemaDrift_DeletedAt_Integration(t *testing.T) {
	type testData struct {
		name       string
		table      string
		assertRead func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool)
	}

	tests := []testData{
		{
			name:  "Success: GetRole handles a stray deleted_at column",
			table: "roles",
			assertRead: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) {
				expected := seededRole(t, ctx, pool)
				retrieved, err := store.GetRole(ctx, expected.ID)
				require.NoError(t, err)
				assert.Equal(t, expected.ID, retrieved.ID)
				assert.Equal(t, expected.Name, retrieved.Name)
			},
		},
		{
			name:  "Success: ListRoles handles a stray deleted_at column",
			table: "roles",
			assertRead: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) {
				expectedCount := seededRoleCount(t, ctx, pool)
				require.NotZero(t, expectedCount, "expected migrations to seed at least one role")
				roles, err := store.ListRoles(ctx, params.Filters{}, params.SortItems{})
				require.NoError(t, err)
				assert.Len(t, roles, expectedCount)
			},
		},
		{
			name:  "Success: GetPermission handles a stray deleted_at column",
			table: "permissions",
			assertRead: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) {
				expected := seededPermission(t, ctx, pool)
				retrieved, err := store.GetPermission(ctx, int(expected.ID))
				require.NoError(t, err)
				assert.Equal(t, expected.ID, retrieved.ID)
				assert.Equal(t, expected.Authority, retrieved.Authority)
				assert.Equal(t, expected.Name, retrieved.Name)
			},
		},
		{
			name:  "Success: ListPermissions handles a stray deleted_at column",
			table: "permissions",
			assertRead: func(t *testing.T, ctx context.Context, store *appdb.Store, pool *pgxpool.Pool) {
				expectedCount := seededPermissionCount(t, ctx, pool)
				permissions, err := store.ListPermissions(ctx, params.Filters{}, params.SortItems{})
				require.NoError(t, err)
				assert.Len(t, permissions, expectedCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
			)

			addDeprecatedDeletedAtColumn(t, ctx, pool, tt.table)
			tt.assertRead(t, ctx, store, pool)
		})
	}
}
