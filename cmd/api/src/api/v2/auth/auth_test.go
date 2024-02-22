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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/api/v2/auth"
	authz "github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/specterops/bloodhound/src/utils/validation"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	samlProviderPathFmt           = "/api/v2/saml/providers/%d"
	updateUserPathFmt             = "/api/v2/auth/users/%s"
	updateUserSecretPathFmt       = "/api/v2/auth/users/%s/secret"
	samlProviderID          int32 = 1234
	samlProviderIDStr             = "1234"
)

func TestManagementResource_PutUserAuthSecret(t *testing.T) {
	var (
		goodUserID        = must.NewUUIDv4()
		badUserID         = must.NewUUIDv4()
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
	)

	defer mockCtrl.Finish()

	mockDB.EXPECT().GetUser(badUserID).Return(model.User{SAMLProviderID: null.Int32From(1)}, nil)
	mockDB.EXPECT().GetUser(goodUserID).Return(model.User{}, nil)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil).Times(1)
	mockDB.EXPECT().CreateAuthSecret(gomock.Any(), gomock.Any()).Return(model.AuthSecret{}, nil).Times(1)

	// Happy path
	test.Request(t).
		WithMethod(http.MethodPut).
		WithHeader(headers.RequestID.String(), "requestID").
		WithURL(fmt.Sprintf(updateUserSecretPathFmt, goodUserID.String())).
		WithURLPathVars(map[string]string{
			"user_id": goodUserID.String(),
		}).
		WithBody(v2.SetUserSecretRequest{
			Secret:             "tesT12345!@#$",
			NeedsPasswordReset: false,
		}).
		OnHandlerFunc(resources.PutUserAuthSecret).
		Require().
		ResponseStatusCode(http.StatusOK)

	// Negative path where a user already has a SAML provider set
	test.Request(t).
		WithMethod(http.MethodPut).
		WithHeader(headers.RequestID.String(), "requestID").
		WithURL(fmt.Sprintf(updateUserSecretPathFmt, badUserID.String())).
		WithURLPathVars(map[string]string{
			"user_id": badUserID.String(),
		}).
		WithBody(v2.SetUserSecretRequest{
			Secret:             "tesT12345!@#$",
			NeedsPasswordReset: false,
		}).
		OnHandlerFunc(resources.PutUserAuthSecret).
		Require().
		ResponseJSONBody(
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors: []api.ErrorDetails{
					{
						Message: "Invalid operation, user is SSO",
					},
				},
			},
		)
}

