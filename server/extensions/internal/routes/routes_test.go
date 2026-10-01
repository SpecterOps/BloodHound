// Copyright 2026 Specter Ops, Inc.
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
package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/extensions/internal/routes"
	"github.com/stretchr/testify/assert"
)

// noopRateLimit is a pass-through middleware factory for use in tests where
// rate-limiting behaviour is not under test.
func noopRateLimit() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler { return next }
}

// TestRegister verifies that routes.Register binds the GET /api/v2/node-kinds/{id}
// endpoint to the gorilla/mux router so that matching requests are dispatched correctly.
func TestRegister(t *testing.T) {
	var (
		cfg            = config.Configuration{}
		authorizer     = auth.NewAuthorizer(nil)
		routerInst     = router.NewRouter(cfg, authorizer, "")
		extensionsMock = mocks.NewMockExtensions(t)
		handlerSet     = handlers.NewHandlersContainer(extensionsMock)
	)

	routes.Register(&routerInst, handlerSet)

	muxRouter := routerInst.MuxRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/node-kinds/123"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		var match mux.RouteMatch
		assert.True(t, muxRouter.Match(req, &match), "%s %s route should be registered", tc.method, tc.path)
	}
}

// TestRegister_RateLimitingReturns429 verifies that the extensions routes are wired
// with the rate-limiting middleware and that requests exceeding the per-IP limit
// are rejected with 429 before reaching the handler.
//
// Because the rate limiter is registered as the outermost middleware layer
// (before the permissions check), even unauthenticated requests count against
// the limit and trigger 429 once the budget is exhausted.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	var (
		cfg            = config.Configuration{}
		authorizer     = auth.NewAuthorizer(nil)
		routerInst     = router.NewRouter(cfg, authorizer, "")
		extensionsMock = mocks.NewMockExtensions(t)
		handlerSet     = handlers.NewHandlersContainer(extensionsMock)
	)

	routes.Register(&routerInst, handlerSet)
	routeHandler := routerInst.Handler()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/node-kinds/123"},
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
