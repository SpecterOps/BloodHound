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
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
)

func TestResources_CypherQuery(t *testing.T) {
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
		setupMocks       func(t *testing.T, mock *mock)
		expected         expected
		dogTagsOverrides dogtags.TestOverrides
	}

	tt := []testData{
		{
			name: "Test Ordered node keys",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{
					Nodes: map[string]model.UnifiedNode{
						"1": {
							Label:      "label",
							Properties: map[string]any{"apple": "snake", "zebra": "elmo", "key": "value", "ball": "value"},
						},
					},
					Edges: []model.UnifiedEdge{
						{
							Properties: map[string]any{"apple": "snake", "zebra": "elmo", "key": "value", "ball": "value"},
							Source:     "source",
						},
					},
					Literals: graph.Literals{},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"node_keys": ["apple", "ball", "key", "zebra"], "edge_keys": ["apple", "ball", "key", "zebra"], "nodes":{"1":{"label":"label","properties": {"apple": "snake", "zebra": "elmo", "key": "value", "ball": "value"},"kind":"","kinds":null, "objectId":"","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z"}},"edges":[{"source":"source","target":"","label":"","properties": {"apple": "snake", "zebra": "elmo", "key": "value", "ball": "value"},"kind":"","lastSeen":"0001-01-01T00:00:00Z"}],"literals":[]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: empty request body - Bad Request",
			buildRequest: func() *http.Request {

				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},

					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"JSON malformed."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GraphQuery.PrepareCypherQuery error - Bad Request",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: HasMutation auth error - Forbidden",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}

				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: true,
				}, nil)
				mocks.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Permission denied: User may not modify the graph."}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: NeoTimeoutError - InternalServerError",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{}, &neo4j.Neo4jError{})
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Neo4jError:  ()"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: empty nodes and edges - Not Found",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"resource not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: nodes and edges returned valid - OK",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{
					Nodes: map[string]model.UnifiedNode{
						"1": {
							Label:      "label",
							Properties: map[string]any{"key": "value"},
							Hidden:     false,
						},
					},
					Edges: []model.UnifiedEdge{
						{
							Source: "source",
						},
					},
					Literals: graph.Literals{},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"edges":[{"kind":"","label":"","lastSeen":"0001-01-01T00:00:00Z","source":"source","target":""}],"literals": [],"node_keys":["key"],"nodes":{"1":{"isOwnedObject":false,"isTierZero":false,"kind":"","kinds":null,"label":"label","lastSeen":"0001-01-01T00:00:00Z","objectId":"","properties":{"key":"value"}}}}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: ETAC enabled, user all envs - OK",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: true,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{
					Nodes: map[string]model.UnifiedNode{
						"1": {
							Label:      "label",
							Properties: map[string]any{"key": "value"},
							Hidden:     false,
						},
					},
					Edges: []model.UnifiedEdge{
						{
							Source: "source",
						},
					},
					Literals: graph.Literals{},
				}, nil)
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"edges":[{"kind":"","label":"","lastSeen":"0001-01-01T00:00:00Z","source":"source","target":""}],"literals":[], "node_keys":["key"],"nodes":{"1":{"isOwnedObject":false,"isTierZero":false,"kind":"","kinds":null,"label":"label","lastSeen":"0001-01-01T00:00:00Z","objectId":"","properties":{"key":"value"}}}}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: ETAC enabled, user filtered node response - OK",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: false,
					EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
						{EnvironmentID: "testenv"},
					},
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{
					Nodes: map[string]model.UnifiedNode{
						"1": {
							Label:      "label",
							Properties: map[string]any{"domainsid": "testenv"},
							Kinds:      []string{"kinds"},
							Hidden:     false,
						},
						"source": {
							Label:      "labelSource",
							Properties: map[string]any{"domainsid": "testenv"},
							Kinds:      []string{"kinds"},
							Hidden:     false,
						},
						"2": {
							Label:      "label2",
							Properties: map[string]any{"domainsid": "value"},
							Kinds:      []string{"kinds"},
							Hidden:     true,
						},
					},
					Edges: []model.UnifiedEdge{
						{Source: "source", Target: "1"},
						{Source: "source", Target: "2"},
						{Source: "2", Target: "1"},
					},
					Literals: graph.Literals{},
				}, nil)
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"edges":[{"kind":"","label":"","lastSeen":"0001-01-01T00:00:00Z","source":"source","target":"1"},{"kind":"HIDDEN","label":"** Hidden Edge **","lastSeen":"0001-01-01T00:00:00Z","source":"source","target":"2"},{"kind":"HIDDEN","label":"** Hidden Edge **","lastSeen":"0001-01-01T00:00:00Z","source":"2","target":"1"}],"literals":[],"node_keys":["domainsid"],"nodes":{"1":{"isOwnedObject":false,"isTierZero":false,"kind":"","kinds":["kinds"],"label":"label","lastSeen":"0001-01-01T00:00:00Z","objectId":"","properties":{"domainsid":"testenv"}},"2":{"hidden":true,"isOwnedObject":false,"isTierZero":false,"kind":"HIDDEN","kinds":[],"label":"** Hidden kinds Object **","lastSeen":"0001-01-01T00:00:00Z","objectId":"HIDDEN"},"source":{"isOwnedObject":false,"isTierZero":false,"kind":"","kinds":["kinds"],"label":"labelSource","lastSeen":"0001-01-01T00:00:00Z","objectId":"","properties":{"domainsid":"testenv"}}}}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: ETAC enabled, user has no access, hidden graph - 200",
			buildRequest: func() *http.Request {
				payload := &v2.CypherQueryPayload{
					Query:             "query",
					IncludeProperties: true,
				}
				jsonPayload, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("error occurred while marshaling payload necessary for test: %v", err)
				}
				user := model.User{
					AllEnvironments: false,
				}
				userCtx := setupUserCtx(user)

				req := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graphs/cypher",
					},
					Body: io.NopCloser(bytes.NewReader(jsonPayload)),
					Header: http.Header{
						headers.ContentType.String(): []string{
							"application/json",
						},
					},
					Method: http.MethodPost,
				}
				req = req.WithContext(userCtx)
				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockGraphQuery.EXPECT().PrepareCypherQuery("query", int64(queries.DefaultQueryFitnessLowerBoundExplore)).Return(queries.PreparedQuery{
					HasMutation: false,
				}, nil)
				mocks.mockGraphQuery.EXPECT().RawCypherQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.UnifiedGraph{
					Nodes: map[string]model.UnifiedNode{
						"1": {
							Label:      "label",
							Properties: map[string]any{"domainsid": "testenv"},
							Kinds:      []string{"kinds"},
						},
						"2": {
							Label:      "label2",
							Properties: map[string]any{"domainsid": "value"},
							Kinds:      []string{"kinds"},
						},
					},
					Edges: []model.UnifiedEdge{
						{
							Source: "source",
							Target: "1",
						},
					},
					Literals: graph.Literals{},
				}, nil)
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"nodes":{"1":{"hidden":true,"isOwnedObject":false,"isTierZero":false,"kind":"HIDDEN","kinds":[],"label":"** Hidden kinds Object **","lastSeen":"0001-01-01T00:00:00Z","objectId":"HIDDEN"},"2":{"hidden":true,"isOwnedObject":false,"isTierZero":false,"kind":"HIDDEN","kinds":[],"label":"** Hidden kinds Object **","lastSeen":"0001-01-01T00:00:00Z","objectId":"HIDDEN"}},"edges":[{"source":"source","target":"1","label":"** Hidden Edge **","kind":"HIDDEN","lastSeen":"0001-01-01T00:00:00Z"}],"literals":[]}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
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
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				Authorizer: auth.NewAuthorizer(mocks.mockDatabase),
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/graphs/cypher", resources.CypherQuery).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
