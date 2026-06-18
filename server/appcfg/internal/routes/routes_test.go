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
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers"
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/appcfg/internal/routes"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		mock       = mocks.NewMockService(t)
		handlerSet = handlers.NewHandlers(mock)
	)

	routes.Register(&routerInst, handlerSet)

	muxRouter := routerInst.MuxRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v2/datapipe/status", nil)
	var match mux.RouteMatch
	assert.True(t, muxRouter.Match(req, &match), "GET /api/v2/datapipe/status route should be registered")
}

// TestRegister_RoutesRequireAuthentication dispatches real requests through the wired
// router to verify that the registered route is guarded by authentication middleware.
// The handlers themselves trust the middleware to enforce this contract; if the route
// wireup ever loses RequireAuth, this test will fail.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		mock       = mocks.NewMockService(t)
		handlerSet = handlers.NewHandlers(mock)
	)

	routes.Register(&routerInst, handlerSet)
	handler := routerInst.Handler()

	var (
		request  = httptest.NewRequest(http.MethodGet, "/api/v2/datapipe/status", nil)
		recorder = httptest.NewRecorder()
	)

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code,
		"unauthenticated GET /api/v2/datapipe/status must be rejected by middleware before reaching the handler")
}
