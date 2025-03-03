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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	uuid2 "github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Can only remove a field at the root of the JSON object
func removeFieldFromJsonString(jsonString string, field string) (string, error) {
	var unmarshaled map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &unmarshaled)
	if err != nil {
		return "", err
	}

	delete(unmarshaled, field)

	modifiedJson, err := json.Marshal(unmarshaled)
	if err != nil {
		return "", err
	}

	return string(modifiedJson), nil
}

func TestResources_ListSavedQueries_SortingError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)
	q := url.Values{}
	q.Add("sort_by", "invalidColumn")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"column format does not support sorting"}]}`, actualWithoutTimestamp)
}

func TestResources_ListSavedQueries_InvalidFilterColumn(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)
	q := url.Values{}
	q.Add("foo", "gt:0")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"the specified column cannot be filtered: foo"}]}`, actualWithoutTimestamp)
}

func TestResources_ListSavedQueries_InvalidFilterPredicate(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)
	q := url.Values{}
	q.Add("name", "gt:0")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"the specified filter predicate is not supported for this column: name gt"}]}`, actualWithoutTimestamp)
}

func TestResources_ListSavedQueries_InvalidSkip(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)
	q := url.Values{}
	q.Add("skip", "-1")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"query parameter \"skip\" is malformed: invalid skip: -1"}]}`, actualWithoutTimestamp)
}

func TestResources_ListSavedQueries_InvalidLimit(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)
	q := url.Values{}
	q.Add("limit", "-1")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"query parameter \"limit\" is malformed: invalid limit: -1"}]}`, actualWithoutTimestamp)
}

func TestResources_ListSavedQueries_DBError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, "", model.SQLFilter{}, 0, 10000).Return(model.SavedQueries{}, 0, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, actualWithoutTimestamp)
}

func TestResources_ListSavedQueries(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, gomock.Any(), gomock.Any(), 1, 10).Return(model.SavedQueries{
		{
			UserID: userId.String(),
			Name:   "myQuery",
			Query:  "Match(n) return n;",
		},
	}, 1, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("skip", "1")
	q.Add("limit", "10")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"count":1,"limit":10,"skip":1,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n;","description":"","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}]}`, response.Body.String())
}

func TestResources_ListSavedQueries_OwnedQueries(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, gomock.Any(), gomock.Any(), 1, 10).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n;",
			Description: "Public query description",
		},
	}, 1, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("skip", "1")
	q.Add("limit", "10")
	q.Add("scope", "owned")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"count":1,"limit":10,"skip":1,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n;","description":"Public query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"owned"}]}`, response.Body.String())
}

func TestResources_ListSavedQueries_PublicQueries(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	mockDB.EXPECT().GetPublicSavedQueries(gomock.Any()).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n",
			Description: "Public query description",
		},
	}, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("scope", "public")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"count":1,"limit":10000,"skip":0,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Public query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"public"}]}`, response.Body.String())
}

func TestResources_ListSavedQueries_SharedQueries(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	mockDB.EXPECT().GetSharedSavedQueries(gomock.Any(), userId).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n",
			Description: "Shared query description",
		},
	}, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("scope", "shared")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"count":1,"limit":10000,"skip":0,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Shared query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"shared"}]}`, response.Body.String())
}

func TestResources_ListSavedQueries_MulitpleScopeQueries(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	mockDB.EXPECT().GetSharedSavedQueries(gomock.Any(), userId).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n",
			Description: "Shared query description",
		},
	}, nil)

	mockDB.EXPECT().GetPublicSavedQueries(gomock.Any()).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n",
			Description: "Public query description",
		},
	}, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("scope", "shared,public")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"count":2,"limit":10000,"skip":0,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Shared query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"shared"},{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Public query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"public"}]}`, response.Body.String())
}

func TestResources_ListSavedQueries_ScopeDBError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, "", model.SQLFilter{}, 0, 10000).Return(model.SavedQueries{}, 0, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("scope", "owned")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, actualWithoutTimestamp)

}

