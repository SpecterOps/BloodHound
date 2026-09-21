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
	"errors"
	"testing"
	"time"

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
