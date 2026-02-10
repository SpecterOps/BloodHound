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

package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	authz "github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	mocks_graph "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
	"github.com/specterops/bloodhound/packages/go/bhlog"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	samlProviderPathFmt     = "/api/v2/saml/providers/%d"
	updateUserSecretPathFmt = "/api/v2/auth/users/%s/secret"
)

func TestManagementResource_PutUserAuthSecret(t *testing.T) {
	var (
		currentPassword      = "currentPassword"
		goodUser             = model.User{AuthSecret: defaultDigestAuthSecret(t, currentPassword), Unique: model.Unique{ID: must.NewUUIDv4()}}
		otherUser            = model.User{Unique: model.Unique{ID: must.NewUUIDv4()}}
		badUser              = model.User{SSOProviderID: null.Int32From(1), Unique: model.Unique{ID: must.NewUUIDv4()}}
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
	)
	defer mockCtrl.Finish()
	bhCtx := ctx.Get(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{}))
	bhCtx.AuthCtx.Owner = goodUser
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	// Happy paths
	mockDB.EXPECT().GetUser(gomock.Any(), goodUser.ID).Return(goodUser, nil).Times(2)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil).Times(2)

	// Change own user secret requires current password
	mockDB.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	test.Request(t).
		WithContext(bhCtx).
		WithMethod(http.MethodPut).
		WithHeader(headers.RequestID.String(), "requestID").
		WithURL(updateUserSecretPathFmt, goodUser.ID.String()).
		WithURLPathVars(map[string]string{"user_id": goodUser.ID.String()}).
		WithBody(v2.SetUserSecretRequest{
			CurrentSecret:      currentPassword,
			Secret:             "tesT12345!@#$",
			NeedsPasswordReset: false,
		}).
		OnHandlerFunc(resources.PutUserAuthSecret).
		Require().
		ResponseStatusCode(http.StatusOK)

	// Change another users secret does not require current password
	mockDB.EXPECT().GetUser(gomock.Any(), otherUser.ID).Return(otherUser, nil)
	mockDB.EXPECT().CreateAuthSecret(gomock.Any(), gomock.Any()).Return(model.AuthSecret{}, nil).Times(1)
	test.Request(t).
		WithContext(bhCtx).
		WithMethod(http.MethodPut).
		WithHeader(headers.RequestID.String(), "requestID").
		WithURL(updateUserSecretPathFmt, goodUser.ID.String()).
		WithURLPathVars(map[string]string{"user_id": otherUser.ID.String()}).
		WithBody(v2.SetUserSecretRequest{
			Secret:             "tesT12345!@#$",
			NeedsPasswordReset: false,
		}).
		OnHandlerFunc(resources.PutUserAuthSecret).
		Require().
		ResponseStatusCode(http.StatusOK)

	// Negative path where a user already has a SAML provider set
	mockDB.EXPECT().GetUser(gomock.Any(), badUser.ID).Return(badUser, nil)
	test.Request(t).
		WithContext(bhCtx).
		WithMethod(http.MethodPut).
		WithHeader(headers.RequestID.String(), "requestID").
		WithURL(updateUserSecretPathFmt, badUser.ID.String()).
		WithURLPathVars(map[string]string{"user_id": badUser.ID.String()}).
		WithBody(v2.SetUserSecretRequest{
			Secret:             "tesT12345!@#$",
			NeedsPasswordReset: false,
		}).
		OnHandlerFunc(resources.PutUserAuthSecret).
		Require().
		ResponseJSONBody(
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: "Invalid operation, user is SSO"}},
			},
		)

	// Negative path where a user uses invalid current password
	test.Request(t).
		WithContext(bhCtx).
		WithMethod(http.MethodPut).
		WithHeader(headers.RequestID.String(), "requestID").
		WithURL(updateUserSecretPathFmt, goodUser.ID.String()).
		WithURLPathVars(map[string]string{"user_id": goodUser.ID.String()}).
		WithBody(v2.SetUserSecretRequest{
			CurrentSecret:      "wrongPassword",
			Secret:             "tesT12345!@#$",
			NeedsPasswordReset: false,
		}).
		OnHandlerFunc(resources.PutUserAuthSecret).
		Require().
		ResponseJSONBody(
			api.ErrorWrapper{
				HTTPStatus: http.StatusUnauthorized,
				Errors:     []api.ErrorDetails{{Message: "Invalid current password"}},
			},
		)
}

func TestManagementResource_EnableUserSAML(t *testing.T) {
	var (
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)

		adminUser  = model.User{Unique: model.Unique{ID: must.NewUUIDv4()}}
		goodRoles  = []int32{0}
		badRoles   = []int32{1}
		goodUserID = must.NewUUIDv4()
		badUserID  = must.NewUUIDv4()

		ssoProviderID     int32 = 123
		samlProviderIDStr       = "1234"

		ssoProvider = model.SSOProvider{
			Serial: model.Serial{ID: ssoProviderID},
			SAMLProvider: &model.SAMLProvider{
				Serial:        model.Serial{ID: 1234},
				SSOProviderID: null.Int32From(ssoProviderID),
			},
			Config: model.SSOProviderConfig{
				AutoProvision: model.SSOProviderAutoProvisionConfig{Enabled: true, RoleProvision: true}},
		}
	)

	bhCtx := ctx.Get(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{}))
	bhCtx.AuthCtx.Owner = adminUser

	defer mockCtrl.Finish()

	t.Run("Successfully update user with deprecated saml provider", func(t *testing.T) {
		mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Eq(goodRoles)).Return(model.Roles{}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), goodUserID).Return(model.User{}, nil)
		mockDB.EXPECT().GetSAMLProvider(gomock.Any(), ssoProvider.SAMLProvider.ID).Return(*ssoProvider.SAMLProvider, nil)
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), ssoProvider.ID).Return(ssoProvider, nil)
		mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)

		test.Request(t).
			WithContext(bhCtx).
			WithURLPathVars(map[string]string{"user_id": goodUserID.String()}).
			WithBody(v2.UpdateUserRequest{
				Principal:      "tester",
				Roles:          goodRoles,
				SAMLProviderID: samlProviderIDStr,
			}).
			OnHandlerFunc(resources.UpdateUser).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("Fails if auth secret set", func(t *testing.T) {
		mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Eq(goodRoles)).Return(model.Roles{}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), badUserID).Return(model.User{AuthSecret: &model.AuthSecret{}}, nil)
		mockDB.EXPECT().GetSAMLProvider(gomock.Any(), ssoProvider.SAMLProvider.ID).Return(*ssoProvider.SAMLProvider, nil)
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), ssoProvider.ID).Return(ssoProvider, nil)
		mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)

		test.Request(t).
			WithContext(bhCtx).
			WithURLPathVars(map[string]string{"user_id": badUserID.String()}).
			WithBody(v2.UpdateUserRequest{
				Principal:      "tester",
				Roles:          goodRoles,
				SAMLProviderID: samlProviderIDStr,
			}).
			OnHandlerFunc(resources.UpdateUser).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("Fails if roles set", func(t *testing.T) {
		mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Eq(badRoles)).Return(model.Roles{model.Role{Serial: model.Serial{ID: 1}}}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), goodUserID).Return(model.User{SSOProviderID: null.Int32From(ssoProviderID), SSOProvider: &ssoProvider, Roles: model.Roles{model.Role{Serial: model.Serial{ID: 0}}}}, nil)
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), ssoProvider.ID).Return(ssoProvider, nil)

		test.Request(t).
			WithContext(bhCtx).
			WithURLPathVars(map[string]string{"user_id": goodUserID.String()}).
			WithBody(v2.UpdateUserRequest{
				Principal:     "tester",
				Roles:         badRoles,
				SSOProviderID: null.Int32From(ssoProviderID),
			}).
			OnHandlerFunc(resources.UpdateUser).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("Successful user update with sso provider-saml", func(t *testing.T) {
		mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Eq(goodRoles)).Return(model.Roles{}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), goodUserID).Return(model.User{}, nil)
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), ssoProvider.ID).Return(ssoProvider, nil)
		mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)

		test.Request(t).
			WithContext(bhCtx).
			WithURLPathVars(map[string]string{"user_id": goodUserID.String()}).
			WithBody(v2.UpdateUserRequest{
				Principal:     "tester",
				Roles:         goodRoles,
				SSOProviderID: null.Int32From(ssoProvider.ID),
			}).
			OnHandlerFunc(resources.UpdateUser).
			Require().
			ResponseStatusCode(http.StatusOK)
	})
}

