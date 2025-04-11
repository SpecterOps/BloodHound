// Copyright 2024 Specter Ops, Inc.
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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetCollectorManifest(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		manifests = config.CollectorManifests{"sharphound": config.CollectorManifest{}, "azurehound": config.CollectorManifest{}}
		resources = v2.Resources{
			CollectorManifests: manifests,
		}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/collectors/%s"

	t.Run("sharphound", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "sharphound"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("azurehound", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "azurehound"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("invalid", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "invalid"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		resources := v2.Resources{CollectorManifests: map[string]config.CollectorManifest{}}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "azurehound"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestManagementResource_DownloadCollectorByVersion(t *testing.T) {
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name                string
		buildRequest        func() *http.Request
		createCollectorFile func(t *testing.T) *os.File
		expected            expected
	}
	tt := []testData{
		{
			name: "Error: invalid collector type - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "",
					},
				}

				return request
			},
			createCollectorFile: func(t *testing.T) *os.File {
				return nil
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   ``,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: collector type does not exist in collector manifest map - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "latest",
					"collector_type": "sharphound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				return nil
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: os.ReadFile error when retrieving collector file - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "latest",
					"collector_type": "azurehound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				return nil
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: download collector not-latest release - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "1.0.0",
					"collector_type": "azurehound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				err := os.Mkdir("azurehound", 0755)
				if err != nil {
					if !errors.Is(err, os.ErrExist) {
						t.Fatalf("error using os.Mkdir to create test file directory: %v", err)
					}
				}

				file, err := os.Create("azurehound/azurehound-1.0.0.zip")
				if err != nil {
					t.Fatalf("error using os.Create to create test file: %v", err)
				}
				return file
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-1.0.0.zip\""}, "Content-Type": []string{"application/octet-stream"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: download collector latest release - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "latest",
					"collector_type": "azurehound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				err := os.Mkdir("azurehound", 0755)
				if err != nil {
					if !errors.Is(err, os.ErrExist) {
						t.Fatalf("error using os.Mkdir to create test file directory: %v", err)
					}
				}

				file, err := os.Create("azurehound/azurehound-latest.zip")
				if err != nil {
					t.Fatalf("error using os.Create to create test file: %v", err)
				}
				return file
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-latest.zip\""}, "Content-Type": []string{"application/octet-stream"}, "Location": []string{"/"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.buildRequest()
			testFile := testCase.createCollectorFile(t)
			if testFile != nil {
				defer func() {
					err := os.RemoveAll("azurehound")
					if err != nil {
						t.Fatalf("error removing test file: %v", err)
					}
				}()
			}

			collectorManifests := map[string]config.CollectorManifest{
				"azurehound": {
					Latest:   "latest",
					Versions: []config.CollectorVersion{},
				},
			}

			resources := v2.Resources{
				CollectorManifests: collectorManifests,
			}

			response := httptest.NewRecorder()

			resources.DownloadCollectorByVersion(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			if body != "" {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			}
		})
	}
}

func TestManagementResource_DownloadCollectorChecksumByVersion(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name                string
		buildRequest        func() *http.Request
		createCollectorFile func(t *testing.T) *os.File
		expected            expected
	}
	tt := []testData{
		{
			name: "Error: invalid collector type - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "",
					},
				}

				return request
			},
			createCollectorFile: func(t *testing.T) *os.File {
				return nil
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: collector type does not exist in collector manifest map - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "latest",
					"collector_type": "sharphound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				return nil
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: os.ReadFile error when retrieving collector file - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "latest",
					"collector_type": "sharphound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				return nil
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: download collector not-latest release - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "1.0.0",
					"collector_type": "azurehound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				err := os.Mkdir("azurehound", 0755)
				if err != nil {
					if !errors.Is(err, os.ErrExist) {
						t.Fatalf("error using os.Mkdir to create test file directory: %v", err)
					}
				}

				file, err := os.Create("azurehound/azurehound-1.0.0.zip.sha256")
				if err != nil {
					t.Fatalf("error using os.Create to create test file: %v", err)
				}
				return file
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-1.0.0.zip.sha256\""}, "Content-Type": []string{"application/octet-stream"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: download collector latest release - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}
				param := map[string]string{
					"release_tag":    "latest",
					"collector_type": "azurehound",
				}

				return mux.SetURLVars(request, param)
			},
			createCollectorFile: func(t *testing.T) *os.File {
				err := os.Mkdir("azurehound", 0755)
				if err != nil {
					if !errors.Is(err, os.ErrExist) {
						t.Fatalf("error using os.Mkdir to create test file directory: %v", err)
					}
				}

				file, err := os.Create("azurehound/azurehound-latest.zip.sha256")
				if err != nil {
					t.Fatalf("error using os.Create to create test file: %v", err)
				}
				return file
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"azurehound-latest.zip.sha256\""}, "Content-Type": []string{"application/octet-stream"}, "Location": []string{"/"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			request := testCase.buildRequest()
			testFile := testCase.createCollectorFile(t)
			if testFile != nil {
				defer func() {
					err := os.RemoveAll("azurehound")
					if err != nil {
						t.Fatalf("error removing test file: %v", err)
					}
				}()
			}

			collectorManifests := map[string]config.CollectorManifest{
				"azurehound": {
					Latest:   "latest",
					Versions: []config.CollectorVersion{},
				},
			}

			resources := v2.Resources{
				CollectorManifests: collectorManifests,
			}

			response := httptest.NewRecorder()

			resources.DownloadCollectorChecksumByVersion(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			if body != "" {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			}
		})
	}
}
