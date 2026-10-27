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
	"github.com/specterops/bloodhound/server/audit/internal/handlers"
	"github.com/specterops/bloodhound/server/audit/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/audit/internal/routes"
	"github.com/stretchr/testify/assert"
)

func TestRegister_RegistersAuditRoute(t *testing.T) {
	t.Parallel()
	var (
		routerInst = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
		handlerSet = handlers.NewHandlersContainer(mocks.NewMockAudit(t))
		match      mux.RouteMatch
	)

	routes.Register(&routerInst, handlerSet)

	assert.True(t, routerInst.MuxRouter().Match(httptest.NewRequest(http.MethodGet, "/api/v2/audit", nil), &match))
}

func TestRegister_AuditRouteRequiresAuthentication(t *testing.T) {
	t.Parallel()
	var (
		routerInst = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
		handlerSet = handlers.NewHandlersContainer(mocks.NewMockAudit(t))
		response   = httptest.NewRecorder()
	)

	routes.Register(&routerInst, handlerSet)
	routerInst.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/audit", nil))

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}
