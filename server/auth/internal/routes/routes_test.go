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
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/auth/internal/handlers"
	"github.com/specterops/bloodhound/server/auth/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/auth/internal/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		authMock   = mocks.NewMockAuth(t)
		handlerSet = handlers.NewHandlersContainer(authMock)
	)

	routes.Register(&routerInst, handlerSet)

	muxRouter := routerInst.MuxRouter()

	for _, tc := range []struct {
		method    string
		path      string
		wantMatch bool
	}{
		{http.MethodPost, "/api/v2/logout", true},
		{http.MethodGet, "/api/v2/logout", false},
		{http.MethodGet, "/api/v2/self", true},
		{http.MethodPost, "/api/v2/self", false},
		{http.MethodGet, "/api/v2/sso-providers", true},
		{http.MethodPost, "/api/v2/sso-providers", false},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		var match mux.RouteMatch
		assert.Equal(t, tc.wantMatch, muxRouter.Match(req, &match), "%s %s route match", tc.method, tc.path)
	}
}

// TestRegister_LogoutDoesNotRequireAuthentication verifies that unauthenticated logout requests
// reach the handler, matching the legacy route. Fails if RequireAuth/RequirePermissions is added.
func TestRegister_LogoutDoesNotRequireAuthentication(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		authMock   = mocks.NewMockAuth(t)
		handlerSet = handlers.NewHandlersContainer(authMock)
		host       = &url.URL{Scheme: "https", Host: "bloodhound.example.com"}
		request    = httptest.NewRequest(http.MethodPost, "/api/v2/logout", nil)
		recorder   = httptest.NewRecorder()
	)

	routes.Register(&routerInst, handlerSet)

	request = bhctx.SetRequestContext(request, &bhctx.Context{Host: host})
	authMock.EXPECT().Logout(mock.Anything, mock.Anything).Return(nil)

	routerInst.Handler().ServeHTTP(recorder, request)

	assert.NotEqual(t, http.StatusUnauthorized, recorder.Code,
		"unauthenticated POST /api/v2/logout must not be rejected by route-level middleware")
}

func TestListAuthProvidersQueryValidation(t *testing.T) {
	for _, query := range []string{"?sort_by=details", "?sort_by=-", "?name=gt:value", "?name=bad:value", "?details=eq:value"} {
		t.Run(query, func(t *testing.T) {
			var (
				routerInstance = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
				serviceMock    = mocks.NewMockAuth(t)
				request        = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/sso-providers"+query, nil), &bhctx.Context{})
				response       = httptest.NewRecorder()
			)

			routes.Register(&routerInstance, handlers.NewHandlersContainer(serviceMock))
			routerInstance.Handler().ServeHTTP(response, request)
			assert.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

// The listing is needed by the login screen before the user authenticates.
func TestListAuthProvidersPublicQuery(t *testing.T) {
	var (
		routerInstance = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
		serviceMock    = mocks.NewMockAuth(t)
		request        = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/sso-providers?name=eq:Example&sort_by=-name", nil), &bhctx.Context{Host: &url.URL{Scheme: "https", Host: "example.com"}})
		response       = httptest.NewRecorder()
		filters        = params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "Example", SetOperator: params.FilterAnd, IsStringData: true}}}
		sortItems      = params.SortItems{{Field: "name", Direction: params.Descending}}
	)

	serviceMock.EXPECT().ListAuthProviders(mock.Anything, filters, sortItems).Return(nil, nil)

	routes.Register(&routerInstance, handlers.NewHandlersContainer(serviceMock))
	routerInstance.Handler().ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
}
