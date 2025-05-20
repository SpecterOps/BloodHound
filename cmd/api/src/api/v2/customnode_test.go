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
	"github.com/specterops/bloodhound/src/auth"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_CreateCustomNodeKindsTest(t *testing.T) {
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
			name: "Error: invalid icon type",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CreateCustomNodeRequest{
					CustomTypes: map[string]model.CustomNodeKindConfig{
						"KindA": {
							Icon: model.CustomNodeIcon{
								Type:  "font-stupid",
								Name:  "coffee",
								Color: "#FFFFFF",
							},
						},
					},
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
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"BadRequest: invalid icon type. only Font Awesome icons are supported"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: invalid hex color string",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CreateCustomNodeRequest{
					CustomTypes: map[string]model.CustomNodeKindConfig{
						"KindA": {
							Icon: model.CustomNodeIcon{
								Type:  "font-awesome",
								Name:  "coffee",
								Color: "FFFFFF",
							},
						},
					},
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
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"BadRequest: icon color must be a valid hexadecimal color string starting with '#' followed by 3 or 6 hex digits"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: created custom node kinds",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.CreateCustomNodeRequest{
					CustomTypes: map[string]model.CustomNodeKindConfig{
						"KindA": {
							Icon: model.CustomNodeIcon{
								Type:  "font-awesome",
								Name:  "coffee",
								Color: "#FFFFFF",
							},
						},
						"KindB": {
							Icon: model.CustomNodeIcon{
								Type:  "font-awesome",
								Name:  "house",
								Color: "#000",
							},
						},
					},
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
				mocks.mockDatabase.EXPECT().CreateCustomNodeKinds(req.Context(), gomock.Any()).Return(model.CustomNodeKinds{
					{
						ID:       1,
						KindName: "KindA",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{
								Type:  "font-awesome",
								Name:  "coffee",
								Color: "#FFFFFF",
							},
						},
					},
					{
						ID:       2,
						KindName: "KindB",
						Config: model.CustomNodeKindConfig{
							Icon: model.CustomNodeIcon{
								Type:  "font-awesome",
								Name:  "house",
								Color: "#000",
							},
						},
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusCreated,
				responseBody:   `{"data":[{"id":1,"kindName":"KindA","config":{"icon":{"type":"font-awesome","name":"coffee","color":"#FFFFFF"}}},{"id":2,"kindName":"KindB","config":{"icon":{"type":"font-awesome","name":"house","color":"#000"}}}]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: dbmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				DB:         mocks.mockDatabase,
				Authorizer: auth.NewAuthorizer(mocks.mockDatabase),
			}

			response := httptest.NewRecorder()
			resources.CreateCustomNodeKind(response, request)

			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_UpdateCustomNodeKindsTest(t *testing.T) {
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
			name: "Error: invalid icon type",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.UpdateCustomNodeKindRequest{
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{
							Type:  "font-stupid",
							Name:  "coffee",
							Color: "#FFFFFF",
						},
					},
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
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"BadRequest: invalid icon type. only Font Awesome icons are supported"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: invalid hex color string",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.UpdateCustomNodeKindRequest{
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{
							Type:  "font-awesome",
							Name:  "coffee",
							Color: "FFFFFF",
						},
					},
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
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"BadRequest: icon color must be a valid hexadecimal color string starting with '#' followed by 3 or 6 hex digits"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: created custom node kinds",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				payload := &v2.UpdateCustomNodeKindRequest{
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{
							Type:  "font-awesome",
							Name:  "coffee",
							Color: "#FFFFFF",
						},
					},
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
				mocks.mockDatabase.EXPECT().UpdateCustomNodeKind(req.Context(), gomock.Any()).Return(model.CustomNodeKind{
					ID:       1,
					KindName: "KindA",
					Config: model.CustomNodeKindConfig{
						Icon: model.CustomNodeIcon{
							Type:  "font-awesome",
							Name:  "coffee",
							Color: "#FFFFFF",
						},
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"id":1,"kindName":"KindA","config":{"icon":{"type":"font-awesome","name":"coffee","color":"#FFFFFF"}}}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: dbmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				DB:         mocks.mockDatabase,
				Authorizer: auth.NewAuthorizer(mocks.mockDatabase),
			}

			response := httptest.NewRecorder()
			resources.UpdateCustomNodeKind(response, request)

			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
