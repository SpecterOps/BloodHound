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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_CypherQuery(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *dbmocks.MockDatabase
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
			name: "Error: empty request body - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"JSON malformed."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: GraphQuery.PrepareCypherQuery error - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}

				request.Header.Add("Content-type", "application/json")
				request.Body = io.NopCloser(bytes.NewReader(jsonPayload))

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: HasMutation auth error - Forbidden",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}

				request.Header.Add("Content-type", "application/json")
				request.Body = io.NopCloser(bytes.NewReader(jsonPayload))

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: true,
				}, nil)
				mocks.mockDatabase.EXPECT().AppendAuditLog(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Permission denied: User may not modify the graph."}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: NeoTimeoutError - InternalServerError",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}

				request.Header.Add("Content-type", "application/json")
				request.Body = io.NopCloser(bytes.NewReader(jsonPayload))

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(req.Context(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{}, &neo4j.Neo4jError{})
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Neo4jError:  ()"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: empty nodes and edges - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}

				request.Header.Add("Content-type", "application/json")
				request.Body = io.NopCloser(bytes.NewReader(jsonPayload))

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(req.Context(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"resource not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: nodes and edges returned valid - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}

				request.Header.Add("Content-type", "application/json")
				request.Body = io.NopCloser(bytes.NewReader(jsonPayload))

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(req.Context(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{
					Nodes: map[string]model.UnifiedNode{
						"1": {
							Label: "label",
						},
					},
					Edges: []model.UnifiedEdge{
						{
							Source: "source",
						},
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"nodes":{"1":{"label":"label","kind":"","objectId":"","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z"}},"edges":[{"source":"source","target":"","label":"","kind":"","lastSeen":"0001-01-01T00:00:00Z"}]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDatabase:   dbmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				Authorizer: auth.NewAuthorizer(mocks.mockDatabase),
			}

			response := httptest.NewRecorder()

			resources.CypherQuery(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
