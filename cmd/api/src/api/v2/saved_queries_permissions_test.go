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
	"database/sql"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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

	// Non-admin owned queries
	t.Run("Non-admin owned, query doesn't belong to user error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Non-admin owned, query shared to self error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrInvalidSelfShare, err)
	})

	t.Run("Non-admin owned, shared query shared to user(s)", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, nil, err)
	})

	t.Run("Non-admin owned, shared query set to public", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, nil, err)
	})

	t.Run("Non-admin owned, shared query set to private", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, nil, err)
	})

	t.Run("Non-admin owned, private query shared to user(s)", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, nil, err)
	})

	t.Run("Non-admin owned, private query set to public", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, nil, err)
	})

	t.Run("Non-admin owned, private query set to private", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, nil, err)
	})

	t.Run("Non-admin owned, public query shared to user(s) error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Non-admin owned, public query set to private error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Non-admin owned, public query set to public error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(nonAdminUser1, true, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	// Admin (non-admin owned) queries
	t.Run("Admin (non-admin owned), public query shared to user(s) incorrectly error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrInvalidPublicShare, err)
	})

	t.Run("Admin (non-admin owned), private query set to public error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Admin (non-admin owned), private query shared to user(s) error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Admin (non-admin owned), private query set to private error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Admin (non-admin owned), shared query set to public error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Admin (non-admin owned), shared query shared to user(s) error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Admin (non-admin owned), shared query set to priavte error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrForbidden, err)
	})

	t.Run("Admin (non-admin owned), public query set to public", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin (non-admin owned), public query set to private", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, false, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	// Admin owned queries
	t.Run("Admin-owned, query shared to self error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{adminUser.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrInvalidSelfShare, err)
	})

	t.Run("Admin-owned, shared query shared to user(s)", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, shared query set to public", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, shared query set to private", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, private query shared to user(s)", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, private query set to public", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, private query set to private", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, public query set to public", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, public query set to private", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Nil(t, err)
	})

	t.Run("Admin-owned, public query shared to user(s) incorrectly error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUser1.ID, nonAdminUser2.ID},
			Public:  false,
		}

		dbSavedQueryScope := database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}

		err := v2.CanUpdateSavedQueriesPermission(adminUser, true, payload, dbSavedQueryScope)
		require.Equal(t, v2.ErrInvalidPublicShare, err)
	})
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

func TestResources_ShareSavedQueriesPermissions_NonAdmin(t *testing.T) {

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

	t.Run("Query set to public and shared to user(s) at same time error", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{userId2},
			Public:  true,
		}

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()

		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "Public cannot be true while user_ids is populated")
	})

	t.Run("Shared query shared to user(s) (and user(s) that already have query shared with them) success", func(t *testing.T) {
		// Request made in order to share to a user for confirming 2nd request below
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{userId2},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userId2).Return([]model.SavedQueriesPermissions{
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(userId2),
			},
		}, nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusCreated, response.Code)

		// Request that we actually care about passing
		payload2 := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{userId2, userId3},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}, nil)
		mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userId2, userId3).Return([]model.SavedQueriesPermissions{
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(userId2),
			},
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(userId3),
			},
		}, nil)

		req2, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload2))
		require.Nil(t, err)
		req2.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		response2 := httptest.NewRecorder()
		// Using the same router as the first request
		router.ServeHTTP(response2, req2)
		require.Equal(t, http.StatusCreated, response2.Code)

		bodyBytes2, err := io.ReadAll(response2.Body)
		require.Nil(t, err)

		var temp2 struct {
			Data v2.ShareSavedQueriesResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes2, &temp2)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, v2.ShareSavedQueriesResponse{
			{
				SharedToUserID: database.NullUUID(userId2),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
			{
				SharedToUserID: database.NullUUID(userId3),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
		}, temp2.Data)
	})

	t.Run("Shared query shared to user(s) success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{userId2},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}, nil)
		mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userId2).Return([]model.SavedQueriesPermissions{
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(userId2),
			},
		}, nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusCreated, response.Code)

		bodyBytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		var temp struct {
			Data v2.ShareSavedQueriesResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &temp)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, v2.ShareSavedQueriesResponse{
			{
				SharedToUserID: database.NullUUID(userId2),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
		}, temp.Data)
	})

	t.Run("Shared query set to public success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
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

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusCreated, response.Code)

		bodyBytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		var temp struct {
			Data v2.ShareSavedQueriesResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &temp)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, v2.ShareSavedQueriesResponse{
			{
				SharedToUserID: uuid.NullUUID{
					UUID:  uuid.UUID{},
					Valid: false,
				},
				QueryID: 1,
				Public:  true,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
		}, temp.Data)
	})

	t.Run("Shared query set to private success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: true,
		}, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("Private query shared to user(s) success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{userId2, userId3},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), userId2, userId3).Return([]model.SavedQueriesPermissions{
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(userId2),
			},
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(userId3),
			},
		}, nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusCreated, response.Code)

		bodyBytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		var temp struct {
			Data v2.ShareSavedQueriesResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &temp)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, v2.ShareSavedQueriesResponse{
			{
				SharedToUserID: database.NullUUID(userId2),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
			{
				SharedToUserID: database.NullUUID(userId3),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
		}, temp.Data)
	})

	t.Run("Private query set to public success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().CreateSavedQueryPermissionToPublic(gomock.Any(), int64(1)).Return(model.SavedQueriesPermissions{
			QueryID: 1,
			SharedToUserID: uuid.NullUUID{
				UUID:  uuid.UUID{},
				Valid: false,
			},
			Public: true,
		}, nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusCreated, response.Code)

		bodyBytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		var temp struct {
			Data v2.ShareSavedQueriesResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &temp)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, v2.ShareSavedQueriesResponse{
			{
				SharedToUserID: uuid.NullUUID{
					UUID:  uuid.UUID{},
					Valid: false,
				},
				QueryID: 1,
				Public:  true,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
		}, temp.Data)
	})

	t.Run("Private query set to private success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any())

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)
	})
}

