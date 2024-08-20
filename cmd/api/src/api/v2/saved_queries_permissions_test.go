/*
 * Copyright 2024 Specter Ops, Inc.
 *
 * Licensed under the Apache License, Version 2.0
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v2_test

import (
	"fmt"
	uuid2 "github.com/gofrs/uuid"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResources_DeleteSavedQueryPermissions(t *testing.T) {

	userId, err := uuid2.NewV4()
	require.Nil(t, err)
	userId2, err := uuid2.NewV4()
	require.Nil(t, err)
	userId3, err := uuid2.NewV4()
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
			UserIds: []uuid2.UUID{userId2, userId3},
		}

		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("user can unshare queries they do not own as an admin", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid2.UUID{userId2, userId3},
		}

		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), gomock.Any()).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("user errors unsharing query from themselves", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			Self: true,
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(true, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUser(gomock.Any(), int64(1), userId).Return(fmt.Errorf("an error"))

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("user can unshare queries shared to them", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			Self: true,
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(true, nil)
		mockDB.EXPECT().DeleteSavedQueryPermissionsForUser(gomock.Any(), int64(1), userId).Return(nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("user cannot unshare a query that wasn't shared to them", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			Self: true,
		}

		mockDB.EXPECT().IsSavedQuerySharedToUser(gomock.Any(), int64(1), userId).Return(false, nil)

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusBadRequest, response.Code)
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
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("error database fails while unsharing to users", func(t *testing.T) {
		var (
			mockCtrl  = gomock.NewController(t)
			mockDB    = mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB}
		)
		defer mockCtrl.Finish()

		payload := v2.DeleteSavedQueryPermissionsRequest{
			UserIds: []uuid2.UUID{userId2, userId3},
		}

		mockDB.EXPECT().DeleteSavedQueryPermissionsForUsers(gomock.Any(), int64(1), gomock.Any()).Return(fmt.Errorf("an error"))

		req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
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
			UserIds: []uuid2.UUID{userId2, userId3},
		}

		req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodDelete, fmt.Sprintf(endpoint, savedQueryId), must.MarshalJSONReader(payload))
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

		response := httptest.NewRecorder()
		handler := http.HandlerFunc(resources.DeleteSavedQueryPermissions)

		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}
