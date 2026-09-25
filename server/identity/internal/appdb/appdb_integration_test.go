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

	"github.com/gofrs/uuid"
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
	type mock struct {
		store *appdb.Store
		pool  *pgxpool.Pool
	}

	type expected struct {
		assertResponse func(t *testing.T, mock mock, permission services.Permission, err error)
	}

	type testData struct {
		name              string
		buildPermissionID func(t *testing.T, mock mock) int
		expected          expected
	}

	tt := []testData{
		{name: "Success: seeded permission is returned", buildPermissionID: func(t *testing.T, mock mock) int {
			return int(seededPermission(t, context.Background(), mock.pool).ID)
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, retrieved services.Permission, err error) {
			expectedPermission := seededPermission(t, context.Background(), mock.pool)
			require.NoError(t, err)
			assert.Equal(t, expectedPermission.ID, retrieved.ID)
			assert.Equal(t, expectedPermission.Authority, retrieved.Authority)
			assert.Equal(t, expectedPermission.Name, retrieved.Name)
		}}},
		{name: "Error: permission does not exist", buildPermissionID: func(_ *testing.T, _ mock) int { return 99999999 }, expected: expected{assertResponse: func(t *testing.T, _ mock, _ services.Permission, err error) {
			assert.ErrorIs(t, err, services.ErrNoPermissionFound)
		}}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
				mock        = mock{store: store, pool: pool}
			)

			permission, err := mock.store.GetPermission(ctx, testCase.buildPermissionID(t, mock))
			testCase.expected.assertResponse(t, mock, permission, err)
		})
	}
}

