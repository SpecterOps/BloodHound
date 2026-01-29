// Copyright 2025 Specter Ops, Inc.
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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_GetEdgeComposition(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraph *graphmocks.MockDatabase
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
			name: "Error: missing edge_type parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/edge-composition",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected edge_type parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: missing source_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected source_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: missing target_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=test&source_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected target_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: edge_type is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=test&edge_type=test2&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one edge_type."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: source_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=test&source_node=test2&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one source_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: target_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=test&target_node=test2&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one target_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid edge_type - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=test&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid edge requested: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid startID for source_node - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=AZBase&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid value for startID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid endID for targetNode - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=AZBase&source_node=1&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid value for endID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: database error fetching edge by start and end - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Could not find edge matching criteria: error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: database error getting edge composition path - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Error getting composition for edge: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: retrieved edge composition - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/edge-composition",
						RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"nodes":{},"edges":[],"literals":[]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph: mocks.mockGraph,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/graphs/edge-composition", resources.GetEdgeComposition).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetEdgeRelayTargets(t *testing.T) {
	t.Parallel()

	type httpValues struct {
		code   int
		header http.Header
		body   string
	}

	cases := []struct {
		name      string
		request   http.Request
		expected  httpValues
		testSetup func(t *testing.T, ctx context.Context, res *v2.Resources)
	}{
		{
			name: "No Parameters",
			request: http.Request{
				URL: &url.URL{},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
				body:   `{"errors":[{"context":"","message":"Expected edge_type parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Missing Parameters",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase"}},
				body:   `{"errors":[{"context":"","message":"Expected source_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Missing Parameters 2",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1"}},
				body:   `{"errors":[{"context":"","message":"Expected target_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Wrong Number of Parameters",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2&edge_type=AZRole",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2&edge_type=AZRole"}},
				body:   `{"errors":[{"context":"","message":"Expected only one edge_type."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Wrong Number of Parameters 2",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2&source_node=3",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2&source_node=3"}},
				body:   `{"errors":[{"context":"","message":"Expected only one source_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Wrong Number of Parameters 3",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2&target_node=3",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2&target_node=3"}},
				body:   `{"errors":[{"context":"","message":"Expected only one target_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Bad Parameter Type",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=LOREMIPSUM&source_node=1&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=LOREMIPSUM&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"Invalid edge requested: LOREMIPSUM"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Bad Parameter Type 2",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=GABAGOOL&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=GABAGOOL&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"Invalid value for startID: GABAGOOL"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Bad Parameter Type 3",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1.67&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1.67&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"Invalid value for startID: 1.67"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Bad Parameter Type 4",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=lorem%20ipsum",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=lorem%20ipsum"}},
				body:   `{"errors":[{"context":"","message":"Invalid value for endID: lorem ipsum"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {},
		},
		{
			name: "Error Trying to get Matching Edge",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"Could not find edge matching criteria: Something went wrong"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {
				t.Helper()
				ctrl := gomock.NewController(t)
				mockGraph := graphmocks.NewMockDatabase(ctrl)
				mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(errors.New("Something went wrong"))

				res.Graph = mockGraph
			},
		},
		{
			name: "Error Trying to get Nodes",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusInternalServerError,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"Error getting composition for edge: Something went wrong"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {
				t.Helper()

				ctrl := gomock.NewController(t)
				mockGraph := graphmocks.NewMockDatabase(ctrl)
				mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)
				mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(errors.New("Something went wrong"))

				res.Graph = mockGraph
			},
		},
		{
			name: "Successful Request",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusOK,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"data":{"nodes":{},"edges":[],"literals":[]}}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, res *v2.Resources) {
				t.Helper()

				ctrl := gomock.NewController(t)
				mockGraph := graphmocks.NewMockDatabase(ctrl)
				mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)
				mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)

				res.Graph = mockGraph
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resources := v2.Resources{}

			testCase.testSetup(t, testCase.request.Context(), &resources)

			response := httptest.NewRecorder()

			resources.GetEdgeRelayTargets(response, &testCase.request)
			mux.NewRouter().ServeHTTP(response, &testCase.request)

			actualCode, actualHeader, actualBody := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.code, actualCode)
			assert.Equal(t, testCase.expected.header, actualHeader)
			assert.Equal(t, testCase.expected.body, actualBody)
		})
	}
}

func TestResources_GetEdgeACLInheritancePath(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraph *graphmocks.MockDatabase
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
			name: "Error: missing edge_type parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/acl-inheritance",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected edge_type parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: missing source_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected source_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: missing target_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=test&source_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected target_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: edge_type is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=test&edge_type=test2&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one edge_type."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: source_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=test&source_node=test2&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one source_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: target_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=test&target_node=test2&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one target_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid edge_type - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=test&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid edge requested: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid startID for source_node - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=GenericAll&source_node=test&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid value for startID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid endID for targetNode - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=GenericAll&source_node=1&target_node=test",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid value for endID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: database error fetching edge by start and end - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=GenericAll&source_node=1&target_node=2",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Could not find edge matching criteria: error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: database error getting ACL inheritance path - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=GenericAll&source_node=1&target_node=2",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Error getting ACL inheritance path for edge: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: retrieved ACL inheritance - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graphs/acl-inheritance",
						RawQuery: "edge_type=GenericAll&source_node=1&target_node=2",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"nodes":{},"edges":[],"literals":[]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph: mocks.mockGraph,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/graphs/acl-inheritance", resources.GetEdgeACLInheritancePath).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
