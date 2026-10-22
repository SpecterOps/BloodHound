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

package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/specterops/bloodhound/server/identity/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetPermission(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		permission services.Permission
		err        error
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx                = context.Background()
		permissionID       = 7
		unexpectedErr      = errors.New("connection refused")
		expectedPermission = services.Permission{
			ID:        7,
			Authority: "app",
			Name:      "ManageProviders",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
	)

	tt := []testData{
		{
			name: "Success: permission is returned - 200",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetPermission(ctx, permissionID).Return(expectedPermission, nil)
			},
			expected: expected{permission: expectedPermission},
		},
		{
			name: "Error: missing permission is propagated - 404",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetPermission(ctx, permissionID).Return(services.Permission{}, services.ErrNoPermissionFound)
			},
			expected: expected{err: services.ErrNoPermissionFound},
		},
		{
			name: "Error: database query fails - 500",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetPermission(ctx, permissionID).Return(services.Permission{}, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})

			result, err := svc.GetPermission(ctx, permissionID)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.permission, result)
			}
		})
	}
}

func TestService_GetRole(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		role services.Role
		err  error
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx           = context.Background()
		roleID        = int32(3)
		unexpectedErr = errors.New("connection refused")
		expectedRole  = services.Role{
			ID:          3,
			Name:        "Administrator",
			Description: "Can manage the application",
			Permissions: []services.Permission{
				{ID: 1, Authority: "app", Name: "ManageProviders"},
			},
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
	)

	tt := []testData{
		{
			name: "Success: role is returned - 200",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetRole(ctx, roleID).Return(expectedRole, nil)
			},
			expected: expected{role: expectedRole},
		},
		{
			name: "Error: missing role is propagated - 404",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetRole(ctx, roleID).Return(services.Role{}, services.ErrNoRoleFound)
			},
			expected: expected{err: services.ErrNoRoleFound},
		},
		{
			name: "Error: database query fails - 500",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetRole(ctx, roleID).Return(services.Role{}, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})

			result, err := svc.GetRole(ctx, roleID)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.role, result)
			}
		})
	}
}

func TestService_ListRoles(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		roles []services.Role
		err   error
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx           = context.Background()
		queryFilters  = params.Filters{"name": {{Operator: params.Equals, Value: "Administrator"}}}
		sortItems     = params.SortItems{{Field: "name", Direction: params.Ascending}}
		unexpectedErr = errors.New("connection refused")
		expectedRoles = []services.Role{
			{
				ID:          3,
				Name:        "Administrator",
				Description: "Can manage the application",
				Permissions: []services.Permission{
					{ID: 1, Authority: "app", Name: "ManageProviders"},
				},
				CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		}
	)

	tt := []testData{
		{
			name: "Success: roles are returned - 200",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListRoles(ctx, queryFilters, sortItems).Return(expectedRoles, nil)
			},
			expected: expected{roles: expectedRoles},
		},
		{
			name: "Success: no roles match - 200",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListRoles(ctx, queryFilters, sortItems).Return([]services.Role{}, nil)
			},
			expected: expected{roles: []services.Role{}},
		},
		{
			name: "Error: database query fails - 500",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListRoles(ctx, queryFilters, sortItems).Return(nil, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})

			result, err := svc.ListRoles(ctx, queryFilters, sortItems)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.roles, result)
			}
		})
	}
}

func TestService_ListPermissions(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		permissions []services.Permission
		err         error
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx                 = context.Background()
		queryFilters        = params.Filters{"authority": {{Operator: params.Equals, Value: "app"}}}
		sortItems           = params.SortItems{{Field: "name", Direction: params.Ascending}}
		unexpectedErr       = errors.New("connection refused")
		expectedPermissions = []services.Permission{{ID: 7, Authority: "app", Name: "ManageProviders"}}
	)

	tt := []testData{
		{
			name: "Success: permissions are returned",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListPermissions(ctx, queryFilters, sortItems).Return(expectedPermissions, nil)
			},
			expected: expected{permissions: expectedPermissions},
		},
		{
			name: "Success: no permissions match",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListPermissions(ctx, queryFilters, sortItems).Return([]services.Permission{}, nil)
			},
			expected: expected{permissions: []services.Permission{}},
		},
		{
			name: "Error: database query fails",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListPermissions(ctx, queryFilters, sortItems).Return(nil, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})

			result, err := svc.ListPermissions(ctx, queryFilters, sortItems)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.permissions, result)
			}
		})
	}
}

