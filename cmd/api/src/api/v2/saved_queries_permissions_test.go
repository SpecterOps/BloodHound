// Copyright 2024 Specter Ops, Inc.
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

package v2_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

var (
	ErrMockDatabaseError = errors.New("mockDatabaseError")
)

// helper to build scopes more succinctly
func newSavedQueryScope(owned, public, shared bool) database.SavedQueryScopeMap {
	return database.SavedQueryScopeMap{
		model.SavedQueryScopeOwned:  owned,
		model.SavedQueryScopePublic: public,
		model.SavedQueryScopeShared: shared,
	}
}

func TestResources_ShareSavedQueriesPermissions_CanUpdateSavedQueriesPermission(t *testing.T) {
	adminUser := model.User{
		Roles: model.Roles{
			{
				Name:        auth.RoleAdministrator,
				Permissions: model.Permissions{auth.Permissions().AuthManageSelf},
			},
		},
	}

	nonAdminUserId1, err := uuid.NewV4()
	require.Nil(t, err)
	nonAdminUser1 := model.User{
		Roles: model.Roles{
			{
				Name: "nonAdminUser1",
			},
		},
		Unique: model.Unique{
			ID: nonAdminUserId1,
		},
	}

	nonAdminUserId2, err := uuid.NewV4()
	require.Nil(t, err)
	nonAdminUser2 := model.User{
		Roles: model.Roles{
			{
				Name: "nonAdminUser2",
			},
		},
		Unique: model.Unique{
			ID: nonAdminUserId2,
		},
	}

	userRoleUserID, err := uuid.NewV4()
	require.Nil(t, err)
	userRoleUser := model.User{
		Roles: model.Roles{
			{
				Name:        auth.RoleUser,
				Permissions: model.Permissions{auth.Permissions().AuthManageSelf},
			},
		},
		Unique: model.Unique{
			ID: userRoleUserID,
		},
	}

	powerUserID, err := uuid.NewV4()
	require.Nil(t, err)
	powerUser := model.User{
		Roles: model.Roles{
			{
				Name:        auth.RolePowerUser,
				Permissions: model.Permissions{auth.Permissions().AuthManageSelf},
			},
		},
		Unique: model.Unique{
			ID: powerUserID,
		},
	}

	tests := []struct {
		name                    string
		comment                 string
		user                    model.User
		savedQueryBelongsToUser bool
		payload                 v2.SavedQueryPermissionRequest
		scope                   database.SavedQueryScopeMap
		expectedErr             error
	}{
		// Non-admin owned queries
		{
			name:                    "Non-admin owned, query doesn't belong to user error",
			comment:                 "Non-privileged user cannot update non-owned, non-public, non-shared query",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, false, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Non-admin owned, query shared to self error",
			comment:                 "Non-privileged user cannot share their own private query to themselves",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: v2.ErrInvalidSelfShare,
		},
		{
			name:                    "Non-admin owned, shared query shared to user(s)",
			comment:                 "Non-privileged user can share their own already-shared private query to others",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, true),
			expectedErr: nil,
		},
		{
			name:                    "Non-admin owned, shared query set to public",
			comment:                 "Non-privileged user can make own shared query public",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(true, false, true),
			expectedErr: nil,
		},
		{
			name:                    "Non-admin owned, shared query set to private",
			comment:                 "Non-privileged user can make own shared query private",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, true),
			expectedErr: nil,
		},
		{
			name:                    "Non-admin owned, private query shared to user(s)",
			comment:                 "Non-privileged user can share their own private query",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "Non-admin owned, private query set to public",
			comment:                 "Non-privileged user can make own private query public",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "Non-admin owned, private query set to private",
			comment:                 "Non-privileged user can keep own private query private (no-op)",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "Non-admin owned, public query shared to user(s) error",
			comment:                 "Non-privileged user cannot share their own public query to specific users",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Non-admin owned, public query set to private error",
			comment:                 "Non-privileged user cannot make own public query private",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Non-admin owned, public query set to public error",
			comment:                 "Non-privileged user cannot 're-set' own public query (forbidden state change rules)",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Non-admin not-owned, public query cannot be made private",
			comment:                 "Non-privileged user cannot make someone else's public query private",
			user:                    nonAdminUser1,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, true, true),
			expectedErr: v2.ErrForbidden,
		},

		// Admin (non-admin owned) queries
		{
			name:                    "Admin (non-admin owned), public query shared to user(s) incorrectly error",
			comment:                 "Admin cannot share a public query to specific users when they don't own it",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, true, false),
			expectedErr: v2.ErrInvalidPublicShare,
		},
		{
			name:                    "Admin (non-admin owned), private query set to public error",
			comment:                 "Admin cannot make someone else's private query public",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(false, false, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Admin (non-admin owned), private query shared to user(s) error",
			comment:                 "Admin cannot share someone else's private query to users",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, false, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Admin (non-admin owned), private query set to private error",
			comment:                 "Admin cannot modify someone else's private query permissions at all",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, false, false),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Admin (non-admin owned), shared query set to public error",
			comment:                 "Admin cannot make someone else's shared query public",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(false, false, true),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Admin (non-admin owned), shared query shared to user(s) error",
			comment:                 "Admin cannot change shares of someone else's shared query",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, false, true),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Admin (non-admin owned), shared query set to private error",
			comment:                 "Admin cannot make someone else's shared query private",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(false, false, true),
			expectedErr: v2.ErrForbidden,
		},
		{
			name:                    "Admin (non-admin owned), public query set to public",
			comment:                 "Admin can leave someone else's public query public (noop)",
			user:                    adminUser,
			savedQueryBelongsToUser: false,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(false, true, false),
			expectedErr: nil,
		},

		// Admin owned queries
		{
			name:                    "Admin-owned, query shared to self error",
			comment:                 "Admin cannot share their own query to themselves",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{adminUser.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: v2.ErrInvalidSelfShare,
		},
		{
			name:                    "Admin-owned, shared query shared to user(s)",
			comment:                 "Admin can share their own shared query to others",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, true),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, shared query set to public",
			comment:                 "Admin can make their own shared query public",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(true, false, true),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, shared query set to private",
			comment:                 "Admin can make their own shared query private",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, true),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, private query shared to user(s)",
			comment:                 "Admin can share their own private query to others",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, private query set to public",
			comment:                 "Admin can make their own private query public",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, private query set to private",
			comment:                 "Admin can leave private query private (noop)",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, public query set to public",
			comment:                 "Admin can leave own public query public",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  true,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, public query set to private",
			comment:                 "Admin can make own public query private",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: nil,
		},
		{
			name:                    "Admin-owned, public query shared to user(s) incorrectly error",
			comment:                 "Admin cannot share own public query to specific users while it is public",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: v2.ErrInvalidPublicShare,
		},
		{
			name:                    "Admin-owned, public & shared query set to private (no shares)",
			comment:                 "Admin can make own public+shared query private with no shares",
			user:                    adminUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, true),
			expectedErr: nil,
		},

		// User role owned
		{
			name:                    "User owned, query shared to self error",
			comment:                 "RoleUser cannot share own query to themselves",
			user:                    userRoleUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{userRoleUser.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: v2.ErrInvalidSelfShare,
		},
		{
			name:                    "User-owned, public query set to private (no shares)",
			comment:                 "RoleUser can make own public query private with no shares",
			user:                    userRoleUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: nil,
		},
		{
			name:                    "User-owned, non-public query set to private (no shares)",
			comment:                 "RoleUser can leave own non-public query private",
			user:                    userRoleUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "User-owned, public & shared query set to private (no shares)",
			comment:                 "RoleUser can make own public+shared query private with no shares",
			user:                    userRoleUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, true),
			expectedErr: nil,
		},
		{
			name:                    "User-owned, public & shared query shared to other users is allowed",
			comment:                 "RoleUser with privileged role can still pass CanUpdate check when public & shared",
			user:                    userRoleUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
				Public:  true, // upper-layer HTTP handler will reject this combination
			},
			scope:       newSavedQueryScope(true, true, true),
			expectedErr: nil,
		},

		// PowerUser owned
		{
			name:                    "PowerUser-owned, query shared to self error",
			comment:                 "PowerUser cannot share own query to themselves",
			user:                    powerUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{powerUser.ID},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: v2.ErrInvalidSelfShare,
		},
		{
			name:                    "PowerUser-owned, public query set to private (no shares)",
			comment:                 "PowerUser can make own public query private with no shares",
			user:                    powerUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, false),
			expectedErr: nil,
		},
		{
			name:                    "PowerUser-owned, non-public query set to private (no shares)",
			comment:                 "PowerUser can leave own non-public query private",
			user:                    powerUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, false, false),
			expectedErr: nil,
		},
		{
			name:                    "PowerUser-owned, public & shared query set to private (no shares)",
			comment:                 "PowerUser can make own public+shared query private with no shares",
			user:                    powerUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{},
				Public:  false,
			},
			scope:       newSavedQueryScope(true, true, true),
			expectedErr: nil,
		},
		{
			name:                    "PowerUser-owned, public & shared query shared to other users is allowed",
			comment:                 "PowerUser with privileged role passes CanUpdate when public & shared",
			user:                    powerUser,
			savedQueryBelongsToUser: true,
			payload: v2.SavedQueryPermissionRequest{
				UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
				Public:  true, // HTTP handler still blocks this, but CanUpdate returns nil
			},
			scope:       newSavedQueryScope(true, true, true),
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			if tc.comment != "" {
				t.Log(tc.comment)
			}

			err := v2.CanUpdateSavedQueriesPermission(tc.user, tc.savedQueryBelongsToUser, tc.payload, tc.scope)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestResources_ShareSavedQueriesPermissions_NonAdmin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocks.NewMockDatabase(mockCtrl)
	resources := v2.Resources{DB: mockDB}

	endpointPattern := "/api/v2/saved-queries/%s/permissions"
	savedQueryID := "1"

	userID := must.NewUUIDv4()
	userID2 := must.NewUUIDv4()
	userID3 := must.NewUUIDv4()

	type testCase struct {
		name               string
		savedQueryID       string
		buildRequest       func(t *testing.T, url string) *http.Request
		setupMocks         func()
		expectedStatus     int
		expectedBodySubstr string
		verifySuccessBody  func(t *testing.T, body []byte)
	}

	tests := []testCase{
		// -------------------------------
		// Early input validation / wiring
		// -------------------------------
		{
			name:         "No associated user found (no context)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2},
					Public:  false,
				}

				req, err := http.NewRequest(http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks:         nil,
			expectedStatus:     http.StatusBadRequest,
			expectedBodySubstr: "No associated user found",
		},
		{
			name:         "Shared Query ID is Invalid.",
			savedQueryID: "malformed",
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks:         nil,
			expectedStatus:     http.StatusBadRequest,
			expectedBodySubstr: "id is malformed",
		},
		{
			name:         "Invalid JSON payload",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				// Same trick as your original test: an array of strings, which cannot be
				// unmarshalled into SavedQueryPermissionRequest.
				payload := []string{
					`{"Public": false}`,
					`Invalid JSON`,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks:         nil,
			expectedStatus:     http.StatusBadRequest,
			expectedBodySubstr: "could not decode limited payload request into value",
		},

		// -------------------------------
		// DB error paths before helper
		// -------------------------------
		{
			name:         "Query does not exist (SavedQueryBelongsToUser -> ErrNotFound)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, database.ErrNotFound)
			},
			expectedStatus:     http.StatusNotFound,
			expectedBodySubstr: "Query does not exist",
		},
		{
			name:         "Database error from SavedQueryBelongsToUser",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, context.DeadlineExceeded)
			},
			expectedStatus:     http.StatusInternalServerError,
			expectedBodySubstr: "request timed out",
		},
		{
			name:         "Database error from GetScopeForSavedQuery",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: false,
					}, ErrMockDatabaseError)
			},
			expectedStatus:     http.StatusInternalServerError,
			expectedBodySubstr: "an internal error has occurred that is preventing the service from servicing this request",
		},

		// -------------------------------
		// CanUpdateSavedQueriesPermission error paths
		// -------------------------------
		{
			name:         "ErrInvalidSelfShare -> 400 (Cannot share query to self)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				// owned, not public, not shared -> helper returns ErrInvalidSelfShare
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: false,
					}, nil)
			},
			expectedStatus:     http.StatusBadRequest,
			expectedBodySubstr: "Cannot share query to self",
		},
		{
			name:         "ErrForbidden from helper (User attempts to share already public query)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				// owned + public -> non-privileged user gets ErrForbidden
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: true,
						model.SavedQueryScopeShared: false,
					}, nil)
			},
			expectedStatus:     http.StatusForbidden,
			expectedBodySubstr: "Forbidden",
		},
		{
			name:         "Query set to public and shared to user(s) at same time error",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2},
					Public:  true,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks:         nil,
			expectedStatus:     http.StatusBadRequest,
			expectedBodySubstr: "Public cannot be true while user_ids is populated",
		},

		// -------------------------------
		// Happy-path non-admin scenarios
		// -------------------------------
		{
			name:         "Shared query shared to user(s) (and user(s) that already have query shared with them) -> success (201, 2 records)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2, userID3},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: false,
					}, nil)

				mockDB.EXPECT().
					CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userID2, userID3).
					Return([]model.SavedQueriesPermissions{
						{
							QueryID:        1,
							Public:         false,
							SharedToUserID: database.NullUUID(userID2),
						},
						{
							QueryID:        1,
							Public:         false,
							SharedToUserID: database.NullUUID(userID3),
						},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			verifySuccessBody: func(t *testing.T, body []byte) {
				var wrapper struct {
					Data v2.ShareSavedQueriesResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &wrapper))

				require.Len(t, wrapper.Data, 2)
				ids := []uuid.UUID{
					wrapper.Data[0].SharedToUserID.UUID,
					wrapper.Data[1].SharedToUserID.UUID,
				}
				assert.ElementsMatch(t, []uuid.UUID{userID2, userID3}, ids)
			},
		},
		{
			name:         "Shared query shared to user(s) success",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: true,
					}, nil)

				mockDB.EXPECT().
					CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userID2).
					Return([]model.SavedQueriesPermissions{
						{
							QueryID:        int64(1),
							Public:         false,
							SharedToUserID: database.NullUUID(userID2),
						},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			verifySuccessBody: func(t *testing.T, body []byte) {
				var wrapper struct {
					Data v2.ShareSavedQueriesResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &wrapper))

				require.Len(t, wrapper.Data, 1)
				ids := []uuid.UUID{
					wrapper.Data[0].SharedToUserID.UUID,
				}
				assert.ElementsMatch(t, []uuid.UUID{userID2}, ids)
			},
		},
		{
			name:         "Shared query set to public success",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{},
					Public:  true,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: true,
					}, nil)

				mockDB.EXPECT().CreateSavedQueryPermissionToPublic(gomock.Any(), int64(1)).Return(model.SavedQueriesPermissions{
					QueryID: 1,
					SharedToUserID: uuid.NullUUID{
						UUID:  uuid.UUID{},
						Valid: false,
					},
					Public: true,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			verifySuccessBody: func(t *testing.T, body []byte) {
				var wrapper struct {
					Data v2.ShareSavedQueriesResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &wrapper))

				require.Equal(t, v2.ShareSavedQueriesResponse{
					{
						SharedToUserID: uuid.NullUUID{
							UUID:  uuid.UUID{},
							Valid: false,
						},
						QueryID: 1,
						Public:  true,
					},
				}, wrapper.Data)
			},
		},
		{
			name:         "Private query shared to user(s) success",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{userID2, userID3},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: false,
					}, nil)
				mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userID2, userID3).Return([]model.SavedQueriesPermissions{
					{
						QueryID:        int64(1),
						Public:         false,
						SharedToUserID: database.NullUUID(userID2),
					},
					{
						QueryID:        int64(1),
						Public:         false,
						SharedToUserID: database.NullUUID(userID3),
					},
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			verifySuccessBody: func(t *testing.T, body []byte) {
				var wrapper struct {
					Data v2.ShareSavedQueriesResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &wrapper))

				require.Equal(t, v2.ShareSavedQueriesResponse{
					{
						SharedToUserID: database.NullUUID(userID2),
						QueryID:        1,
						Public:         false,
					},
					{
						SharedToUserID: database.NullUUID(userID3),
						QueryID:        1,
						Public:         false,
					},
				}, wrapper.Data)
			},
		},
		{
			name:         "Private query set to public success (201)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{},
					Public:  true,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: false,
					}, nil)
				mockDB.EXPECT().
					CreateSavedQueryPermissionToPublic(gomock.Any(), int64(1)).
					Return(model.SavedQueriesPermissions{
						QueryID:        1,
						SharedToUserID: uuid.NullUUID{},
						Public:         true,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			verifySuccessBody: func(t *testing.T, body []byte) {
				var wrapper struct {
					Data v2.ShareSavedQueriesResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &wrapper))

				require.Len(t, wrapper.Data, 1)
				assert.True(t, wrapper.Data[0].Public)
				assert.False(t, wrapper.Data[0].SharedToUserID.Valid)
			},
		},
		{
			name:         "Shared query set to private (204, DeleteSavedQueryPermissionsForUsers)",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: true,
					}, nil)
				mockDB.EXPECT().
					DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:         "Private query set to private success",
			savedQueryID: savedQueryID,
			buildRequest: func(t *testing.T, url string) *http.Request {
				payload := v2.SavedQueryPermissionRequest{
					UserIDs: []uuid.UUID{},
					Public:  false,
				}

				req, err := http.NewRequestWithContext(createContextWithOwnerId(userID), http.MethodPut, url, must.MarshalJSONReader(payload))
				require.NoError(t, err)
				req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				return req
			},
			setupMocks: func() {
				mockDB.EXPECT().
					SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil)
				mockDB.EXPECT().
					GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(database.SavedQueryScopeMap{
						model.SavedQueryScopeOwned:  true,
						model.SavedQueryScopePublic: false,
						model.SavedQueryScopeShared: false,
					}, nil)
				mockDB.EXPECT().
					DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods(http.MethodPut)

			url := fmt.Sprintf(endpointPattern, tc.savedQueryID)
			req := tc.buildRequest(t, url)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedStatus, resp.Code, "unexpected status code")

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tc.expectedBodySubstr != "" {
				assert.Contains(t, string(bodyBytes), tc.expectedBodySubstr)
			}

			if tc.verifySuccessBody != nil {
				tc.verifySuccessBody(t, bodyBytes)
			}
		})
	}
}

