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

package appdb_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedGetPermissionSQL is the literal SQL the Store issues for GetPermission.
const expectedGetPermissionSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions WHERE id = $1 LIMIT $2`

// expectedGetRoleSQL is the literal SQL the Store issues for the roles query in GetRole.
const expectedGetRoleSQL = `SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1 LIMIT $2`

// expectedGetRolePermissionsSQL is the literal SQL the Store issues for the
// per-role permissions query in GetRole.
const expectedGetRolePermissionsSQL = `SELECT p.id, p.authority, p.name, p.created_at, p.updated_at FROM permissions p JOIN roles_permissions rp ON rp.permission_id = p.id WHERE rp.role_id = $1`

// expectedListRolePermissionsSQL is the literal SQL the Store issues to batch-load
// permissions for every listed role in a single query.
const expectedListRolePermissionsSQL = `SELECT rp.role_id, p.id, p.authority, p.name, p.created_at, p.updated_at FROM permissions p JOIN roles_permissions rp ON rp.permission_id = p.id WHERE rp.role_id IN ($1)`

// expectedListRolePermissionsTwoSQL is the batched permissions query for two listed roles.
const expectedListRolePermissionsTwoSQL = `SELECT rp.role_id, p.id, p.authority, p.name, p.created_at, p.updated_at FROM permissions p JOIN roles_permissions rp ON rp.permission_id = p.id WHERE rp.role_id IN ($1, $2)`

// expectedListRolesSQL is the literal SQL the Store issues for the roles query in
// ListRoles when no filters or sorts are supplied.
const expectedListRolesSQL = `SELECT id, name, description, created_at, updated_at FROM roles`

// expectedListRolesSortedSQL is the literal SQL the Store issues when a single
// ascending sort on name is supplied.
const expectedListRolesSortedSQL = `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name ASC`

// expectedListRolesFilteredSQL is the literal SQL the Store issues when a single
// equality filter on name is supplied.
const expectedListRolesFilteredSQL = `SELECT id, name, description, created_at, updated_at FROM roles WHERE (name = $1)`

// expectedListRolesFilteredByIDSQL is the literal SQL the Store issues when a
// single greater-than filter on the numeric id column is supplied.
const expectedListRolesFilteredByIDSQL = `SELECT id, name, description, created_at, updated_at FROM roles WHERE (id > $1)`

const expectedListPermissionsSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions`

const expectedListPermissionsSortedSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions ORDER BY name ASC`

const expectedListPermissionsFilteredSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions WHERE (authority = $1)`

const expectedListPermissionsFilteredByIDSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions WHERE (id > $1)`

const expectedListPermissionsCreatedAtNullSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions WHERE (created_at IS NULL)`

const expectedListPermissionsCreatedAtNotNullSQL = `SELECT id, authority, name, created_at, updated_at FROM permissions WHERE (created_at IS NOT NULL)`

// expectedListUsersSQL is the literal SQL the Store issues for the users query in
// ListUsers when no filters or sorts are supplied.
const expectedListUsersSQL = `SELECT id, sso_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, all_environments, eula_accepted, created_at, updated_at FROM users WHERE support_account = $1`

// expectedListUsersSortedSQL is the literal SQL issued when a single ascending
// sort on principal_name is supplied.
const expectedListUsersSortedSQL = expectedListUsersSQL + ` ORDER BY principal_name ASC`

// expectedListUsersFilteredSQL is the literal SQL issued when a single equality
// filter on first_name is supplied alongside the support_account exclusion.
const expectedListUsersFilteredSQL = `SELECT id, sso_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, all_environments, eula_accepted, created_at, updated_at FROM users WHERE support_account = $1 AND (first_name = $2)`

// expectedListUsersFilteredNullSQL is the literal SQL issued for an eq:null
// filter, which must render IS NULL rather than binding "null" as a parameter.
const expectedListUsersFilteredNullSQL = `SELECT id, sso_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, all_environments, eula_accepted, created_at, updated_at FROM users WHERE support_account = $1 AND (last_login IS NULL)`

