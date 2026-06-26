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
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	databaseMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/graphdb/internal/routes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// noopRateLimit is a pass-through middleware factory for use in tests where
// rate-limiting behaviour is not under test.
func noopRateLimit() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler { return next }
}

// TestRegister verifies that routes.Register binds the GET /api/v2/relationships/{id}
// endpoint to the gorilla/mux router so that matching requests are dispatched correctly.
func TestRegister(t *testing.T) {
	var (
		cfg         = config.Configuration{}
		authorizer  = auth.NewAuthorizer(nil)
		routerInst  = router.NewRouter(cfg, authorizer, "")
		graphDBMock = mocks.NewMockGraphDB(t)
		handlerSet  = handlers.NewHandlersContainer(graphDBMock)
	)

	routes.Register(&routerInst, handlerSet, noopRateLimit)

	muxRouter := routerInst.MuxRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/relationships/123"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		var match mux.RouteMatch
		assert.True(t, muxRouter.Match(req, &match), "%s %s route should be registered", tc.method, tc.path)
	}
}

// TestRegister_RateLimitingReturns429 verifies that the graphdb routes are wired
// with the rate-limiting middleware and that requests exceeding the per-IP limit
// are rejected with 429 before reaching the handler.
//
// Because the rate limiter is registered as the outermost middleware layer
// (before the permissions check), even unauthenticated requests count against
// the limit and trigger 429 once the budget is exhausted.
func TestRegister_RateLimitingReturns429(t *testing.T) {
	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = databaseMocks.NewMockDatabase(mockCtrl)
		cfg         = config.Configuration{}
		authorizer  = auth.NewAuthorizer(nil)
		routerInst  = router.NewRouter(cfg, authorizer, "")
		graphDBMock = mocks.NewMockGraphDB(t)
		handlerSet  = handlers.NewHandlersContainer(graphDBMock)
	)

	// Stub the trusted-proxies DB call so the rate limiter can extract the client IP.
	// Returning an empty parameter causes GetTrustedProxiesParameters to return 0,
	// meaning the direct RemoteAddr is used as the rate-limit key.
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()

	// Strict factory: 1 request per second per IP.
	strictRateLimitFactory := func() mux.MiddlewareFunc {
		return middleware.RateLimitMiddleware(mockDB, 1)
	}

	routes.Register(&routerInst, handlerSet, strictRateLimitFactory)

	var (
		routeHandler = routerInst.Handler()
		// All requests come from the same synthetic IP so they share one bucket.
		firstRequest  = httptest.NewRequest(http.MethodGet, "/api/v2/relationships/123", nil)
		secondRequest = httptest.NewRequest(http.MethodGet, "/api/v2/relationships/123", nil)
	)

	// Both requests use the same RemoteAddr so they share the same rate-limit bucket.
	firstRequest.RemoteAddr = "192.0.2.1:1234"
	secondRequest.RemoteAddr = "192.0.2.1:1234"

	// First request: consumes the single allowed slot, then stopped by auth → 401.
	firstRecorder := httptest.NewRecorder()
	routeHandler.ServeHTTP(firstRecorder, firstRequest)
	assert.Equal(t, http.StatusUnauthorized, firstRecorder.Code,
		"first request should pass rate limit and be rejected by auth middleware")

	// Second request: rate limit bucket is exhausted → 429 before auth is reached.
	secondRecorder := httptest.NewRecorder()
	routeHandler.ServeHTTP(secondRecorder, secondRequest)
	assert.Equal(t, http.StatusTooManyRequests, secondRecorder.Code,
		"second request should be rejected by the rate limiter with 429")
}

// TestRegister_RoutesRequireAuthentication dispatches real unauthenticated
// requests through the wired router to verify that every registered graphdb
// route is guarded by authentication middleware. If the route wireup ever
// loses RequirePermissions(), this test will fail.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	var (
		cfg         = config.Configuration{}
		authorizer  = auth.NewAuthorizer(nil)
		routerInst  = router.NewRouter(cfg, authorizer, "")
		graphDBMock = mocks.NewMockGraphDB(t)
		handlerSet  = handlers.NewHandlersContainer(graphDBMock)
	)

	routes.Register(&routerInst, handlerSet, noopRateLimit)
	routeHandler := routerInst.Handler()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/relationships/123"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var (
				request  = httptest.NewRequest(tc.method, tc.path, nil)
				recorder = httptest.NewRecorder()
			)
			routeHandler.ServeHTTP(recorder, request)
			assert.Equal(t, http.StatusUnauthorized, recorder.Code,
				"unauthenticated %s %s must be rejected by middleware before reaching the handler",
				tc.method, tc.path)
		})
	}
}
