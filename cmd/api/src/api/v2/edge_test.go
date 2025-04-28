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
	graphmocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_GetEdgeComposition(t *testing.T) {
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
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: missing edge_type parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected edge_type parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: missing source_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected source_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test"}},
			},
		},
		{
			name: "Error: missing target_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&source_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected target_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&source_node=test"}},
			},
		},
		{
			name: "Error: edge_type is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&edge_type=test2&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one edge_type."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&edge_type=test2&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: source_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&source_node=test2&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one source_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&source_node=test2&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: target_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&target_node=test2&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Expected only one target_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&target_node=test2&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: invalid edge_type - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid edge requested: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: invalid startID for source_node - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=AZBase&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid value for startID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: invalid endID for targetNode - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=AZBase&source_node=1&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid value for endID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=test"}},
			},
		},
		{
			name: "Error: database error fetching edge by start and end - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Could not find edge matching criteria: error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
			},
		},
		{
			name: "Error: database error getting edge composition path - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Error getting composition for edge: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
			},
		},
		{
			name: "Success: retrieved edge composition - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"nodes":{},"edges":[]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
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
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				Graph: mocks.mockGraph,
			}

			response := httptest.NewRecorder()

			resources.GetEdgeComposition(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetEdgeRelayTargets_BadParameters(t *testing.T) {
	t.Parallel()

	type httpValues struct {
		Code   int
		Header http.Header
		Body   string
	}

	cases := []struct {
		Name     string
		Request  http.Request
		Expected httpValues
	}{
		{
			Name: "No Parameters",
			Request: http.Request{
				URL: &url.URL{},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
				Body:   `{"errors":[{"context":"","message":"Expected edge_type parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Missing Parameters",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase"}},
				Body:   `{"errors":[{"context":"","message":"Expected source_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Missing Parameters 2",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1"}},
				Body:   `{"errors":[{"context":"","message":"Expected target_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Wrong Number of Parameters",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2&edge_type=AZRole",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2&edge_type=AZRole"}},
				Body:   `{"errors":[{"context":"","message":"Expected only one edge_type."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Wrong Number of Parameters 2",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2&source_node=3",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2&source_node=3"}},
				Body:   `{"errors":[{"context":"","message":"Expected only one source_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Wrong Number of Parameters 3",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2&target_node=3",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2&target_node=3"}},
				Body:   `{"errors":[{"context":"","message":"Expected only one target_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Bad Parameter Type",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=LOREMIPSUM&source_node=1&target_node=2",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=LOREMIPSUM&source_node=1&target_node=2"}},
				Body:   `{"errors":[{"context":"","message":"Invalid edge requested: LOREMIPSUM"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Bad Parameter Type 2",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=GABAGOOL&target_node=2",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=GABAGOOL&target_node=2"}},
				Body:   `{"errors":[{"context":"","message":"Invalid value for startID: GABAGOOL"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Bad Parameter Type 3",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1.67&target_node=2",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1.67&target_node=2"}},
				Body:   `{"errors":[{"context":"","message":"Invalid value for startID: 1.67"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			Name: "Bad Parameter Type 4",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=lorem%20ipsum",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=lorem%20ipsum"}},
				Body:   `{"errors":[{"context":"","message":"Invalid value for endID: lorem ipsum"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
	}

	setupInternalState := func() v2.Resources {
		t.Helper()
		return v2.Resources{}
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			resources := setupInternalState()

			response := httptest.NewRecorder()

			resources.GetEdgeRelayTargets(response, &testCase.Request)
			mux.NewRouter().ServeHTTP(response, &testCase.Request)

			actualCode, actualHeader, actualBody := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.Expected.Code, actualCode)
			assert.Equal(t, testCase.Expected.Header, actualHeader)
			assert.Equal(t, testCase.Expected.Body, actualBody)
		})
	}
}

func TestResources_GetEdgeRelayTargets_CannotMatchEdge(t *testing.T) {
	t.Parallel()

	type httpValues struct {
		Code   int
		Header http.Header
		Body   string
	}

	cases := []struct {
		Name     string
		Request  http.Request
		Expected httpValues
	}{
		{
			Name: "Error Trying to get Matching Edge",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			Expected: httpValues{
				Code:   http.StatusBadRequest,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				Body:   `{"errors":[{"context":"","message":"Could not find edge matching criteria: Something went wrong"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
	}

	setupInternalState := func(ctx context.Context) v2.Resources {
		t.Helper()

		ctrl := gomock.NewController(t)
		mockGraph := graphmocks.NewMockDatabase(ctrl)
		mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(errors.New("Something went wrong"))

		res := v2.Resources{
			Graph: mockGraph,
		}

		return res
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			resources := setupInternalState(testCase.Request.Context())

			response := httptest.NewRecorder()

			resources.GetEdgeRelayTargets(response, &testCase.Request)
			mux.NewRouter().ServeHTTP(response, &testCase.Request)

			actualCode, actualHeader, actualBody := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.Expected.Code, actualCode)
			assert.Equal(t, testCase.Expected.Header, actualHeader)
			assert.Equal(t, testCase.Expected.Body, actualBody)
		})
	}
}

func TestResources_GetEdgeRelayTargets_CannotGetNodes(t *testing.T) {
	t.Parallel()

	type httpValues struct {
		Code   int
		Header http.Header
		Body   string
	}

	cases := []struct {
		Name     string
		Request  http.Request
		Expected httpValues
	}{
		{
			Name: "Error Trying to get Nodes",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			Expected: httpValues{
				Code:   http.StatusInternalServerError,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				Body:   `{"errors":[{"context":"","message":"Error getting composition for edge: Something went wrong"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
	}

	setupInternalState := func(ctx context.Context) v2.Resources {
		t.Helper()

		ctrl := gomock.NewController(t)
		mockGraph := graphmocks.NewMockDatabase(ctrl)
		mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)
		mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(errors.New("Something went wrong"))

		res := v2.Resources{
			Graph: mockGraph,
		}

		return res
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			resources := setupInternalState(testCase.Request.Context())

			response := httptest.NewRecorder()

			resources.GetEdgeRelayTargets(response, &testCase.Request)
			mux.NewRouter().ServeHTTP(response, &testCase.Request)

			actualCode, actualHeader, actualBody := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.Expected.Code, actualCode)
			assert.Equal(t, testCase.Expected.Header, actualHeader)
			assert.Equal(t, testCase.Expected.Body, actualBody)
		})
	}
}

func TestResources_GetEdgeRelayTargets_PositiveTest(t *testing.T) {
	t.Parallel()

	type httpValues struct {
		Code   int
		Header http.Header
		Body   string
	}

	cases := []struct {
		Name     string
		Request  http.Request
		Expected httpValues
	}{
		{
			Name: "Successful Request",
			Request: http.Request{
				URL: &url.URL{
					RawQuery: "edge_type=AZBase&source_node=1&target_node=2",
				},
			},
			Expected: httpValues{
				Code:   http.StatusOK,
				Header: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=AZBase&source_node=1&target_node=2"}},
				Body:   `{"data":{"nodes":{},"edges":[]}}`,
			},
		},
	}

	setupInternalState := func(ctx context.Context) v2.Resources {
		t.Helper()

		ctrl := gomock.NewController(t)
		mockGraph := graphmocks.NewMockDatabase(ctrl)
		mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)
		mockGraph.EXPECT().ReadTransaction(ctx, gomock.Any()).Return(nil)

		res := v2.Resources{
			Graph: mockGraph,
		}

		return res
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			resources := setupInternalState(testCase.Request.Context())

			response := httptest.NewRecorder()

			resources.GetEdgeRelayTargets(response, &testCase.Request)
			mux.NewRouter().ServeHTTP(response, &testCase.Request)

			actualCode, actualHeader, actualBody := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.Expected.Code, actualCode)
			assert.Equal(t, testCase.Expected.Header, actualHeader)
			assert.Equal(t, testCase.Expected.Body, actualBody)
		})
	}
}