func TestManagementResource_DeleteSAMLProvider(t *testing.T) {
	var (
		goodSAMLProvider = model.SAMLProvider{
			SSOProviderID: null.Int32From(1),
			Serial: model.Serial{
				ID: 1,
			},
		}

		samlProviderWithUsers = model.SAMLProvider{
			SSOProviderID: null.Int32From(2),
			Serial: model.Serial{
				ID: 2,
			},
		}

		samlEnabledUser = model.User{
			Unique: model.Unique{
				ID: must.NewUUIDv4(),
			},
		}

		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
	)

	defer mockCtrl.Finish()

	t.Run("successfully deletes saml provider", func(t *testing.T) {
		mockDB.EXPECT().GetSAMLProvider(gomock.Any(), goodSAMLProvider.ID).Return(goodSAMLProvider, nil)
		mockDB.EXPECT().DeleteSSOProvider(gomock.Any(), gomock.Eq(int(goodSAMLProvider.SSOProviderID.Int32))).Return(nil)
		mockDB.EXPECT().GetSSOProviderUsers(gomock.Any(), int(goodSAMLProvider.ID)).Return(nil, nil)
		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(samlProviderPathFmt, goodSAMLProvider.ID).
			WithURLPathVars(map[string]string{
				api.URIPathVariableSAMLProviderID: fmt.Sprintf("%d", goodSAMLProvider.ID),
			}).
			OnHandlerFunc(resources.DeleteSAMLProvider).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("fails when  provider has attached users", func(t *testing.T) {
		mockDB.EXPECT().GetSAMLProvider(gomock.Any(), samlProviderWithUsers.ID).Return(samlProviderWithUsers, nil)
		mockDB.EXPECT().DeleteSSOProvider(gomock.Any(), gomock.Eq(int(samlProviderWithUsers.SSOProviderID.Int32))).Return(nil)
		mockDB.EXPECT().GetSSOProviderUsers(gomock.Any(), int(samlProviderWithUsers.ID)).Return(model.Users{samlEnabledUser}, nil)
		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(samlProviderPathFmt, samlProviderWithUsers.ID).
			WithURLPathVars(map[string]string{
				api.URIPathVariableSAMLProviderID: fmt.Sprintf("%d", samlProviderWithUsers.ID),
			}).
			OnHandlerFunc(resources.DeleteSAMLProvider).
			Require().
			ResponseStatusCode(http.StatusOK)
	})
}

func TestManagementResource_ListPermissions_SortingError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/permissions"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListPermissions).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestManagementResource_ListPermissions_InvalidFilterPredicate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/permissions"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "invalidPredicate:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListPermissions).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	}
}

func TestManagementResource_ListPermissions_PredicateMismatchWithColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/permissions"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListPermissions).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestManagementResource_ListPermissions_DBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/permissions"
	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllPermissions(gomock.Any(), "authority desc, name", model.SQLFilter{SQLString: "name = 'foo'"}).Return(model.Permissions{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "-authority")
		q.Add("sort_by", "name")
		q.Add("name", "eq:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListPermissions).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestManagementResource_ListPermissions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/permissions"

	perm1 := model.Permission{
		Authority: "a",
		Name:      "a",
		Serial: model.Serial{
			Basic: model.Basic{
				CreatedAt: time.Time{},
			},
		},
	}

	perm2 := model.Permission{
		Authority: "b",
		Name:      "b",
		Serial: model.Serial{
			Basic: model.Basic{
				CreatedAt: time.Time{},
			},
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllPermissions(gomock.Any(), "authority desc, name", model.SQLFilter{SQLString: "name = 'a'"}).Return(model.Permissions{perm1, perm2}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "-authority")
		q.Add("sort_by", "name")
		q.Add("name", "eq:a")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListPermissions).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respPermissions := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respPermissions)
		require.Nil(t, err)

		require.Equal(t, perm1.Authority, respPermissions["data"].(map[string]any)["permissions"].([]any)[0].(map[string]any)["authority"])
		require.Equal(t, perm2.Authority, respPermissions["data"].(map[string]any)["permissions"].([]any)[1].(map[string]any)["authority"])
	}
}
func TestManagementResource_GetPermission(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Invalid permission ID format - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/permissions/invalid",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed"}]}`,
			},
		},
		{
			name: "Error: Database Error GetPermission - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/permissions/123",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetPermission(gomock.Any(), 123).Return(model.Permission{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Permission found - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/permissions/123",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				expectedPermission := model.Permission{
					Authority: "read:users",
					Name:      "Read Users",
					Serial: model.Serial{
						ID: 123,
					},
				}
				mock.mockDatabase.EXPECT().GetPermission(gomock.Any(), 123).Return(expectedPermission, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"authority":"read:users", "created_at":"0001-01-01T00:00:00Z", "deleted_at":{"Time":"0001-01-01T00:00:00Z", "Valid":false}, "id":123, "name":"Read Users", "updated_at":"0001-01-01T00:00:00Z"}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, authz.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/permissions/{%s}", api.URIPathVariablePermissionID), resources.GetPermission).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_ListRoles_SortingError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestManagementResource_ListRoles_InvalidColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("foo", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "column cannot be filtered")
	}
}

func TestManagementResource_ListRoles_InvalidFilterPredicate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "invalidPredicate:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	}
}

func TestManagementResource_ListRoles_PredicateMismatchWithColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestManagementResource_ListRoles_DBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"
	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllRoles(gomock.Any(), "description desc, name", model.SQLFilter{}).Return(model.Roles{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "-description")
		q.Add("sort_by", "name")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestManagementResource_ListRoles(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"

	role1 := model.Role{
		Name:        "a",
		Description: "a",
		Permissions: nil,
		Serial:      model.Serial{},
	}

	role2 := model.Role{
		Name:        "b",
		Description: "b",
		Permissions: nil,
		Serial:      model.Serial{},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllRoles(gomock.Any(), "description desc, name", model.SQLFilter{}).Return(model.Roles{role1, role2}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "-description")
		q.Add("sort_by", "name")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respPermissions := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respPermissions)
		require.Nil(t, err)

		require.Equal(t, role1.Name, respPermissions["data"].(map[string]any)["roles"].([]any)[0].(map[string]any)["name"])
		require.Equal(t, role2.Name, respPermissions["data"].(map[string]any)["roles"].([]any)[1].(map[string]any)["name"])
	}
}

func TestManagementResource_ListRoles_Filtered(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"

	role1 := model.Role{
		Name:        "a",
		Description: "a",
		Permissions: nil,
		Serial:      model.Serial{},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllRoles(gomock.Any(), "", model.SQLFilter{SQLString: "name = 'a'"}).Return(model.Roles{role1}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "eq:a")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListRoles).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respPermissions := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respPermissions)
		require.Nil(t, err)
		require.Equal(t, role1.Name, respPermissions["data"].(map[string]any)["roles"].([]any)[0].(map[string]any)["name"])
	}
}

func TestManagementResource_GetRole(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Invalid role ID format - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/roles/invalid",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed"}]}`,
			},
		},
		{
			name: "Error: Database Error GetRole - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/roles/123",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(123)).Return(model.Role{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Role found - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/roles/123",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				permissionRead := model.Permission{
					Authority: "read:users",
					Name:      "Read Users",
					Serial: model.Serial{
						ID: 1,
					},
				}
				permissionWrite := model.Permission{
					Authority: "write:users",
					Name:      "Write Users",
					Serial: model.Serial{
						ID: 2,
					},
				}

				expectedRole := model.Role{
					Name:        "Administrator",
					Description: "System administrator role",
					Permissions: model.Permissions{permissionRead, permissionWrite},
					Serial: model.Serial{
						ID: 123,
					},
				}
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(123)).Return(expectedRole, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"created_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"description":"System administrator role","id":123,"name":"Administrator","permissions":[{"authority":"read:users","created_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"id":1,"name":"Read Users","updated_at":"0001-01-01T00:00:00Z"},{"authority":"write:users","created_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"id":2,"name":"Write Users","updated_at":"0001-01-01T00:00:00Z"}],"updated_at":"0001-01-01T00:00:00Z"}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, authz.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/roles/{%s}", api.URIPathVariableRoleID), resources.GetRole).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestExpireUserAuthSecret_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/secret"

	badUserId := uuid.NullUUID{}
	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(gomock.Any(), badUserId.UUID).Return(model.User{}, fmt.Errorf("db failure"))

	type Input struct {
		UserId string
	}

	cases := []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{"notauuid"},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsIDMalformed}},
			},
		},
		{
			Input{badUserId.UUID.String()},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
	}

	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		if req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf(endpoint, tc.Input.UserId), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/auth/users/{user_id}/secret", resources.ExpireUserAuthSecret).Methods("DELETE")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.HTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.HTTPStatus)
			}

			var body any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			require.Equal(t, len(tc.Expected.Errors), 1)
			if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected.Errors[0].Message) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected.Errors[0].Message)
			}
		}
	}
}

func TestExpireUserAuthSecret_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/secret"
	userId, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(gomock.Any(), userId).Return(model.User{AuthSecret: &model.AuthSecret{}}, nil)
	mockDB.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf(endpoint, userId), nil); err != nil {
		t.Fatal(err)
	} else {
		router := mux.NewRouter()
		router.HandleFunc("/api/v2/auth/users/{user_id}/secret", resources.ExpireUserAuthSecret).Methods("DELETE")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, rr.Code, http.StatusOK)
	}
}

func TestManagementResource_ListUsers_SortingError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestManagementResource_ListUsers_InvalidColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("foo", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "column cannot be filtered")
	}
}

func TestManagementResource_ListUsers_InvalidFilterPredicate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("first_name", "invalidPredicate:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	}
}

func TestManagementResource_ListUsers_PredicateMismatchWithColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("first_name", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestManagementResource_ListUsers_DBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllUsers(gomock.Any(), "first_name desc, last_name", model.SQLFilter{}).Return(model.Users{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "-first_name")
		q.Add("sort_by", "last_name")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestManagementResource_ListUsers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	user1 := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
	}

	user2 := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "Jane", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "janedoe@gmail.com", Valid: true}},
		PrincipalName: "Jane",
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllUsers(gomock.Any(), "first_name desc, last_name", model.SQLFilter{}).Return(model.Users{user1, user2}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "-first_name")
		q.Add("sort_by", "last_name")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respUsers := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respUsers)
		require.Nil(t, err)

		require.Equal(t, user1.FirstName.String, respUsers["data"].(map[string]any)["users"].([]any)[0].(map[string]any)["first_name"])
		require.Equal(t, user2.FirstName.String, respUsers["data"].(map[string]any)["users"].([]any)[1].(map[string]any)["first_name"])
	}
}

func TestManagementResource_ListUsers_Filtered(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	user1 := model.User{
		FirstName: null.String{
			NullString: sql.NullString{
				String: "a",
				Valid:  true,
			},
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllUsers(gomock.Any(), "", model.SQLFilter{SQLString: "first_name = 'a'"}).Return(model.Users{user1}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("first_name", "eq:a")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListUsers).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respPermissions := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respPermissions)
		require.Nil(t, err)
		require.Equal(t, user1.FirstName.String, respPermissions["data"].(map[string]any)["users"].([]any)[0].(map[string]any)["first_name"])
	}
}