// expectedListUsersFilteredNotNullSQL is the literal SQL issued for a neq:null
// filter, which must render IS NOT NULL rather than binding "null".
const expectedListUsersFilteredNotNullSQL = `SELECT id, sso_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, all_environments, eula_accepted, created_at, updated_at FROM users WHERE support_account = $1 AND (last_login IS NOT NULL)`

// expectedRolesForUsersOneSQL / TwoSQL are the batched roles query for one and
// two listed users respectively.
const expectedRolesForUsersOneSQL = `SELECT ur.user_id, r.id, r.name, r.description, r.created_at, r.updated_at FROM roles r JOIN users_roles ur ON ur.role_id = r.id WHERE ur.user_id IN ($1)`
const expectedRolesForUsersTwoSQL = `SELECT ur.user_id, r.id, r.name, r.description, r.created_at, r.updated_at FROM roles r JOIN users_roles ur ON ur.role_id = r.id WHERE ur.user_id IN ($1, $2)`

// expectedETACForUsersOneSQL / TwoSQL are the batched environment access control
// query for one and two listed users respectively.
const expectedETACForUsersOneSQL = `SELECT id, user_id, environment_id, created_at, updated_at FROM environment_targeted_access_control WHERE user_id IN ($1)`
const expectedETACForUsersTwoSQL = `SELECT id, user_id, environment_id, created_at, updated_at FROM environment_targeted_access_control WHERE user_id IN ($1, $2)`

// expectedAuthSecretsForUsersOneSQL / TwoSQL are the batched auth secrets query
// for one and two listed users respectively.
const expectedAuthSecretsForUsersOneSQL = `SELECT id, user_id, digest_method, expires_at, totp_activated, created_at, updated_at FROM auth_secrets WHERE user_id IN ($1)`
const expectedAuthSecretsForUsersTwoSQL = `SELECT id, user_id, digest_method, expires_at, totp_activated, created_at, updated_at FROM auth_secrets WHERE user_id IN ($1, $2)`

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func permissionRowColumns() []string {
	return []string{"authority", "name", "id", "created_at", "updated_at"}
}

