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
}