func TestManagementResource_EnableUserSAML(t *testing.T) {
	var (
		goodRoles         = []int32{0}
		goodUserID        = must.NewUUIDv4()
		badUserID         = must.NewUUIDv4()
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
	)

	defer mockCtrl.Finish()

	mockDB.EXPECT().GetRoles(gomock.Eq(goodRoles)).Return(model.Roles{}, nil).AnyTimes()
	mockDB.EXPECT().GetUser(badUserID).Return(model.User{AuthSecret: &model.AuthSecret{}}, nil)
	mockDB.EXPECT().GetUser(goodUserID).Return(model.User{}, nil)
	mockDB.EXPECT().GetSAMLProvider(samlProviderID).Return(model.SAMLProvider{}, nil).Times(2)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDB.EXPECT().DeleteAuthSecret(gomock.Any(), gomock.Any()).Return(nil)

	// Happy path
	test.Request(t).
		WithMethod(http.MethodPut).
		WithURL(fmt.Sprintf(updateUserPathFmt, goodUserID.String())).
		WithURLPathVars(map[string]string{
			"user_id": goodUserID.String(),
		}).
		WithBody(v2.UpdateUserRequest{
			Principal:      "tester",
			Roles:          goodRoles,
			SAMLProviderID: samlProviderIDStr,
		}).
		OnHandlerFunc(resources.UpdateUser).
		Require().
		ResponseStatusCode(http.StatusOK)

	// Negative path where a user already has an auth secret set
	test.Request(t).
		WithMethod(http.MethodPut).
		WithURL(fmt.Sprintf(updateUserPathFmt, badUserID.String())).
		WithURLPathVars(map[string]string{
			"user_id": badUserID.String(),
		}).
		WithBody(v2.UpdateUserRequest{
			Principal:      "tester",
			Roles:          goodRoles,
			SAMLProviderID: samlProviderIDStr,
		}).
		OnHandlerFunc(resources.UpdateUser).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestManagementResource_DeleteSAMLProvider(t *testing.T) {
	var (
		goodSAMLProvider = model.SAMLProvider{
			Serial: model.Serial{
				ID: 1,
			},
		}

		samlProviderWithUsers = model.SAMLProvider{
			Serial: model.Serial{
				ID: 2,
			},
		}

		samlEnabledUser = model.User{
			Unique: model.Unique{
				ID: must.NewUUIDv4(),
			},
		}

		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
	)

	defer mockCtrl.Finish()

	mockDB.EXPECT().GetSAMLProvider(goodSAMLProvider.ID).Return(goodSAMLProvider, nil)
	mockDB.EXPECT().GetSAMLProvider(samlProviderWithUsers.ID).Return(samlProviderWithUsers, nil)
	mockDB.EXPECT().DeleteSAMLProvider(gomock.Any(), gomock.Eq(goodSAMLProvider)).Return(nil)
	mockDB.EXPECT().DeleteSAMLProvider(gomock.Any(), gomock.Eq(samlProviderWithUsers)).Return(nil)
	mockDB.EXPECT().GetSAMLProviderUsers(goodSAMLProvider.ID).Return(nil, nil)
	mockDB.EXPECT().GetSAMLProviderUsers(samlProviderWithUsers.ID).Return(model.Users{samlEnabledUser}, nil)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(samlEnabledUser)).Return(nil)

	// Happy path
	test.Request(t).
		WithMethod(http.MethodDelete).
		WithURL(fmt.Sprintf(samlProviderPathFmt, goodSAMLProvider.ID)).
		WithURLPathVars(map[string]string{
			api.URIPathVariableSAMLProviderID: fmt.Sprintf("%d", goodSAMLProvider.ID),
		}).
		OnHandlerFunc(resources.DeleteSAMLProvider).
		Require().
		ResponseStatusCode(http.StatusOK)

	// Negative path where a provider has attached users
	test.Request(t).
		WithMethod(http.MethodDelete).
		WithURL(fmt.Sprintf(samlProviderPathFmt, samlProviderWithUsers.ID)).
		WithURLPathVars(map[string]string{
			api.URIPathVariableSAMLProviderID: fmt.Sprintf("%d", samlProviderWithUsers.ID),
		}).
		OnHandlerFunc(resources.DeleteSAMLProvider).
		Require().
		ResponseStatusCode(http.StatusOK)
}

func TestManagementResource_ListPermissions_SortingError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/permissions"
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAllPermissions("authority desc, name", model.SQLFilter{SQLString: "name = ?", Params: []any{"foo"}}).Return(model.Permissions{}, fmt.Errorf("foo"))

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllPermissions("authority desc, name", model.SQLFilter{SQLString: "name = ?", Params: []any{"a"}}).Return(model.Permissions{perm1, perm2}, nil)

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

func TestManagementResource_ListRoles_SortingError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/roles"
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAllRoles("description desc, name", model.SQLFilter{}).Return(model.Roles{}, fmt.Errorf("foo"))

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllRoles("description desc, name", model.SQLFilter{}).Return(model.Roles{role1, role2}, nil)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllRoles("", model.SQLFilter{SQLString: "name = ?", Params: []any{"a"}}).Return(model.Roles{role1}, nil)

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

func TestExpireUserAuthSecret_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/secret"

	badUserId := uuid.NullUUID{}
	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(badUserId.UUID).Return(model.User{}, fmt.Errorf("db failure"))

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(userId).Return(model.User{AuthSecret: &model.AuthSecret{}}, nil)
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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAllUsers("first_name desc, last_name", model.SQLFilter{}).Return(model.Users{}, fmt.Errorf("foo"))

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllUsers("first_name desc, last_name", model.SQLFilter{}).Return(model.Users{user1, user2}, nil)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllUsers("", model.SQLFilter{SQLString: "first_name = ?", Params: []any{"a"}}).Return(model.Users{user1}, nil)

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
		Roles:         model.Roles{},
		PrincipalName: "Bad User",
		FirstName:     null.StringFrom("bad"),
		LastName:      null.StringFrom("bad"),
		EmailAddress:  null.StringFrom("bad"),
		EULAAccepted:  true,
	}

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil).AnyTimes()
	mockDB.EXPECT().GetRoles(badRole).Return(model.Roles{}, fmt.Errorf("db error"))
	mockDB.EXPECT().GetRoles(gomock.Not(badRole)).Return(model.Roles{}, nil).AnyTimes()
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
				Errors:     []api.ErrorDetails{{Message: auth.ErrorResponseDetailsNumRoles}},
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
					EmailAddress: badUser.EmailAddress.String,
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

