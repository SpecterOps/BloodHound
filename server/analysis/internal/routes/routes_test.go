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
	"github.com/specterops/bloodhound/server/analysis/internal/handlers"
	"github.com/specterops/bloodhound/server/analysis/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/analysis/internal/routes"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		mock       = mocks.NewMockAnalysis(t)
		handlerSet = handlers.NewHandlersContainer(mock)
	)

	routes.Register(&routerInst, handlerSet)

	muxRouter := routerInst.MuxRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/analysis/status"},
		{http.MethodPut, "/api/v2/analysis"},
		{http.MethodDelete, "/api/v2/analysis"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		var match mux.RouteMatch
		assert.True(t, muxRouter.Match(req, &match), "%s %s route should be registered", tc.method, tc.path)
	}
}

// TestRegister_RoutesRequireAuthentication dispatches real requests through the wired
// router to verify that the registered routes are guarded by authentication middleware.
// The handlers themselves trust the middleware to enforce this contract; if the route
// wireup ever loses RequirePermissions/RequireAuth, this test will fail.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		mock       = mocks.NewMockAnalysis(t)
		handlerSet = handlers.NewHandlersContainer(mock)
	)

	routes.Register(&routerInst, handlerSet)
	handler := routerInst.Handler()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/analysis/status"},
		{http.MethodPut, "/api/v2/analysis"},
		{http.MethodDelete, "/api/v2/analysis"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var (
				request  = httptest.NewRequest(tc.method, tc.path, nil)
				recorder = httptest.NewRecorder()
			)

			handler.ServeHTTP(recorder, request)

			assert.Equal(t, http.StatusUnauthorized, recorder.Code,
				"unauthenticated %s %s must be rejected by middleware before reaching the handler", tc.method, tc.path)
		})
	}
}
