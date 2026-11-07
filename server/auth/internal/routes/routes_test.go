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
