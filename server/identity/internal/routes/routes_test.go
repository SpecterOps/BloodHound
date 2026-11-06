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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	databaseMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/identity/internal/routes"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegister_ListAuthTokens_QueryParameters(t *testing.T) {
	var (
		userID = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
		user   = model.User{
			Unique: model.Unique{ID: userID},
			Roles: model.Roles{{Permissions: model.Permissions{
				auth.Permissions().AuthCreateToken,
				auth.Permissions().AuthManageSelf,
				auth.Permissions().AuthManageUsers,
			}}},
		}
		tests = []struct {
			name         string
			query        string
			filters      params.Filters
			sortItems    params.SortItems
			responseCode int
		}{
			{
				name:         "no query parameters returns an unpaginated token list",
				responseCode: http.StatusOK,
			},
			{
				name:         "pagination parameters are ignored",
				query:        "?limit=1&skip=2",
				responseCode: http.StatusOK,
			},
			{
				name:         "empty name is applied as a filter",
				query:        "?name=eq:",
				filters:      params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "", SetOperator: params.FilterAnd, IsStringData: true}}},
				responseCode: http.StatusOK,
			},
			{
				name:         "punctuation in a name is preserved",
				query:        "?name=eq:foo%21bar%2Bbaz",
				filters:      params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "foo!bar+baz", SetOperator: params.FilterAnd, IsStringData: true}}},
				responseCode: http.StatusOK,
			},
			{
				name:  "repeated timestamp filters are combined with AND",
				query: "?expires_at=gte:2026-01-01T00:00:00Z&expires_at=lt:2027-01-01T00:00:00Z",
				filters: params.Filters{"expires_at": {
					{Field: "expires_at", Operator: params.GreaterThanOrEquals, Value: "2026-01-01T00:00:00Z", SetOperator: params.FilterAnd},
					{Field: "expires_at", Operator: params.LessThan, Value: "2027-01-01T00:00:00Z", SetOperator: params.FilterAnd},
				}},
				responseCode: http.StatusOK,
			},
			{
				name:         "deleted_at sort is rejected",
				query:        "?sort_by=deleted_at",
				responseCode: http.StatusBadRequest,
			},
			{
				name:         "empty sort is rejected",
				query:        "?sort_by=",
				responseCode: http.StatusBadRequest},
			{
				name:         "key filter is rejected",
				query:        "?key=eq:secret",
				responseCode: http.StatusBadRequest},
			{
				name:         "deleted_at filter is rejected",
				query:        "?deleted_at=eq:null",
				responseCode: http.StatusBadRequest},
			{
				name:         "unsupported name operator is rejected",
				query:        "?name=gt:token",
				responseCode: http.StatusBadRequest},
			{
				name:         "unknown filter operator is rejected",
				query:        "?name=invalid:token",
				responseCode: http.StatusBadRequest},
		}
	)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				identityMock = mocks.NewMockIdentity(t)
				routerInst   = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
				request      = bhctx.SetRequestContext(
					httptest.NewRequest(http.MethodGet, "/api/v2/tokens"+testCase.query, nil),
					&bhctx.Context{AuthCtx: auth.Context{Owner: user}},
				)
				recorder  = httptest.NewRecorder()
				filters   = testCase.filters
				sortItems = testCase.sortItems
			)

			if testCase.responseCode == http.StatusOK {
				if filters == nil {
					filters = params.Filters{}
				}
				if sortItems == nil {
					sortItems = params.SortItems{}
				}
				identityMock.EXPECT().ListAuthTokens(
					testifyMock.Anything,
					services.AuthTokenRequester{UserID: userID, IsUser: true, CanManageUsers: true},
					filters, sortItems,
				).Return([]services.AuthToken{}, nil).Once()
			}

			routes.Register(&routerInst, handlers.NewHandlersContainer(identityMock))
			routerInst.Handler().ServeHTTP(recorder, request)

			require.Equal(t, testCase.responseCode, recorder.Code, recorder.Body.String())
			if testCase.responseCode == http.StatusOK {
				assert.JSONEq(t, `{"data":{"tokens":[]}}`, recorder.Body.String())
			} else {
				var envelope struct {
					HTTPStatus int `json:"http_status"`
					Errors     []struct {
						Message string `json:"message"`
					} `json:"errors"`
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
				assert.Equal(t, testCase.responseCode, envelope.HTTPStatus)
				require.Len(t, envelope.Errors, 1)
				assert.NotEmpty(t, envelope.Errors[0].Message)
			}
		})
	}
}