func TestCreateUser_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	goodUser := model.User{PrincipalName: "good user"}

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
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

	log.ConfigureDefaults()

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any()).Return(model.User{}, fmt.Errorf("foo"))

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, fmt.Errorf("foo"))

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

func TestManagementResource_UpdateUser_SelfDisable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users"
	// logged in user has ID 00000000-0000-0000-0000-000000000000
	// leaving ID blank here will make goodUser have the same ID, so this should fail
	goodUser := model.User{PrincipalName: "good user"}

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{model.Role{
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
		IsDisabled: true,
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{model.Role{
		Name:        "admin",
		Description: "admin",
		Permissions: model.Permissions{model.Permission{
			Authority: "admin",
			Name:      "admin",
			Serial:    model.Serial{},
		}},
		Serial: model.Serial{},
	}}, nil)
	mockDB.EXPECT().LookupActiveSessionsByUser(gomock.Any()).Return([]model.UserSession{}, fmt.Errorf("foo"))

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
		IsDisabled: true,
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{model.Role{
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

func TestManagementResource_DeleteUser_BadUserID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"
	userID := "badUserID"

	resources, _ := apitest.NewAuthManagementResource(mockCtrl)

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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(userID).Return(model.User{}, database.ErrNotFound)

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

func TestManagementResource_DeleteUser_GetUserError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(userID).Return(model.User{}, fmt.Errorf("foo"))

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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	user := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: userID,
		},
	}

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(userID).Return(user, nil)
	mockDB.EXPECT().DeleteUser(gomock.Any(), user).Return(fmt.Errorf("foo"))

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

func TestManagementResource_DeleteUser_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/bloodhound-users"

	userID, err := uuid.NewV4()
	require.Nil(t, err)

	user := model.User{
		PrincipalName: "good user",
		Unique: model.Unique{
			ID: userID,
		},
	}

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(userID).Return(user, nil)
	mockDB.EXPECT().DeleteUser(gomock.Any(), user).Return(nil)

	ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	require.Nil(t, err)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetConfigurationParameter(appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{
		Key: appcfg.PasswordExpirationWindow,
		Value: must.NewJSONBObject(appcfg.PasswordExpiration{
			Duration: appcfg.DefaultPasswordExpirationWindow,
		}),
	}, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{}, nil)
	mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(goodUser, nil).AnyTimes()
	mockDB.EXPECT().GetUser(gomock.Any()).Return(goodUser, nil)
	mockDB.EXPECT().GetRoles(gomock.Any()).Return(model.Roles{model.Role{
		Name:        "admin",
		Description: "admin",
		Permissions: model.Permissions{model.Permission{
			Authority: "admin",
			Name:      "admin",
			Serial:    model.Serial{},
		}},
		Serial: model.Serial{},
	}}, nil)
	mockDB.EXPECT().LookupActiveSessionsByUser(gomock.Any()).Return([]model.UserSession{}, nil)
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
		IsDisabled: true,
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

	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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
	mockDB := dbmocks.NewMockDatabase(mockCtrl)

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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

	mockDB := dbmocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAllAuthTokens("name, last_access desc", model.SQLFilter{SQLString: "user_id = ?", Params: []any{user.ID.String()}}).Return(model.AuthTokens{}, fmt.Errorf("foo"))

	config, err := config.NewDefaultConfiguration()
	require.Nilf(t, err, "Failed to create default configuration: %v", err)
	config.Crypto.Argon2.NumIterations = 1
	config.Crypto.Argon2.NumThreads = 1
	resources := auth.NewManagementResource(config, mockDB, authz.NewAuthorizer(mockDB))

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllAuthTokens("name, last_access desc", model.SQLFilter{}).Return(allAuthTokens, nil)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetAllAuthTokens("name, last_access desc", model.SQLFilter{SQLString: "user_id = ?", Params: []any{user.ID.String()}}).Return(user.AuthTokens, nil)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	// The filters are stored in a map before parsing, which means we don't know what order the resulted SQLFilter will be in.
	// Mock out both possibilities to catch both cases.
	mockDB.EXPECT().GetAllAuthTokens("", model.SQLFilter{SQLString: "name = ? AND user_id = ?", Params: []any{"a", user.ID.String()}}).AnyTimes().Return(model.AuthTokens{authToken1}, nil)
	mockDB.EXPECT().GetAllAuthTokens("", model.SQLFilter{SQLString: "user_id = ? AND name = ?", Params: []any{user.ID.String(), "a"}}).AnyTimes().Return(model.AuthTokens{authToken1}, nil)

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

func TestEnrollMFA(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa"
	missingUserId, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	userId, _ := uuid.NewV4()
	activatedId := test.NewUUIDv4(t)
	badPassId := test.NewUUIDv4(t)
	ssoId := test.NewUUIDv4(t)
	genTOTPFailId := test.NewUUIDv4(t)

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(missingUserId).Return(model.User{}, database.ErrNotFound)
	mockDB.EXPECT().GetUser(activatedId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: true}}, nil)
	mockDB.EXPECT().GetUser(badPassId).Return(model.User{AuthSecret: &model.AuthSecret{}}, nil)
	mockDB.EXPECT().GetUser(ssoId).Return(model.User{SAMLProviderID: null.Int32From(1)}, nil)
	mockDB.EXPECT().GetUser(genTOTPFailId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)

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
				Errors:     []api.ErrorDetails{{Message: api.ErrorContentTypeJson.Error()}},
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
			Input{activatedId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: auth.ErrorResponseDetailsMFAActivated}},
			},
		},
		{
			Input{badPassId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: auth.ErrorResponseDetailsInvalidCurrentPassword}},
			},
		},
		{
			Input{ssoId.String(), nil},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: "Invalid operation, user is SSO"}},
			},
		},
		{
			Input{genTOTPFailId.String(), auth.MFAEnrollmentRequest{"password"}},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
		// Note: a non-trivial refactor is required to make green path testing possible
	}
	for _, tc := range cases {
		ctx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{Host: &url.URL{}})
		if payload, err := json.Marshal(tc.Input.Body); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(endpoint, tc.Input.UserId), bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			test.TestV2HandlerFailure(t, []string{"POST"}, fmt.Sprintf(endpoint, "{user_id}"), resources.EnrollMFA, *req, tc.Expected)
		}
	}
}

