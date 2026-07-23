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