func TestCreateUser_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	badRole := []int32{3}

	badUser := model.User{
		Roles:           model.Roles{},
		PrincipalName:   "Bad User",
		FirstName:       null.StringFrom("bad"),
		LastName:        null.StringFrom("bad"),
		EmailAddress:    null.StringFrom("bad"),
		EULAAccepted:    true,
		AllEnvironments: true,
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil).AnyTimes()
	mockDB.EXPECT().GetRoles(gomock.Any(), badRole).Return(model.Roles{}, fmt.Errorf("db error"))
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Not(badRole)).Return(model.Roles{}, nil).AnyTimes()
	mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockDB.EXPECT().CreateUser(gomock.Any(), badUser).Return(model.User{}, fmt.Errorf("db error"))

	type Input struct {
		Body v2.CreateUserRequest
	}

	cases := []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Roles: []int32{2, 3},
				},
			}},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: auth.ErrResponseDetailsNumRoles}},
			},
		},
		{
			Input{v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Roles: badRole,
				},
			}},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
		{
			Input{v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Roles:        []int32{},
					Principal:    badUser.PrincipalName,
					FirstName:    badUser.FirstName.String,
					LastName:     badUser.LastName.String,
					EmailAddress: badUser.EmailAddress.ValueOrZero(),
				},
			}},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
		{
			Input{v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal: "good user",
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "badPASS12345",
					NeedsPasswordReset: false,
				},
			}},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf("Secret: "+validation.ErrorPasswordSpecial, 1)}},
			},
		},
	}

	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		if payload, err := json.Marshal(tc.Input.Body); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.HTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.HTTPStatus)
			}

			require.Equal(t, len(tc.Expected.Errors), 1)
			if rr.Body.Bytes() != nil {
				var body any
				if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatal("failed to unmarshal response body")
				}

				if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected.Errors[0].Message) {
					t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected.Errors[0].Message)
				}
			} else if tc.Expected.Errors[0].Message != "" {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, rr.Body, tc.Expected.Errors[0].Message)
			}
		}
	}
}

func TestCreateUser_FailureDuplicateEmail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(model.User{}, database.ErrDuplicateEmail)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	if payload, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusConflict)
		require.Contains(t, rr.Body.String(), "email must be unique")
	}
}

func TestCreateUser_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	goodUser := model.User{PrincipalName: "good user"}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	if payload, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusOK)
		require.Contains(t, rr.Body.String(), "good user")
	}
}

func TestCreateUser_ETAC(t *testing.T) {
	endpoint := "/api/v2/auth/users"

	tests := []struct {
		name           string
		goodUser       model.User
		createReq      v2.CreateUserRequest
		expectMocks    func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph)
		expectedStatus int
		returnedRoles  model.Roles
		assertBody     func(t *testing.T, body string)
	}{
		{
			name: "Success setting all_environments on user",
			goodUser: model.User{
				PrincipalName:   "good user",
				AllEnvironments: true,
			},
			createReq: v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal:                        "good user",
					AllEnvironments:                  null.BoolFrom(true),
					EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{},
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, `"all_environments":true`)
				assert.Contains(t, body, `"environment_targeted_access_control":null`)
			},
		},
		{
			name: "Success creating an etac list on new user",
			goodUser: model.User{
				PrincipalName: "good user",
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "12345"},
					{EnvironmentID: "54321"},
				},
			},
			createReq: v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal: "good user",
					EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
						Environments: []v2.UpdateEnvironmentRequest{
							{
								EnvironmentID: "12345",
							},
							{
								EnvironmentID: "54321",
							},
						},
					},
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockGraphDB.EXPECT().FetchNodesByObjectIDsAndKinds(gomock.Any(), graph.Kinds{ad.Domain, azure.Tenant}, []string{"12345", "54321"}).Return(graph.NodeSet{
					graph.ID(1): {
						Properties: graph.AsProperties(map[string]any{
							common.ObjectID.String(): "12345",
						}),
					},
					graph.ID(2): {
						Properties: graph.AsProperties(map[string]any{
							common.ObjectID.String(): "54321",
						}),
					},
				}, nil)
				mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, `"environment_id":"12345"`)
				assert.Contains(t, body, `"environment_id":"54321"`)
				assert.Contains(t, body, `"all_environments":false`)
			},
		},
		{
			name: "Success when ETAC enabled and list omitted defaults to all environments",
			goodUser: model.User{
				PrincipalName:   "good user",
				AllEnvironments: true,
			},
			createReq: v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal: "good user",
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, `"all_environments":true`)
			},
		},
		{
			name: "Error setting etac for ineligible role",
			goodUser: model.User{
				PrincipalName: "good user",
			},
			createReq: v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal: "good user",
					EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
						Environments: []v2.UpdateEnvironmentRequest{
							{
								EnvironmentID: "12345",
							},
							{
								EnvironmentID: "54321",
							},
						},
					},
					Roles: []int32{1},
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			},
			returnedRoles: []model.Role{
				{
					Name: authz.RoleAdministrator,
				},
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
			},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, api.ErrorResponseETACInvalidRoles)
			},
		},
		{
			name: "Error setting both environment list and all environments to true",
			goodUser: model.User{
				PrincipalName: "good user",
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "12345"},
					{EnvironmentID: "54321"},
				},
			},
			createReq: v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal: "good user",
					EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
						Environments: []v2.UpdateEnvironmentRequest{
							{
								EnvironmentID: "12345",
							},
							{
								EnvironmentID: "54321",
							},
						},
					},
					AllEnvironments: null.BoolFrom(true),
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
			},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, api.ErrorResponseETACBadRequest)
			},
		},
		{
			name: "Error setting etac list on user when environment does not exist",
			goodUser: model.User{
				PrincipalName: "good user",
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "12345"},
					{EnvironmentID: "54321"},
				},
			},
			createReq: v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{
					Principal: "good user",
					EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
						Environments: []v2.UpdateEnvironmentRequest{
							{
								EnvironmentID: "12345",
							},
							{
								EnvironmentID: "54321",
							},
						},
					},
				},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockGraphDB.EXPECT().FetchNodesByObjectIDsAndKinds(gomock.Any(), graph.Kinds{ad.Domain, azure.Tenant}, []string{"12345", "54321"}).Return(graph.NodeSet{
					graph.ID(1): {
						Properties: graph.AsProperties(map[string]any{
							common.ObjectID.String(): "12345",
						}),
					},
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "domain or tenant not found: 54321")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			resources, mockDB, mockGraphDB := apitest.NewAuthManagementResource(mockCtrl)

			// common mocks
			mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
				Key: appcfg.PasswordExpirationWindow,
				Value: must.NewJSONBObject(appcfg.PasswordExpiration{
					Duration: appcfg.DefaultPasswordExpirationWindow,
				}),
			}, nil)
			mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(tc.returnedRoles, nil)

			// case-specific mocks
			tc.expectMocks(mockDB, tc.goodUser, mockGraphDB)

			resources.DogTags = dogtags.NewTestService(dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			})

			// request/response
			ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
			payload, err := json.Marshal(tc.createReq)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
			require.NoError(t, err)
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.CreateUser).Methods(http.MethodPost)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)
			tc.assertBody(t, rr.Body.String())
		})
	}
}

func TestCreateUser_ResetPassword(t *testing.T) {
	goodUser := model.User{
		PrincipalName: "good user",
		AuthSecret:    &model.AuthSecret{ExpiresAt: time.Time{}},
	}
	goodUserMap, err := utils.MarshalToMap(goodUser)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := "/api/v2/auth/users"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)

	input := struct {
		Body v2.CreateUserRequest
	}{
		v2.CreateUserRequest{
			UpdateUserRequest: v2.UpdateUserRequest{
				Principal: "good user",
			},
			SetUserSecretRequest: v2.SetUserSecretRequest{
				Secret:             "abcDEF123456$$",
				NeedsPasswordReset: true,
			},
		},
	}

	expected := struct {
		Code int
		Body any
	}{
		http.StatusOK,
		goodUserMap,
	}

	bhlog.ConfigureDefaultText(os.Stdout)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	payload, err := json.Marshal(input.Body)
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, expected.Code, rr.Code)

	responseStructure := api.ResponseWrapper{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseStructure)
	require.Nil(t, err)

	expiresAt := responseStructure.Data.(map[string]any)["AuthSecret"].(map[string]any)["expires_at"]
	require.Equal(t, time.Time{}.Format(time.RFC3339), expiresAt)
}

func TestManagementResource_UpdateUser_IDMalformed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	goodUser := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: goodUserID,
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusOK)

	payload, err = json.Marshal(v2.UpdateUserRequest{})
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", goodUserID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsIDMalformed)
}

func TestManagementResource_UpdateUser_GetUserError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	goodUser := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: goodUserID,
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusOK)

	payload, err = json.Marshal(v2.UpdateUserRequest{})
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", goodUserID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: goodUserID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestManagementResource_UpdateUser_GetRolesError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	goodUser := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: goodUserID,
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusOK)

	payload, err = json.Marshal(v2.UpdateUserRequest{})
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", goodUserID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: goodUserID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestManagementResource_UpdateUser_DuplicateEmailError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{EmailAddress: null.StringFrom("")}, nil)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(database.ErrDuplicateEmail)

	reqCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

	payload, err := json.Marshal(v2.UpdateUserRequest{EmailAddress: "different"})
	require.Nil(t, err)

	endpoint := fmt.Sprintf("/api/v2/bloodhound-users/%v", goodUserID)
	req, err := http.NewRequestWithContext(reqCtx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: goodUserID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusConflict, response.Code)
	require.Contains(t, response.Body.String(), "email must be unique")
}

func TestManagementResource_UpdateUser_SelfDisable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		endpoint = "/api/v2/auth/users"
		// logged in user has ID 00000000-0000-0000-0000-000000000000
		// leaving ID blank here will make goodUser have the same ID, so this should fail
		goodUser             = model.User{PrincipalName: "good user"}
		isDisabled           = true
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
	)

	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{model.Role{
		Name:        "admin",
		Description: "admin",
		Permissions: model.Permissions{model.Permission{
			Authority: "admin",
			Name:      "admin",
			Serial:    model.Serial{},
		}},
		Serial: model.Serial{},
	}}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusOK)

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	payload, err = json.Marshal(v2.UpdateUserRequest{
		IsDisabled: &isDisabled,
	})
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", userID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseUserSelfDisable)
}

