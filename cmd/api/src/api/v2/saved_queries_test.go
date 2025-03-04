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
	"bytes"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Can only replace a field value at the root of the JSON object
// if field does not exist in jsonString, will effectively no-op
func replaceFieldValueInJsonString(jsonString string, field string, value any) (string, error) {
	var unmarshaled map[string]any
	err := json.Unmarshal([]byte(jsonString), &unmarshaled)
	if err != nil {
		return "", err
	}

	if _, exists := unmarshaled[field]; exists {
		unmarshaled[field] = value
	}

	modifiedJson, err := json.Marshal(unmarshaled)
	if err != nil {
		return "", err
	}

	return string(modifiedJson), nil
}

type TestData struct {
	testName             string
	expectedResponseBody string
	expectedResponseCode int
	queryParams          map[string]string
}

func getTestArgs() []TestData {
	return []TestData{
		{
			testName:             "SortingError",
			expectedResponseBody: `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"column format does not support sorting"}]}`,
			expectedResponseCode: http.StatusBadRequest,
			queryParams:          map[string]string{"sort_by": "invalidColumn"},
		},
		{
			testName:             "InvalidFilterColumn",
			expectedResponseBody: `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"the specified column cannot be filtered: foo"}]}`,
			expectedResponseCode: http.StatusBadRequest,
			queryParams:          map[string]string{"foo": "gt:0"},
		},
	}
}

// // Matrix testing experiment
// Pro:  Unify test logic - later maintenance requires only one test modification, or splitting off new tests when necessary
// Con:  Poor IDE integration, added complexity, test writer must determine when tests should and shouldn't be consolidated
func TestResources_ListSavedQueries_SortingError(t *testing.T) {
	// move method for getargs into test
	testData := getTestArgs()
	for _, testArgs := range testData {
		// New test context with test name
		// t.Run(testArgs.testName, t)
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
		for key, val := range testArgs.queryParams {
			q.Add(key, val)
		}

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		// Act
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		// Assert
		responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
		require.NoError(t, err, fmt.Sprintf("Test: %s", testArgs.testName))
		assert.Equal(t, testArgs.expectedResponseCode, response.Code, fmt.Sprintf("Test: %s", testArgs.testName))
		assert.JSONEq(t, testArgs.expectedResponseBody, responseBodyWithDefaultTimestamp, fmt.Sprintf("Test: %s", testArgs.testName))
	}
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"the specified column cannot be filtered: foo"}]}`, responseBodyWithDefaultTimestamp)
}

func TestResources_ListSavedQueries_NotAUserAuth(t *testing.T) {
	// Setup
	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Owner: model.Role{},
		},
		Host: nil,
	}
	notAUserCtx := bhCtx.ConstructGoContext()

	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"

	req, err := http.NewRequestWithContext(notAUserCtx, "GET", endpoint, nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"errors":[{"context":"","message":"No associated user found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`, responseBodyWithDefaultTimestamp)
}

func TestResources_ListSavedQueries_InvalidQueryParameterFilters(t *testing.T) {
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
	q.Add("name", "notAnOperator:0")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"the specified filter predicate is not supported for this column: name gt"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query parameter \"skip\" is malformed: invalid skip: -1"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query parameter \"limit\" is malformed: invalid limit: -1"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, responseBodyWithDefaultTimestamp)
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
	q.Add("sort_by", "name")
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"count":1,"limit":10,"skip":1,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n;","description":"","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}]}`, response.Body.String())
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"count":1,"limit":10,"skip":1,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n;","description":"Public query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"owned"}]}`, response.Body.String())
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"count":1,"limit":10000,"skip":0,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Public query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"public"}]}`, response.Body.String())
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"count":1,"limit":10000,"skip":0,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Shared query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"shared"}]}`, response.Body.String())
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"count":2,"limit":10000,"skip":0,"data":[{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Shared query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"shared"},{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myQuery","query":"Match(n) return n","description":"Public query description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"scope":"public"}]}`, response.Body.String())
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, responseBodyWithDefaultTimestamp)

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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"invalid scope param"}]}`, responseBodyWithDefaultTimestamp)
}

