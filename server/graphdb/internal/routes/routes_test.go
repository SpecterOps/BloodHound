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
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/graphdb/internal/routes"
	"github.com/stretchr/testify/assert"
)

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

	routes.Register(&routerInst, handlerSet)

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

	routes.Register(&routerInst, handlerSet)
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