func TestManagementResource_UpdateUser_UserSelfModify(t *testing.T) {
	var (
		adminRole = model.Role{
			Serial: model.Serial{
				ID: 1,
			},
		}
		goodRoles = []int32{1}
		badRole   = model.Role{
			Serial: model.Serial{
				ID: 2,
			},
		}
		badRoles             = []int32{2}
		adminUser            = model.User{AuthSecret: defaultDigestAuthSecret(t, "currentPassword"), Unique: model.Unique{ID: must.NewUUIDv4()}, Roles: model.Roles{adminRole}}
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
	)

	bhCtx := ctx.Get(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{}))
	bhCtx.AuthCtx.Owner = adminUser

	defer mockCtrl.Finish()

	t.Run("Prevent users from changing their own SSO provider", func(t *testing.T) {
		mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{adminRole}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), adminUser.ID).Return(adminUser, nil)
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{}, nil)

		test.Request(t).
			WithContext(bhCtx).
			WithURLPathVars(map[string]string{"user_id": adminUser.ID.String()}).
			WithBody(v2.UpdateUserRequest{
				Principal:     "tester",
				Roles:         goodRoles,
				SSOProviderID: null.Int32From(1),
			}).
			OnHandlerFunc(resources.UpdateUser).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("Prevent users from changing their own roles", func(t *testing.T) {
		mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{badRole}, nil)
		mockDB.EXPECT().GetUser(gomock.Any(), adminUser.ID).Return(adminUser, nil)

		test.Request(t).
			WithContext(bhCtx).
			WithURLPathVars(map[string]string{"user_id": adminUser.ID.String()}).
			WithBody(v2.UpdateUserRequest{
				Principal: "tester",
				Roles:     badRoles,
			}).
			OnHandlerFunc(resources.UpdateUser).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})
}

func TestManagementResource_UpdateUser_LookupActiveSessionsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	goodUser := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: goodUserID,
		},
	}

	isDisabled := true

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{model.Role{
		Name:        "admin",
		Description: "admin",
		Permissions: model.Permissions{model.Permission{
			Authority: "admin",
			Name:      "admin",
			Serial:    model.Serial{},
		}},
		Serial: model.Serial{},
	}}, nil)
	mockDB.EXPECT().LookupActiveSessionsByUser(gomock.Any(), gomock.Any()).Return([]model.UserSession{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusOK)

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	payload, err = json.Marshal(v2.UpdateUserRequest{
		IsDisabled: &isDisabled,
	})
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", userID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
}