func TestResources_CreateSavedQuery_NotAUserAuth(t *testing.T) {
	// Setup
	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Owner: model.Role{},
		},
		Host: nil,
	}
	notAUserCtx := bhCtx.ConstructGoContext()

	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	payload := "foobar"

	req, err := http.NewRequestWithContext(notAUserCtx, "POST", endpoint, must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"errors":[{"context":"","message":"No associated user found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"could not decode limited payload request into value: json: cannot unmarshal string into Go value of type v2.CreateSavedQueryRequest"}]}`, responseBodyWithDefaultTimestamp)
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

	payload := map[string]any{
		"query": "",
		"name":  "",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"the name and/or query field is empty"}]}`, responseBodyWithDefaultTimestamp)
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

	payload := map[string]any{
		"query": "Match(n) return n",
		"name":  "myQuery",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("duplicate key value violates unique constraint \"idx_saved_queries_composite_index\""))

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"duplicate name for saved query: please choose a different name"}]}`, responseBodyWithDefaultTimestamp)
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

	payload := map[string]any{
		"query":       "Match(n) return n",
		"name":        "myCustomQuery1",
		"description": "An example description",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), userId, payload["name"], payload["query"], payload["description"]).Return(model.SavedQuery{}, fmt.Errorf("foo"))

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"foo"}]}`, responseBodyWithDefaultTimestamp)
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

	payload := map[string]any{
		"query":       "Match(n) return n",
		"name":        "myCustomQuery1",
		"description": "An example description",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), userId, payload["name"], payload["query"], payload["description"]).Return(model.SavedQuery{
		UserID:      userId.String(),
		Name:        fmt.Sprintf("%v", payload["name"]),
		Query:       fmt.Sprintf("%v", payload["query"]),
		Description: fmt.Sprintf("%v", payload["description"]),
	}, nil)

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"myCustomQuery1","query":"Match(n) return n","description":"An example description","id":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_UpdateSavedQuery_NotAUserAuth(t *testing.T) {
	// Setup
	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Owner: model.Role{},
		},
		Host: nil,
	}
	notAUserCtx := bhCtx.ConstructGoContext()

	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	payload := "foobar"

	req, err := http.NewRequestWithContext(notAUserCtx, "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.UpdateSavedQuery).Methods("PUT")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"errors":[{"context":"","message":"No associated user found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"could not decode limited payload request into value: json: cannot unmarshal string into Go value of type v2.CreateSavedQueryRequest"}]}`, responseBodyWithDefaultTimestamp)
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

	var payload any

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "NotAValidUUID"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.UpdateSavedQuery).Methods("PUT")

	// Act
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed."}]}`, responseBodyWithDefaultTimestamp)
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

	var payload any
	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"foo"}]}`, responseBodyWithDefaultTimestamp)
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

	var payload any

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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, responseBodyWithDefaultTimestamp)
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

	var payload any

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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, responseBodyWithDefaultTimestamp)
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

	var payload any

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, responseBodyWithDefaultTimestamp)
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

	var payload any

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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"foo"}]}`, responseBodyWithDefaultTimestamp)
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

	payload := map[string]any{
		"query":       "notFoo",
		"name":        "notBar",
		"description": "notBaz",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, responseBodyWithDefaultTimestamp)
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

	payload := map[string]any{
		"query":       "notFoo",
		"name":        "notBar",
		"description": "notBaz",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
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

	payload := map[string]any{
		"query":       "notFoo",
		"name":        "notBar",
		"description": "notBaz",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// user is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
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

	payload := map[string]any{
		"query":       "notFoo",
		"name":        "notBar",
		"description": "notBaz",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
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

	payload := map[string]any{
		"query":       "notFoo",
		"name":        "notBar",
		"description": "notBaz",
	}

	marshalledPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// context owner is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), bytes.NewReader(marshalledPayload))
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"data":{"user_id":"ac83d188-cb30-430b-953a-9e0ecab45e2c","name":"foo","query":"bar","description":"baz","id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`, response.Body.String())
}

func TestResources_DeleteSavedQuery_NotAUserAuth(t *testing.T) {
	// Setup
	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Owner: model.Role{},
		},
		Host: nil,
	}
	notAUserCtx := bhCtx.ConstructGoContext()

	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	req, err := http.NewRequestWithContext(notAUserCtx, "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.NoError(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	// Act
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)

	// Assert
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"errors":[{"context":"","message":"No associated user found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed."}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.JSONEq(t, `{"http_status":403,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"User does not have permission to delete this query"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.JSONEq(t, `{"http_status":403,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"User does not have permission to delete this query"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"query does not exist"}]}`, responseBodyWithDefaultTimestamp)
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
	responseBodyWithDefaultTimestamp, err := replaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":""}`, responseBodyWithDefaultTimestamp)
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
	assert.Equal(t, http.StatusNoContent, response.Code)
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
	assert.Equal(t, http.StatusNoContent, response.Code)
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
