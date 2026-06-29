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

package modules_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/modules"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
)

// noopRateLimit is a pass-through middleware factory for use in tests where
// rate-limiting behaviour is not under test.
func noopRateLimit() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler { return next }
}

func TestRegister_PanicsOnNilRouter(t *testing.T) {
	assert.Panics(t, func() {
		modules.Register(modules.Deps{
			Router:              nil,
			Pool:                new(pgxpool.Pool),
			Graph:               &graph.DatabaseSwitch{},
			RateLimitMiddleware: noopRateLimit,
		})
	})
}

func TestRegister_PanicsOnNilPool(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	assert.Panics(t, func() {
		modules.Register(modules.Deps{
			Router:              &routerInst,
			Pool:                nil,
			Graph:               &graph.DatabaseSwitch{},
			RateLimitMiddleware: noopRateLimit,
		})
	})
}

func TestRegister_PanicsOnNilGraph(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	assert.Panics(t, func() {
		modules.Register(modules.Deps{
			Router:              &routerInst,
			Pool:                new(pgxpool.Pool),
			Graph:               nil,
			RateLimitMiddleware: noopRateLimit,
		})
	})
}

func TestRegister_PanicsOnNilRateLimitMiddleware(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	assert.Panics(t, func() {
		modules.Register(modules.Deps{
			Router:              &routerInst,
			Pool:                new(pgxpool.Pool),
			Graph:               &graph.DatabaseSwitch{},
			RateLimitMiddleware: nil,
		})
	})
}

// TestRegister_WiresFeatureRoutes verifies that the composition root correctly
// attaches the feature module routes to the shared router. Matching a
// representative route from each module proves that Register successfully
// delegated to the feature modules.
func TestRegister_WiresFeatureModuleRoutes(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		deps       = modules.Deps{
			Router:              &routerInst,
			Pool:                new(pgxpool.Pool),
			Graph:               &graph.DatabaseSwitch{},
			RateLimitMiddleware: noopRateLimit,
		}
	)

	modules.Register(deps)

	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{"analysis status", http.MethodGet, "/api/v2/analysis/status"},
		{"feature flags list", http.MethodGet, "/api/v2/features"},
		{"feature flag toggle", http.MethodPut, "/api/v2/features/1/toggle"},
		{"relationship request", http.MethodGet, "/api/v2/relationships/1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				muxRouter           = routerInst.MuxRouter()
				request = httptest.NewRequest(tc.method, tc.path, nil)
				match   mux.RouteMatch
			)

			assert.True(t, muxRouter.Match(request, &match), "%s %s route should be registered by Register", tc.method, tc.path)
		})
	}
}
