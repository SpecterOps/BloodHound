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
	"time"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	querymocks "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_GetEdgeComposition(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraph *graphmocks.MockDatabase
		mockDb    *mocks.MockDatabase
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
			name: "Error: GetPrimaryDisplayKindsError",
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
				mock.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
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
				mock.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
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

			mock := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
				mockDb:    mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mock)

			resources := v2.Resources{
				Graph: mock.mockGraph,
				DB:    mock.mockDb,
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

	type mock struct {
		ctrl           *gomock.Controller
		mockGraph      *graphmocks.MockDatabase
		mockGraphQuery *querymocks.MockGraph
		mockDb         *mocks.MockDatabase
	}

	cases := []struct {
		name             string
		request          http.Request
		expected         httpValues
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		testSetup        func(t *testing.T, ctx context.Context, mocks mock)
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {},
		},
		{
			name: "Error Fetching Source Node",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"source node error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(nil, errors.New("source node error"))
			},
		},
		{
			name: "Error Fetching Target Node",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"target node error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				sourceNode := graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
					common.ObjectID.String(): "source-object-id",
				}), ad.User)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(nil, errors.New("target node error"))
			},
		},
		{
			name: "Source Node Missing Object ID",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"property objectid: property not found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(graph.NewNode(graph.ID(1), graph.NewProperties(), ad.User), nil)
				targetNode := graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{
					common.ObjectID.String(): "target-object-id",
				}), ad.Computer)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
			},
		},
		{
			name: "Target Node Missing Object ID",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"property objectid: property not found"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				sourceNode := graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
					common.ObjectID.String(): "source-object-id",
				}), ad.User)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(graph.NewNode(graph.ID(2), graph.NewProperties(), ad.Computer), nil)
			},
		},
		{
			name: "Error Checking Source Node Access",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
			},
			user: model.User{EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "S-1-5-21-ALLOWED"}}},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"source access error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "source-object-id",
					}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "target-object-id",
					}), ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "source-object-id", ad.User).Return(nil, errors.New("source access error"))
			},
		},
		{
			name: "Source Node Access Denied",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
			},
			user: model.User{EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "S-1-5-21-ALLOWED"}}},
			expected: httpValues{
				code:   http.StatusNotFound,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "source-object-id",
					}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "target-object-id",
					}), ad.Computer)
					deniedSourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						ad.DomainSID.String(): "S-1-5-21-DENIED",
					}), ad.Entity, ad.User)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "source-object-id", ad.User).Return(deniedSourceNode, nil)
			},
		},
		{
			name: "Error Checking Target Node Access",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
			},
			user: model.User{EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "S-1-5-21-ALLOWED"}}},
			expected: httpValues{
				code:   http.StatusBadRequest,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"target access error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "source-object-id",
					}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "target-object-id",
					}), ad.Computer)
					allowedSourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						ad.DomainSID.String(): "S-1-5-21-ALLOWED",
					}), ad.Entity, ad.User)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "source-object-id", ad.User).Return(allowedSourceNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "target-object-id", ad.Computer).Return(nil, errors.New("target access error"))
			},
		},
		{
			name: "Target Node Access Denied",
			request: http.Request{
				URL: &url.URL{RawQuery: "edge_type=AZBase&source_node=1&target_node=2"},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
			},
			user: model.User{EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "S-1-5-21-ALLOWED"}}},
			expected: httpValues{
				code:   http.StatusNotFound,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "source-object-id",
					}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "target-object-id",
					}), ad.Computer)
					allowedSourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{
						ad.DomainSID.String(): "S-1-5-21-ALLOWED",
					}), ad.Entity, ad.User)
					deniedTargetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{
						ad.DomainSID.String(): "S-1-5-21-DENIED",
					}), ad.Entity, ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "source-object-id", ad.User).Return(allowedSourceNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "target-object-id", ad.Computer).Return(deniedTargetNode, nil)
			},
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{common.ObjectID.String(): "source-object-id"}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{common.ObjectID.String(): "target-object-id"}), ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(errors.New("Something went wrong"))
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{common.ObjectID.String(): "source-object-id"}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{common.ObjectID.String(): "target-object-id"}), ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(errors.New("Something went wrong"))
			},
		},
		{
			name: "GetPrimaryDisplayKindsError",
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			expected: httpValues{
				code:   http.StatusInternalServerError,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				body:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{common.ObjectID.String(): "source-object-id"}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{common.ObjectID.String(): "target-object-id"}), ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil).Times(2)
				mocks.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
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
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()
				var (
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{common.ObjectID.String(): "source-object-id"}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{common.ObjectID.String(): "target-object-id"}), ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil).Times(2)
				mocks.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
		},
		{
			name: `ETAC enabled, nodes hidden outside of assigned environment`,
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=CoerceAndRelayNTLMToLDAP&source_node=1&target_node=2",
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "S-1-5-21-ALLOWED"},
				},
			},
			expected: httpValues{
				code:   http.StatusOK,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=CoerceAndRelayNTLMToLDAP&source_node=1&target_node=2"}},
				body:   `{"data":{"nodes":{"3":{"label":"** Hidden Computer Object **","kind":"HIDDEN","kinds":[],"objectId":"HIDDEN","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z","hidden":true}},"edges":[],"literals":[]}}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()

				var (
					edge              = graph.NewRelationship(graph.ID(1), graph.ID(1), graph.ID(2), graph.NewProperties(), ad.CoerceAndRelayNTLMToLDAP)
					startNode         = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{ad.DomainSID.String(): "S-1-5-21-NOTALLOWED"}), ad.Entity)
					relayNode         = graph.NewNode(graph.ID(3), graph.AsProperties(map[string]any{ad.DomainSID.String(): "S-1-5-21-NOTALLOWED"}), ad.Computer)
					sourceNode        = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{common.ObjectID.String(): "source-object-id"}), ad.User)
					targetNode        = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{common.ObjectID.String(): "target-object-id"}), ad.Computer)
					allowedSourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{ad.DomainSID.String(): "S-1-5-21-ALLOWED"}), ad.Entity, ad.User)
					allowedTargetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{ad.DomainSID.String(): "S-1-5-21-ALLOWED"}), ad.Entity, ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "source-object-id", ad.User).Return(allowedSourceNode, nil)
				mocks.mockGraphQuery.EXPECT().GetEntityByObjectId(ctx, "target-object-id", ad.Computer).Return(allowedTargetNode, nil)

				// analysis.FetchEdgeByStartAndEnd
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					mockTransaction := graphmocks.NewMockTransaction(mocks.ctrl)
					mockRelationshipQuery := graphmocks.NewMockRelationshipQuery(mocks.ctrl)
					mockTransaction.EXPECT().Relationships().Return(mockRelationshipQuery)
					mockRelationshipQuery.EXPECT().Filter(gomock.Any()).Return(mockRelationshipQuery)
					mockRelationshipQuery.EXPECT().First().Return(edge, nil)
					return delegate(mockTransaction)
				})

				// ad.GetRelayTargets
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					return delegate(nil)
				})

				// GetVulnerableDomainControllersForRelayNTLMtoLDAP
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					mockTransaction := graphmocks.NewMockTransaction(mocks.ctrl)
					mockNodeQuery := graphmocks.NewMockNodeQuery(mocks.ctrl)
					mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
					mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery)
					mockNodeQuery.EXPECT().First().Return(startNode, nil)
					return delegate(mockTransaction)
				})

				// GetVulnerableDomainControllersForRelayNTLMtoLDAP
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					mockTransaction := graphmocks.NewMockTransaction(mocks.ctrl)
					mockNodeQuery := graphmocks.NewMockNodeQuery(mocks.ctrl)
					mockCursor := graphmocks.NewMockCursor[*graph.Node](mocks.ctrl)
					mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
					mockNodeQuery.EXPECT().Filter(gomock.Any()).Return(mockNodeQuery)
					mockNodeQuery.EXPECT().Fetch(gomock.Any()).DoAndReturn(func(cursorDelegate func(graph.Cursor[*graph.Node]) error, _ ...graph.Criteria) error {
						nodeChannel := make(chan *graph.Node, 1)
						nodeChannel <- relayNode
						close(nodeChannel)
						mockCursor.EXPECT().Chan().Return(nodeChannel)
						mockCursor.EXPECT().Error().Return(nil)
						return cursorDelegate(mockCursor)
					})
					return delegate(mockTransaction)
				})

				mocks.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
		},
		{
			name: `ETAC disabled, nodes not hidden`,
			request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=CoerceAndRelayNTLMToLDAP&source_node=1&target_node=2",
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: false,
				},
			},
			user: model.User{
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "S-1-5-21-ALLOWED"},
				},
			},
			expected: httpValues{
				code:   http.StatusOK,
				header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=CoerceAndRelayNTLMToLDAP&source_node=1&target_node=2"}},
				body:   `{"data":{"nodes":{"3":{"label":"COMPUTER1","kind":"Computer","kinds":["Computer"],"objectId":"S-1-5-21-NOTALLOWED-1000","isTierZero":false,"isOwnedObject":false,"lastSeen":"2025-01-01T00:00:00Z","properties":{"domainsid":"S-1-5-21-NOTALLOWED","lastseen":"2025-01-01T00:00:00Z","name":"COMPUTER1","objectid":"S-1-5-21-NOTALLOWED-1000"}}},"edges":[],"literals":[]}}`,
			},
			testSetup: func(t *testing.T, ctx context.Context, mocks mock) {
				t.Helper()

				var (
					edge       = graph.NewRelationship(graph.ID(1), graph.ID(1), graph.ID(2), graph.NewProperties(), ad.CoerceAndRelayNTLMToLDAP)
					startNode  = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{ad.DomainSID.String(): "S-1-5-21-NOTALLOWED"}), ad.Entity)
					sourceNode = graph.NewNode(graph.ID(1), graph.AsProperties(map[string]any{common.ObjectID.String(): "source-object-id"}), ad.User)
					targetNode = graph.NewNode(graph.ID(2), graph.AsProperties(map[string]any{common.ObjectID.String(): "target-object-id"}), ad.Computer)
					relayNode  = graph.NewNode(graph.ID(3), graph.AsProperties(map[string]any{
						common.ObjectID.String(): "S-1-5-21-NOTALLOWED-1000",
						common.Name.String():     "COMPUTER1",
						common.LastSeen.String(): time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						ad.DomainSID.String():    "S-1-5-21-NOTALLOWED",
					}), ad.Computer)
				)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(1)).Return(sourceNode, nil)
				mocks.mockGraphQuery.EXPECT().FetchNodeByGraphId(ctx, graph.ID(2)).Return(targetNode, nil)

				// analysis.FetchEdgeByStartAndEnd
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					mockTransaction := graphmocks.NewMockTransaction(mocks.ctrl)
					mockRelationshipQuery := graphmocks.NewMockRelationshipQuery(mocks.ctrl)
					mockTransaction.EXPECT().Relationships().Return(mockRelationshipQuery)
					mockRelationshipQuery.EXPECT().Filter(gomock.Any()).Return(mockRelationshipQuery)
					mockRelationshipQuery.EXPECT().First().Return(edge, nil)
					return delegate(mockTransaction)
				})

				// ad.GetRelayTargets
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					return delegate(nil)
				})

				// GetVulnerableDomainControllersForRelayNTLMtoLDAP
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					mockTransaction := graphmocks.NewMockTransaction(mocks.ctrl)
					mockNodeQuery := graphmocks.NewMockNodeQuery(mocks.ctrl)
					mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
					mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery)
					mockNodeQuery.EXPECT().First().Return(startNode, nil)
					return delegate(mockTransaction)
				})

				// GetVulnerableDomainControllersForRelayNTLMtoLDAP
				mocks.mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...any) error {
					mockTransaction := graphmocks.NewMockTransaction(mocks.ctrl)
					mockNodeQuery := graphmocks.NewMockNodeQuery(mocks.ctrl)
					mockCursor := graphmocks.NewMockCursor[*graph.Node](mocks.ctrl)
					mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
					mockNodeQuery.EXPECT().Filter(gomock.Any()).Return(mockNodeQuery)
					mockNodeQuery.EXPECT().Fetch(gomock.Any()).DoAndReturn(func(cursorDelegate func(graph.Cursor[*graph.Node]) error, _ ...graph.Criteria) error {
						nodeChannel := make(chan *graph.Node, 1)
						nodeChannel <- relayNode
						close(nodeChannel)
						mockCursor.EXPECT().Chan().Return(nodeChannel)
						mockCursor.EXPECT().Error().Return(nil)
						return cursorDelegate(mockCursor)
					})
					return delegate(mockTransaction)
				})

				mocks.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl           = gomock.NewController(t)
				mockGraph      = graphmocks.NewMockDatabase(ctrl)
				mockGraphQuery = querymocks.NewMockGraph(ctrl)
				mockDb         = mocks.NewMockDatabase(ctrl)
				resources      = v2.Resources{
					Graph:      mockGraph,
					GraphQuery: mockGraphQuery,
					DB:         mockDb,
					DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
				}
				request = testCase.request.WithContext(setupUserCtx(testCase.user))
			)

			testCase.testSetup(t, request.Context(), mock{ctrl: ctrl, mockGraph: mockGraph, mockGraphQuery: mockGraphQuery, mockDb: mockDb})

			response := httptest.NewRecorder()

			resources.GetEdgeRelayTargets(response, request)
			mux.NewRouter().ServeHTTP(response, request)

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
		mockDb    *mocks.MockDatabase
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
			name: "Error: GetPrimaryDisplayKindsError",
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
				mock.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
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
				mock.mockDb.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
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

			mock := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
				mockDb:    mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mock)

			resources := v2.Resources{
				Graph: mock.mockGraph,
				DB:    mock.mockDb,
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