func TestStore_GetRole_Integration(t *testing.T) {
	type mock struct {
		store *appdb.Store
		pool  *pgxpool.Pool
	}

	type expected struct {
		assertResponse func(t *testing.T, mock mock, role services.Role, err error)
	}

	type testData struct {
		name        string
		buildRoleID func(t *testing.T, mock mock) int32
		expected    expected
	}

	tt := []testData{
		{name: "Success: seeded role with permissions is returned", buildRoleID: func(t *testing.T, mock mock) int32 {
			return seededRole(t, context.Background(), mock.pool).ID
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, retrieved services.Role, err error) {
			expectedRole := seededRole(t, context.Background(), mock.pool)
			require.NoError(t, err)
			assert.Equal(t, expectedRole.ID, retrieved.ID)
			assert.Equal(t, expectedRole.Name, retrieved.Name)
			assert.NotEmpty(t, retrieved.Permissions, "expected the seeded role to have at least one permission")
		}}},
		{name: "Error: role does not exist", buildRoleID: func(_ *testing.T, _ mock) int32 { return 99999999 }, expected: expected{assertResponse: func(t *testing.T, _ mock, _ services.Role, err error) {
			assert.ErrorIs(t, err, services.ErrNoRoleFound)
		}}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
				mock        = mock{store: store, pool: pool}
			)

			role, err := mock.store.GetRole(ctx, testCase.buildRoleID(t, mock))
			testCase.expected.assertResponse(t, mock, role, err)
		})
	}
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
	type mock struct {
		store *appdb.Store
		pool  *pgxpool.Pool
	}

	type expected struct {
		assertResponse func(t *testing.T, mock mock, roles []services.Role, err error)
	}

	type testData struct {
		name       string
		buildQuery func(t *testing.T, mock mock) (params.Filters, params.SortItems)
		expected   expected
	}

	const epoch = "2000-01-01T00:00:00Z"

	tt := []testData{
		{name: "Success: every seeded role with permissions is returned", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, roles []services.Role, err error) {
			expectedCount := seededRoleCount(t, context.Background(), mock.pool)
			require.NotZero(t, expectedCount, "expected migrations to seed at least one role")
			require.NoError(t, err)
			assert.Len(t, roles, expectedCount)
			for _, role := range roles {
				if len(role.Permissions) > 0 {
					return
				}
			}
			t.Fatal("expected at least one role to preload its permissions")
		}}},
		{name: "Success: roles are sorted by name", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{}, params.SortItems{{Field: "name", Direction: params.Ascending}}
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, roles []services.Role, err error) {
			require.NoError(t, err)
			require.Len(t, roles, seededRoleCount(t, context.Background(), mock.pool))
			for i := 1; i < len(roles); i++ {
				assert.LessOrEqual(t, roles[i-1].Name, roles[i].Name, "roles should be sorted by name ascending")
			}
		}}},
		{name: "Success: roles are filtered by name", buildQuery: func(t *testing.T, mock mock) (params.Filters, params.SortItems) {
			target := seededRole(t, context.Background(), mock.pool)
			return params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: target.Name, SetOperator: params.FilterAnd}}}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, roles []services.Role, err error) {
			target := seededRole(t, context.Background(), mock.pool)
			require.NoError(t, err)
			require.Len(t, roles, 1)
			assert.Equal(t, target.Name, roles[0].Name)
		}}},
		{name: "Success: no roles match", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "does-not-exist", SetOperator: params.FilterAnd}}}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, _ mock, roles []services.Role, err error) {
			require.NoError(t, err)
			assert.Empty(t, roles)
		}}},
		{name: "Success: roles created after the epoch are returned", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{"created_at": {{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, roles []services.Role, err error) {
			require.NoError(t, err)
			assert.Len(t, roles, seededRoleCount(t, context.Background(), mock.pool))
		}}},
		{name: "Success: no roles were created before the epoch", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{"created_at": {{Field: "created_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, _ mock, roles []services.Role, err error) {
			require.NoError(t, err)
			assert.Empty(t, roles)
		}}},
		{name: "Success: roles updated after the epoch are returned", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{"updated_at": {{Field: "updated_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, mock mock, roles []services.Role, err error) {
			require.NoError(t, err)
			assert.Len(t, roles, seededRoleCount(t, context.Background(), mock.pool))
		}}},
		{name: "Success: no roles were updated before the epoch", buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
			return params.Filters{"updated_at": {{Field: "updated_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
		}, expected: expected{assertResponse: func(t *testing.T, _ mock, roles []services.Role, err error) {
			require.NoError(t, err)
			assert.Empty(t, roles)
		}}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
				mock        = mock{store: store, pool: pool}
			)

			filters, sortItems := testCase.buildQuery(t, mock)
			roles, err := mock.store.ListRoles(ctx, filters, sortItems)
			testCase.expected.assertResponse(t, mock, roles, err)
		})
	}
}

func TestStore_ListPermissions_Integration(t *testing.T) {
	type mock struct {
		store *appdb.Store
		pool  *pgxpool.Pool
	}

	type expected struct {
		assertResponse func(t *testing.T, mock mock, permissions []services.Permission)
	}

	type testData struct {
		name       string
		buildQuery func(t *testing.T, mock mock) (params.Filters, params.SortItems)
		expected   expected
	}

	const epoch = "2000-01-01T00:00:00Z"

	tt := []testData{
		{
			name: "Success: all seeded permissions are returned",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, mock mock, permissions []services.Permission) {
				expectedCount := seededPermissionCount(t, context.Background(), mock.pool)
				require.NotZero(t, expectedCount, "expected migrations to seed at least one permission")
				assert.Len(t, permissions, expectedCount)
			}},
		},
		{
			name: "Success: permissions are sorted by name",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{}, params.SortItems{{Field: "name", Direction: params.Ascending}}
			},
			expected: expected{assertResponse: func(t *testing.T, mock mock, permissions []services.Permission) {
				expectedCount := seededPermissionCount(t, context.Background(), mock.pool)
				require.Len(t, permissions, expectedCount)
				for i := 1; i < len(permissions); i++ {
					assert.LessOrEqual(t, permissions[i-1].Name, permissions[i].Name, "permissions should be sorted by name ascending")
				}
			}},
		},
		{
			name: "Success: permissions are filtered by authority",
			buildQuery: func(t *testing.T, mock mock) (params.Filters, params.SortItems) {
				target := seededPermission(t, context.Background(), mock.pool)
				return params.Filters{"authority": {{Field: "authority", Operator: params.Equals, Value: target.Authority, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, mock mock, permissions []services.Permission) {
				target := seededPermission(t, context.Background(), mock.pool)
				require.NotEmpty(t, permissions)
				for _, permission := range permissions {
					assert.Equal(t, target.Authority, permission.Authority)
				}
			}},
		},
		{
			name: "Success: no permissions match",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "does-not-exist", SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, _ mock, permissions []services.Permission) {
				assert.Empty(t, permissions)
			}},
		},
		{
			name: "Success: created_at supports date comparisons after the epoch",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{"created_at": {{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, mock mock, permissions []services.Permission) {
				assert.Len(t, permissions, seededPermissionCount(t, context.Background(), mock.pool))
			}},
		},
		{
			name: "Success: created_at supports date comparisons before the epoch",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{"created_at": {{Field: "created_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, _ mock, permissions []services.Permission) {
				assert.Empty(t, permissions)
			}},
		},
		{
			name: "Success: updated_at supports date comparisons after the epoch",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{"updated_at": {{Field: "updated_at", Operator: params.GreaterThanOrEquals, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, mock mock, permissions []services.Permission) {
				assert.Len(t, permissions, seededPermissionCount(t, context.Background(), mock.pool))
			}},
		},
		{
			name: "Success: updated_at supports date comparisons before the epoch",
			buildQuery: func(_ *testing.T, _ mock) (params.Filters, params.SortItems) {
				return params.Filters{"updated_at": {{Field: "updated_at", Operator: params.LessThan, Value: epoch, SetOperator: params.FilterAnd}}}, params.SortItems{}
			},
			expected: expected{assertResponse: func(t *testing.T, _ mock, permissions []services.Permission) {
				assert.Empty(t, permissions)
			}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
				mock        = mock{store: store, pool: pool}
			)

			filters, sortItems := testCase.buildQuery(t, mock)
			permissions, err := mock.store.ListPermissions(ctx, filters, sortItems)
			require.NoError(t, err)
			testCase.expected.assertResponse(t, mock, permissions)
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
	type mock struct {
		store *appdb.Store
		pool  *pgxpool.Pool
	}

	type expected struct {
		assertResponse func(t *testing.T, ctx context.Context, mock mock)
	}

	type testData struct {
		name     string
		table    string
		expected expected
	}

	tt := []testData{
		{
			name:  "Success: GetRole handles a stray deleted_at column",
			table: "roles",
			expected: expected{assertResponse: func(t *testing.T, ctx context.Context, mock mock) {
				expectedRole := seededRole(t, ctx, mock.pool)
				retrieved, err := mock.store.GetRole(ctx, expectedRole.ID)
				require.NoError(t, err)
				assert.Equal(t, expectedRole.ID, retrieved.ID)
				assert.Equal(t, expectedRole.Name, retrieved.Name)
			}},
		},
		{
			name:  "Success: ListRoles handles a stray deleted_at column",
			table: "roles",
			expected: expected{assertResponse: func(t *testing.T, ctx context.Context, mock mock) {
				expectedCount := seededRoleCount(t, ctx, mock.pool)
				require.NotZero(t, expectedCount, "expected migrations to seed at least one role")
				roles, err := mock.store.ListRoles(ctx, params.Filters{}, params.SortItems{})
				require.NoError(t, err)
				assert.Len(t, roles, expectedCount)
			}},
		},
		{
			name:  "Success: GetPermission handles a stray deleted_at column",
			table: "permissions",
			expected: expected{assertResponse: func(t *testing.T, ctx context.Context, mock mock) {
				expectedPermission := seededPermission(t, ctx, mock.pool)
				retrieved, err := mock.store.GetPermission(ctx, int(expectedPermission.ID))
				require.NoError(t, err)
				assert.Equal(t, expectedPermission.ID, retrieved.ID)
				assert.Equal(t, expectedPermission.Authority, retrieved.Authority)
				assert.Equal(t, expectedPermission.Name, retrieved.Name)
			}},
		},
		{
			name:  "Success: ListPermissions handles a stray deleted_at column",
			table: "permissions",
			expected: expected{assertResponse: func(t *testing.T, ctx context.Context, mock mock) {
				expectedCount := seededPermissionCount(t, ctx, mock.pool)
				permissions, err := mock.store.ListPermissions(ctx, params.Filters{}, params.SortItems{})
				require.NoError(t, err)
				assert.Len(t, permissions, expectedCount)
			}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			var (
				ctx         = context.Background()
				store, pool = setupStoreAndPool(t)
				mock        = mock{store: store, pool: pool}
			)

			addDeprecatedDeletedAtColumn(t, ctx, mock.pool, testCase.table)
			testCase.expected.assertResponse(t, ctx, mock)
		})
	}
}

// seedUser inserts a user row directly via the pool with every non-defaulted
// column populated so the strict pgx scanners in ListUsers do not encounter
// NULLs, and returns its generated id.
func seedUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, principalName string, supportAccount bool) uuid.UUID {
	t.Helper()

	id, err := uuid.NewV4()
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, principal_name, first_name, last_name, email_address, last_login, is_disabled, all_environments, eula_accepted, support_account, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, now(), false, true, false, $6, now(), now())`,
		id.String(), principalName, principalName+"-first", principalName+"-last", principalName+"@example.com", supportAccount,
	)
	require.NoError(t, err)

	return id
}

// seedUserWithNullLastLogin inserts a user row whose last_login is left NULL,
// exercising the nullable-timestamp scan path in ListUsers, and returns its id.
func seedUserWithNullLastLogin(t *testing.T, ctx context.Context, pool *pgxpool.Pool, principalName string) uuid.UUID {
	t.Helper()

	id, err := uuid.NewV4()
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, principal_name, first_name, last_name, email_address, last_login, is_disabled, all_environments, eula_accepted, support_account, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NULL, false, true, false, false, now(), now())`,
		id.String(), principalName, principalName+"-first", principalName+"-last", principalName+"@example.com",
	)
	require.NoError(t, err)

	return id
}

