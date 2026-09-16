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
	var (
		ctx          = context.Background()
		dbErr        = errors.New("connection refused")
		permissionID = 5
		expected     = services.Permission{
			Authority: "clients",
			Name:      "ReadClients",
			ID:        5,
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
	)

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      services.Permission
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns the permission on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					pool.NewRows(permissionRowColumns()).AddRow(
						expected.Authority,
						expected.Name,
						expected.ID,
						expected.CreatedAt,
						expected.UpdatedAt,
					),
				)
			},
			wantResult: expected,
		},
		{
			name: "maps CollectOneRow pgx.ErrNoRows to services.ErrNoPermissionFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				// Query succeeds but returns zero rows; CollectOneRow returns pgx.ErrNoRows
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					pool.NewRows(permissionRowColumns()),
				)
			},
			wantErr: services.ErrNoPermissionFound,
		},
		{
			name: "wraps CollectOneRow iteration error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				// The rows object carries a close error that pgx.CollectOneRow surfaces
				// via rows.Err() when Next() returns false.
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnRows(
					pool.NewRows(permissionRowColumns()).CloseError(errors.New("forced iteration error")),
				)
			},
			wantErrContains: "finding permission:",
		},
		{
			name: "propagates other database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetPermissionSQL).WithArgs(permissionID, 1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.GetPermission(ctx, permissionID)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
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
	var (
		ctx       = context.Background()
		dbErr     = errors.New("connection refused")
		roleID    = int32(3)
		createdAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedAt = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		expected  = services.Role{
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

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      services.Role
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns the role with permissions on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnRows(
					pool.NewRows(roleRowColumns()).AddRow(
						expected.ID,
						expected.Name,
						expected.Description,
						expected.CreatedAt,
						expected.UpdatedAt,
					),
				)
				pool.ExpectQuery(expectedGetRolePermissionsSQL).WithArgs(roleID).WillReturnRows(
					pool.NewRows(rolePermissionRowColumns()).AddRow(
						expected.Permissions[0].ID,
						expected.Permissions[0].Authority,
						expected.Permissions[0].Name,
						expected.Permissions[0].CreatedAt,
						expected.Permissions[0].UpdatedAt,
					),
				)
			},
			wantResult: expected,
		},
		{
			name: "maps CollectOneRow pgx.ErrNoRows to services.ErrNoRoleFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnRows(
					pool.NewRows(roleRowColumns()),
				)
			},
			wantErr: services.ErrNoRoleFound,
		},
		{
			name: "propagates role query database error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name: "wraps permissions query error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetRoleSQL).WithArgs(roleID, 1).WillReturnRows(
					pool.NewRows(roleRowColumns()).AddRow(
						expected.ID, expected.Name, expected.Description, expected.CreatedAt, expected.UpdatedAt,
					),
				)
				pool.ExpectQuery(expectedGetRolePermissionsSQL).WithArgs(roleID).WillReturnError(dbErr)
			},
			wantErrContains: "querying permissions for role:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.GetRole(ctx, roleID)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_ListRoles(t *testing.T) {
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

	tests := []struct {
		name            string
		filters         params.Filters
		sortItems       params.SortItems
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      []services.Role
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns every role with permissions on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListRolesSQL).WithArgs().WillReturnRows(expectRoleRows(pool, admin, readOnly))
				expectPermissionsFor(pool, expectedListRolePermissionsTwoSQL, admin, readOnly)
			},
			wantResult: []services.Role{admin, readOnly},
		},
		{
			name:      "issues an ORDER BY clause for a sorted request",
			sortItems: params.SortItems{{Field: "name", Direction: params.Ascending}},
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListRolesSortedSQL).WithArgs().WillReturnRows(expectRoleRows(pool, admin))
				expectPermissionsFor(pool, expectedListRolePermissionsSQL, admin)
			},
			wantResult: []services.Role{admin},
		},
		{
			name:    "issues a WHERE clause for a filtered request",
			filters: params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "Administrator", SetOperator: params.FilterAnd}}},
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListRolesFilteredSQL).WithArgs("Administrator").WillReturnRows(expectRoleRows(pool, admin))
				expectPermissionsFor(pool, expectedListRolePermissionsSQL, admin)
			},
			wantResult: []services.Role{admin},
		},
		{
			name:    "issues a WHERE clause for a numeric filter on id",
			filters: params.Filters{"id": {{Field: "id", Operator: params.GreaterThan, Value: "1", SetOperator: params.FilterAnd}}},
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListRolesFilteredByIDSQL).WithArgs("1").WillReturnRows(expectRoleRows(pool, readOnly))
				expectPermissionsFor(pool, expectedListRolePermissionsSQL, readOnly)
			},
			wantResult: []services.Role{readOnly},
		},
		{
			name:    "returns an error for an unknown filter field",
			filters: params.Filters{"nope": {{Field: "nope", Operator: params.Equals, Value: "x", SetOperator: params.FilterAnd}}},
			expectations: func(pool pgxmock.PgxPoolIface) {
			},
			wantErrContains: "unknown field",
		},
		{
			name:      "returns an error for an unknown sort field",
			sortItems: params.SortItems{{Field: "nope", Direction: params.Ascending}},
			expectations: func(pool pgxmock.PgxPoolIface) {
			},
			wantErrContains: "unknown field",
		},
		{
			name: "propagates the roles query database error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListRolesSQL).WithArgs().WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name: "wraps the permissions query error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListRolesSQL).WithArgs().WillReturnRows(expectRoleRows(pool, admin))
				pool.ExpectQuery(expectedListRolePermissionsSQL).WithArgs(admin.ID).WillReturnError(dbErr)
			},
			wantErr:         dbErr,
			wantErrContains: "querying permissions for roles:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.ListRoles(ctx, tt.filters, tt.sortItems)
			switch {
			case tt.wantErr != nil || tt.wantErrContains != "":
				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				if tt.wantErrContains != "" {
					assert.ErrorContains(t, err, tt.wantErrContains)
				}
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

// expectedListUsersSQL is the literal SQL the Store issues for the users query in
// ListUsers when no filters or sorts are supplied.
const expectedListUsersSQL = `SELECT id, sso_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, all_environments, eula_accepted, created_at, updated_at FROM users WHERE support_account = $1`

// expectedListUsersSortedSQL is the literal SQL issued when a single ascending
// sort on principal_name is supplied.
const expectedListUsersSortedSQL = expectedListUsersSQL + ` ORDER BY principal_name ASC`

// expectedListUsersFilteredSQL is the literal SQL issued when a single equality
// filter on first_name is supplied alongside the support_account exclusion.
const expectedListUsersFilteredSQL = `SELECT id, sso_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, all_environments, eula_accepted, created_at, updated_at FROM users WHERE support_account = $1 AND (first_name = $2)`

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

	tests := []struct {
		name            string
		filters         params.Filters
		sortItems       params.SortItems
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      []services.User
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns every user with associations on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				userRows := pool.NewRows(userRowColumns())
				addUserRow(userRows, user1)
				addUserRow(userRows, user2)
				pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnRows(userRows)

				pool.ExpectQuery(expectedRolesForUsersTwoSQL).WithArgs(user1ID.String(), user2ID.String()).WillReturnRows(
					pool.NewRows(userRoleRowColumns()).
						AddRow(user1ID, role1.ID, role1.Name, role1.Description, role1.CreatedAt, role1.UpdatedAt).
						AddRow(user2ID, role2.ID, role2.Name, role2.Description, role2.CreatedAt, role2.UpdatedAt),
				)
				pool.ExpectQuery(expectedListRolePermissionsTwoSQL).WithArgs(role1.ID, role2.ID).WillReturnRows(
					pool.NewRows(listRolePermissionRowColumns()).
						AddRow(role1.ID, permission1.ID, permission1.Authority, permission1.Name, permission1.CreatedAt, permission1.UpdatedAt),
				)
				pool.ExpectQuery(expectedETACForUsersTwoSQL).WithArgs(user1ID.String(), user2ID.String()).WillReturnRows(
					pool.NewRows(userETACRowColumns()).
						AddRow(etac1.ID, user1ID, etac1.EnvironmentID, etac1.CreatedAt, sql.NullTime{Time: etac1.UpdatedAt, Valid: true}),
				)
				pool.ExpectQuery(expectedAuthSecretsForUsersTwoSQL).WithArgs(user1ID.String(), user2ID.String()).WillReturnRows(
					pool.NewRows(userAuthSecretRowColumns()).
						AddRow(secret1.ID, user1ID, secret1.DigestMethod, sql.NullTime{Time: secret1.ExpiresAt, Valid: true}, secret1.TOTPActivated, secret1.CreatedAt, secret1.UpdatedAt),
				)
			},
			wantResult: []services.User{user1, user2},
		},
		{
			name:      "issues an ORDER BY clause for a sorted request",
			sortItems: params.SortItems{{Field: "principal_name", Direction: params.Ascending}},
			expectations: func(pool pgxmock.PgxPoolIface) {
				userRows := pool.NewRows(userRowColumns())
				addUserRow(userRows, user2)
				pool.ExpectQuery(expectedListUsersSortedSQL).WithArgs(false).WillReturnRows(userRows)

				pool.ExpectQuery(expectedRolesForUsersOneSQL).WithArgs(user2ID.String()).WillReturnRows(pool.NewRows(userRoleRowColumns()))
				pool.ExpectQuery(expectedETACForUsersOneSQL).WithArgs(user2ID.String()).WillReturnRows(pool.NewRows(userETACRowColumns()))
				pool.ExpectQuery(expectedAuthSecretsForUsersOneSQL).WithArgs(user2ID.String()).WillReturnRows(pool.NewRows(userAuthSecretRowColumns()))
			},
			wantResult: []services.User{{
				ID:              user2ID,
				SSOProviderID:   sql.NullInt32{Int32: 5, Valid: true},
				PrincipalName:   "sso-user",
				LastLogin:       lastLogin,
				AllEnvironments: true,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			}},
		},
		{
			name:    "issues a WHERE clause for a filtered request",
			filters: params.Filters{"first_name": {{Field: "first_name", Operator: params.Equals, Value: "Ada", SetOperator: params.FilterAnd}}},
			expectations: func(pool pgxmock.PgxPoolIface) {
				userRows := pool.NewRows(userRowColumns())
				addUserRow(userRows, user1)
				pool.ExpectQuery(expectedListUsersFilteredSQL).WithArgs(false, "Ada").WillReturnRows(userRows)

				pool.ExpectQuery(expectedRolesForUsersOneSQL).WithArgs(user1ID.String()).WillReturnRows(
					pool.NewRows(userRoleRowColumns()).
						AddRow(user1ID, role1.ID, role1.Name, role1.Description, role1.CreatedAt, role1.UpdatedAt),
				)
				pool.ExpectQuery(expectedListRolePermissionsSQL).WithArgs(role1.ID).WillReturnRows(
					pool.NewRows(listRolePermissionRowColumns()).
						AddRow(role1.ID, permission1.ID, permission1.Authority, permission1.Name, permission1.CreatedAt, permission1.UpdatedAt),
				)
				pool.ExpectQuery(expectedETACForUsersOneSQL).WithArgs(user1ID.String()).WillReturnRows(
					pool.NewRows(userETACRowColumns()).
						AddRow(etac1.ID, user1ID, etac1.EnvironmentID, etac1.CreatedAt, sql.NullTime{Time: etac1.UpdatedAt, Valid: true}),
				)
				pool.ExpectQuery(expectedAuthSecretsForUsersOneSQL).WithArgs(user1ID.String()).WillReturnRows(
					pool.NewRows(userAuthSecretRowColumns()).
						AddRow(secret1.ID, user1ID, secret1.DigestMethod, sql.NullTime{Time: secret1.ExpiresAt, Valid: true}, secret1.TOTPActivated, secret1.CreatedAt, secret1.UpdatedAt),
				)
			},
			wantResult: []services.User{user1},
		},
		{
			name: "returns an empty slice when no users match",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnRows(pool.NewRows(userRowColumns()))
			},
			wantResult: []services.User{},
		},
		{
			name:    "returns an error for an unknown filter field",
			filters: params.Filters{"nope": {{Field: "nope", Operator: params.Equals, Value: "x", SetOperator: params.FilterAnd}}},
			expectations: func(pool pgxmock.PgxPoolIface) {
			},
			wantErrContains: "unknown field",
		},
		{
			name:      "returns an error for an unknown sort field",
			sortItems: params.SortItems{{Field: "nope", Direction: params.Ascending}},
			expectations: func(pool pgxmock.PgxPoolIface) {
			},
			wantErrContains: "unknown field",
		},
		{
			name: "propagates the users query database error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name: "wraps the roles query error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				userRows := pool.NewRows(userRowColumns())
				addUserRow(userRows, user1)
				pool.ExpectQuery(expectedListUsersSQL).WithArgs(false).WillReturnRows(userRows)
				pool.ExpectQuery(expectedRolesForUsersOneSQL).WithArgs(user1ID.String()).WillReturnError(dbErr)
			},
			wantErr:         dbErr,
			wantErrContains: "querying roles for users:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.ListUsers(ctx, tt.filters, tt.sortItems)
			switch {
			case tt.wantErr != nil || tt.wantErrContains != "":
				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				if tt.wantErrContains != "" {
					assert.ErrorContains(t, err, tt.wantErrContains)
				}
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