func TestManagementResource_UpdateUser_DBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	goodUser := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: goodUserID,
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{model.Role{
		Name:        "admin",
		Description: "admin",
		Permissions: model.Permissions{model.Permission{
			Authority: "admin",
			Name:      "admin",
			Serial:    model.Serial{},
		}},
		Serial: model.Serial{},
	}}, nil)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK)
	require.Contains(t, rr.Body.String(), "good user")

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	payload, err = json.Marshal(v2.UpdateUserRequest{})
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", userID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestManagementResource_GetUser(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Invalid User ID format - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/invalid",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed"}]}`,
			},
		},
		{
			name: "Error: Database Error GetUser - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				userID, err := uuid.FromString("00000000-0000-0000-0000-000000000001")
				if err != nil {
					t.Fatalf("uuid required for test name %s", t.Name())
				}

				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), userID).Return(model.User{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: User found - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				userID, err := uuid.FromString("00000000-0000-0000-0000-000000000001")
				if err != nil {
					t.Fatalf("uuid required for test name %s", t.Name())
				}

				expectedUser := model.User{
					FirstName:     null.StringFrom("John"),
					LastName:      null.StringFrom("Doe"),
					EmailAddress:  null.StringFrom("john.doe@example.com"),
					PrincipalName: "john.doe",
					Unique: model.Unique{
						ID: userID,
					},
				}
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), userID).Return(expectedUser, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"AuthSecret":null, "all_environments":false, "created_at":"0001-01-01T00:00:00Z", "deleted_at":{"Time":"0001-01-01T00:00:00Z", "Valid":false}, "email_address":"john.doe@example.com", "environment_targeted_access_control":null, "eula_accepted":false, "first_name":"John", "id":"00000000-0000-0000-0000-000000000001", "is_disabled":false, "last_login":"0001-01-01T00:00:00Z", "last_name":"Doe", "principal_name":"john.doe", "roles":null, "sso_provider_id":null, "updated_at":"0001-01-01T00:00:00Z"}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, authz.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), resources.GetUser).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetSelf(t *testing.T) {
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		expected     expected
	}
	tt := []testData{
		{
			name: "Success: Get authenticated user - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/self",
					},
					Method: http.MethodGet,
				}

				userID := uuid.FromStringOrNil("id")
				user := model.User{
					FirstName:     null.StringFrom("John"),
					LastName:      null.StringFrom("Doe"),
					EmailAddress:  null.StringFrom("john.doe@example.com"),
					PrincipalName: "john.doe",
					Roles: model.Roles{
						{
							Name:        "Big Boy",
							Description: "The big boy.",
							Permissions: model.Permissions{},
						},
					},
					Unique: model.Unique{
						ID: userID,
					},
				}

				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.AuthCtx.Owner = user

				return request.WithContext(userContext)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"AuthSecret":null, "all_environments":false, "created_at":"0001-01-01T00:00:00Z", "deleted_at":{"Time":"0001-01-01T00:00:00Z", "Valid":false}, "email_address":"john.doe@example.com", "environment_targeted_access_control":null, "eula_accepted":false, "first_name":"John", "id":"00000000-0000-0000-0000-000000000000", "is_disabled":false, "last_login":"0001-01-01T00:00:00Z", "last_name":"Doe", "principal_name":"john.doe", "roles":[{"created_at":"0001-01-01T00:00:00Z", "deleted_at":{"Time":"0001-01-01T00:00:00Z", "Valid":false}, "description":"The big boy.", "id":0, "name":"Big Boy", "permissions":[], "updated_at":"0001-01-01T00:00:00Z"}], "sso_provider_id":null, "updated_at":"0001-01-01T00:00:00Z"}}`,
			},
		},
		{
			name: "Success: Get empty authenticated user - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/self",
					},
					Method: http.MethodGet,
				}

				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.AuthCtx.Owner = model.User{}

				return request.WithContext(userContext)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"AuthSecret":null, "all_environments":false, "created_at":"0001-01-01T00:00:00Z", "deleted_at":{"Time":"0001-01-01T00:00:00Z", "Valid":false}, "email_address":null, "environment_targeted_access_control":null, "eula_accepted":false, "first_name":null, "id":"00000000-0000-0000-0000-000000000000", "is_disabled":false, "last_login":"0001-01-01T00:00:00Z", "last_name":null, "principal_name":"", "roles":null, "sso_provider_id":null, "updated_at":"0001-01-01T00:00:00Z"}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.buildRequest()

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, &database.BloodhoundDB{}, authz.NewAuthorizer(&database.BloodhoundDB{}), api.NewAuthenticator(config.Configuration{}, &database.BloodhoundDB{}, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc(request.URL.Path, resources.GetSelf).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_DeleteUser_BadUserID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"
	userID := "badUserID"

	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusBadRequest)
}

func TestManagementResource_DeleteUser_UserNotFound(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		endpoint = "/api/v2/bloodhound-users"
	)
	defer mockCtrl.Finish()

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), userID).Return(model.User{}, database.ErrNotFound)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusNotFound)
}

func TestManagementResource_DeleteUser_UserBhCtxNotFound(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		endpoint = "/api/v2/bloodhound-users"
	)

	defer mockCtrl.Finish()

	userID, err := uuid.NewV4()
	require.NoError(t, err)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), userID).Return(model.User{
		Unique: model.Unique{
			ID: userID,
		},
	}, nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "No associated user found")
}

func TestManagementResource_DeleteUser_UserCannotSelfDelete(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		adminRole = model.Role{
			Serial: model.Serial{
				ID: 1,
			},
		}
		endpoint = "/api/v2/bloodhound-users"
	)

	defer mockCtrl.Finish()

	userID, err := uuid.NewV4()
	require.NoError(t, err)

	adminUser := model.User{AuthSecret: defaultDigestAuthSecret(t, "currentPassword"), Unique: model.Unique{ID: userID}, Roles: model.Roles{adminRole}}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), userID).Return(model.User{
		Unique: model.Unique{
			ID: userID,
		},
	}, nil)

	bhCtx := ctx.Get(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{}))
	bhCtx.AuthCtx.Owner = adminUser

	req, err := http.NewRequestWithContext(bhCtx.ConstructGoContext(), "DELETE", endpoint, nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "User cannot delete themselves")
}

func TestManagementResource_DeleteUser_GetUserError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), userID).Return(model.User{}, fmt.Errorf("foo"))

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusInternalServerError)
}

func TestManagementResource_DeleteUser_DeleteUserError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		adminRole = model.Role{
			Serial: model.Serial{
				ID: 1,
			},
		}
		endpoint = "/api/v2/bloodhound-users"
	)

	defer mockCtrl.Finish()

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	user := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: userID,
		},
	}

	adminUser := model.User{AuthSecret: defaultDigestAuthSecret(t, "currentPassword"), Unique: model.Unique{ID: must.NewUUIDv4()}, Roles: model.Roles{adminRole}}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), userID).Return(user, nil)
	mockDB.EXPECT().DeleteUser(gomock.Any(), user).Return(fmt.Errorf("foo"))

	bhCtx := ctx.Get(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{}))
	bhCtx.AuthCtx.Owner = adminUser
	req, err := http.NewRequestWithContext(bhCtx.ConstructGoContext(), "DELETE", endpoint, nil)
	require.NoError(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusInternalServerError)
}

func TestManagementResource_DeleteUser_Success(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		adminRole = model.Role{
			Serial: model.Serial{
				ID: 1,
			},
		}
		endpoint = "/api/v2/bloodhound-users"
	)
	defer mockCtrl.Finish()

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	user := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: userID,
		},
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), userID).Return(user, nil)
	mockDB.EXPECT().DeleteUser(gomock.Any(), user).Return(nil)

	adminUser := model.User{AuthSecret: defaultDigestAuthSecret(t, "currentPassword"), Unique: model.Unique{ID: must.NewUUIDv4()}, Roles: model.Roles{adminRole}}

	bhCtx := ctx.Get(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{}))
	bhCtx.AuthCtx.Owner = adminUser
	req, err := http.NewRequestWithContext(bhCtx.ConstructGoContext(), "DELETE", endpoint, nil)
	require.NoError(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.DeleteUser)
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK)
}

func TestManagementResource_UpdateUser_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"

	goodUserID, err := uuid.NewV4()
	require.Nil(t, err)

	goodUser := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: goodUserID,
		},
	}

	isDisabled := true

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{model.Role{
		Name:        "admin",
		Description: "admin",
		Permissions: model.Permissions{model.Permission{
			Authority: "admin",
			Name:      "admin",
			Serial:    model.Serial{},
		}},
		Serial: model.Serial{},
	}}, nil)
	mockDB.EXPECT().LookupActiveSessionsByUser(gomock.Any(), gomock.Any()).Return([]model.UserSession{}, nil)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	input := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal: "good user",
		},
		SetUserSecretRequest: v2.SetUserSecretRequest{
			Secret:             "abcDEF123456$$",
			NeedsPasswordReset: true,
		},
	}

	payload, err := json.Marshal(input)
	require.Nil(t, err)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK)
	require.Contains(t, rr.Body.String(), "good user")

	userID, err := uuid.NewV4()
	require.Nil(t, err)
	updateUserRequest := v2.UpdateUserRequest{
		IsDisabled: &isDisabled,
	}

	payload, err = json.Marshal(updateUserRequest)
	require.Nil(t, err)

	endpoint = fmt.Sprintf("/api/v2/bloodhound-users/%v", userID)
	req, err = http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(payload))
	require.Nil(t, err)

	req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: userID.String()})
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(resources.UpdateUser)
	handler.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestManagementResource_UpdateUser_ETAC(t *testing.T) {
	endpoint := "/api/v2/auth/users"

	type testCase struct {
		name           string
		setupUser      func(uuid.UUID) model.User
		updateRequest  v2.UpdateUserRequest
		expectedStatus int
		returnedRoles  model.Roles
		assertBody     func(t *testing.T, body string)
		expectMocks    func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph)
	}

	isDisabled := true

	tests := []testCase{
		{
			name: "Success updating a user to all environments",
			setupUser: func(id uuid.UUID) model.User {
				return model.User{
					PrincipalName: "good user",
					Unique:        model.Unique{ID: id},
				}
			},
			updateRequest: v2.UpdateUserRequest{
				IsDisabled:                       &isDisabled,
				AllEnvironments:                  null.BoolFrom(true),
				EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{},
			},
			expectedStatus: http.StatusOK,
			assertBody:     func(t *testing.T, _ string) {},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "Success updating a user to specific environments",
			setupUser: func(id uuid.UUID) model.User {
				return model.User{
					PrincipalName: "good user",
					Unique:        model.Unique{ID: id},
				}
			},
			updateRequest: v2.UpdateUserRequest{
				IsDisabled:      &isDisabled,
				AllEnvironments: null.BoolFrom(false),
				EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
					Environments: []v2.UpdateEnvironmentRequest{
						{
							EnvironmentID: "12345",
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			assertBody:     func(t *testing.T, _ string) {},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockGraphDB.EXPECT().FetchNodesByObjectIDsAndKinds(gomock.Any(), graph.Kinds{ad.Domain, azure.Tenant}, []string{"12345"}).Return(graph.NodeSet{
					graph.ID(1): {
						Properties: graph.AsProperties(map[string]any{
							common.ObjectID.String(): "12345",
						}),
					},
				}, nil)
				mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "Error when current role is Administrator and roles omitted",
			setupUser: func(id uuid.UUID) model.User {
				return model.User{
					PrincipalName: "good user",
					Unique:        model.Unique{ID: id},
					Roles:         model.Roles{{Name: authz.RoleAdministrator}},
				}
			},
			updateRequest: v2.UpdateUserRequest{
				IsDisabled: &isDisabled,
				EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
					Environments: []v2.UpdateEnvironmentRequest{
						{
							EnvironmentID: "12345",
						},
					},
				},
			},
			returnedRoles:  nil,
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, api.ErrorResponseETACInvalidRoles)
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {},
		},
		{
			name: "Error attempting to set both all_environments true and set access to specific environments",
			setupUser: func(id uuid.UUID) model.User {
				return model.User{
					PrincipalName:   "good user",
					Unique:          model.Unique{ID: id},
					AllEnvironments: true,
				}
			},
			updateRequest: v2.UpdateUserRequest{
				IsDisabled:      &isDisabled,
				AllEnvironments: null.BoolFrom(true),
				EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
					Environments: []v2.UpdateEnvironmentRequest{
						{
							EnvironmentID: "12345",
						},
						{
							EnvironmentID: "54321",
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, api.ErrorResponseETACBadRequest)
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				// no DeleteEnvironmentTargetedAccessControlForUser or UpdateUser expected here
			},
		},
		{
			name: "Error setting etac list on user when environment does not exist ",
			setupUser: func(id uuid.UUID) model.User {
				return model.User{
					PrincipalName: "good user",
					Unique:        model.Unique{ID: id},
				}
			},
			updateRequest: v2.UpdateUserRequest{
				IsDisabled:      &isDisabled,
				AllEnvironments: null.BoolFrom(false),
				EnvironmentTargetedAccessControl: &v2.UpdateUserETACRequest{
					Environments: []v2.UpdateEnvironmentRequest{
						{
							EnvironmentID: "12345",
						},
						{
							EnvironmentID: "54321",
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "domain or tenant not found: 54321")
			},
			expectMocks: func(mockDB *mocks.MockDatabase, goodUser model.User, mockGraphDB *mocks_graph.MockGraph) {
				mockGraphDB.EXPECT().FetchNodesByObjectIDsAndKinds(gomock.Any(), graph.Kinds{ad.Domain, azure.Tenant}, []string{"12345", "54321"}).Return(graph.NodeSet{
					graph.ID(1): {
						Properties: graph.AsProperties(map[string]any{
							common.ObjectID.String(): "12345",
						}),
					},
				}, nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			goodUserID, err := uuid.NewV4()
			require.Nil(t, err)

			goodUser := tc.setupUser(goodUserID)

			resources, mockDB, mockGraphDB := apitest.NewAuthManagementResource(mockCtrl)

			// common mocks
			mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
				Key: appcfg.PasswordExpirationWindow,
				Value: must.NewJSONBObject(appcfg.PasswordExpiration{
					Duration: appcfg.DefaultPasswordExpirationWindow,
				}),
			}, nil)
			mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(model.Roles{}, nil)
			mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
			mockDB.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(goodUser, nil)
			mockDB.EXPECT().GetRoles(gomock.Any(), gomock.Any()).Return(tc.returnedRoles, nil)
			mockDB.EXPECT().LookupActiveSessionsByUser(gomock.Any(), gomock.Any()).Return([]model.UserSession{}, nil)

			// case-specific expectations
			tc.expectMocks(mockDB, goodUser, mockGraphDB)

			ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

			resources.DogTags = dogtags.NewTestService(dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			})

			// create user first
			createInput := v2.CreateUserRequest{
				UpdateUserRequest: v2.UpdateUserRequest{Principal: "good user"},
				SetUserSecretRequest: v2.SetUserSecretRequest{
					Secret:             "abcDEF123456$$",
					NeedsPasswordReset: true,
				},
			}
			payload, err := json.Marshal(createInput)
			require.Nil(t, err)
			req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
			require.Nil(t, err)
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.CreateUser).Methods("POST")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusOK)

			// update user
			updatePayload, err := json.Marshal(tc.updateRequest)
			require.Nil(t, err)
			updateEndpoint := fmt.Sprintf("/api/v2/bloodhound-users/%v", goodUserID)
			req, err = http.NewRequestWithContext(ctx, http.MethodPatch, updateEndpoint, bytes.NewReader(updatePayload))
			require.Nil(t, err)
			req = mux.SetURLVars(req, map[string]string{api.URIPathVariableUserID: goodUserID.String()})
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			response := httptest.NewRecorder()
			http.HandlerFunc(resources.UpdateUser).ServeHTTP(response, req)

			require.Equal(t, tc.expectedStatus, response.Code)
			tc.assertBody(t, response.Body.String())
		})
	}
}

func TestManagementResource_ListAuthTokens_SortingError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	user := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}

	authToken1 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "auth_token_1", Valid: true}},
		Key:        "abcd",
		HmacMethod: "method",
		LastAccess: time.Now().Add(100 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
			},
		},
	}

	authToken2 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "auth_token_2", Valid: true}},
		Key:        "abcde",
		HmacMethod: "method",
		LastAccess: time.Now().Add(150 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(200 * time.Second),
					Valid: true,
				},
			},
		},
	}

	user.AuthTokens = model.AuthTokens{authToken1, authToken2}

	c := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(c)
	bhCtx.AuthCtx.Owner = user
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/tokens"
	if req, err := http.NewRequestWithContext(c, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestManagementResource_ListAuthTokens_InvalidColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/tokens"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("foo", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "column cannot be filtered")
	}
}

func TestManagementResource_ListAuthTokens_InvalidFilterPredicate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/tokens"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("first_name", "invalidPredicate:foo")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	}
}

func TestManagementResource_ListAuthTokens_PredicateMismatchWithColumn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/tokens"
	resources, _, _ := apitest.NewAuthManagementResource(mockCtrl)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestManagementResource_ListAuthTokens_DBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	user := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}

	user.AuthTokens = model.AuthTokens{}

	c := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(c)
	bhCtx.AuthCtx.Owner = user
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllAuthTokens(gomock.Any(), "name, last_access desc", model.SQLFilter{SQLString: "user_id = '" + user.ID.String() + "'"}).Return(model.AuthTokens{}, fmt.Errorf("foo"))

	endpoint := "/api/v2/auth/tokens"
	if req, err := http.NewRequestWithContext(c, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "name")
		q.Add("sort_by", "-last_access")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestManagementResource_ListAuthTokens_Admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	user := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}

	otherUser := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "Other", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "User", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "otheruser@gmail.com", Valid: true}},
		PrincipalName: "OtherUser",
		AuthTokens:    model.AuthTokens{},
	}

	authToken1 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "auth_token_1", Valid: true}},
		Key:        "abcd",
		HmacMethod: "method",
		LastAccess: time.Now().Add(100 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
			},
		},
	}

	authToken2 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "auth_token_2", Valid: true}},
		Key:        "abcde",
		HmacMethod: "method",
		LastAccess: time.Now().Add(150 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(200 * time.Second),
					Valid: true,
				},
			},
		},
	}

	otherUserToken := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  otherUser.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "other_user_token_1", Valid: true}},
		Key:        "abcdef",
		HmacMethod: "method",
		LastAccess: time.Now().Add(150 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(200 * time.Second),
					Valid: true,
				},
			},
		},
	}

	user.AuthTokens = model.AuthTokens{authToken1, authToken2}
	otherUser.AuthTokens = model.AuthTokens{otherUserToken}
	allAuthTokens := model.AuthTokens{authToken1, authToken2, otherUserToken}

	c := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(c)
	bhCtx.AuthCtx.Owner = user
	bhCtx.AuthCtx.PermissionOverrides = authz.PermissionOverrides{
		Enabled: true,
		Permissions: model.Permissions{
			authz.Permissions().AuthManageUsers,
		},
	}
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllAuthTokens(gomock.Any(), "name, last_access desc", model.SQLFilter{}).Return(allAuthTokens, nil)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1

	endpoint := "/api/v2/auth/tokens"
	if req, err := http.NewRequestWithContext(c, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "name")
		q.Add("sort_by", "-last_access")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respTokens := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respTokens)
		require.Nil(t, err)

		require.Equal(t, authToken1.Name.String, respTokens["data"].(map[string]any)["tokens"].([]any)[0].(map[string]any)["name"])
		require.Equal(t, authToken2.Name.String, respTokens["data"].(map[string]any)["tokens"].([]any)[1].(map[string]any)["name"])
		require.Equal(t, otherUserToken.Name.String, respTokens["data"].(map[string]any)["tokens"].([]any)[2].(map[string]any)["name"])
	}
}

func TestManagementResource_ListAuthTokens_NonAdmin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	user := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}

	otherUser := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "Other", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "User", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "otheruser@gmail.com", Valid: true}},
		PrincipalName: "OtherUser",
		AuthTokens:    model.AuthTokens{},
	}

	authToken1 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "auth_token_1", Valid: true}},
		Key:        "abcd",
		HmacMethod: "method",
		LastAccess: time.Now().Add(100 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
			},
		},
	}

	authToken2 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "auth_token_2", Valid: true}},
		Key:        "abcde",
		HmacMethod: "method",
		LastAccess: time.Now().Add(150 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(200 * time.Second),
					Valid: true,
				},
			},
		},
	}

	otherUserToken := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  otherUser.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "other_user_token_1", Valid: true}},
		Key:        "abcdef",
		HmacMethod: "method",
		LastAccess: time.Now().Add(150 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(200 * time.Second),
					Valid: true,
				},
			},
		},
	}

	user.AuthTokens = model.AuthTokens{authToken1, authToken2}
	otherUser.AuthTokens = model.AuthTokens{otherUserToken}

	c := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(c)
	bhCtx.AuthCtx.Owner = user
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllAuthTokens(gomock.Any(), "name, last_access desc", model.SQLFilter{SQLString: "user_id = '" + user.ID.String() + "'"}).Return(user.AuthTokens, nil)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1

	endpoint := "/api/v2/auth/tokens"
	if req, err := http.NewRequestWithContext(c, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "name")
		q.Add("sort_by", "-last_access")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respTokens := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respTokens)
		require.Nil(t, err)

		require.Equal(t, len(respTokens["data"].(map[string]any)["tokens"].([]any)), 2)
		require.Equal(t, authToken1.Name.String, respTokens["data"].(map[string]any)["tokens"].([]any)[0].(map[string]any)["name"])
		require.Equal(t, authToken2.Name.String, respTokens["data"].(map[string]any)["tokens"].([]any)[1].(map[string]any)["name"])
	}
}

func TestManagementResource_ListAuthTokens_Filtered(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/tokens"

	user := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}

	authToken1 := model.AuthToken{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Name:       null.String{NullString: sql.NullString{String: "a", Valid: true}},
		Key:        "abcd",
		HmacMethod: "method",
		LastAccess: time.Now().Add(100 * time.Second),
		Unique: model.Unique{
			Basic: model.Basic{
				CreatedAt: time.Now(),
			},
		},
	}

	user.AuthTokens = model.AuthTokens{authToken1}

	c := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(c)
	bhCtx.AuthCtx.Owner = user
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	// The filters are stored in a map before parsing, which means we don't know what order the resulted SQLFilter will be in.
	// Mock out both possibilities to catch both cases.
	mockDB.EXPECT().GetAllAuthTokens(gomock.Any(), "", model.SQLFilter{SQLString: "name = 'a' and user_id = '" + user.ID.String() + "'"}).AnyTimes().Return(model.AuthTokens{authToken1}, nil)
	mockDB.EXPECT().GetAllAuthTokens(gomock.Any(), "", model.SQLFilter{SQLString: "user_id = '" + user.ID.String() + "' and name = 'a'"}).AnyTimes().Return(model.AuthTokens{authToken1}, nil)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1

	if req, err := http.NewRequestWithContext(c, "GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "eq:a")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthTokens).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		respPermissions := map[string]any{}
		err := json.Unmarshal(response.Body.Bytes(), &respPermissions)
		require.Nil(t, err)
		require.Equal(t, authToken1.Name.String, respPermissions["data"].(map[string]any)["tokens"].([]any)[0].(map[string]any)["name"])
	}
}

func defaultDigestAuthSecretWithTOTP(t *testing.T, value, totpSecret string) *model.AuthSecret {
	authSecret := defaultDigestAuthSecret(t, value)
	authSecret.TOTPSecret = totpSecret

	return authSecret
}

func defaultDigestAuthSecret(t *testing.T, value string) *model.AuthSecret {
	var (
		digester = config.Argon2Configuration{
			MemoryKibibytes: 1024 * 1024,
			NumIterations:   1,
			NumThreads:      1,
		}.NewDigester()

		digest, err = digester.Digest(value)
	)

	if err != nil {
		t.Fatal(err)
	}

	return &model.AuthSecret{
		DigestMethod: digester.Method(),
		Digest:       digest.String(),
	}
}

func TestManagementResource_CreateAuthToken(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: API Keys are disabled - Forbidden Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Method: http.MethodPost,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: appcfg.APITokensParameter{Enabled: false},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":403,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"API key creation is disabled"}]}`,
			},
		},
		{
			name: "Error: GetUserFromAuthCtx unable to get user from ctx - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Method: http.MethodPost,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: ReadJSONRequestPayloadLimited invalid header - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), "invalid")
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Header: header,
					Method: http.MethodPost,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"error unmarshalling JSON payload"}]}`,
			},
		},
		{
			name: "Error: ReadJSONRequestPayloadLimited ErrNoRequestBody - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Header: header,
					Method: http.MethodPost,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"error unmarshalling JSON payload"}]}`,
			},
		},
		{
			name: "Error: Database Error GetUser - Internal Server Error",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"token_name":"name","user_id":"id"}`))),
					Header: header,
					Method: http.MethodPost,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")).Return(model.User{}, errors.New("error"))
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: invalid userID verifyUserID - Forbidden",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Method: http.MethodPost,
					Header: header,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"token_name":"name","user_id":"1"}`))),
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")).Return(model.User{}, nil)
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":403,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"missing permission to create tokens for other users"}]}`,
			},
		},
		{
			name: "Error: auth.NewUserAuthToken error creating auth token - Internal Server Error",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Method: http.MethodPost,
					Header: header,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"token_name":"name","user_id":"id"}`))),
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								model.NewPermission("auth", "ManageUsers"),
							},
						},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")).Return(model.User{}, nil)
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: database error db.CreateAuthToken - Internal Server Error",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Method: http.MethodPost,
					Header: header,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"token_name":"name","user_id":"00000000-0000-0000-0000-000000000000"}`))),
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Permissions: model.Permissions{model.NewPermission("auth", "ManageUsers")},
								},
							},
						},
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								model.NewPermission("auth", "ManageUsers"),
							},
						},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")).Return(model.User{}, nil)
				mock.mockDatabase.EXPECT().CreateAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, errors.New("error"))
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: created auth token; user has correct permissions - OK",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens",
					},
					Method: http.MethodPost,
					Header: header,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"token_name":"name","user_id":"00000000-0000-0000-0000-000000000000"}`))),
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Permissions: model.Permissions{model.NewPermission("auth", "ManageUsers")},
								},
							},
						},
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								model.NewPermission("auth", "ManageUsers"),
							},
						},
					},
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000")).Return(model.User{}, nil)
				mock.mockDatabase.EXPECT().CreateAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
					UserID:     uuid.NullUUID{UUID: uuid.FromStringOrNil("id")},
					ClientID:   uuid.NullUUID{UUID: uuid.FromStringOrNil("id")},
					Name:       null.StringFrom("name"),
					Key:        "key",
					HmacMethod: "hmac-sha2-256",
					LastAccess: time.Time{},
					Unique: model.Unique{
						ID: uuid.FromStringOrNil("id"),
						Basic: model.Basic{
							CreatedAt: time.Time{},
							UpdatedAt: time.Time{},
							DeletedAt: sql.NullTime{},
						},
					},
				}, nil)
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(appcfg.Parameter{Key: appcfg.APITokens, Value: types.JSONBObject{
					Object: &appcfg.APITokensParameter{Enabled: true},
				}}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"created_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"hmac_method":"hmac-sha2-256","id":"00000000-0000-0000-0000-000000000000","key":"key","last_access":"0001-01-01T00:00:00Z","name":"name","updated_at":"0001-01-01T00:00:00Z","user_id":null}}`,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, authz.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/tokens", resources.CreateAuthToken).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_EnrollMFA(t *testing.T) {
	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Invalid Request - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/{%s}/mfa",
					},
					Method: http.MethodPost,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"unable to parse request parameters"}]}`,
			},
		},
		{
			name: "Error: Invalid User ID - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/id/mfa",
					},
					Method:   http.MethodPost,
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed"}]}`,
			},
		},
		{
			name: "Error: ReadJSONRequestPayloadLimited invalid header - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), "invalid")
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Header:   header,
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"content type must be application/json"}]}`,
			},
		},
		{
			name: "Error: ReadJSONRequestPayloadLimited ErrNoRequestBody - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Header:   header,
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"content type must be application/json"}]}`,
			},
		},
		{
			name: "Error: Database error db.GetUser - Internal Server Error",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Body:     io.NopCloser(bytes.NewReader([]byte(`{"secret":"valid"}`))),
					Header:   header,
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: Invalid Operation user.SSOProviderID.Valid - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method: http.MethodPost,
					Header: header,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"secret":"valid"}`))),
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: 9,
							Valid: true,
						},
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"Invalid operation, user is SSO"}]}`,
			},
		},
		{
			name: "Error: Multi Factor Activated user.AuthSecret.TOTPActivated - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Header:   header,
					Body:     io.NopCloser(bytes.NewReader([]byte(`{"secret":"valid"}`))),
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: 9,
							Valid: false,
						},
					},
					AuthSecret: &model.AuthSecret{
						TOTPActivated: true,
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"multi-factor authentication already active"}]}`,
			},
		},
		{
			name: "Error: Invalid secret - Bad Request",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Header:   header,
					Body:     io.NopCloser(bytes.NewReader([]byte(`{"secret":"valid"}`))),
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: 9,
							Valid: false,
						},
					},
					AuthSecret: &model.AuthSecret{
						TOTPActivated: false,
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"unable to verify current password"}]}`,
			},
		},
		{
			name: "Error: auth.GenerateTOTPSecret - Internal Server Error",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				input := &auth.MFAEnrollmentRequest{"password"}
				mfaBytes, err := json.Marshal(input)
				if err != nil {
					t.Fatal("error occurred while marshaling mfa enrollment request")
				}
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method: http.MethodPost,
					Header: header,
					Body:   io.NopCloser(bytes.NewReader(mfaBytes)),
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: 9,
							Valid: false,
						},
					},
					AuthSecret: defaultDigestAuthSecretWithTOTP(t, "password", "password"),
					Unique: model.Unique{
						ID: uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000"),
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: Database Error db.UpdateAuthSecret - Internal Server Error",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				input := &auth.MFAEnrollmentRequest{"password"}
				mfaBytes, err := json.Marshal(input)
				if err != nil {
					t.Fatal("error occurred while marshaling mfa enrollment request")
				}
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Header:   header,
					Body:     io.NopCloser(bytes.NewReader(mfaBytes)),
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: 9,
							Valid: false,
						},
					},
					PrincipalName: "name",
					AuthSecret:    defaultDigestAuthSecretWithTOTP(t, "password", "password"),
					Unique: model.Unique{
						ID: uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000"),
					},
				}, nil)
				mock.mockDatabase.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: MFAEnrollmentReponse - OK",
			buildRequest: func() *http.Request {
				header := http.Header{}
				header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				input := &auth.MFAEnrollmentRequest{"password"}
				mfaBytes, err := json.Marshal(input)
				if err != nil {
					t.Fatal("error occurred while marshaling mfa enrollment request")
				}
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/bloodhound-users/00000000-0000-0000-0000-000000000000/mfa",
					},
					Method:   http.MethodPost,
					Header:   header,
					Body:     io.NopCloser(bytes.NewReader(mfaBytes)),
					PostForm: url.Values{},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					Host: request.URL,
				}))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: 9,
							Valid: false,
						},
					},
					PrincipalName: "name",
					AuthSecret:    defaultDigestAuthSecretWithTOTP(t, "password", "password"),
					Unique: model.Unique{
						ID: uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000"),
					},
				}, nil)
				mock.mockDatabase.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, authz.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{
				Crypto: config.CryptoConfiguration{
					Argon2: config.Argon2Configuration{},
				},
			}, mocks.mockDatabase, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa", api.URIPathVariableUserID), resources.EnrollMFA).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusOK {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				// For success cases, the MFAEnrollmentReponse contains a QR Code and TOTP Secret.
				// The QR code image data & TOTP Secret change every time they are generated.
				// Because of this, we can't assert on the exact value of the QR code string.
				// Instead, we verify the structure by checking that it:
				// QR Code
				// - is not empty
				// - has the correct data URI prefix
				// - decodes properly from base64
				// - decodes into a valid PNG image with expected dimensions
				// TOTP Secret
				// - is not empty
				// - uses helper function ValidateTOTPSecret to validate TOTP Secret
				// This approach ensures the QR code is valid without relying on fixed content.
				var resp api.BasicResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err, "Failed to decode top-level JSON response")

				var mfaResp auth.MFAEnrollmentReponse
				err = json.Unmarshal([]byte(resp.Data), &mfaResp)
				assert.NoError(t, err, "Failed to decode MFA enrollment response JSON")

				const prefix = "data:image/png;base64,"

				// Validate QR Code
				// Assert QR code exists and has correct prefix
				assert.NotEmpty(t, mfaResp.QrCode, "QR code should not be empty")
				assert.True(t, strings.HasPrefix(mfaResp.QrCode, prefix), "QR code should start with the base64 PNG image prefix")

				// Extract base64 data and decode
				base64Data := mfaResp.QrCode[len(prefix):]
				decoded, err := base64.StdEncoding.DecodeString(base64Data)
				assert.NoError(t, err, "Base64 decoding should succeed")

				// Decode PNG image from decoded bytes
				img, err := png.Decode(bytes.NewReader(decoded))
				assert.NoError(t, err, "PNG decoding should succeed")
				assert.NotNil(t, img, "Decoded image should not be nil")

				// Check image dimensions
				bounds := img.Bounds()
				assert.Equal(t, 200, bounds.Dx(), "Image width should be 200px")
				assert.Equal(t, 200, bounds.Dy(), "Image height should be 200px")

				// Validate TOTP secret
				assert.NotEmpty(t, mfaResp.TOTPSecret, "TOTP secret should not be empty")
				err = authz.ValidateTOTPSecret(mfaResp.TOTPSecret, model.AuthSecret{})
				assert.NoError(t, err, "TOTP secret should be valid")
			}
		})
	}
}