func TestDisenrollMFA_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa"

	missingUserId := test.NewUUIDv4(t)
	userId := test.NewUUIDv4(t)

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(missingUserId).Return(model.User{}, database.ErrNotFound)

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
				Errors:     []api.ErrorDetails{{Message: api.ErrorContentTypeJson.Error()}},
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"
	userId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(userId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"
	nonAdminId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(nonAdminId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)
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

func TestDisenrollMFA_Admin_FailureIncorrectPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	endpoint := "/api/v2/auth/users/%s/mfa"
	nonAdminId := test.NewUUIDv4(t)

	mockDB.EXPECT().GetUser(nonAdminId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password"), Unique: model.Unique{ID: nonAdminId}}, nil).AnyTimes()

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
				Errors:     []api.ErrorDetails{{Message: auth.ErrorResponseDetailsInvalidCurrentPassword}},
			},
		)
	}
}

func TestGetMFAActivationStatus_Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/auth/users/%s/mfa-activation"

	missingId := test.NewUUIDv4(t)

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	mockDB.EXPECT().GetUser(missingId).Return(model.User{}, database.ErrNotFound)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(activatedId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: true, TOTPSecret: "imasharedsecret"}}, nil)
	mockDB.EXPECT().GetUser(pendingId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: false, TOTPSecret: "imasharedsecret"}}, nil)
	mockDB.EXPECT().GetUser(deactivatedId).Return(model.User{AuthSecret: &model.AuthSecret{TOTPActivated: false}}, nil)

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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

	mockDB.EXPECT().GetUser(missingUserId).Return(model.User{}, database.ErrNotFound)
	mockDB.EXPECT().GetUser(unenrolledId).Return(model.User{AuthSecret: defaultDigestAuthSecret(t, "password")}, nil)

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
				Errors:     []api.ErrorDetails{{Message: api.ErrorContentTypeJson.Error()}},
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
				Errors:     []api.ErrorDetails{{Message: auth.ErrorResponseDetailsMFAEnrollmentRequired}},
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

	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
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
	mockDB.EXPECT().GetUser(userId).Return(model.User{AuthSecret: defaultDigestAuthSecretWithTOTP(t, "password", totpSecret.Secret())}, nil)
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
