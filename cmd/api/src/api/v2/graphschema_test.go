// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_ListEdgeTypes(t *testing.T) {
	t.Parallel()
	type mock struct {
		mockDatabase *mocks.MockDatabase
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
			name: "Error: bad query parameter filter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graph-schema/edges",
						RawQuery: "badqueryparam=unknown:bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
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
						Path:     "/api/v2/graph-schema/edges",
						RawQuery: "badqueryparam=eq:bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
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
						Path:     "/api/v2/graph-schema/edges",
						RawQuery: "is_traversable=gt:true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"the specified filter predicate is not supported for this column: is_traversable gt"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/graph-schema/edges",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetGraphSchemaRelationshipKindsWithSchemaName(gomock.Any(), model.Filters{}, model.Sort{}, 0, 0).Return(model.GraphSchemaRelationshipKindsWithNamedSchema{}, 0, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success: list edge types - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/graph-schema/edges",
						RawQuery: "schema_names=eq:extension_a&is_traversable=eq:true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetGraphSchemaRelationshipKindsWithSchemaName(gomock.Any(), model.Filters{
					"schema.name": []model.Filter{{
						Operator:    model.Equals,
						Value:       "extension_a",
						SetOperator: model.FilterOr,
					}}, "is_traversable": []model.Filter{{
						Operator: model.Equals,
						Value:    "true",
					}},
				}, model.Sort{}, 0, 0).Return(model.GraphSchemaRelationshipKindsWithNamedSchema{
					model.GraphSchemaRelationshipKindWithNamedSchema{ID: 1, Name: "Edge_Kind_1", Description: "Edge Kind 1", IsTraversable: true, SchemaName: "extension_a"}, model.GraphSchemaRelationshipKindWithNamedSchema{ID: 2, Name: "Edge_Kind_2", Description: "Edge Kind 2", IsTraversable: true, SchemaName: "extension_a"}, model.GraphSchemaRelationshipKindWithNamedSchema{ID: 3, Name: "Edge_Kind_3", Description: "Edge Kind 3", IsTraversable: true, SchemaName: "extension_a"},
				}, 3, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody: `{
                    "data": [
                        {
                            "id": 1,
                            "name": "Edge_Kind_1",
                            "description": "Edge Kind 1",
                            "is_traversable": true,
                            "schema_name": "extension_a"
                        },
                        {
                            "id": 2,
                            "name": "Edge_Kind_2",
                            "description": "Edge Kind 2",
                            "is_traversable": true,
                            "schema_name": "extension_a"
                        },
                        {
                            "id": 3,
                            "name": "Edge_Kind_3",
                            "description": "Edge Kind 3",
                            "is_traversable": true,
                            "schema_name": "extension_a"
                        }]}`,
			},
		}}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB:         mocks.mockDatabase,
				Authorizer: auth.NewAuthorizer(mocks.mockDatabase),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/graph-schema/edges", resources.ListEdgeTypes).Methods("GET")

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)

		})
	}
}
