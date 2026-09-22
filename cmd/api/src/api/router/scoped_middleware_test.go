// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	testutil "github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_WithRouteMiddleware(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode    int
		responseBody    string
		middlewareCalls int
	}
	type testData struct {
		name           string
		path           string
		registerRoutes func(*router.Router, func() mux.MiddlewareFunc) error
		expected       expected
	}

	tt := []testData{
		{
			name: "Success: migrated route receives one scoped middleware - 200",
			path: "/api/v2/migrated",
			registerRoutes: func(routerInst *router.Router, scopedMiddleware func() mux.MiddlewareFunc) error {
				return routerInst.WithRouteMiddleware(scopedMiddleware, func() error {
					routerInst.GET("/api/v2/migrated", func(response http.ResponseWriter, _ *http.Request) {
						_, _ = response.Write([]byte("migrated"))
					})
					return nil
				})
			},
			expected: expected{
				responseCode:    http.StatusOK,
				responseBody:    "migrated",
				middlewareCalls: 1,
			},
		},
		{
			name: "Success: legacy route receives no scoped middleware - 200",
			path: "/api/v2/legacy",
			registerRoutes: func(routerInst *router.Router, _ func() mux.MiddlewareFunc) error {
				routerInst.GET("/api/v2/legacy", func(response http.ResponseWriter, _ *http.Request) {
					_, _ = response.Write([]byte("legacy"))
				})
				return nil
			},
			expected: expected{
				responseCode:    http.StatusOK,
				responseBody:    "legacy",
				middlewareCalls: 0,
			},
		},
	}

	for _, testCase := range tt {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var middlewareCalls int
			routerInst := router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
			scopedMiddleware := func() mux.MiddlewareFunc {
				return func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
						middlewareCalls++
						next.ServeHTTP(response, request)
					})
				}
			}
			require.NoError(t, testCase.registerRoutes(&routerInst, scopedMiddleware))

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			response := httptest.NewRecorder()
			routerInst.Handler().ServeHTTP(response, request)

			status, _, body := testutil.ProcessResponse(t, response)
			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseBody, body)
			assert.Equal(t, testCase.expected.middlewareCalls, middlewareCalls)
		})
	}
}
