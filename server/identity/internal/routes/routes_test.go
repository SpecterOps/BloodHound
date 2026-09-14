// Copyright 2026 Specter Ops, Inc.
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

package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/identity/internal/routes"
	"github.com/stretchr/testify/assert"
)

func TestRegister_RegistersRoutes(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		routeRegistered bool
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	tests := []testData{
		{
			name: "Success: GET /api/v2/roles is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/roles/{role_id} is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/permissions is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/permissions/{permission_id} is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				cfg        = config.Configuration{}
				authorizer = auth.NewAuthorizer(nil)
				routerInst = router.NewRouter(cfg, authorizer, "")
				mock       = mock{identity: mocks.NewMockIdentity(t)}
				handlerSet = handlers.NewHandlersContainer(mock.identity)
				match      mux.RouteMatch
			)

			routes.Register(&routerInst, handlerSet)
			testCase.setupMocks(mock)

			assert.Equal(t, testCase.expected.routeRegistered, routerInst.MuxRouter().Match(testCase.buildRequest(), &match))
		})
	}
}

// TestRegister_RoutesRequireAuthentication dispatches real unauthenticated requests
// through the wired router to verify that the registered routes are guarded by
// authentication middleware. If the route wireup ever loses RequirePermissions,
// this test will fail.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode int
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	tests := []testData{
		{
			name: "Error: unauthenticated GET /api/v2/roles - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/roles/{role_id} - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/permissions - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/permissions/{permission_id} - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				cfg        = config.Configuration{}
				authorizer = auth.NewAuthorizer(nil)
				routerInst = router.NewRouter(cfg, authorizer, "")
				mock       = mock{identity: mocks.NewMockIdentity(t)}
				handlerSet = handlers.NewHandlersContainer(mock.identity)
				recorder   = httptest.NewRecorder()
			)

			routes.Register(&routerInst, handlerSet)
			testCase.setupMocks(mock)
			routerInst.Handler().ServeHTTP(recorder, testCase.buildRequest())

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
		})
	}
}
