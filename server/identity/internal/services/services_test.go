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
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/specterops/bloodhound/server/identity/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetPermission(t *testing.T) {
	var (
		ctx           = context.Background()
		permissionID  = 7
		unexpectedErr = errors.New("connection refused")
		expected      = services.Permission{
			ID:        7,
			Authority: "app",
			Name:      "ManageProviders",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
	)

	tests := []struct {
		name       string
		dbResult   services.Permission
		dbErr      error
		wantResult services.Permission
		wantErr    error
	}{
		{
			name:       "returns the permission on success",
			dbResult:   expected,
			wantResult: expected,
		},
		{
			name:    "propagates ErrNoPermissionFound",
			dbErr:   services.ErrNoPermissionFound,
			wantErr: services.ErrNoPermissionFound,
		},
		{
			name:    "propagates unexpected database errors",
			dbErr:   unexpectedErr,
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			databaseMock.EXPECT().GetPermission(ctx, permissionID).Return(tt.dbResult, tt.dbErr)

			result, err := svc.GetPermission(ctx, permissionID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestService_GetRole(t *testing.T) {
	var (
		ctx           = context.Background()
		roleID        = int32(3)
		unexpectedErr = errors.New("connection refused")
		expected      = services.Role{
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

	tests := []struct {
		name       string
		dbResult   services.Role
		dbErr      error
		wantResult services.Role
		wantErr    error
	}{
		{
			name:       "returns the role on success",
			dbResult:   expected,
			wantResult: expected,
		},
		{
			name:    "propagates ErrNoRoleFound",
			dbErr:   services.ErrNoRoleFound,
			wantErr: services.ErrNoRoleFound,
		},
		{
			name:    "propagates unexpected database errors",
			dbErr:   unexpectedErr,
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			databaseMock.EXPECT().GetRole(ctx, roleID).Return(tt.dbResult, tt.dbErr)

			result, err := svc.GetRole(ctx, roleID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestService_ListRoles(t *testing.T) {
	var (
		ctx           = context.Background()
		queryFilters  = params.Filters{"name": {{Operator: params.Equals, Value: "Administrator"}}}
		sortItems     = params.SortItems{{Field: "name", Direction: params.Ascending}}
		unexpectedErr = errors.New("connection refused")
		expected      = []services.Role{
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

	tests := []struct {
		name       string
		dbResult   []services.Role
		dbErr      error
		wantResult []services.Role
		wantErr    error
	}{
		{
			name:       "returns the roles on success",
			dbResult:   expected,
			wantResult: expected,
		},
		{
			name:       "returns an empty slice when no roles match",
			dbResult:   []services.Role{},
			wantResult: []services.Role{},
		},
		{
			name:    "propagates unexpected database errors",
			dbErr:   unexpectedErr,
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			databaseMock.EXPECT().ListRoles(ctx, queryFilters, sortItems).Return(tt.dbResult, tt.dbErr)

			result, err := svc.ListRoles(ctx, queryFilters, sortItems)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
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

	tests := []testData{
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

	for _, testCase := range tests {
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