func TestResources_ShareSavedQueriesPermissions_Admin(t *testing.T) {

	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)

	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/%s/permissions"
	savedQueryId := "1"
	adminUserId, err := uuid.NewV4()
	require.Nil(t, err)
	nonAdminUserId, err := uuid.NewV4()
	require.Nil(t, err)
	nonAdminUserId2, err := uuid.NewV4()
	require.Nil(t, err)

	t.Run("Admin, public query set to private success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(adminUserId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	// Test cases where admin is making operations against their own query
	t.Run("Admin owned, public query set to public success", func(t *testing.T) {
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  true,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}, nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(adminUserId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("Admin owned, public query shared to user(s) success", func(t *testing.T) {
		// First have public query set to private
		payload := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(adminUserId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/saved-queries/{saved_query_id}/permissions", resources.ShareSavedQueries).Methods("PUT")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusNoContent, response.Code)

		// Now have private query shared to users
		payload2 := v2.SavedQueryPermissionRequest{
			UserIDs: []uuid.UUID{nonAdminUserId, nonAdminUserId2},
			Public:  false,
		}

		mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockDB.EXPECT().GetScopeForSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(database.SavedQueryScopeMap{
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeShared: false,
		}, nil)
		mockDB.EXPECT().CreateSavedQueryPermissionsToUsers(gomock.Any(), gomock.Any(), nonAdminUserId, nonAdminUserId2).Return([]model.SavedQueriesPermissions{
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(nonAdminUserId),
			},
			{
				QueryID:        int64(1),
				Public:         false,
				SharedToUserID: database.NullUUID(nonAdminUserId2),
			},
		}, nil)

		req2, err := http.NewRequestWithContext(createContextWithAdminOwnerId(adminUserId), "PUT", fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload2))
		require.Nil(t, err)
		req2.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		response2 := httptest.NewRecorder()
		// Using the same router as the first request
		router.ServeHTTP(response2, req2)
		require.Equal(t, http.StatusCreated, response2.Code)

		bodyBytes2, err := io.ReadAll(response2.Body)
		require.Nil(t, err)

		var temp2 struct {
			Data v2.ShareSavedQueriesResponse `json:"data"`
		}
		err = json.Unmarshal(bodyBytes2, &temp2)
		require.Nil(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		require.Nil(t, err)

		require.Equal(t, v2.ShareSavedQueriesResponse{
			{
				SharedToUserID: database.NullUUID(nonAdminUserId),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
			{
				SharedToUserID: database.NullUUID(nonAdminUserId2),
				QueryID:        1,
				Public:         false,
				BigSerial: model.BigSerial{
					ID: 0,
					Basic: model.Basic{
						CreatedAt: parsedTime,
						UpdatedAt: parsedTime,
						DeletedAt: sql.NullTime{
							Time:  parsedTime,
							Valid: false,
						},
					},
				},
			},
		}, temp2.Data)
	})
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