func TestRegister_RegistersRoutes(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		routeRegistered bool
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	tests := []testData{
		{
			name: "Success: GET /api/v2/roles is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/roles/{role_id} is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/permissions is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/permissions/{permission_id} is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/bloodhound-users is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/bloodhound-users", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/tokens is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/tokens", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/bloodhound-users-minimal is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/bloodhound-users-minimal", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{routeRegistered: true},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				cfg        = config.Configuration{}
				authorizer = auth.NewAuthorizer(nil)
				routerInst = router.NewRouter(cfg, authorizer, "")
				mock       = mock{identity: mocks.NewMockIdentity(t)}
				handlerSet = handlers.NewHandlersContainer(mock.identity)
				match      mux.RouteMatch
			)

			routes.Register(&routerInst, handlerSet)
			testCase.setupMocks(mock)

			assert.Equal(t, testCase.expected.routeRegistered, routerInst.MuxRouter().Match(testCase.buildRequest(), &match))
		})
	}
}

// TestRegister_RoutesRequireAuthentication dispatches real unauthenticated requests
// through the wired router to verify that the registered routes are guarded by
// authentication middleware. If the route wireup ever loses RequirePermissions,
// this test will fail.
func TestRegister_RoutesRequireAuthentication(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode int
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	tests := []testData{
		{
			name: "Error: unauthenticated GET /api/v2/roles - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/roles/{role_id} - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/permissions - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/permissions/{permission_id} - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/1", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/bloodhound-users - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/bloodhound-users", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/tokens - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/tokens", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
		{
			name: "Error: unauthenticated GET /api/v2/bloodhound-users-minimal - 401",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/bloodhound-users-minimal", nil)
			},
			setupMocks: func(mock) {},
			expected:   expected{responseCode: http.StatusUnauthorized},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				cfg        = config.Configuration{}
				authorizer = auth.NewAuthorizer(nil)
				routerInst = router.NewRouter(cfg, authorizer, "")
				mock       = mock{identity: mocks.NewMockIdentity(t)}
				handlerSet = handlers.NewHandlersContainer(mock.identity)
				recorder   = httptest.NewRecorder()
			)

			routes.Register(&routerInst, handlerSet)
			testCase.setupMocks(mock)
			routerInst.Handler().ServeHTTP(recorder, testCase.buildRequest())

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
		})
	}
}

func TestRegister_RateLimitingReturns429(t *testing.T) {
	type mock struct {
		database *databaseMocks.MockDatabase
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCodes []int
	}

	type testData struct {
		name          string
		buildRequests func() []*http.Request
		setupMocks    func(mock mock)
		expected      expected
	}

	newRequest := func() *http.Request {
		request := httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		return request
	}

	tests := []testData{
		{
			name: "Success: first request passes rate limit - 401",
			buildRequests: func() []*http.Request {
				return []*http.Request{newRequest()}
			},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
			},
			expected: expected{responseCodes: []int{http.StatusUnauthorized}},
		},
		{
			name: "Error: rate limit exceeded - 429",
			buildRequests: func() []*http.Request {
				return []*http.Request{newRequest(), newRequest()}
			},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
			},
			expected: expected{responseCodes: []int{http.StatusUnauthorized, http.StatusTooManyRequests}},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				mockController = gomock.NewController(t)
				mock           = mock{
					database: databaseMocks.NewMockDatabase(mockController),
					identity: mocks.NewMockIdentity(t),
				}
				cfg        = config.Configuration{}
				authorizer = auth.NewAuthorizer(nil)
				routerInst = router.NewRouter(cfg, authorizer, "")
				handlerSet = handlers.NewHandlersContainer(mock.identity)
			)

			testCase.setupMocks(mock)
			assert.NoError(t, routerInst.WithRouteMiddleware(func() mux.MiddlewareFunc {
				return middleware.RateLimitMiddleware(mock.database, 1)
			}, func() error {
				routes.Register(&routerInst, handlerSet)
				return nil
			}))

			for index, request := range testCase.buildRequests() {
				recorder := httptest.NewRecorder()
				routerInst.Handler().ServeHTTP(recorder, request)
				assert.Equal(t, testCase.expected.responseCodes[index], recorder.Code)
			}
		})
	}
}