func TestDisenrollMFA_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa"

	missingUserId := test.NewUUIDv4(t)
	userId := test.NewUUIDv4(t)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), missingUserId).Return(model.User{}, database.ErrNotFound)

	type Input struct {
		UserId string
		Body   any
	}

	cases := []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{"notauuid", nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsIDMalformed}},
			},
		},
		{
			Input{userId.String(), "imnotjson"},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrContentTypeJson.Error()}},
			},
		},
		{
			Input{missingUserId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusNotFound,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsResourceNotFound}},
			},
		},
	}
	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		if payload, err := json.Marshal(tc.Input.Body); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf(endpoint, tc.Input.UserId), bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			test.TestV2HandlerFailure(t, []string{"DELETE"}, fmt.Sprintf(endpoint, "{user_id}"), resources.DisenrollMFA, *req, tc.Expected)
		}
	}
}

func TestDisenrollMFA_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"
	userId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(gomock.Any(), userId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)
	mockDB.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil)

	input := auth.MFAEnrollmentRequest{"password"}

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	if payload, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf(endpoint, userId.String()), bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(fmt.Sprintf(endpoint, "{user_id}"), resources.DisenrollMFA).Methods("DELETE")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusOK)
		require.Contains(t, rr.Body.String(), auth.MFADeactivated)
	}
}

