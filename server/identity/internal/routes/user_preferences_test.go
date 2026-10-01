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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/identity/internal/routes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type preferencesAuditLogger struct{}

func (preferencesAuditLogger) AppendAuditLog(context.Context, model.AuditEntry) error { return nil }

func TestPreferencesRoutes_Authorization(t *testing.T) {
	var (
		userID      = uuid.Must(uuid.NewV4())
		otherID     = uuid.Must(uuid.NewV4())
		permissions = auth.Permissions()
		payload     = []byte{0, 255, 128}
	)
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		for _, testCase := range []struct {
			name          string
			authenticated bool
			target        uuid.UUID
			permissions   model.Permissions
			status        int
		}{
			{"unauthenticated", false, userID, nil, 401},
			{"missing permission", true, userID, nil, 403},
			{"ManageUsers alone", true, userID, model.Permissions{permissions.AuthManageUsers}, 403},
			{"owner with ManageSelf alone", true, userID, model.Permissions{permissions.AuthManageSelf}, 200},
			{"another user with ManageSelf", true, otherID, model.Permissions{permissions.AuthManageSelf}, 403},
			{"administrator cannot access another user", true, otherID, model.Permissions{permissions.AuthManageSelf, permissions.AuthManageUsers}, 403},
		} {
			t.Run(method+"/"+testCase.name, func(t *testing.T) {
				var (
					service        = mocks.NewMockIdentity(t)
					routerInst     = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(preferencesAuditLogger{}), "")
					request        = httptest.NewRequest(method, "/api/v2/bloodhound-users/"+testCase.target.String()+"/preferences", bytes.NewReader(payload))
					response       = httptest.NewRecorder()
					requestContext = &bhctx.Context{}
					rateLimited    bool
					expectedStatus = testCase.status
				)
				if testCase.authenticated {
					requestContext.AuthCtx.Owner = model.User{
						Unique: model.Unique{ID: userID},
						Roles:  model.Roles{{Permissions: testCase.permissions}},
					}
				}
				request = request.WithContext(bhctx.Set(request.Context(), requestContext))
				request.Header.Set("Content-Type", "application/octet-stream")
				routes.Register(&routerInst, handlers.NewHandlersContainer(service), func() mux.MiddlewareFunc {
					return func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
							rateLimited = true
							next.ServeHTTP(response, request)
						})
					}
				})
				if expectedStatus == 200 {
					if method == http.MethodGet {
						service.EXPECT().GetUserPreferences(mock.Anything, userID).Return(payload, nil).Once()
					} else {
						expectedStatus = http.StatusNoContent
						service.EXPECT().UpsertUserPreferences(mock.Anything, userID, payload).Return(nil).Once()
					}
				}
				routerInst.Handler().ServeHTTP(response, request)
				require.Equal(t, expectedStatus, response.Code)
				require.True(t, rateLimited, "both preference routes must use the injected rate limiter")
				if expectedStatus == http.StatusOK {
					require.Equal(t, payload, response.Body.Bytes())
				}
			})
		}
	}
}

func TestPreferencesRoutes_RateLimitRejectsBeforeStorage(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			var (
				service    = mocks.NewMockIdentity(t)
				routerInst = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(preferencesAuditLogger{}), "")
				response   = httptest.NewRecorder()
				request    = httptest.NewRequest(method, "/api/v2/bloodhound-users/"+uuid.Must(uuid.NewV4()).String()+"/preferences", nil)
			)
			routes.Register(&routerInst, handlers.NewHandlersContainer(service), func() mux.MiddlewareFunc {
				return func(http.Handler) http.Handler {
					return http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
						response.WriteHeader(http.StatusTooManyRequests)
					})
				}
			})
			routerInst.Handler().ServeHTTP(response, request)
			require.Equal(t, http.StatusTooManyRequests, response.Code)
		})
	}
}
