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

package registration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	databaseMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegistration_WithRouteMiddlewareUsesIndependentRateLimitsAcrossRoutes(t *testing.T) {
	t.Parallel()

	type mock struct {
		database *databaseMocks.MockDatabase
	}
	type expected struct {
		responseCode int
	}
	type testData struct {
		name         string
		warmupPath   string
		buildRequest func() *http.Request
		setupMocks   func(*mock)
		expected     expected
	}

	tt := []testData{
		{
			name:       "Success: second migrated route has an independent rate limit - 204",
			warmupPath: "/api/v2/first",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/second", nil)
				request.RemoteAddr = "192.0.2.1:1234"
				return request
			},
			setupMocks: func(m *mock) {
				m.database.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).
					Return(appcfg.Parameter{}, nil).
					Times(56)
			},
			expected: expected{responseCode: http.StatusNoContent},
		},
		{
			name:       "Success: first migrated route has an independent rate limit - 204",
			warmupPath: "/api/v2/second",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/first", nil)
				request.RemoteAddr = "192.0.2.1:1234"
				return request
			},
			setupMocks: func(m *mock) {
				m.database.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).
					Return(appcfg.Parameter{}, nil).
					Times(56)
			},
			expected: expected{responseCode: http.StatusNoContent},
		},
	}

	for _, testCase := range tt {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				mockController = gomock.NewController(t)
				mockDatabase   = databaseMocks.NewMockDatabase(mockController)
				routerInst     = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
			)

			m := &mock{database: mockDatabase}
			testCase.setupMocks(m)
			require.NoError(t, routerInst.WithRouteMiddleware(func() mux.MiddlewareFunc {
				return middleware.RateLimitMiddleware(mockDatabase, 55)
			}, func() error {
				routerInst.GET("/api/v2/first", func(response http.ResponseWriter, _ *http.Request) {
					response.WriteHeader(http.StatusNoContent)
				})
				routerInst.GET("/api/v2/second", func(response http.ResponseWriter, _ *http.Request) {
					response.WriteHeader(http.StatusNoContent)
				})
				return nil
			}))

			for requestNumber := 0; requestNumber < 55; requestNumber++ {
				request := httptest.NewRequest(http.MethodGet, testCase.warmupPath, nil)
				request.RemoteAddr = "192.0.2.1:1234"
				response := httptest.NewRecorder()

				routerInst.Handler().ServeHTTP(response, request)

				require.Equal(t, http.StatusNoContent, response.Code, "request %d should pass the first route's rate limit", requestNumber+1)
			}

			request := testCase.buildRequest()
			response := httptest.NewRecorder()
			routerInst.Handler().ServeHTTP(response, request)

			assert.Equal(t, testCase.expected.responseCode, response.Code)
		})
	}
}