func TestDisenrollMFA_Admin_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	admin := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "Admin", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "User", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "admin@gmail.com", Valid: true}},
		PrincipalName: "AdminUser",
		AuthSecret:    defaultDigestAuthSecret(t, "adminpassword"),
	}

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"
	nonAdminId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(gomock.Any(), nonAdminId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)
	mockDB.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil)

	adminContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(adminContext)
	bhCtx.AuthCtx.Owner = admin
	bhCtx.AuthCtx.PermissionOverrides = authz.PermissionOverrides{
		Enabled: true,
		Permissions: model.Permissions{
			authz.Permissions().AuthManageUsers,
		},
	}
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	input := auth.MFAEnrollmentRequest{"adminpassword"}
	if payload, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(adminContext, "DELETE", fmt.Sprintf(endpoint, nonAdminId.String()), bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(fmt.Sprintf(endpoint, "{user_id}"), resources.DisenrollMFA).Methods("DELETE")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), auth.MFADeactivated)
	}
}

func TestDisenrollMFA_Admin_SuccessNoPasswordSSO(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"

	admin := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "Admin", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "User", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "admin@gmail.com", Valid: true}},
		PrincipalName: "AdminUser",
		SSOProviderID: null.Int32{
			NullInt32: sql.NullInt32{
				Int32: 9,
				Valid: true,
			},
		},
	}

	nonAdminId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(gomock.Any(), nonAdminId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password"), Unique: model.Unique{ID: nonAdminId}}, nil).AnyTimes()
	mockDB.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil)

	adminContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(adminContext)
	bhCtx.AuthCtx.Owner = admin
	bhCtx.AuthCtx.PermissionOverrides = authz.PermissionOverrides{
		Enabled: true,
		Permissions: model.Permissions{
			authz.Permissions().AuthManageUsers,
		},
	}
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	input := auth.MFAEnrollmentRequest{}
	if payload, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(adminContext, "DELETE", fmt.Sprintf(endpoint, nonAdminId.String()), bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(fmt.Sprintf(endpoint, "{user_id}"), resources.DisenrollMFA).Methods("DELETE")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), auth.MFADeactivated)
	}
}