// seedUserRole associates a seeded user with an existing role via the join table.
func seedUserRole(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, roleID int32) {
	t.Helper()

	_, err := pool.Exec(ctx, `INSERT INTO users_roles (user_id, role_id) VALUES ($1, $2)`, userID.String(), roleID)
	require.NoError(t, err)
}

// seedETAC inserts an environment-targeted access control row for a user. It
// intentionally omits updated_at so the nullable-timestamp scan path is covered.
func seedETAC(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, environmentID string) {
	t.Helper()

	_, err := pool.Exec(ctx,
		`INSERT INTO environment_targeted_access_control (user_id, environment_id, created_at) VALUES ($1, $2, now())`,
		userID.String(), environmentID,
	)
	require.NoError(t, err)
}

// seedAuthSecret inserts an auth secret for a user, letting the id sequence default.
func seedAuthSecret(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) {
	t.Helper()

	_, err := pool.Exec(ctx,
		`INSERT INTO auth_secrets (user_id, digest, digest_method, expires_at, totp_secret, totp_activated, created_at, updated_at)
		 VALUES ($1, 'digest', 'argon2', now(), '', false, now(), now())`,
		userID.String(),
	)
	require.NoError(t, err)
}

func TestStore_ListUsers_Integration(t *testing.T) {
	t.Run("returns non-support users with preloaded associations", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
			role        = seededRole(t, ctx, pool)
		)

		secretUserID := seedUser(t, ctx, pool, "list-users-secret", false)
		seedUserRole(t, ctx, pool, secretUserID, role.ID)
		seedETAC(t, ctx, pool, secretUserID, "env-list-users")
		seedAuthSecret(t, ctx, pool, secretUserID)

		ssoUserID := seedUser(t, ctx, pool, "list-users-sso", false)
		supportUserID := seedUser(t, ctx, pool, "list-users-support", true)

		users, err := store.ListUsers(ctx, params.Filters{}, params.SortItems{})
		require.NoError(t, err)

		byID := make(map[uuid.UUID]services.User, len(users))
		for _, user := range users {
			byID[user.ID] = user
		}

		_, hasSupport := byID[supportUserID]
		assert.False(t, hasSupport, "support accounts must be excluded from the list")

		secretUser, ok := byID[secretUserID]
		require.True(t, ok, "expected the seeded secret user to be returned")
		require.NotEmpty(t, secretUser.Roles, "expected the seeded user to preload its role")
		assert.NotEmpty(t, secretUser.Roles[0].Permissions, "expected the preloaded role to carry its permissions")
		require.Len(t, secretUser.EnvironmentTargetedAccessControl, 1)
		assert.Equal(t, "env-list-users", secretUser.EnvironmentTargetedAccessControl[0].EnvironmentID)
		require.NotNil(t, secretUser.AuthSecret, "expected the seeded user to preload its auth secret")
		assert.Equal(t, "argon2", secretUser.AuthSecret.DigestMethod)

		ssoUser, ok := byID[ssoUserID]
		require.True(t, ok, "expected the seeded sso user to be returned")
		assert.Nil(t, ssoUser.AuthSecret, "expected the user without a secret to have a nil AuthSecret")
		assert.Empty(t, ssoUser.Roles)
		assert.Empty(t, ssoUser.EnvironmentTargetedAccessControl)
	})

	t.Run("returns users filtered by principal_name", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
			principal   = "filter-target-user"
		)

		seedUser(t, ctx, pool, principal, false)
		seedUser(t, ctx, pool, "filter-other-user", false)

		users, err := store.ListUsers(ctx, params.Filters{
			"principal_name": {{Field: "principal_name", Operator: params.Equals, Value: principal, SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)
		require.Len(t, users, 1)
		assert.Equal(t, principal, users[0].PrincipalName)
	})

	t.Run("returns a user with a NULL last_login as a zero-value timestamp", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		userID := seedUserWithNullLastLogin(t, ctx, pool, "null-last-login-user")

		// The eq:null filter selects exactly the NULL rows, exercising the strict
		// pgx scan path that previously failed on a NULL last_login.
		users, err := store.ListUsers(ctx, params.Filters{
			"last_login": {{Field: "last_login", Operator: params.Equals, Value: "null", SetOperator: params.FilterAnd}},
		}, params.SortItems{})
		require.NoError(t, err)

		byID := make(map[uuid.UUID]services.User, len(users))
		for _, user := range users {
			byID[user.ID] = user
		}

		user, ok := byID[userID]
		require.True(t, ok, "expected the user with a NULL last_login to be returned")
		assert.True(t, user.LastLogin.IsZero(), "a NULL last_login should scan to the zero-value time, matching legacy gorm coercion")
	})

	t.Run("returns users sorted by principal_name ascending", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		seedUser(t, ctx, pool, "zzz-sort-user", false)
		seedUser(t, ctx, pool, "aaa-sort-user", false)

		users, err := store.ListUsers(ctx, params.Filters{}, params.SortItems{{Field: "principal_name", Direction: params.Ascending}})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(users), 2)

		for i := 1; i < len(users); i++ {
			assert.LessOrEqual(t, users[i-1].PrincipalName, users[i].PrincipalName, "users should be sorted by principal_name ascending")
		}
	})
}
