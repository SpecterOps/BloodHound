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
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers"
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/featureflags/internal/routes"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	type args struct {
		method string
		path   string
	}
	type want struct {
		routeRegistered bool
	}

	var (
		cfg             = config.Configuration{}
		authorizer      = auth.NewAuthorizer(nil)
		routerInst      = router.NewRouter(cfg, authorizer, "")
		featureFlagMock = mocks.NewMockFeatureFlag(t)
		handlerSet      = handlers.NewHandlersContainer(featureFlagMock)
		tests           = []struct {
			name string
			args args
			want want
		}{
			{
				name: "Success: registers the GET features route",
				args: args{method: http.MethodGet, path: "/api/v2/features"},
				want: want{routeRegistered: true},
			},
			{
				name: "Success: registers the PUT toggle route",
				args: args{method: http.MethodPut, path: "/api/v2/features/1/toggle"},
				want: want{routeRegistered: true},
			},
		}
	)

	routes.Register(&routerInst, handlerSet)

	muxRouter := routerInst.MuxRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var match mux.RouteMatch
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)

			assert.Equal(t, tt.want.routeRegistered, muxRouter.Match(request, &match), "%s %s route should be registered", tt.args.method, tt.args.path)
		})
	}
}

// TestRegister_RoutesRequireAuthentication dispatches real requests through the wired
// router to verify that the registered routes are guarded by authentication middleware.
// The handlers themselves trust the middleware to enforce this contract; if the route
// wireup ever loses RequirePermissions/RequireAuth, this test will fail.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		args struct {
			method string
			path   string
		}
		want struct {
			statusCode int
		}
	}

	var (
		cfg             = config.Configuration{}
		authorizer      = auth.NewAuthorizer(nil)
		routerInst      = router.NewRouter(cfg, authorizer, "")
		featureFlagMock = mocks.NewMockFeatureFlag(t)
		handlerSet      = handlers.NewHandlersContainer(featureFlagMock)
		tests           = []testCase{
			{
				name: "Error: unauthenticated GET features route - 401",
				args: struct {
					method string
					path   string
				}{http.MethodGet, "/api/v2/features"},
				want: struct{ statusCode int }{http.StatusUnauthorized},
			},
			{
				name: "Error: unauthenticated PUT toggle route - 401",
				args: struct {
					method string
					path   string
				}{http.MethodPut, "/api/v2/features/1/toggle"},
				want: struct{ statusCode int }{http.StatusUnauthorized},
			},
		}
	)

	routes.Register(&routerInst, handlerSet)
	handler := routerInst.Handler()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				request  = httptest.NewRequest(testCase.args.method, testCase.args.path, nil)
				recorder = httptest.NewRecorder()
			)

			handler.ServeHTTP(recorder, request)

			assert.Equal(t, testCase.want.statusCode, recorder.Code,
				"unauthenticated %s %s must be rejected by middleware before reaching the handler", testCase.args.method, testCase.args.path)
		})
	}
}