func TestResources_ShareSavedQueriesPermissions_SavingPermissionsErrors(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/%s/permissions"
	savedQueryId := "1"
	userId, err := uuid.NewV4()
	require.Nil(t, err)
	userId2, err := uuid.NewV4()
	require.Nil(t, err)
	userId3, err := uuid.NewV4()
	require.Nil(t, err)

	payload := v2.SavedQueryPermissionRequest{
		UserIDs: []uuid.UUID{userId2, userId3},
		Public:  false,
	}

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeShared: false,
	}, nil)
	mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userId2, userId3).Return(nil, fmt.Errorf("Error!"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResources_ShareSavedQueriesPermissions_Admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocks.NewMockDatabase(mockCtrl)
	resources := v2.Resources{DB: mockDB}

	endpointPattern := "/api/v2/saved-queries/%s/permissions"
	savedQueryID := "1"

	adminUserID := must.NewUUIDv4()
	nonAdminUserID := must.NewUUIDv4()
	nonAdminUserID2 := must.NewUUIDv4()

	type step struct {
		name               string
		buildRequest       func(t *testing.T, url string) *http.Request
		setupMocks         func()
		expectedStatus     int
		expectedBodySubstr string
		verifySuccessBody  func(t *testing.T, body []byte)
	}

	type testCase struct {
		name         string
		savedQueryID string
		steps        []step
	}

	tests := []testCase{
		{
			name:         "Admin, non-owned public query set to private (DeleteSavedQueryPermissionsForUsers)",
			savedQueryID: savedQueryID,
			steps: []step{
				{
					name: "public -> private",
					buildRequest: func(t *testing.T, url string) *http.Request {
						payload := v2.SavedQueryPermissionRequest{
							UserIDs: []uuid.UUID{},
							Public:  false,
						}

						req, err := http.NewRequestWithContext(
							createContextWithAdminOwnerId(adminUserID),
							http.MethodPut,
							url,
							must.MarshalJSONReader(payload),
						)
						require.NoError(t, err)
						req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
						return req
					},
					setupMocks: func() {
						mockDB.EXPECT().
							SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(false, nil)
						mockDB.EXPECT().
							GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(database.SavedQueryScopeMap{
								model.SavedQueryScopeOwned:  false,
								model.SavedQueryScopePublic: true,
								model.SavedQueryScopeShared: false,
							}, nil)
						mockDB.EXPECT().
							DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).
							Return(nil)
					},
					expectedStatus: http.StatusNoContent,
				},
			},
		},
		{
			name:         "Admin-owned public query set to public (success, 204)",
			savedQueryID: savedQueryID,
			steps: []step{
				{
					name: "noop public -> public",
					buildRequest: func(t *testing.T, url string) *http.Request {
						payload := v2.SavedQueryPermissionRequest{
							UserIDs: []uuid.UUID{},
							Public:  true,
						}

						req, err := http.NewRequestWithContext(
							createContextWithAdminOwnerId(adminUserID),
							http.MethodPut,
							url,
							must.MarshalJSONReader(payload),
						)
						require.NoError(t, err)
						req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
						return req
					},
					setupMocks: func() {
						mockDB.EXPECT().
							SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(true, nil)
						mockDB.EXPECT().
							GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(database.SavedQueryScopeMap{
								model.SavedQueryScopeOwned:  true,
								model.SavedQueryScopePublic: true,
								model.SavedQueryScopeShared: false,
							}, nil)
						// CanUpdateSavedQueriesPermission returns nil; handler sees already-public
						// and just writes 204 with no additional DB calls.
					},
					expectedStatus: http.StatusNoContent,
				},
			},
		},
		{
			name:         "Admin-owned public query attempted to be shared to users in one step -> ErrInvalidPublicShare mapped to 400",
			savedQueryID: savedQueryID,
			steps: []step{
				{
					name: "public shared -> invalid",
					buildRequest: func(t *testing.T, url string) *http.Request {
						payload := v2.SavedQueryPermissionRequest{
							UserIDs: []uuid.UUID{nonAdminUserID},
							Public:  false,
						}

						req, err := http.NewRequestWithContext(
							createContextWithAdminOwnerId(adminUserID),
							http.MethodPut,
							url,
							must.MarshalJSONReader(payload),
						)
						require.NoError(t, err)
						req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
						return req
					},
					setupMocks: func() {
						mockDB.EXPECT().
							SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(true, nil)
						mockDB.EXPECT().
							GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(database.SavedQueryScopeMap{
								model.SavedQueryScopeOwned:  true,
								model.SavedQueryScopePublic: true,
								model.SavedQueryScopeShared: false,
							}, nil)
						// CanUpdateSavedQueriesPermission will return ErrInvalidPublicShare
					},
					expectedStatus:     http.StatusBadRequest,
					expectedBodySubstr: "Public query cannot be shared to users. You must set your query to private first",
				},
			},
		},
		{
			name:         "Admin-owned public query shared to users in two steps -> success",
			savedQueryID: savedQueryID,
			steps: []step{
				{
					name: "step1: public query -> private query (no shares)",
					buildRequest: func(t *testing.T, url string) *http.Request {
						// First turn public query into private with no shares
						payload := v2.SavedQueryPermissionRequest{
							UserIDs: []uuid.UUID{},
							Public:  false,
						}

						req, err := http.NewRequestWithContext(
							createContextWithAdminOwnerId(adminUserID),
							http.MethodPut,
							url,
							must.MarshalJSONReader(payload),
						)
						require.NoError(t, err)
						req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
						return req
					},
					setupMocks: func() {
						mockDB.EXPECT().
							SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(true, nil)
						mockDB.EXPECT().
							GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(database.SavedQueryScopeMap{
								model.SavedQueryScopeOwned:  true,
								model.SavedQueryScopePublic: true,
								model.SavedQueryScopeShared: false,
							}, nil)
						mockDB.EXPECT().
							DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).
							Return(nil)
					},
					expectedStatus: http.StatusNoContent,
				},
				{
					name: "step2: private query -> shared to 2 users",
					buildRequest: func(t *testing.T, url string) *http.Request {
						payload := v2.SavedQueryPermissionRequest{
							UserIDs: []uuid.UUID{nonAdminUserID, nonAdminUserID2},
							Public:  false,
						}

						req, err := http.NewRequestWithContext(
							createContextWithAdminOwnerId(adminUserID),
							http.MethodPut,
							url,
							must.MarshalJSONReader(payload),
						)
						require.NoError(t, err)
						req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
						return req
					},
					setupMocks: func() {
						mockDB.EXPECT().
							SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(true, nil)
						mockDB.EXPECT().
							GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(database.SavedQueryScopeMap{
								model.SavedQueryScopeOwned:  true,
								model.SavedQueryScopePublic: false,
								model.SavedQueryScopeShared: false,
							}, nil)
						mockDB.EXPECT().
							CreateSavedQueryPermissionsToUsers(
								gomock.Any(),
								gomock.Any(),
								nonAdminUserID,
								nonAdminUserID2,
							).
							Return([]model.SavedQueriesPermissions{
								{
									QueryID:        1,
									Public:         false,
									SharedToUserID: database.NullUUID(nonAdminUserID),
								},
								{
									QueryID:        1,
									Public:         false,
									SharedToUserID: database.NullUUID(nonAdminUserID2),
								},
							}, nil)
					},
					expectedStatus: http.StatusCreated,
					verifySuccessBody: func(t *testing.T, body []byte) {
						var wrapper struct {
							Data v2.ShareSavedQueriesResponse `json:"data"`
						}
						require.NoError(t, json.Unmarshal(body, &wrapper))

						require.Len(t, wrapper.Data, 2)

						ids := []uuid.UUID{
							wrapper.Data[0].SharedToUserID.UUID,
							wrapper.Data[1].SharedToUserID.UUID,
						}
						assert.ElementsMatch(t, []uuid.UUID{nonAdminUserID, nonAdminUserID2}, ids)
					},
				},
			},
		},
		{
			name:         "Admin-owned private query shared to users (201, two records)",
			savedQueryID: savedQueryID,
			steps: []step{
				{
					name: "single step: private -> shared",
					buildRequest: func(t *testing.T, url string) *http.Request {
						payload := v2.SavedQueryPermissionRequest{
							UserIDs: []uuid.UUID{nonAdminUserID, nonAdminUserID2},
							Public:  false,
						}

						req, err := http.NewRequestWithContext(
							createContextWithAdminOwnerId(adminUserID),
							http.MethodPut,
							url,
							must.MarshalJSONReader(payload),
						)
						require.NoError(t, err)
						req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
						return req
					},
					setupMocks: func() {
						mockDB.EXPECT().
							SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(true, nil)
						mockDB.EXPECT().
							GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(database.SavedQueryScopeMap{
								model.SavedQueryScopeOwned:  true,
								model.SavedQueryScopePublic: false,
								model.SavedQueryScopeShared: false,
							}, nil)
						mockDB.EXPECT().
							CreateSavedQueryPermissionsToUsers(
								gomock.Any(),
								gomock.Any(),
								nonAdminUserID,
								nonAdminUserID2,
							).
							Return([]model.SavedQueriesPermissions{
								{
									QueryID:        1,
									Public:         false,
									SharedToUserID: database.NullUUID(nonAdminUserID),
								},
								{
									QueryID:        1,
									Public:         false,
									SharedToUserID: database.NullUUID(nonAdminUserID2),
								},
							}, nil)
					},
					expectedStatus: http.StatusCreated,
					verifySuccessBody: func(t *testing.T, body []byte) {
						var wrapper struct {
							Data v2.ShareSavedQueriesResponse `json:"data"`
						}
						require.NoError(t, json.Unmarshal(body, &wrapper))

						require.Len(t, wrapper.Data, 2)
						ids := []uuid.UUID{
							wrapper.Data[0].SharedToUserID.UUID,
							wrapper.Data[1].SharedToUserID.UUID,
						}
						assert.ElementsMatch(t, []uuid.UUID{nonAdminUserID, nonAdminUserID2}, ids)
					},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc(
				"/api/v2/saved-queries/{saved_query_id}/permissions",
				resources.ShareSavedQueries,
			).Methods(http.MethodPut)

			url := fmt.Sprintf(endpointPattern, tc.savedQueryID)

			for _, step := range tc.steps {
				if step.setupMocks != nil {
					step.setupMocks()
				}

				req := step.buildRequest(t, url)

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req)

				require.Equal(t, step.expectedStatus, resp.Code)

				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				if step.expectedBodySubstr != "" {
					assert.Contains(t, string(bodyBytes), step.expectedBodySubstr)
				}
				if step.verifySuccessBody != nil {
					step.verifySuccessBody(t, bodyBytes)
				}
			}
		})
	}
}