func TestService_ListUsers(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		users []services.User
		err   error
	}

	type testData struct {
		name       string
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx           = context.Background()
		queryFilters  = params.Filters{"email_address": {{Operator: params.Equals, Value: "ada@example.com"}}}
		sortItems     = params.SortItems{{Field: "principal_name", Direction: params.Ascending}}
		unexpectedErr = errors.New("connection refused")
		expectedUsers = []services.User{
			{
				PrincipalName: "ada",
				EmailAddress:  sql.NullString{String: "ada@example.com", Valid: true},
				FirstName:     sql.NullString{String: "Ada", Valid: true},
				Roles: []services.Role{
					{ID: 3, Name: "Administrator"},
				},
				CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		}
	)

	tt := []testData{
		{
			name: "Success: users are returned - 200",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListUsers(ctx, queryFilters, sortItems).Return(expectedUsers, nil)
			},
			expected: expected{users: expectedUsers},
		},
		{
			name: "Success: no users match - 200",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListUsers(ctx, queryFilters, sortItems).Return([]services.User{}, nil)
			},
			expected: expected{users: []services.User{}},
		},
		{
			name: "Error: database query fails - 500",
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListUsers(ctx, queryFilters, sortItems).Return(nil, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})

			result, err := svc.ListUsers(ctx, queryFilters, sortItems)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.users, result)
			}
		})
	}
}

func TestService_ListActiveUsersMinimal(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		users []services.UserMinimal
		err   error
	}

	type testData struct {
		name       string
		sortItems  params.SortItems
		setupMocks func(mock mock)
		expected   expected
	}

	var (
		ctx              = context.Background()
		queryFilters     = params.Filters{"first_name": {{Operator: params.ApproximatelyEquals, Value: "ada"}}}
		requestedSort    = params.SortItems{{Field: "first_name", Direction: params.Descending}}
		tieBrokenSort    = params.SortItems{{Field: "first_name", Direction: params.Descending}, {Field: "email_address", Direction: params.Ascending}}
		emailSort        = params.SortItems{{Field: "last_name", Direction: params.Ascending}, {Field: "email_address", Direction: params.Descending}}
		defaultSort      = params.SortItems{{Field: "email_address", Direction: params.Ascending}}
		unexpectedErr    = errors.New("connection refused")
		expectedUserList = []services.UserMinimal{
			{
				ID:           uuid.Must(uuid.FromString("4b7c3f0e-2f7a-4d8e-9a43-6d1b1a2c3e4f")),
				EmailAddress: sql.NullString{String: "ada@example.com", Valid: true},
				FirstName:    sql.NullString{String: "Ada", Valid: true},
				LastName:     sql.NullString{String: "Lovelace", Valid: true},
			},
		}
	)

	tt := []testData{
		{
			name:      "Success: email_address ascending is appended as a tie-breaker to the requested sort - 200",
			sortItems: requestedSort,
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListActiveUsersMinimal(ctx, queryFilters, tieBrokenSort).Return(expectedUserList, nil)
			},
			expected: expected{users: expectedUserList},
		},
		{
			name:      "Success: requested sort already containing email_address is passed through unchanged - 200",
			sortItems: emailSort,
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListActiveUsersMinimal(ctx, queryFilters, emailSort).Return(expectedUserList, nil)
			},
			expected: expected{users: expectedUserList},
		},
		{
			name:      "Success: users are sorted by email_address ascending when no sort is requested - 200",
			sortItems: params.SortItems{},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListActiveUsersMinimal(ctx, queryFilters, defaultSort).Return(expectedUserList, nil)
			},
			expected: expected{users: expectedUserList},
		},
		{
			name:      "Success: users are sorted by email_address ascending when sort is nil - 200",
			sortItems: nil,
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListActiveUsersMinimal(ctx, queryFilters, defaultSort).Return(expectedUserList, nil)
			},
			expected: expected{users: expectedUserList},
		},
		{
			name:      "Success: no users match - 200",
			sortItems: requestedSort,
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListActiveUsersMinimal(ctx, queryFilters, tieBrokenSort).Return([]services.UserMinimal{}, nil)
			},
			expected: expected{users: []services.UserMinimal{}},
		},
		{
			name:      "Error: database query fails - 500",
			sortItems: requestedSort,
			setupMocks: func(mock mock) {
				mock.database.EXPECT().ListActiveUsersMinimal(ctx, queryFilters, tieBrokenSort).Return(nil, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})

			result, err := svc.ListActiveUsersMinimal(ctx, queryFilters, testCase.sortItems)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected.users, result)
			}
		})
	}
}
