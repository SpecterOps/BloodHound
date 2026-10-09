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
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	databaseMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRouter_EnsureMatchedRouteRateLimitInstallsOnceBeforeAuthentication(t *testing.T) {
	var (
		mockController        = gomock.NewController(t)
		mockDatabase          = databaseMocks.NewMockDatabase(mockController)
		routerInst            = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
		rateLimitFactoryCalls int
		authenticationCalls   int
	)

	mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
	routerInst.UsePanicRecovery(middleware.PanicHandler)
	routerInst.UseAuthenticationMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			authenticationCalls++
			next.ServeHTTP(response, request)
		})
	})
	// Model BHE middleware registered after authentication.
	routerInst.UsePostrouting(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			next.ServeHTTP(response, request)
		})
	})
	rateLimitFactory := func() mux.MiddlewareFunc {
		rateLimitFactoryCalls++
		return middleware.RateLimitMiddleware(mockDatabase, 1)
	}
	require.NoError(t, routerInst.EnsureMatchedRouteRateLimit(rateLimitFactory))
	require.NoError(t, routerInst.EnsureMatchedRouteRateLimit(rateLimitFactory))
	routerInst.GET("/api/v2/test", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})

	for requestNumber, expectedStatus := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/api/v2/test", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()

		routerInst.Handler().ServeHTTP(response, request)

		require.Equal(t, expectedStatus, response.Code, "request %d should have the expected rate-limit result", requestNumber+1)
	}
	require.Equal(t, 1, rateLimitFactoryCalls, "the limiter factory should run once even when registries ensure the limiter")
	require.Equal(t, 1, authenticationCalls, "the rejected request should not reach authentication")
}

func TestRouter_PanicRecoveryWrapsMatchedRouteRateLimit(t *testing.T) {
	routerInst := router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
	routerInst.UsePanicRecovery(middleware.PanicHandler)
	require.NoError(t, routerInst.EnsureMatchedRouteRateLimit(func() mux.MiddlewareFunc {
		return func(http.Handler) http.Handler {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				panic("rate limiter panic")
			})
		}
	}))
	routerInst.GET("/api/v2/test", func(http.ResponseWriter, *http.Request) {})

	response := httptest.NewRecorder()
	routerInst.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/test", nil))

	require.Equal(t, http.StatusOK, response.Code, "the panic handler should recover a panic from the limiter")
}
