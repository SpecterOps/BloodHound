// Copyright 2023 Specter Ops, Inc.
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

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func permissionsCheckAllHandler(db *dbmocks.MockDatabase, internalHandler http.HandlerFunc, permissions ...model.Permission) http.Handler {
	return PermissionsCheckAll(auth.NewAuthorizer(db), permissions...)(internalHandler)
}

func permissionsCheckAtLeastOneHandler(db *dbmocks.MockDatabase, internalHandler http.HandlerFunc, permissions ...model.Permission) http.Handler {
	return PermissionsCheckAtLeastOne(auth.NewAuthorizer(db), permissions...)(internalHandler)
}

func auditEntryAndContext(bhCtx ctx.Context, action model.AuditLogAction, fields model.AuditData, status model.AuditLogEntryStatus) (context.Context, model.AuditEntry) {
	testCtx := context.Background()
	testCtx = ctx.Set(testCtx, &bhCtx)

	auditEntry := model.AuditEntry{
		CommitID: uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111"),
		Action:   action,
		Model:    fields,
		Status:   status,
	}

	return testCtx, auditEntry
}

func Test_parseAuthorizationHeader(t *testing.T) {
	var (
		expectedTime = time.Now()
		expectedID   = must.NewUUIDv4()
		request      = must.NewHTTPRequest(http.MethodGet, "http://example.com/", nil)
	)

	request.Header.Set(headers.Authorization.String(), "bhesignature "+expectedID.String())
	request.Header.Set(headers.RequestDate.String(), expectedTime.Format(time.RFC3339Nano))

	authScheme, schemeParameter, err := parseAuthorizationHeader(request)

	require.Equal(t, api.AuthorizationSchemeBHESignature, authScheme)
	require.Equal(t, expectedID.String(), schemeParameter)
	require.Nil(t, err)
}

func TestPermissionsCheckAll(t *testing.T) {
	var (
		handlerReturn200 = func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(http.StatusOK)
		}
		mockCtrl   = gomock.NewController(t)
		mockDB     = dbmocks.NewMockDatabase(mockCtrl)
		noPermsCtx = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					EmailAddress:  null.StringFrom("no@permissions.com"),
					PrincipalName: "noPermissions",
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
						},
					},
					Unique: model.Unique{
						ID: uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222"),
					},
				},
				Session: model.UserSession{},
			},
		}
		allPermsCtx = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					EmailAddress:  null.StringFrom("all@permissions.com"),
					PrincipalName: "allPermissions",
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: auth.Permissions().All(),
						},
					},
					Unique: model.Unique{
						ID: uuid.FromStringOrNil("33333333-3333-3333-3333-333333333333"),
					},
				},
				Session: model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(0) // No audit logs should be created on successful login or missing auth context
	test.Request(t).
		WithURL("http://example.com/test").
		WithHeader(headers.RequestID.String(), "requestID").
		WithContext(&ctx.Context{}).
		OnHandler(permissionsCheckAllHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusUnauthorized)

	test.Request(t).
		WithURL("http://example.com/test").
		WithHeader(headers.RequestID.String(), "requestID").
		WithContext(&allPermsCtx).
		OnHandler(permissionsCheckAllHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusOK)

	auditContext, noPermsEntry := auditEntryAndContext(
		noPermsCtx, model.AuditLogActionUnauthorizedAccessAttempt,
		model.AuditData{"endpoint": "POST /test"},
		model.AuditLogStatusFailure,
	)
	mockDB.EXPECT().AppendAuditLog(auditContext, noPermsEntry).Times(1)
	test.Request(t).
		WithURL("http://example.com/test").
		WithMethod(http.MethodPost).
		WithContext(&noPermsCtx).
		OnHandler(permissionsCheckAllHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusForbidden)
}

func TestPermissionsCheckAtLeastOne(t *testing.T) {
	var (
		handlerReturn200 = func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(http.StatusOK)
		}
		mockCtrl        = gomock.NewController(t)
		mockDB          = dbmocks.NewMockDatabase(mockCtrl)
		missingPermsCtx = ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: model.Permissions{auth.Permissions().AuthManageSelf, auth.Permissions().GraphDBRead},
						},
					},
				},
				Session: model.UserSession{},
			},
		}
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(0)
	test.Request(t).
		WithURL("http://example.com/test").
		WithContext(&ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: model.Permissions{auth.Permissions().AuthManageSelf},
						},
					},
				},
				Session: model.UserSession{},
			},
		}).
		OnHandler(permissionsCheckAtLeastOneHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusOK)

	test.Request(t).
		WithURL("http://example.com/test").
		WithContext(&ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: model.Permissions{auth.Permissions().AuthManageSelf, auth.Permissions().GraphDBRead},
						},
					},
				},
				Session: model.UserSession{},
			},
		}).
		OnHandler(permissionsCheckAtLeastOneHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusOK)

	test.Request(t).
		WithURL("http://example.com/test").
		WithContext(&ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: model.Permissions{auth.Permissions().AuthManageSelf, auth.Permissions().GraphDBRead},
						},
					},
				},
				Session: model.UserSession{},
			},
		}).
		OnHandler(permissionsCheckAtLeastOneHandler(mockDB, handlerReturn200, auth.Permissions().GraphDBRead)).
		Require().
		ResponseStatusCode(http.StatusOK)

	test.Request(t).
		WithURL("http://example.com/test").
		WithHeader(headers.RequestID.String(), "requestID").
		WithContext(&ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: auth.Permissions().All(),
						},
					},
				},
				Session: model.UserSession{},
			},
		}).
		OnHandler(permissionsCheckAtLeastOneHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusOK)

	auditContext, missingPermsEntry := auditEntryAndContext(
		missingPermsCtx, model.AuditLogActionUnauthorizedAccessAttempt,
		model.AuditData{"endpoint": "PUT /test"},
		model.AuditLogStatusFailure,
	)
	mockDB.EXPECT().AppendAuditLog(auditContext, missingPermsEntry).Times(1)
	test.Request(t).
		WithURL("http://example.com/test").
		WithMethod(http.MethodPut).
		WithContext(&missingPermsCtx).
		OnHandler(permissionsCheckAtLeastOneHandler(mockDB, handlerReturn200, auth.Permissions().GraphDBWrite)).
		Require().
		ResponseStatusCode(http.StatusForbidden)
}

func Test_AuthMiddleware(t *testing.T) {
	t.Run("test the basic functionality of the AuthMiddleware using bhesignature", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		mockCtrl := gomock.NewController(t)
		mockAuth := mocks.NewMockAuthenticator(mockCtrl)
		token := must.NewUUIDv4()

		mockAuth.EXPECT().ValidateRequestSignature(token, gomock.Any(), gomock.Any()).Return(auth.Context{}, http.StatusOK, nil).AnyTimes()

		if request, err := http.NewRequestWithContext(ctx, http.MethodOptions, "/foo", nil); err != nil {
			t.Error(err)
		} else {
			request.Header.Set("Authorization", api.AuthorizationSchemeBHESignature+" "+token.String())
			response := httptest.NewRecorder()
			wrapHandler := AuthMiddleware(mockAuth)(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))
			wrapHandler.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code)
		}
	})
}
