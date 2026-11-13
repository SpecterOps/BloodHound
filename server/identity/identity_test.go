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

package identity_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_RegistersRoutes(t *testing.T) {
	type expected struct {
		routeRegistered bool
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		expected     expected
	}

	tests := []testData{
		{
			name: "Success: GET /api/v2/roles is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
			},
			expected: expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/roles/{role_id} is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles/1", nil)
			},
			expected: expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/permissions is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
			},
			expected: expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/permissions/{permission_id} is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/1", nil)
			},
			expected: expected{routeRegistered: true},
		},
		{
			name: "Success: GET /api/v2/tokens is registered",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/tokens", nil)
			},
			expected: expected{routeRegistered: true},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				cfg        = config.Configuration{}
				authorizer = auth.NewAuthorizer(nil)
				routerInst = router.NewRouter(cfg, authorizer, "")
				pool       = new(pgxpool.Pool)
				match      mux.RouteMatch
			)

			require.NotPanics(t, func() {
				identity.Register(&routerInst, pool)
			})

			assert.Equal(t, testCase.expected.routeRegistered, routerInst.MuxRouter().Match(testCase.buildRequest(), &match))
		})
	}
}

type authTokenReaderFunc func(context.Context, params.Filters, params.SortItems) ([]identity.AuthToken, error)

func (s authTokenReaderFunc) ListAuthTokens(ctx context.Context, filters params.Filters, sortItems params.SortItems) ([]identity.AuthToken, error) {
	return s(ctx, filters, sortItems)
}

func TestRegister_WithAuthTokenReader(t *testing.T) {
	var (
		ownerID     = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
		otherID     = uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222")
		tokenID     = uuid.FromStringOrNil("33333333-3333-3333-3333-333333333333")
		permissions = auth.Permissions()
	)

	for _, testCase := range []struct {
		name        string
		admin       bool
		query       string
		readerError error
		wantStatus  int
		wantCalled  bool
		wantFilters params.Filters
		wantSort    params.SortItems
	}{
		{
			name:  "unfiltered admin still uses application reader",
			admin: true, wantStatus: http.StatusOK, wantCalled: true,
			wantFilters: params.Filters{}, wantSort: params.SortItems{},
		},
		{
			name:  "non-admin is scoped before application reader",
			query: "?name=eq:visible-token&sort_by=-name", wantStatus: http.StatusOK, wantCalled: true,
			wantFilters: params.Filters{
				"name":    {{Field: "name", Operator: params.Equals, Value: "visible-token", SetOperator: params.FilterAnd, IsStringData: true}},
				"user_id": {{Field: "user_id", Operator: params.Equals, Value: ownerID.String(), SetOperator: params.FilterAnd}},
			},
			wantSort: params.SortItems{{Field: "name", Direction: params.Descending}},
		},
		{
			name:  "other user's tokens rejected before application reader",
			query: "?user_id=eq:" + otherID.String(), wantStatus: http.StatusForbidden,
		},
		{
			name:  "invalid filters rejected before application reader",
			admin: true, query: "?key=eq:secret", wantStatus: http.StatusBadRequest,
		},
		{
			name:  "reader failure does not fall back to unrestricted store",
			admin: true, readerError: errors.New("application query failed"), wantStatus: http.StatusInternalServerError,
			wantCalled: true, wantFilters: params.Filters{}, wantSort: params.SortItems{},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				routerInst      = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
				userPermissions = model.Permissions{permissions.AuthCreateToken, permissions.AuthManageSelf}
				request         = httptest.NewRequest(http.MethodGet, "/api/v2/tokens"+testCase.query, nil)
				recorder        = httptest.NewRecorder()
				called          bool
				reader          = authTokenReaderFunc(func(ctx context.Context, filters params.Filters, sortItems params.SortItems) ([]identity.AuthToken, error) {
					called = true
					assert.Equal(t, testCase.wantFilters, filters)
					assert.Equal(t, testCase.wantSort, sortItems)
					assert.Equal(t, ownerID, bhctx.Get(ctx).AuthCtx.Owner.(model.User).ID)
					return []identity.AuthToken{{ID: tokenID}}, testCase.readerError
				})
			)
			if testCase.admin {
				userPermissions = append(userPermissions, permissions.AuthManageUsers)
			}
			request = bhctx.SetRequestContext(request, &bhctx.Context{AuthCtx: auth.Context{Owner: model.User{
				Unique: model.Unique{ID: ownerID}, Roles: model.Roles{{Permissions: userPermissions}},
			}}})
			identity.Register(&routerInst, new(pgxpool.Pool), identity.WithAuthTokenReader(reader))
			routerInst.Handler().ServeHTTP(recorder, request)
			assert.Equal(t, testCase.wantStatus, recorder.Code, recorder.Body.String())
			assert.Equal(t, testCase.wantCalled, called)
			if testCase.wantStatus == http.StatusOK {
				var envelope struct {
					Data struct {
						Tokens []struct {
							ID uuid.UUID `json:"id"`
						} `json:"tokens"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
				require.Len(t, envelope.Data.Tokens, 1)
				assert.Equal(t, tokenID, envelope.Data.Tokens[0].ID)
			}
		})
	}
}

func TestRegister_AuthTokenScopingUsesConfiguredAuthorizer(t *testing.T) {
	var (
		ownerID     = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
		permissions = auth.Permissions()
	)

	for _, testCase := range []struct {
		name            string
		storedAdmin     bool
		authorizedAdmin bool
	}{
		{name: "custom authorizer restricts stored admin", storedAdmin: true},
		{name: "custom authorizer grants management access", authorizedAdmin: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				storedPermissions    = model.Permissions{permissions.AuthCreateToken, permissions.AuthManageSelf}
				effectivePermissions = model.Permissions{permissions.AuthCreateToken, permissions.AuthManageSelf}
				called               bool
			)
			if testCase.storedAdmin {
				storedPermissions = append(storedPermissions, permissions.AuthManageUsers)
			}
			if testCase.authorizedAdmin {
				effectivePermissions = append(effectivePermissions, permissions.AuthManageUsers)
			}

			var (
				authorizer = auth.NewCustomAuthorizer(nil, func(auth.Context) (model.Permissions, bool) {
					return effectivePermissions, true
				})
				routerInst = router.NewRouter(config.Configuration{}, authorizer, "")
				request    = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/tokens", nil), &bhctx.Context{AuthCtx: auth.Context{Owner: model.User{
					Unique: model.Unique{ID: ownerID}, Roles: model.Roles{{Permissions: storedPermissions}},
				}}})
				recorder = httptest.NewRecorder()
				reader   = authTokenReaderFunc(func(ctx context.Context, filters params.Filters, sortItems params.SortItems) ([]identity.AuthToken, error) {
					called = true
					if testCase.authorizedAdmin {
						assert.Empty(t, filters)
					} else {
						assert.Equal(t, params.Filters{
							"user_id": {{Field: "user_id", Operator: params.Equals, Value: ownerID.String(), SetOperator: params.FilterAnd}},
						}, filters)
					}
					return []identity.AuthToken{}, nil
				})
			)

			identity.Register(&routerInst, new(pgxpool.Pool), identity.WithAuthTokenReader(reader))
			routerInst.Handler().ServeHTTP(recorder, request)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.True(t, called)
		})
	}
}