func TestStore_GetPermission(t *testing.T) {
	type mock struct {
		pool pgxmock.PgxPoolIface
	}

	type expected struct {
		permission  services.Permission
		err         error
		errContains string
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx                = context.Background()
		dbErr              = errors.New("connection refused")
		permissionID       = 5
		expectedPermission = services.Permission{
			Authority: "clients",
			Name:      "ReadClients",
			ID:        5,
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
	)

	tt := []testData{
		{
			name: "Success: permission is returned - 200",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					mock.pool.NewRows(permissionRowColumns()).AddRow(
						expectedPermission.Authority,
						expectedPermission.Name,
						expectedPermission.ID,
						expectedPermission.CreatedAt,
						expectedPermission.UpdatedAt,
					),
				)
			},
			expected: expected{permission: expectedPermission},
		},
		{
			name: "Error: missing permission maps to ErrNoPermissionFound - 404",
			setupMocks: func(mock mock) {
				// Query succeeds but returns zero rows; CollectOneRow returns pgx.ErrNoRows
				mock.pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					mock.pool.NewRows(permissionRowColumns()),
				)
			},
			expected: expected{err: services.ErrNoPermissionFound},
		},
		{
			name: "Error: permission rows cannot be collected - 500",
			setupMocks: func(mock mock) {
				// The rows object carries a close error that pgx.CollectOneRow surfaces
				// via rows.Err() when Next() returns false.
				mock.pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					mock.pool.NewRows(permissionRowColumns()).CloseError(errors.New("forced iteration error")),
				)
			},
			expected: expected{errContains: "finding permission:"},
		},
		{
			name: "Error: database query fails - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnError(dbErr)
			},
			expected: expected{err: dbErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			testCase.setupMocks(mock{pool: pool})

			result, err := store.GetPermission(ctx, permissionID)
			switch {
			case testCase.expected.err != nil:
				assert.ErrorIs(t, err, testCase.expected.err)
			case testCase.expected.errContains != "":
				assert.ErrorContains(t, err, testCase.expected.errContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.permission, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func roleRowColumns() []string {
	return []string{"id", "name", "description", "created_at", "updated_at"}
}

func rolePermissionRowColumns() []string {
	return []string{"id", "authority", "name", "created_at", "updated_at"}
}

func listRolePermissionRowColumns() []string {
	return []string{"role_id", "id", "authority", "name", "created_at", "updated_at"}
}

func TestStore_GetRole(t *testing.T) {
	type mock struct {
		pool pgxmock.PgxPoolIface
	}

	type expected struct {
		role        services.Role
		err         error
		errContains string
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx          = context.Background()
		dbErr        = errors.New("connection refused")
		roleID       = int32(3)
		createdAt    = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedAt    = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		expectedRole = services.Role{
			ID:          3,
			Name:        "Administrator",
			Description: "Can manage the application",
			Permissions: []services.Permission{
				{ID: 1, Authority: "auth", Name: "ManageProviders", CreatedAt: createdAt, UpdatedAt: updatedAt},
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	)

	tt := []testData{
		{
			name: "Success: role with permissions is returned - 200",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnRows(
					mock.pool.NewRows(roleRowColumns()).AddRow(
						expectedRole.ID,
						expectedRole.Name,
						expectedRole.Description,
						expectedRole.CreatedAt,
						expectedRole.UpdatedAt,
					),
				)
				mock.pool.ExpectQuery(expectedGetRolePermissionsSQL).WithArgs(roleID).WillReturnRows(
					mock.pool.NewRows(rolePermissionRowColumns()).AddRow(
						expectedRole.Permissions[0].ID,
						expectedRole.Permissions[0].Authority,
						expectedRole.Permissions[0].Name,
						expectedRole.Permissions[0].CreatedAt,
						expectedRole.Permissions[0].UpdatedAt,
					),
				)
			},
			expected: expected{role: expectedRole},
		},
		{
			name: "Error: missing role maps to ErrNoRoleFound - 404",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnRows(
					mock.pool.NewRows(roleRowColumns()),
				)
			},
			expected: expected{err: services.ErrNoRoleFound},
		},
		{
			name: "Error: role query fails - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnError(dbErr)
			},
			expected: expected{err: dbErr},
		},
		{
			name: "Error: role permissions query fails - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnRows(
					mock.pool.NewRows(roleRowColumns()).AddRow(
						expectedRole.ID, expectedRole.Name, expectedRole.Description, expectedRole.CreatedAt, expectedRole.UpdatedAt,
					),
				)
				mock.pool.ExpectQuery(expectedGetRolePermissionsSQL).WithArgs(roleID).WillReturnError(dbErr)
			},
			expected: expected{errContains: "querying permissions for role:"},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			testCase.setupMocks(mock{pool: pool})

			result, err := store.GetRole(ctx, roleID)
			switch {
			case testCase.expected.err != nil:
				assert.ErrorIs(t, err, testCase.expected.err)
			case testCase.expected.errContains != "":
				assert.ErrorContains(t, err, testCase.expected.errContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.role, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_ListRoles(t *testing.T) {
	type mock struct {
		pool pgxmock.PgxPoolIface
	}

	type expected struct {
		roles       []services.Role
		err         error
		errContains string
	}

	type testData struct {
		name       string
		filters    params.Filters
		sortItems  params.SortItems
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx       = context.Background()
		dbErr     = errors.New("connection refused")
		createdAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedAt = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		admin     = services.Role{
			ID:          1,
			Name:        "Administrator",
			Description: "Can manage the application",
			Permissions: []services.Permission{
				{ID: 1, Authority: "auth", Name: "ManageProviders", CreatedAt: createdAt, UpdatedAt: updatedAt},
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		readOnly = services.Role{
			ID:          2,
			Name:        "Read-Only",
			Description: "Read only access",
			Permissions: []services.Permission{},
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
	)

	expectRoleRows := func(pool pgxmock.PgxPoolIface, roles ...services.Role) *pgxmock.Rows {
		rows := pool.NewRows(roleRowColumns())
		for _, r := range roles {
			rows.AddRow(r.ID, r.Name, r.Description, r.CreatedAt, r.UpdatedAt)
		}
		return rows
	}

	expectPermissionsFor := func(pool pgxmock.PgxPoolIface, expectedSQL string, roles ...services.Role) {
		rows := pool.NewRows(listRolePermissionRowColumns())
		args := make([]any, 0, len(roles))
		for _, r := range roles {
			args = append(args, r.ID)
			for _, p := range r.Permissions {
				rows.AddRow(r.ID, p.ID, p.Authority, p.Name, p.CreatedAt, p.UpdatedAt)
			}
		}
		pool.ExpectQuery(expectedSQL).WithArgs(args...).WillReturnRows(rows)
	}

	tt := []testData{
		{
			name: "Success: every role with permissions is returned - 200",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListRolesSQL).WithArgs().WillReturnRows(expectRoleRows(mock.pool, admin, readOnly))
				expectPermissionsFor(mock.pool, expectedListRolePermissionsTwoSQL, admin, readOnly)
			},
			expected: expected{roles: []services.Role{admin, readOnly}},
		},
		{
			name:      "Success: roles are sorted by name - 200",
			sortItems: params.SortItems{{Field: "name", Direction: params.Ascending}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListRolesSortedSQL).WithArgs().WillReturnRows(expectRoleRows(mock.pool, admin))
				expectPermissionsFor(mock.pool, expectedListRolePermissionsSQL, admin)
			},
			expected: expected{roles: []services.Role{admin}},
		},
		{
			name:    "Success: roles are filtered by name - 200",
			filters: params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "Administrator", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListRolesFilteredSQL).WithArgs("Administrator").WillReturnRows(expectRoleRows(mock.pool, admin))
				expectPermissionsFor(mock.pool, expectedListRolePermissionsSQL, admin)
			},
			expected: expected{roles: []services.Role{admin}},
		},
		{
			name:    "Success: roles are filtered by ID - 200",
			filters: params.Filters{"id": {{Field: "id", Operator: params.GreaterThan, Value: "1", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListRolesFilteredByIDSQL).WithArgs("1").WillReturnRows(expectRoleRows(mock.pool, readOnly))
				expectPermissionsFor(mock.pool, expectedListRolePermissionsSQL, readOnly)
			},
			expected: expected{roles: []services.Role{readOnly}},
		},
		{
			name:    "Error: filter field is unknown - 400",
			filters: params.Filters{"nope": {{Field: "nope", Operator: params.Equals, Value: "x", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
			},
			expected: expected{errContains: "unknown field"},
		},
		{
			name:      "Error: sort field is unknown - 400",
			sortItems: params.SortItems{{Field: "nope", Direction: params.Ascending}},
			setupMocks: func(mock mock) {
			},
			expected: expected{errContains: "unknown field"},
		},
		{
			name: "Error: roles query fails - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListRolesSQL).WithArgs().WillReturnError(dbErr)
			},
			expected: expected{err: dbErr},
		},
		{
			name: "Error: role permissions query fails - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListRolesSQL).WithArgs().WillReturnRows(expectRoleRows(mock.pool, admin))
				mock.pool.ExpectQuery(expectedListRolePermissionsSQL).WithArgs(admin.ID).WillReturnError(dbErr)
			},
			expected: expected{err: dbErr, errContains: "querying permissions for roles:"},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			testCase.setupMocks(mock{pool: pool})

			result, err := store.ListRoles(ctx, testCase.filters, testCase.sortItems)
			switch {
			case testCase.expected.err != nil || testCase.expected.errContains != "":
				if testCase.expected.err != nil {
					assert.ErrorIs(t, err, testCase.expected.err)
				}
				if testCase.expected.errContains != "" {
					assert.ErrorContains(t, err, testCase.expected.errContains)
				}
			default:
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.roles, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_ListPermissions(t *testing.T) {
	type mock struct {
		pool pgxmock.PgxPoolIface
	}

	type expected struct {
		permissions []services.Permission
		err         error
		errContains string
	}

	type testData struct {
		name       string
		filters    params.Filters
		sortItems  params.SortItems
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx                = context.Background()
		dbErr              = errors.New("connection refused")
		createdAt          = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedAt          = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		expectedPermission = services.Permission{ID: 7, Authority: "app", Name: "ManageProviders", CreatedAt: createdAt, UpdatedAt: updatedAt}
	)

	expectPermissionRows := func(pool pgxmock.PgxPoolIface, permissions ...services.Permission) *pgxmock.Rows {
		rows := pool.NewRows(permissionRowColumns())
		for _, permission := range permissions {
			rows.AddRow(permission.Authority, permission.Name, permission.ID, permission.CreatedAt, permission.UpdatedAt)
		}
		return rows
	}

	tt := []testData{
		{
			name: "Success: every permission is returned",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsSQL).WithArgs().WillReturnRows(expectPermissionRows(mock.pool, expectedPermission))
			},
			expected: expected{permissions: []services.Permission{expectedPermission}},
		},
		{
			name:      "Success: permissions are sorted by name",
			sortItems: params.SortItems{{Field: "name", Direction: params.Ascending}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsSortedSQL).WithArgs().WillReturnRows(expectPermissionRows(mock.pool, expectedPermission))
			},
			expected: expected{permissions: []services.Permission{expectedPermission}},
		},
		{
			name:    "Success: permissions are filtered by authority",
			filters: params.Filters{"authority": {{Field: "authority", Operator: params.Equals, Value: "app", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsFilteredSQL).WithArgs("app").WillReturnRows(expectPermissionRows(mock.pool, expectedPermission))
			},
			expected: expected{permissions: []services.Permission{expectedPermission}},
		},
		{
			name:    "Success: permissions are filtered by ID",
			filters: params.Filters{"id": {{Field: "id", Operator: params.GreaterThan, Value: "1", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsFilteredByIDSQL).WithArgs("1").WillReturnRows(expectPermissionRows(mock.pool, expectedPermission))
			},
			expected: expected{permissions: []services.Permission{expectedPermission}},
		},
		{
			name:    "Success: null equality uses IS NULL",
			filters: params.Filters{"created_at": {{Operator: params.Equals, Value: "null", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsCreatedAtNullSQL).WithArgs().WillReturnRows(expectPermissionRows(mock.pool))
			},
			expected: expected{permissions: []services.Permission{}},
		},
		{
			name:    "Success: null inequality uses IS NOT NULL",
			filters: params.Filters{"created_at": {{Operator: params.NotEquals, Value: "null", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsCreatedAtNotNullSQL).WithArgs().WillReturnRows(expectPermissionRows(mock.pool))
			},
			expected: expected{permissions: []services.Permission{}},
		},
		{
			name:       "Error: filter field is unknown",
			filters:    params.Filters{"nope": {{Field: "nope", Operator: params.Equals, Value: "x", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock) {},
			expected:   expected{errContains: "unknown field"},
		},
		{
			name:       "Error: sort field is unknown",
			sortItems:  params.SortItems{{Field: "nope", Direction: params.Ascending}},
			setupMocks: func(mock) {},
			expected:   expected{errContains: "unknown field"},
		},
		{
			name: "Error: database query fails",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsSQL).WithArgs().WillReturnError(dbErr)
			},
			expected: expected{err: dbErr},
		},
		{
			name: "Error: permission rows cannot be collected",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListPermissionsSQL).WithArgs().WillReturnRows(
					mock.pool.NewRows(permissionRowColumns()).CloseError(dbErr),
				)
			},
			expected: expected{errContains: "collecting permissions:"},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			testCase.setupMocks(mock{pool: pool})

			permissions, err := store.ListPermissions(ctx, testCase.filters, testCase.sortItems)
			switch {
			case testCase.expected.err != nil:
				assert.ErrorIs(t, err, testCase.expected.err)
			case testCase.expected.errContains != "":
				assert.ErrorContains(t, err, testCase.expected.errContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.permissions, permissions)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func userRowColumns() []string {
	return []string{"id", "sso_provider_id", "first_name", "last_name", "email_address", "principal_name", "last_login", "is_disabled", "all_environments", "eula_accepted", "created_at", "updated_at"}
}

func userRoleRowColumns() []string {
	return []string{"user_id", "id", "name", "description", "created_at", "updated_at"}
}

func userETACRowColumns() []string {
	return []string{"id", "user_id", "environment_id", "created_at", "updated_at"}
}

func userAuthSecretRowColumns() []string {
	return []string{"id", "user_id", "digest_method", "expires_at", "totp_activated", "created_at", "updated_at"}
}

func TestStore_ListUsers(t *testing.T) {
	type mock struct {
		pool pgxmock.PgxPoolIface
	}

	type expected struct {
		users       []services.User
		err         error
		errContains string
	}

	type testData struct {
		name       string
		filters    params.Filters
		sortItems  params.SortItems
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx       = context.Background()
		dbErr     = errors.New("connection refused")
		createdAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedAt = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		lastLogin = time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
		expiresAt = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		user1ID   = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
		user2ID   = uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222")

		permission1 = services.Permission{ID: 1, Authority: "auth", Name: "ManageProviders", CreatedAt: createdAt, UpdatedAt: updatedAt}
		role1       = services.Role{ID: 1, Name: "Administrator", Description: "Can manage the application", Permissions: []services.Permission{permission1}, CreatedAt: createdAt, UpdatedAt: updatedAt}
		role2       = services.Role{ID: 2, Name: "Read-Only", Description: "Read only access", Permissions: []services.Permission{}, CreatedAt: createdAt, UpdatedAt: updatedAt}
		etac1       = services.EnvironmentAccessControl{ID: 10, UserID: user1ID.String(), EnvironmentID: "env-1", CreatedAt: createdAt, UpdatedAt: updatedAt}
		secret1     = services.AuthSecret{ID: 100, DigestMethod: "argon2", ExpiresAt: expiresAt, TOTPActivated: false, CreatedAt: createdAt, UpdatedAt: updatedAt}

		user1 = services.User{
			ID:                               user1ID,
			SSOProviderID:                    sql.NullInt32{},
			FirstName:                        sql.NullString{String: "Ada", Valid: true},
			LastName:                         sql.NullString{String: "Lovelace", Valid: true},
			EmailAddress:                     sql.NullString{String: "ada@example.com", Valid: true},
			PrincipalName:                    "ada",
			LastLogin:                        lastLogin,
			IsDisabled:                       false,
			AllEnvironments:                  false,
			EULAAccepted:                     true,
			Roles:                            []services.Role{role1},
			EnvironmentTargetedAccessControl: []services.EnvironmentAccessControl{etac1},
			AuthSecret:                       &secret1,
			CreatedAt:                        createdAt,
			UpdatedAt:                        updatedAt,
		}
		user2 = services.User{
			ID:              user2ID,
			SSOProviderID:   sql.NullInt32{Int32: 5, Valid: true},
			FirstName:       sql.NullString{},
			LastName:        sql.NullString{},
			EmailAddress:    sql.NullString{},
			PrincipalName:   "sso-user",
			LastLogin:       lastLogin,
			IsDisabled:      false,
			AllEnvironments: true,
			EULAAccepted:    false,
			Roles:           []services.Role{role2},
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
		}
	)

	addUserRow := func(rows *pgxmock.Rows, user services.User) {
		rows.AddRow(user.ID, user.SSOProviderID, user.FirstName, user.LastName, user.EmailAddress, user.PrincipalName, user.LastLogin, user.IsDisabled, user.AllEnvironments, user.EULAAccepted, user.CreatedAt, user.UpdatedAt)
	}

	tt := []testData{
		{
			name: "Success: every user with associations is returned - 200",
			setupMocks: func(mock mock) {
				userRows := mock.pool.NewRows(userRowColumns())
				addUserRow(userRows, user1)
				addUserRow(userRows, user2)
				mock.pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnRows(userRows)

				mock.pool.ExpectQuery(expectedRolesForUsersTwoSQL).WithArgs(user1ID.String(), user2ID.String()).WillReturnRows(
					mock.pool.NewRows(userRoleRowColumns()).
						AddRow(user1ID, role1.ID, role1.Name, role1.Description, role1.CreatedAt, role1.UpdatedAt).
						AddRow(user2ID, role2.ID, role2.Name, role2.Description, role2.CreatedAt, role2.UpdatedAt),
				)
				mock.pool.ExpectQuery(expectedListRolePermissionsTwoSQL).WithArgs(role1.ID, role2.ID).WillReturnRows(
					mock.pool.NewRows(listRolePermissionRowColumns()).
						AddRow(role1.ID, permission1.ID, permission1.Authority, permission1.Name, permission1.CreatedAt, permission1.UpdatedAt),
				)
				mock.pool.ExpectQuery(expectedETACForUsersTwoSQL).WithArgs(user1ID.String(), user2ID.String()).WillReturnRows(
					mock.pool.NewRows(userETACRowColumns()).
						AddRow(etac1.ID, user1ID, etac1.EnvironmentID, etac1.CreatedAt, sql.NullTime{Time: etac1.UpdatedAt, Valid: true}),
				)
				mock.pool.ExpectQuery(expectedAuthSecretsForUsersTwoSQL).WithArgs(user1ID.String(), user2ID.String()).WillReturnRows(
					mock.pool.NewRows(userAuthSecretRowColumns()).
						AddRow(secret1.ID, user1ID, secret1.DigestMethod, sql.NullTime{Time: secret1.ExpiresAt, Valid: true}, secret1.TOTPActivated, secret1.CreatedAt, secret1.UpdatedAt),
				)
			},
			expected: expected{users: []services.User{user1, user2}},
		},
		{
			name:      "Success: users are sorted by principal_name - 200",
			sortItems: params.SortItems{{Field: "principal_name", Direction: params.Ascending}},
			setupMocks: func(mock mock) {
				userRows := mock.pool.NewRows(userRowColumns())
				addUserRow(userRows, user2)
				mock.pool.ExpectQuery(expectedListUsersSortedSQL).WithArgs(false).WillReturnRows(userRows)

				mock.pool.ExpectQuery(expectedRolesForUsersOneSQL).WithArgs(user2ID.String()).WillReturnRows(mock.pool.NewRows(userRoleRowColumns()))
				mock.pool.ExpectQuery(expectedETACForUsersOneSQL).WithArgs(user2ID.String()).WillReturnRows(mock.pool.NewRows(userETACRowColumns()))
				mock.pool.ExpectQuery(expectedAuthSecretsForUsersOneSQL).WithArgs(user2ID.String()).WillReturnRows(mock.pool.NewRows(userAuthSecretRowColumns()))
			},
			expected: expected{users: []services.User{{
				ID:              user2ID,
				SSOProviderID:   sql.NullInt32{Int32: 5, Valid: true},
				PrincipalName:   "sso-user",
				LastLogin:       lastLogin,
				AllEnvironments: true,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			}}},
		},
		{
			name:    "Success: users are filtered by first_name - 200",
			filters: params.Filters{"first_name": {{Field: "first_name", Operator: params.Equals, Value: "Ada", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				userRows := mock.pool.NewRows(userRowColumns())
				addUserRow(userRows, user1)
				mock.pool.ExpectQuery(expectedListUsersFilteredSQL).WithArgs(false, "Ada").WillReturnRows(userRows)

				mock.pool.ExpectQuery(expectedRolesForUsersOneSQL).WithArgs(user1ID.String()).WillReturnRows(
					mock.pool.NewRows(userRoleRowColumns()).
						AddRow(user1ID, role1.ID, role1.Name, role1.Description, role1.CreatedAt, role1.UpdatedAt),
				)
				mock.pool.ExpectQuery(expectedListRolePermissionsSQL).WithArgs(role1.ID).WillReturnRows(
					mock.pool.NewRows(listRolePermissionRowColumns()).
						AddRow(role1.ID, permission1.ID, permission1.Authority, permission1.Name, permission1.CreatedAt, permission1.UpdatedAt),
				)
				mock.pool.ExpectQuery(expectedETACForUsersOneSQL).WithArgs(user1ID.String()).WillReturnRows(
					mock.pool.NewRows(userETACRowColumns()).
						AddRow(etac1.ID, user1ID, etac1.EnvironmentID, etac1.CreatedAt, sql.NullTime{Time: etac1.UpdatedAt, Valid: true}),
				)
				mock.pool.ExpectQuery(expectedAuthSecretsForUsersOneSQL).WithArgs(user1ID.String()).WillReturnRows(
					mock.pool.NewRows(userAuthSecretRowColumns()).
						AddRow(secret1.ID, user1ID, secret1.DigestMethod, sql.NullTime{Time: secret1.ExpiresAt, Valid: true}, secret1.TOTPActivated, secret1.CreatedAt, secret1.UpdatedAt),
				)
			},
			expected: expected{users: []services.User{user1}},
		},
		{
			name:    "Success: null equality uses IS NULL - 200",
			filters: params.Filters{"last_login": {{Field: "last_login", Operator: params.Equals, Value: "null", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListUsersFilteredNullSQL).WithArgs(false).WillReturnRows(mock.pool.NewRows(userRowColumns()))
			},
			expected: expected{users: []services.User{}},
		},
		{
			name:    "Success: null inequality uses IS NOT NULL - 200",
			filters: params.Filters{"last_login": {{Field: "last_login", Operator: params.NotEquals, Value: "null", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListUsersFilteredNotNullSQL).WithArgs(false).WillReturnRows(mock.pool.NewRows(userRowColumns()))
			},
			expected: expected{users: []services.User{}},
		},
		{
			name: "Success: no users match - 200",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnRows(mock.pool.NewRows(userRowColumns()))
			},
			expected: expected{users: []services.User{}},
		},
		{
			name:       "Error: filter field is unknown - 400",
			filters:    params.Filters{"nope": {{Field: "nope", Operator: params.Equals, Value: "x", SetOperator: params.FilterAnd}}},
			setupMocks: func(mock) {},
			expected:   expected{errContains: "unknown field"},
		},
		{
			name:       "Error: sort field is unknown - 400",
			sortItems:  params.SortItems{{Field: "nope", Direction: params.Ascending}},
			setupMocks: func(mock) {},
			expected:   expected{errContains: "unknown field"},
		},
		{
			name: "Error: users query fails - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnError(dbErr)
			},
			expected: expected{err: dbErr},
		},
		{
			name: "Error: roles query fails - 500",
			setupMocks: func(mock mock) {
				userRows := mock.pool.NewRows(userRowColumns())
				addUserRow(userRows, user1)
				mock.pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnRows(userRows)
				mock.pool.ExpectQuery(expectedRolesForUsersOneSQL).WithArgs(user1ID.String()).WillReturnError(dbErr)
			},
			expected: expected{err: dbErr, errContains: "querying roles for users:"},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			testCase.setupMocks(mock{pool: pool})

			result, err := store.ListUsers(ctx, testCase.filters, testCase.sortItems)
			switch {
			case testCase.expected.err != nil || testCase.expected.errContains != "":
				if testCase.expected.err != nil {
					assert.ErrorIs(t, err, testCase.expected.err)
				}
				if testCase.expected.errContains != "" {
					assert.ErrorContains(t, err, testCase.expected.errContains)
				}
			default:
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.users, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
