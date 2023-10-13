// Copyright 2023 Specter Ops, Inc.
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database/mocks"

	"github.com/specterops/bloodhound/src/model"
)

func TestResources_ListAuditLogs_SortingError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)

		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestResources_ListAuditLogs_InvalidColumn(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("foo", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "column cannot be filtered")
	}
}

func TestResources_ListAuditLogs_InvalidPredicate(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("actor_name", "invalidPredicate:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	}
}

func TestResources_ListAuditLogs_PredicateMismatchWithColumn(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("actor_name", "gte:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestResources_ListAuditLogs_DBError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().ListAuditLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "id, actor_name desc", model.SQLFilter{}).Return(model.AuditLogs{}, 0, fmt.Errorf("foo"))

	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "id")
		q.Add("sort_by", "-actor_name")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestResources_ListAuditLogs(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().ListAuditLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "id, actor_name desc", model.SQLFilter{}).Return(model.AuditLogs{}, 1000, nil)

	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "id")
		q.Add("sort_by", "-actor_name")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
	}
}

func TestResources_ListAuditLogs_Filtered(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().ListAuditLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "", model.SQLFilter{SQLString: "actor_name = ?", Params: []any{"foo"}}).Return(model.AuditLogs{}, 1000, nil)
	endpoint := "/api/v2/audit"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("actor_name", "eq:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuditLogs).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
	}
}