func TestDisenrollMFA_Admin_FailureIncorrectPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"
	nonAdminId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(gomock.Any(), nonAdminId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password"), Unique: model.Unique{ID: nonAdminId}}, nil).AnyTimes()

	admin := model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "Admin", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "User", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "admin@gmail.com", Valid: true}},
		PrincipalName: "AdminUser",
		AuthSecret:    defaultDigestAuthSecret(t, "adminpassword"),
	}

	adminContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	bhCtx := ctx.Get(adminContext)
	bhCtx.AuthCtx.Owner = admin
	bhCtx.AuthCtx.PermissionOverrides = authz.PermissionOverrides{
		Enabled: true,
		Permissions: model.Permissions{
			authz.Permissions().AuthManageUsers,
		},
	}
	_, isUser := authz.GetUserFromAuthCtx(bhCtx.AuthCtx)
	require.True(t, isUser)

	// Make the request with the same password as the user we are are attempting to disenroll to ensure the logic remains correct
	if payload, err := json.Marshal(auth.MFAEnrollmentRequest{"password"}); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(adminContext, "DELETE", fmt.Sprintf(endpoint, nonAdminId.String()), bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		test.TestV2HandlerFailure(
			t,
			[]string{"DELETE"},
			fmt.Sprintf(endpoint, "{user_id}"),
			resources.DisenrollMFA,
			*req,
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: auth.ErrResponseDetailsInvalidCurrentPassword}},
			},
		)
	}
}

func TestGetMFAActivationStatus_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa-activation"

	missingId := test.NewUUIDv4(t)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(gomock.Any(), missingId).Return(model.User{}, database.ErrNotFound)

	type Input struct {
		UserId string
		Body   any
	}

	cases := []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{"notauuid", nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsIDMalformed}},
			},
		},
		{
			Input{missingId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusNotFound,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsResourceNotFound}},
			},
		},
	}
	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		if payload, err := json.Marshal(tc.Input.Body); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(endpoint, tc.Input.UserId), bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			test.TestV2HandlerFailure(t, []string{"GET"}, fmt.Sprintf(endpoint, "{user_id}"), resources.GetMFAActivationStatus, *req, tc.Expected)
		}
	}
}

func TestGetMFAActivationStatus_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa-activation"

	activatedId := test.NewUUIDv4(t)
	pendingId := test.NewUUIDv4(t)
	deactivatedId := test.NewUUIDv4(t)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(gomock.Any(), activatedId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: true, TOTPSecret: "imasharedsecret"}}, nil)
	mockDB.EXPECT().GetUser(gomock.Any(), pendingId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: false, TOTPSecret: "imasharedsecret"}}, nil)
	mockDB.EXPECT().GetUser(gomock.Any(), deactivatedId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: false}}, nil)

	type Input struct {
		UserId string
		Body   any
	}

	type Map = map[string]any
	cases := []struct {
		Input    Input
		Expected test.ExpectedResponse
	}{
		{
			Input{activatedId.String(), nil},
			test.ExpectedResponse{
				Code: http.StatusOK,
				Body: Map{"status": string(auth.MFAActivated)},
			},
		},
		{
			Input{pendingId.String(), nil},
			test.ExpectedResponse{
				Code: http.StatusOK,
				Body: Map{"status": string(auth.MFAPending)},
			},
		},
		{
			Input{deactivatedId.String(), nil},
			test.ExpectedResponse{
				Code: http.StatusOK,
				Body: Map{"status": string(auth.MFADeactivated)},
			},
		},
	}
	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		if payload, err := json.Marshal(tc.Input.Body); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(endpoint, tc.Input.UserId), bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf(endpoint, "{user_id}"), resources.GetMFAActivationStatus).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			require.Equal(t, tc.Expected.Code, rr.Code)
			require.Contains(t, rr.Body.String(), tc.Expected.Body.(Map)["status"])
		}
	}
}

func TestActivateMFA_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa-activation"
	missingUserId := test.NewUUIDv4(t)
	unenrolledId := test.NewUUIDv4(t)

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(gomock.Any(), missingUserId).Return(model.User{}, database.ErrNotFound)
	mockDB.EXPECT().GetUser(gomock.Any(), unenrolledId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)

	type Input struct {
		UserId string
		Body   any
	}

	cases := []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{"notauuid", nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsIDMalformed}},
			},
		},
		{
			Input{unenrolledId.String(), "imnotjson"},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrContentTypeJson.Error()}},
			},
		},
		{
			Input{missingUserId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusNotFound,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsResourceNotFound}},
			},
		},
		{
			Input{unenrolledId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: auth.ErrResponseDetailsMFAEnrollmentRequired}},
			},
		},
	}
	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		if payload, err := json.Marshal(tc.Input.Body); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(endpoint, tc.Input.UserId), bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			test.TestV2HandlerFailure(t, []string{"POST"}, fmt.Sprintf(endpoint, "{user_id}"), resources.ActivateMFA, *req, tc.Expected)
		}
	}
}

func TestActivateMFA_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resources, mockDB, _ := apitest.NewAuthManagementResource(mockCtrl)
	totpSecret, err := authz.GenerateTOTPSecret("https://example.com", "foo@bar.baz")
	if err != nil {
		t.Fatal(err)
	}
	passcode, err := totp.GenerateCode(totpSecret.Secret(), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	endpoint := "/api/v2/auth/users/%s/mfa-activation"
	userId := test.NewUUIDv4(t)
	mockDB.EXPECT().GetUser(gomock.Any(), userId).Return(model.User{AuthSecret: defaultDigestAuthSecretWithTOTP(t, "password", totpSecret.Secret())}, nil)
	mockDB.EXPECT().UpdateAuthSecret(gomock.Any(), gomock.Any()).Return(nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	inputBody := auth.MFAActivationRequest{passcode}
	if payload, err := json.Marshal(inputBody); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(endpoint, userId.String()), bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(fmt.Sprintf(endpoint, "{user_id}"), resources.ActivateMFA).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, rr.Code, http.StatusOK)
		require.Contains(t, rr.Body.String(), auth.MFAActivated)
	}
}

func TestManagementResource_DeleteAuthToken(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: GetUserFromAuthCtx unable to get user from ctx - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/{%s}",
					},
					Method: http.MethodDelete,
				}
			},

			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: invalid token_id - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/invalid",
					},
					Method: http.MethodDelete,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								model.NewPermission("auth", "ManageUsers"),
							},
						},
					},
				}))
			},

			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"id is malformed"}]}`,
			},
		},
		{
			name: "Error: Database error db.GetAuthToken - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodDelete,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								model.NewPermission("auth", "ManageUsers"),
							},
						},
					},
				}))
			},

			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: Database Error db.AppendAuditLog - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodDelete,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{},
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								model.NewPermission("auth", "ManageUsers"),
							},
						},
					},
				}))
			},

			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: request user ID != auth token user - Forbidden",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodDelete,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{
							Unique: model.Unique{
								ID: must.NewUUIDv4(),
							},
						},
						// No AuthManageUsers Permission
						PermissionOverrides: authz.PermissionOverrides{
							Enabled:     true,
							Permissions: model.Permissions{},
						},
					},
				}))
			},

			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
					UserID: uuid.NullUUID{
						Valid: true,
					},
				}, nil)
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil)
				// purpose of failed audit log is to add code coverage
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":403,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"Forbidden"}]}`,
			},
		},
		{
			name: "Error: Database error db.DeleteAuthToken - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodDelete,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{
							Unique: model.Unique{
								ID: must.NewUUIDv4(),
							},
						},
						// AuthManageUsers Permission
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								authz.Permissions().AuthManageUsers,
							},
						},
					},
				}))
			},

			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
					UserID: uuid.NullUUID{
						Valid: true,
					},
				}, nil)
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockDatabase.EXPECT().DeleteAuthToken(gomock.Any(), model.AuthToken{
					UserID: uuid.NullUUID{
						Valid: true,
					},
				}).Return(errors.New("error"))
				// purpose of failed audit log is to add code coverage
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(database.ErrNotFound)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Auth token deleted - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/tokens/00000000-0000-0000-0000-000000000001",
					},
					Method: http.MethodDelete,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
					AuthCtx: authz.Context{
						Owner: model.User{
							Unique: model.Unique{
								ID: must.NewUUIDv4(),
							},
						},
						// AuthManageUsers Permission
						PermissionOverrides: authz.PermissionOverrides{
							Enabled: true,
							Permissions: model.Permissions{
								authz.Permissions().AuthManageUsers,
							},
						},
					},
				}))
			},

			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
					UserID: uuid.NullUUID{
						Valid: true,
					},
				}, nil)
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockDatabase.EXPECT().DeleteAuthToken(gomock.Any(), model.AuthToken{
					UserID: uuid.NullUUID{
						Valid: true,
					},
				}).Return(nil)
				// purpose of failed audit log is to add code coverage
				mock.mockDatabase.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			response := httptest.NewRecorder()

			resources := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, authz.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/tokens/{%s}", api.URIPathVariableTokenID), resources.DeleteAuthToken).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if body != "" {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}
