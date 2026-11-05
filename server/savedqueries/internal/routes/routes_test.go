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
	"github.com/specterops/bloodhound/server/savedqueries/internal/handlers"
	"github.com/specterops/bloodhound/server/savedqueries/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/savedqueries/internal/routes"
	"github.com/specterops/bloodhound/server/users"
	"github.com/stretchr/testify/assert"
)

func TestRegister_GetSavedQuery(t *testing.T) {
	type testData struct {
		name   string
		assert func(*testing.T, *router.Router)
	}

	tt := []testData{
		{
			name: "Success: GET route is registered",
			assert: func(t *testing.T, routerInst *router.Router) {
				var match mux.RouteMatch
				request := httptest.NewRequest(http.MethodGet, "/api/v2/saved-queries/42", nil)

				assert.True(t, routerInst.MuxRouter().Match(request, &match))
			},
		},
		{
			name: "Error: unauthenticated request is rejected - 401",
			assert: func(t *testing.T, routerInst *router.Router) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(http.MethodGet, "/api/v2/saved-queries/42", nil)
				)

				routerInst.Handler().ServeHTTP(recorder, request)
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				configuration = config.Configuration{}
				authorizer    = auth.NewAuthorizer(nil)
				routerInst    = router.NewRouter(configuration, authorizer, "")
				serviceMock   = mocks.NewMockSavedQueries(t)
				handlerSet    = handlers.NewHandlersContainer(serviceMock, users.NewResolver())
			)

			routes.Register(&routerInst, handlerSet)
			testCase.assert(t, &routerInst)
		})
	}
}
