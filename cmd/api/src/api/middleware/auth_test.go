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
	"net/http"
	"testing"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func permissionsCheckAllHandler(db *dbmocks.MockDatabase, internalHandler http.HandlerFunc, permissions ...model.Permission) http.Handler {
	return PermissionsCheckAll(auth.NewAuthorizer(db), permissions...)(internalHandler)
}

func permissionsCheckAtLeastOneHandler(db *dbmocks.MockDatabase, internalHandler http.HandlerFunc, permissions ...model.Permission) http.Handler {
	return PermissionsCheckAtLeastOne(auth.NewAuthorizer(db), permissions...)(internalHandler)
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
		mockCtrl = gomock.NewController(t)
		mockDB   = dbmocks.NewMockDatabase(mockCtrl)
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	test.Request(t).
		WithURL("http//example.com").
		WithHeader(headers.RequestID.String(), "requestID").
		WithContext(&ctx.Context{}).
		OnHandler(permissionsCheckAllHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusUnauthorized)

	test.Request(t).
		WithURL("http//example.com").
		WithContext(&ctx.Context{
			AuthCtx: auth.Context{
				PermissionOverrides: auth.PermissionOverrides{},
				Owner: model.User{
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
						},
					},
				},
				Session: model.UserSession{},
			},
		}).
		OnHandler(permissionsCheckAllHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusForbidden)

	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(0) // No audit logs should be created on successful login
	test.Request(t).
		WithURL("http//example.com").
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
		OnHandler(permissionsCheckAllHandler(mockDB, handlerReturn200, auth.Permissions().AuthManageSelf)).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestPermissionsCheckAtLeastOne(t *testing.T) {
	var (
		handlerReturn200 = func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(http.StatusOK)
		}
		mockCtrl = gomock.NewController(t)
		mockDB   = dbmocks.NewMockDatabase(mockCtrl)
	)
	defer mockCtrl.Finish()

	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(0)
	test.Request(t).
		WithURL("http//example.com").
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
		WithURL("http//example.com").
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
		WithURL("http//example.com").
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
		WithURL("http//example.com").
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

	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	test.Request(t).
		WithURL("http//example.com").
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
		OnHandler(permissionsCheckAtLeastOneHandler(mockDB, handlerReturn200, auth.Permissions().GraphDBWrite)).
		Require().
		ResponseStatusCode(http.StatusForbidden)
}
