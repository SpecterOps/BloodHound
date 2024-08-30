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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_ListSavedQueries_SortingError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	if req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)

		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestResources_ListSavedQueries_InvalidFilterColumn(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	if req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("foo", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)

		require.Contains(t, response.Body.String(), "column cannot be filtered")
	}
}

func TestResources_ListSavedQueries_InvalidFilterPredicate(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	if req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestResources_ListSavedQueries_InvalidSkip(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	if req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("skip", "-1")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "invalid skip")
	}
}

func TestResources_ListSavedQueries_InvalidLimit(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	if req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("limit", "-1")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "invalid limit")
	}
}

func TestResources_ListSavedQueries_DBError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, "", model.SQLFilter{}, 0, 10000).Return(model.SavedQueries{}, 0, fmt.Errorf("foo"))

	if req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestResources_ListSavedQueries(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, gomock.Any(), gomock.Any(), 1, 10).Return(model.SavedQueries{
		{
			UserID: userId.String(),
			Name:   "myQuery",
			Query:  "Match(n) return n;",
		},
	}, 1, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("skip", "1")
	q.Add("limit", "10")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)

}

func TestResources_ListSavedQueries_OwnedQueries(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, gomock.Any(), gomock.Any(), 1, 10).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n;",
			Description: "Public query description",
		},
	}, 1, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)

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

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

}

func TestResources_ListSavedQueries_PublicQueries(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().GetPublicSavedQueries(gomock.Any()).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n",
			Description: "Public query description",
		},
	}, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("scope", "public")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	// Set up the router and serve the request
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusOK, response.Code)

	var actualResponse struct {
		Data []model.SavedQueryResponse `json:"data"`
	}

	err = json.NewDecoder(response.Body).Decode(&actualResponse)
	require.Nil(t, err)

	expectedResponse := []model.SavedQueryResponse{
		{
			SavedQuery: model.SavedQuery{
				UserID:      userId.String(),
				Name:        "myQuery",
				Query:       "Match(n) return n",
				Description: "Public query description",
			},
			Scope: "public",
		},
	}

	assert.Equal(t, expectedResponse, actualResponse.Data)

}

func TestResources_ListSavedQueries_SharedQueries(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().GetSharedSavedQueries(gomock.Any(), userId).Return(model.SavedQueries{
		{
			UserID:      userId.String(),
			Name:        "myQuery",
			Query:       "Match(n) return n",
			Description: "Shared query description",
		},
	}, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("scope", "shared")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	// Set up the router and serve the request
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusOK, response.Code)

	var actualResponse struct {
		Data []model.SavedQueryResponse `json:"data"`
	}

	err = json.NewDecoder(response.Body).Decode(&actualResponse)
	require.Nil(t, err)

	expectedResponse := []model.SavedQueryResponse{
		{
			SavedQuery: model.SavedQuery{
				UserID:      userId.String(),
				Name:        "myQuery",
				Query:       "Match(n) return n",
				Description: "Shared query description",
			},
			Scope: "shared",
		},
	}

	assert.Equal(t, expectedResponse, actualResponse.Data)

}

func TestResources_ListSavedQueries_MulitpleScopeQueries(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	q := url.Values{}
	q.Add("sort_by", "-name")
	q.Add("name", "eq:myQuery")
	q.Add("scope", "shared,public")

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	// Set up the router and serve the request
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusOK, response.Code)

	var actualResponse struct {
		Data []model.SavedQueryResponse `json:"data"`
	}

	err = json.NewDecoder(response.Body).Decode(&actualResponse)
	require.Nil(t, err)

	expectedResponse := []model.SavedQueryResponse{
		{
			SavedQuery: model.SavedQuery{
				UserID:      userId.String(),
				Name:        "myQuery",
				Query:       "Match(n) return n",
				Description: "Shared query description",
			},
			Scope: "shared",
		},
		{
			SavedQuery: model.SavedQuery{
				UserID:      userId.String(),
				Name:        "myQuery",
				Query:       "Match(n) return n",
				Description: "Public query description",
			},
			Scope: "public",
		},
	}

	assert.Equal(t, expectedResponse, actualResponse.Data)

}

func TestResources_ListSavedQueries_ScopeDBError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().ListSavedQueries(gomock.Any(), userId, "", model.SQLFilter{}, 0, 10000).Return(model.SavedQueries{}, 0, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)

	q := url.Values{}
	q.Add("scope", "owned")
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)

}

