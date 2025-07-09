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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	fsmocks "github.com/specterops/bloodhound/cmd/api/src/services/fs/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_DownloadCollectorByVersion(t *testing.T) {
	type mock struct {
		mockFS *fsmocks.MockService
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
			name: "Error: invalid collector type - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/InvalidCollectorType/latest",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid collector type: InvalidCollectorType"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: collector type does not exist in collector manifest map - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/sharphound/latest",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: os.ReadFile error when retrieving collector file - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/azurehound/latest",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockFS.EXPECT().ReadFile("azurehound/azurehound-latest.zip").Return([]byte{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: download collector not-latest release - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/azurehound/v1.0.0",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockFS.EXPECT().ReadFile("azurehound/azurehound-v1.0.0.zip").Return([]byte{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-v1.0.0.zip\""}, "Content-Type": []string{"application/octet-stream"}},
			},
		},
		{
			name: "Success: download collector latest release - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/azurehound/latest",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockFS.EXPECT().ReadFile("azurehound/azurehound-latest.zip").Return([]byte{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-latest.zip\""}, "Content-Type": []string{"application/octet-stream"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.buildRequest()

			ctrl := gomock.NewController(t)
			mock := &mock{
				mockFS: fsmocks.NewMockService(ctrl),
			}
			testCase.setupMocks(t, mock)

			collectorManifests := map[string]config.CollectorManifest{
				"azurehound": {
					Latest:   "latest",
					Versions: []config.CollectorVersion{},
				},
			}

			resources := v2.Resources{
				CollectorManifests: collectorManifests,
				FileService:        mock.mockFS,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/collectors/{%s}/{%s:v[0-9]+.[0-9]+.[0-9]+|latest}", v2.CollectorTypePathParameterName, v2.CollectorReleaseTagPathParameterName), resources.DownloadCollectorByVersion).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if body != "" {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}

func TestResources_DownloadCollectorChecksumByVersion(t *testing.T) {
	type mock struct {
		mockFS *fsmocks.MockService
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
			name: "Error: invalid collector type - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/InvalidCollectorType/latest/checksum",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Invalid collector type: InvalidCollectorType"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: collector type does not exist in collector manifest map - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/sharphound/latest/checksum",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: os.ReadFile error when retrieving collector file - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/azurehound/latest/checksum",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockFS.EXPECT().ReadFile("azurehound/azurehound-latest.zip.sha256").Return([]byte{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: download collector not-latest release - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/azurehound/v1.0.0/checksum",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockFS.EXPECT().ReadFile("azurehound/azurehound-v1.0.0.zip.sha256").Return([]byte{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-v1.0.0.zip.sha256\""}, "Content-Type": []string{"application/octet-stream"}},
			},
		},
		{
			name: "Success: download collector latest release - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/collectors/azurehound/latest/checksum",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockFS.EXPECT().ReadFile("azurehound/azurehound-latest.zip.sha256").Return([]byte{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-latest.zip.sha256\""}, "Content-Type": []string{"application/octet-stream"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.buildRequest()

			collectorManifests := map[string]config.CollectorManifest{
				"azurehound": {
					Latest:   "latest",
					Versions: []config.CollectorVersion{},
				},
			}

			ctrl := gomock.NewController(t)
			mock := &mock{
				mockFS: fsmocks.NewMockService(ctrl),
			}
			testCase.setupMocks(t, mock)

			resources := v2.Resources{
				CollectorManifests: collectorManifests,
				FileService:        mock.mockFS,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/collectors/{%s}/{%s:v[0-9]+.[0-9]+.[0-9]+|latest}/checksum", v2.CollectorTypePathParameterName, v2.CollectorReleaseTagPathParameterName), resources.DownloadCollectorChecksumByVersion).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if body != "" {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}