func TestResources_ListSavedQueries_InvalidScope(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.NoError(t, err)

	q := url.Values{}
	q.Add("scope", "foo")
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"invalid scope param"}]}`, actualWithoutTimestamp)
}

func TestResources_CreateSavedQuery_InvalidBody(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	payload := "foobar"

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"could not decode limited payload request into value: json: cannot unmarshal string into Go value of type v2.CreateSavedQueryRequest"}]}`, actualWithoutTimestamp)
}

func TestResources_CreateSavedQuery_EmptyBody(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"the name and/or query field is empty"}]}`, actualWithoutTimestamp)
}

func TestResources_CreateSavedQuery_DuplicateName(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("duplicate key value violates unique constraint \"idx_saved_queries_composite_index\""))

	payload := v2.CreateSavedQueryRequest{
		Query: "Match(n) return n",
		Name:  "myQuery",
	}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"duplicate name for saved query: please choose a different name"}]}`, actualWithoutTimestamp)
}

func TestResources_CreateSavedQuery_CreateFailure(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	payload := v2.CreateSavedQueryRequest{
		Query:       "Match(n) return n",
		Name:        "myCustomQuery1",
		Description: "An example description",
	}

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), userId, payload.Name, payload.Query, payload.Description).Return(model.SavedQuery{}, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"foo"}]}`, actualWithoutTimestamp)
}

func TestResources_CreateSavedQuery(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	payload := v2.CreateSavedQueryRequest{
		Query:       "Match(n) return n",
		Name:        "myCustomQuery1",
		Description: "An example description",
	}

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), userId, payload.Name, payload.Query, payload.Description).Return(model.SavedQuery{
		UserID:      userId.String(),
		Name:        payload.Name,
		Query:       payload.Query,
		Description: "An example description",
	}, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusCreated, response.Code)
	require.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myCustomQuery1","query":"Match(n) return n","description":"An example description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_UpdateSavedQuery_InvalidBody(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	payload := "foobar"

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.UpdateSavedQuery).Methods("PUT")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"could not decode limited payload request into value: json: cannot unmarshal string into Go value of type v2.CreateSavedQueryRequest"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_InvalidID(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "NotAValidUUID"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.UpdateSavedQuery).Methods("PUT")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"id is malformed."}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_GetSavedQueryError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("foo"))

	payload := v2.CreateSavedQueryRequest{}
	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"foo"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_QueryBelongsToAnotherUser(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	payload := v2.CreateSavedQueryRequest{}

	// context owner is not an admin
	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.JSONEq(t, `{"http_status":404,"request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_Admin_NonPublicQuery(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	// query is not public
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, nil)

	payload := v2.CreateSavedQueryRequest{}

	// context owner is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.JSONEq(t, `{"http_status":404,"request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_NoQueryMatch(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, nil)

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.JSONEq(t, `{"http_status":404,"request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_ErrorFetchingPublicStatus(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	// query is public
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("foo"))

	payload := v2.CreateSavedQueryRequest{}

	// context owner is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"foo"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_UpdateFailed(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	savedQuery := model.SavedQuery{
		UserID:      userId.String(),
		Name:        "foo",
		Query:       "bar",
		Description: "baz",
		BigSerial: model.BigSerial{
			ID: int64(1),
		},
	}

	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)
	mockDB.EXPECT().UpdateSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("random error"))

	payload := v2.CreateSavedQueryRequest{Name: "notFoo", Query: "notBar", Description: "notBaz"}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, actualWithoutTimestamp)
}

func TestResources_UpdateSavedQuery_OwnPrivateQuery_Success(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	savedQuery := model.SavedQuery{
		UserID:      userId.String(),
		Name:        "foo",
		Query:       "bar",
		Description: "baz",
		BigSerial: model.BigSerial{
			ID: int64(1),
		},
	}

	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)
	mockDB.EXPECT().UpdateSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)

	payload := v2.CreateSavedQueryRequest{Name: "notFoo", Query: "notBar", Description: "notBaz"}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_UpdateSavedQuery_AdminPrivateQuery_Success(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	savedQuery := model.SavedQuery{
		UserID:      userId.String(),
		Name:        "foo",
		Query:       "bar",
		Description: "baz",
		BigSerial: model.BigSerial{
			ID: int64(1),
		},
	}

	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)
	mockDB.EXPECT().UpdateSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)

	payload := v2.CreateSavedQueryRequest{Name: "notFoo", Query: "notBar", Description: "notBaz"}

	// user is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_UpdateSavedQuery_OwnPublicQuery_Success(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	savedQuery := model.SavedQuery{
		UserID:      userId.String(),
		Name:        "foo",
		Query:       "bar",
		Description: "baz",
		BigSerial: model.BigSerial{
			ID: int64(1),
		},
	}

	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)
	mockDB.EXPECT().UpdateSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)

	payload := v2.CreateSavedQueryRequest{Name: "notFoo", Query: "notBar", Description: "notBaz"}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_UpdateSavedQuery_AdminPublicQuery_Success(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.FromString("ac83d188-cb30-430b-953a-9e0ecab45e2c")
	require.NoError(t, err)

	savedQuery := model.SavedQuery{
		UserID:      userId.String(),
		Name:        "foo",
		Query:       "bar",
		Description: "baz",
		BigSerial: model.BigSerial{
			ID: int64(1),
		},
	}

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	// query is public
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(true, nil)

	mockDB.EXPECT().UpdateSavedQuery(gomock.Any(), gomock.Any()).Return(savedQuery, nil)

	payload := v2.CreateSavedQueryRequest{Name: "notFoo", Query: "notBar", Description: "notBaz"}

	// context owner is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_DeleteSavedQuery_IDMalformed(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", "/api/v2/saved-queries/-1", nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc("/api/v2/saved-queries/-1", resources.DeleteSavedQuery).Methods("DELETE")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"http_status":400,"request_id":"","errors":[{"context":"","message":"id is malformed."}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_DBError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_UserNotAdmin(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.JSONEq(t, `{"http_status":403,"request_id":"","errors":[{"context":"","message":"User does not have permission to delete this query"}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_IsPublicSavedQueryDBError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/%s"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("error"))

	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"http_status":500,"request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_NotPublicQueryAndUserIsAdmin(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/%s"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, nil)

	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.JSONEq(t, `{"http_status":403,"request_id":"","errors":[{"context":"","message":"User does not have permission to delete this query"}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_RecordNotFound(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, database.ErrNotFound)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.JSONEq(t, `{"http_status":404,"request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_RecordNotFound_EdgeCase(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(database.ErrNotFound)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.JSONEq(t, `{"http_status":404,"request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_DeleteError(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	actualWithoutTimestamp, err := removeFieldFromJsonString(response.Body.String(), "timestamp")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":""}`, actualWithoutTimestamp)
}

func TestResources_DeleteSavedQuery_PublicQueryAndUserIsAdmin(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/%s"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(nil)

	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "", response.Body.String())
}

func TestResources_DeleteSavedQuery(t *testing.T) {
	// Setup
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.NoError(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "", response.Body.String())
}

func createContextWithOwnerId(id uuid2.UUID) context.Context {
	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Owner: model.User{
				Unique: model.Unique{
					ID: id,
				},
			},
		},
		Host: nil,
	}
	return bhCtx.ConstructGoContext()
}

func createContextWithAdminOwnerId(id uuid2.UUID) context.Context {
	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Owner: model.User{
				Unique: model.Unique{
					ID: id,
				},
				Roles: model.Roles{{
					Name:        auth.RoleAdministrator,
					Description: "Can manage users, clients, and application configuration",
					Permissions: auth.Permissions().All(),
				}},
			},
		},
		Host: nil,
	}
	return bhCtx.ConstructGoContext()
}