func TestResources_ListSavedQueries_InvalidScope(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "GET", endpoint, nil)
	require.Nil(t, err)

	q := url.Values{}
	q.Add("scope", "foo")
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req.URL.RawQuery = q.Encode()

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), "invalid scope param")

}

func TestResources_CreateSavedQuery_InvalidBody(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	payload := "foobar"

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestResources_CreateSavedQuery_EmptyBody(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "field is empty")
}

func TestResources_CreateSavedQuery_DuplicateName(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("duplicate key value violates unique constraint \"idx_saved_queries_composite_index\""))

	payload := v2.CreateSavedQueryRequest{
		Query: "Match(n) return n",
		Name:  "myQuery",
	}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "duplicate")
}

func TestResources_CreateSavedQuery_CreateFailure(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	payload := v2.CreateSavedQueryRequest{
		Query:       "Match(n) return n",
		Name:        "myCustomQuery1",
		Description: "An example description",
	}

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), userId, payload.Name, payload.Query, payload.Description).Return(model.SavedQuery{}, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResources_CreateSavedQuery(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusCreated, response.Code)
}

func TestResources_UpdateSavedQuery_InvalidBody(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	payload := "foobar"

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.UpdateSavedQuery).Methods("PUT")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestResources_UpdateSavedQuery_InvalidID(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "NotAValidUUID"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.UpdateSavedQuery).Methods("PUT")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
}

func TestResources_UpdateSavedQuery_GetSavedQueryError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("foo"))

	payload := v2.CreateSavedQueryRequest{}
	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "foo")
}

func TestResources_UpdateSavedQuery_QueryBelongsToAnotherUser(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	payload := v2.CreateSavedQueryRequest{}

	// context owner is not an admin
	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "query does not exist")
}

func TestResources_UpdateSavedQuery_Admin_NonPublicQuery(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	// query is not public
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, nil)

	payload := v2.CreateSavedQueryRequest{}

	// context owner is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "query does not exist")
}

func TestResources_UpdateSavedQuery_NoQueryMatch(t *testing.T) {
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
	require.Nil(t, err)

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "query does not exist")
}

func TestResources_UpdateSavedQuery_ErrorFetchingPublicStatus(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	// query belongs to another user
	mockDB.EXPECT().GetSavedQuery(gomock.Any(), gomock.Any()).Return(model.SavedQuery{UserID: "notThisUser"}, nil)

	// query is public
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("foo"))

	payload := v2.CreateSavedQueryRequest{}

	// context owner is an admin
	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "PUT", fmt.Sprintf(endpoint, "1"), must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "foo")
}

func TestResources_UpdateSavedQuery_UpdateFailed(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResources_UpdateSavedQuery_OwnPrivateQuery_Success(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
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

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestResources_UpdateSavedQuery_OwnPublicQuery_Success(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestResources_UpdateSavedQuery_AdminPublicQuery_Success(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

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
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestResources_DeleteSavedQuery_IDMalformed(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", "/api/v2/saved-queries/-1", nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc("/api/v2/saved-queries/-1", resources.DeleteSavedQuery).Methods("DELETE")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
}

func TestResources_DeleteSavedQuery_DBError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResources_DeleteSavedQuery_UserNotAdmin(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Contains(t, response.Body.String(), "User does not have permission to delete this query")
}

func TestResources_DeleteSavedQuery_IsPublicSavedQueryDBError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/%s"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("error"))

	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResources_DeleteSavedQuery_NotPublicQueryAndUserIsAdmin(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/%s"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(false, nil)

	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Contains(t, response.Body.String(), "User does not have permission to delete this query")
}

func TestResources_DeleteSavedQuery_RecordNotFound(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, database.ErrNotFound)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "query does not exist")
}

func TestResources_DeleteSavedQuery_RecordNotFound_EdgeCase(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(database.ErrNotFound)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "query does not exist")
}

func TestResources_DeleteSavedQuery_DeleteError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(fmt.Errorf("foo"))

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
}

func TestResources_DeleteSavedQuery_PublicQueryAndUserIsAdmin(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/%s"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	mockDB.EXPECT().IsSavedQueryPublic(gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(nil)

	req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestResources_DeleteSavedQuery(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	endpoint := "/api/v2/saved-queries/{%s}"
	savedQueryId := "1"

	mockDB.EXPECT().SavedQueryBelongsToUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockDB.EXPECT().DeleteSavedQuery(gomock.Any(), gomock.Any()).Return(nil)

	req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), "DELETE", fmt.Sprintf(endpoint, savedQueryId), nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableSavedQueryID: savedQueryId})

	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteSavedQuery)

	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusNoContent, response.Code)
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
