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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
)

func TestResources_GetUIConfiguration(t *testing.T) {
	t.Parallel()

	type args struct {
		enableUserAnalytics bool
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		args         args
		buildRequest func() *http.Request
		expected     expected
	}

	tt := []testData{
		{
			name: "Success: Retrieved enabled configuration - OK",
			args: args{
				enableUserAnalytics: true,
			},
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/ui-configuration",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"enable_user_analytics":true}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: Retrieved disabled configuration - OK",
			args: args{
				enableUserAnalytics: false,
			},
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/ui-configuration",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"enable_user_analytics":false}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := testCase.buildRequest()

			resources := v2.Resources{
				Config: config.Configuration{
					UI: config.UIConfiguration{
						EnableUserAnalytics: testCase.args.enableUserAnalytics,
					},
				},
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/ui-configuration", resources.GetUIConfiguration).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
