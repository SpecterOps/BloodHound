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

package v2_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_ListKinds(t *testing.T) {
	t.Parallel()
	type mock struct {
		mockGraph *graphmocks.MockDatabase
		mockDB    *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	var (
		customNodeKind = graph.StringKind("Villain")
		mockKindsResp  = graph.Kinds{
			ad.User,
			ad.DCSync,
			tiering.KindTagTierZero,
			common.MigrationData,
			azure.Entity,
			azure.HasRole,
			customNodeKind,
		}

		mockSchemaNodeKindsResp = model.GraphSchemaNodeKinds{{Name: ad.User.String()}}
		mockCustomNodeKindsResp = []model.CustomNodeKind{{KindName: customNodeKind.String()}}
		mockSrcKindsResp        = []model.SourceKind{{ID: 1, Name: azure.Entity.String()}}
	)

	tt := []testData{
		{
			name: "Error: FetchKinds database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL:    &url.URL{Path: "/api/v2/graphs/kinds"},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: bad query parameter filter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=unknown:bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any())
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: column not filterable - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "badqueryparam=eq:bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any())
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"the specified column cannot be filtered: badqueryparam"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: predicate not supported - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=gt:node",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any())
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"the specified filter predicate is not supported for this column: type gt"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: GetGraphSchemaNodeKinds database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=eq:node",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any())
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(nil, 0, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: GetCustomNodeKinds database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=eq:node",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any())
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0)
				mock.mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: GetSourceKinds database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=eq:node",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any())
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0)
				mock.mockDB.EXPECT().GetCustomNodeKinds(gomock.Any())
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success: no filters returns all kinds alpha sorted - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL:    &url.URL{Path: "/api/v2/graphs/kinds"},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(mockKindsResp, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"kinds":["AZBase", "AZHasRole", "DCSync", "MigrationData", "Tag_Tier_Zero", "User", "Villain"]}}`,
			},
		},
		{
			name: "Success: filter type=eq:node returns only node kinds - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=eq:node",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(mockKindsResp, nil)
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(mockSchemaNodeKindsResp, 0, nil)
				mock.mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return(mockCustomNodeKindsResp, nil)
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return(mockSrcKindsResp, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"kinds":["AZBase","MigrationData","Tag_Tier_Zero","User","Villain"]}}`,
			},
		},
		{
			name: "Success: filter type=eq:edge returns only edge kinds - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=eq:edge",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(mockKindsResp, nil)
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(mockSchemaNodeKindsResp, 0, nil)
				mock.mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return(mockCustomNodeKindsResp, nil)
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return(mockSrcKindsResp, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"kinds":["AZHasRole", "DCSync"]}}`,
			},
		},
		{
			name: "Success: filter type=neq:node returns edges - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=neq:node",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(mockKindsResp, nil)
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(mockSchemaNodeKindsResp, 0, nil)
				mock.mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return(mockCustomNodeKindsResp, nil)
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return(mockSrcKindsResp, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"kinds":["AZHasRole","DCSync"]}}`,
			},
		},
		{
			name: "Success: unknown filter value returns empty list - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/kinds",
						RawQuery: "type=eq:bupkis",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(mockKindsResp, nil)
				mock.mockDB.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(mockSchemaNodeKindsResp, 0, nil)
				mock.mockDB.EXPECT().GetCustomNodeKinds(gomock.Any()).Return(mockCustomNodeKindsResp, nil)
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return(mockSrcKindsResp, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"kinds":[]}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mock := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
				mockDB:    mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mock)

			resources := v2.Resources{
				Graph: mock.mockGraph,
				DB:    mock.mockDB,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/graphs/kinds", resources.ListKinds).Methods("GET")

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_ListSourceKinds(t *testing.T) {
	t.Parallel()
	type mock struct {
		mockDB *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: GetSourceKinds database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL:    &url.URL{Path: "/api/v2/graphs/source-kinds"},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success: returns kinds with Sourceless appended - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL:    &url.URL{Path: "/api/v2/graphs/source-kinds"},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return([]model.SourceKind{
					{ID: 1, Name: ad.Entity.String()},
					{ID: 2, Name: azure.Entity.String()},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"kinds":[{"id":1,"name":"Base"},{"id":2,"name":"AZBase"},{"id":0,"name":"Sourceless"}]}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mock := &mock{
				mockDB: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mock)

			resources := v2.Resources{DB: mock.mockDB}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/graphs/source-kinds", resources.ListSourceKinds).Methods("GET")

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