func TestResources_DeleteSavedQueryPermissions(t *testing.T) {

	userId, err := uuid.NewV4()
	require.Nil(t, err)
	userId2, err := uuid.NewV4()
	require.Nil(t, err)
	userId3, err := uuid.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}/permissions"
	savedQueryId := "1"

	t.Run("user can unshare their owned saved query", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), userId, int64(1)).Return(true, nil)

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId2, userId3},
		}

		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("user can unshare queries they do not own as an admin", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId2, userId3},
		}

		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("user errors unsharing query from themselves", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId},
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(true, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), []uuid.UUID{userId}).Return(fmt.Errorf("an error"))

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("user can unshare queries shared to them", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId},
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(true, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), []uuid.UUID{userId}).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("user cannot unshare a query that wasn't shared to them", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId},
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(false, nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("error checking if query is shared with user", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId},
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(false, fmt.Errorf("an error"))

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("error user unsharing saved query that does not belong to them", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), userId, int64(1)).Return(false, nil)

		var userIds []uuid.UUID
		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(userIds))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusForbidden, response.Code)
	})

	t.Run("error database fails while unsharing to users", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId2, userId3},
		}

		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), gomock.Any()).Return(fmt.Errorf("an error"))

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("error database fails while checking saved query ownership", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), userId, int64(1)).Return(false, fmt.Errorf("an error"))

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid.UUID{userId2, userId3},
		}

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestResources_GetPermissionsForSavedQuery(t *testing.T) {
	t.Parallel()

	// testQuery1 owner
	user1Id, err := uuid.NewV4()
	require.NoError(t, err)
	// testQuery 1 shared to user
	user2Id, err := uuid.NewV4()
	require.NoError(t, err)
	// admin
	user3Id, err := uuid.NewV4()
	require.NoError(t, err)

	// Setup
	var (
		testSavedQuery1 = model.SavedQuery{
			UserID:      user1Id.String(),
			Name:        "Test Query 1",
			Query:       "Match (n:Base) return n",
			Description: "test query",
			BigSerial: model.BigSerial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
				},
			},
		}
		testSavedQuery1Permissions = model.SavedQueriesPermissions{
			SharedToUserID: uuid.NullUUID{
				UUID:  user2Id,
				Valid: true,
			},
			QueryID: 1,
			Public:  false,
			BigSerial: model.BigSerial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
				},
			},
		}
		testSavedQuery2Permissions = model.SavedQueriesPermissions{
			QueryID: 2,
			Public:  true,
			BigSerial: model.BigSerial{
				ID: 2,
				Basic: model.Basic{
					CreatedAt: time.Now(),
				},
			},
		}
	)

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}

	type expected struct {
		responseCode   int
		responseBody   string
		responseHeader http.Header
	}

	type fields struct {
		setupMocks func(t *testing.T, mock *mock)
	}

	type args struct {
		buildRequest func() *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		expect expected
	}{
		{
			name: "fail - not a user",
			fields: fields{
				// required otherwise the build mocks func will panic
				setupMocks: func(t *testing.T, mock *mock) {},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequest(http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.Nil(t, err)
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"no associated user found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name: "fail - saved query URI parameter not an int",
			fields: fields{
				// required otherwise the build mocks func will panic
				setupMocks: func(t *testing.T, mock *mock) {},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user1Id), http.MethodGet, "/api/v2/saved-queries/not-an-int/permissions", nil)
					require.Nil(t, err)
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"id is malformed"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name: "fail - error retrieving saved query permissions",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(1)).Return([]model.SavedQueriesPermissions{}, fmt.Errorf("error returning saved query"))
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user1Id), http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "1"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name: "fail - query does not have shared permissions",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(1)).Return([]model.SavedQueriesPermissions{}, nil)
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user1Id), http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "1"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"no query permissions exist for saved query"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - error asserting if user owns query",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(1)).Return([]model.SavedQueriesPermissions{testSavedQuery1Permissions}, nil)
					mock.mockDatabase.EXPECT().GetSavedQuery(gomock.Any(), int64(1)).Return(model.SavedQuery{}, fmt.Errorf("error returning saved query"))
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user3Id), http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "1"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - user cannot access saved query permissions",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(1)).Return([]model.SavedQueriesPermissions{testSavedQuery1Permissions}, nil)
					mock.mockDatabase.EXPECT().GetSavedQuery(gomock.Any(), int64(1)).Return(testSavedQuery1, nil)
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user3Id), http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "1"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"no query permissions exist for saved query"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "success - user owns query",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(1)).Return([]model.SavedQueriesPermissions{testSavedQuery1Permissions}, nil)
					mock.mockDatabase.EXPECT().GetSavedQuery(gomock.Any(), int64(1)).Return(testSavedQuery1, nil)
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user1Id), http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "1"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   fmt.Sprintf(`{"data":{"query_id":1,"public":false,"shared_to_user_ids":["%s"]}}`, user2Id),
			},
		},
		{
			name: "success - admin access query",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(1)).Return([]model.SavedQueriesPermissions{testSavedQuery1Permissions}, nil)
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(user3Id), http.MethodGet, "/api/v2/saved-queries/1/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "1"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   fmt.Sprintf(`{"data":{"query_id":1,"public":false,"shared_to_user_ids":["%s"]}}`, user2Id),
			},
		},
		{
			name: "success - public query",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					t.Helper()
					mock.mockDatabase.EXPECT().GetSavedQueryPermissions(gomock.Any(), int64(2)).Return([]model.SavedQueriesPermissions{testSavedQuery2Permissions}, nil)
				},
			},
			args: args{
				buildRequest: func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(user2Id), http.MethodGet, "/api/v2/saved-queries/2/permissions", nil)
					require.NoError(t, err)
					req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: "2"})
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"query_id":2,"public":true,"shared_to_user_ids":[]}}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			databaseMock := &mock{
				mockDatabase: mocks.NewMockDatabase(mockCtrl),
			}

			defer mockCtrl.Finish()

			tt.fields.setupMocks(t, databaseMock)
			s := v2.Resources{
				DB: databaseMock.mockDatabase,
			}

			response := httptest.NewRecorder()
			request := tt.args.buildRequest()
			s.GetSavedQueryPermissions(response, request)
			statusCode, header, body := test.ProcessResponse(t, response)
			assert.Equal(t, tt.expect.responseCode, statusCode)
			assert.Equal(t, tt.expect.responseHeader, header)
			assert.JSONEq(t, tt.expect.responseBody, body)
		})
	}
}
