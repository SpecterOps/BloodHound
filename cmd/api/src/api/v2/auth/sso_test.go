// Copyright 2024 Specter Ops, Inc.
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
	"context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/database/mocks"

	"github.com/pkg/errors"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/api/v2/auth"
	bhceAuth "github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_ListAuthProviders(t *testing.T) {
	const endpoint = "/api/v2/sso-providers"

	var (
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
		reqCtx            = &ctx.Context{Host: &url.URL{}}

		oidcProvider = model.OIDCProvider{
			SSOProviderID: 1,
			ClientID:      "client-id-1",
			Issuer:        "https://issuer1.com",
		}
		samlProvider = model.SAMLProvider{
			Serial:        model.Serial{ID: 2},
			Name:          "SAML Provider 1",
			DisplayName:   "SAML Provider One",
			IssuerURI:     "https://saml-issuer1.com",
			SSOProviderID: null.Int32From(2),
		}
		ssoProviders = []model.SSOProvider{
			{
				Serial:       model.Serial{ID: 1},
				Name:         "OIDC Provider 1",
				Slug:         "oidc-provider-1",
				Type:         model.SessionAuthProviderOIDC,
				OIDCProvider: &oidcProvider,
				Config: model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 3,
						RoleProvision: true,
					},
				},
			},
			{
				Serial:       model.Serial{ID: 2},
				Name:         "SAML Provider 1",
				Slug:         "saml-provider-1",
				Type:         model.SessionAuthProviderSAML,
				SAMLProvider: &samlProvider,
				Config: model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 2,
						RoleProvision: false,
					},
				},
			},
		}
	)
	defer mockCtrl.Finish()

	t.Run("successfully list auth providers without query parameters", func(t *testing.T) {
		// default ordering and no filters
		mockDB.EXPECT().GetAllSSOProviders(gomock.Any(), "created_at", model.SQLFilter{}).Return(ssoProviders, nil)

		test.Request(t).
			WithMethod(http.MethodGet).
			WithContext(reqCtx).
			WithURL(endpoint).
			OnHandlerFunc(resources.ListAuthProviders).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("successfully list auth providers with sorting", func(t *testing.T) {
		// sorting by name descending
		mockDB.EXPECT().GetAllSSOProviders(gomock.Any(), "name desc", model.SQLFilter{SQLString: "", Params: nil}).Return(ssoProviders, nil)
		const reqUrl = endpoint + "?sort_by=-name"

		test.Request(t).
			WithMethod(http.MethodGet).
			WithContext(reqCtx).
			WithURL(reqUrl).
			OnHandlerFunc(resources.ListAuthProviders).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("successfully list auth providers with filtering", func(t *testing.T) {
		// filtering by name
		mockDB.EXPECT().GetAllSSOProviders(gomock.Any(), "created_at", model.SQLFilter{
			SQLString: "name = ?",
			Params:    []interface{}{"OIDC Provider 1"},
		}).Return([]model.SSOProvider{ssoProviders[0]}, nil)
		const reqUrl = endpoint + "?name=eq:OIDC Provider 1"

		test.Request(t).
			WithMethod(http.MethodGet).
			WithContext(reqCtx).
			WithURL(reqUrl).
			OnHandlerFunc(resources.ListAuthProviders).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("fail to list auth providers with invalid sort field", func(t *testing.T) {
		const reqUrl = endpoint + "?sort_by=invalid_field"

		test.Request(t).
			WithMethod(http.MethodGet).
			WithContext(reqCtx).
			WithURL(reqUrl).
			OnHandlerFunc(resources.ListAuthProviders).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("fail to list auth providers with invalid filter predicate", func(t *testing.T) {
		const reqUrl = endpoint + "?name=invalid_predicate:Provider"

		test.Request(t).
			WithMethod(http.MethodGet).
			WithContext(reqCtx).
			WithURL(reqUrl).
			OnHandlerFunc(resources.ListAuthProviders).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})
}

func TestManagementResource_DeleteOIDCProvider(t *testing.T) {
	var (
		ssoDeleteURL      = "/api/v2/sso-providers/%s"
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
	)

	t.Run("successfully delete an SSOProvider", func(t *testing.T) {
		expectedUser := model.User{
			PrincipalName: "tester",
			SSOProviderID: null.Int32From(1),
		}

		mockDB.EXPECT().DeleteSSOProvider(gomock.Any(), 1).Return(nil)
		mockDB.EXPECT().GetSSOProviderUsers(gomock.Any(), 1).Return(model.Users{
			expectedUser,
		}, nil)

		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(ssoDeleteURL, api.URIPathVariableSSOProviderID).
			WithURLPathVars(map[string]string{api.URIPathVariableSSOProviderID: "1"}).
			OnHandlerFunc(resources.DeleteSSOProvider).
			Require().
			ResponseStatusCode(http.StatusOK).
			ResponseJSONBody(auth.DeleteSSOProviderResponse{AffectedUsers: model.Users{
				expectedUser,
			}})
	})

	t.Run("error invalid sso_provider_id format", func(t *testing.T) {
		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(ssoDeleteURL, api.URIPathVariableSSOProviderID).
			WithURLPathVars(map[string]string{api.URIPathVariableSSOProviderID: "bloodhound"}).
			OnHandlerFunc(resources.DeleteSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error user cannot delete their own SSO provider", func(t *testing.T) {
		test.Request(t).
			WithMethod(http.MethodDelete).
			WithContext(&ctx.Context{AuthCtx: bhceAuth.Context{
				Owner: model.User{SSOProviderID: null.Int32From(1)},
			}}).
			WithURL(ssoDeleteURL, api.URIPathVariableSSOProviderID).
			WithURLPathVars(map[string]string{api.URIPathVariableSSOProviderID: "1"}).
			OnHandlerFunc(resources.DeleteSSOProvider).
			Require().
			ResponseStatusCode(http.StatusConflict)
	})

	t.Run("error when retrieving users of the sso provider", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderUsers(gomock.Any(), 1).Return(model.Users{}, errors.New("an error"))

		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(ssoDeleteURL, api.URIPathVariableSSOProviderID).
			WithURLPathVars(map[string]string{api.URIPathVariableSSOProviderID: "1"}).
			OnHandlerFunc(resources.DeleteSSOProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})

	t.Run("error when deleting sso providers", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderUsers(gomock.Any(), 1).Return(model.Users{}, nil)
		mockDB.EXPECT().DeleteSSOProvider(gomock.Any(), 1).Return(errors.New("an error"))

		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(ssoDeleteURL, api.URIPathVariableSSOProviderID).
			WithURLPathVars(map[string]string{api.URIPathVariableSSOProviderID: "1"}).
			OnHandlerFunc(resources.DeleteSSOProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})

	t.Run("error could not find sso_provider by id", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderUsers(gomock.Any(), 1).Return(model.Users{}, nil)
		mockDB.EXPECT().DeleteSSOProvider(gomock.Any(), 1).Return(database.ErrNotFound)

		test.Request(t).
			WithMethod(http.MethodDelete).
			WithURL(ssoDeleteURL, api.URIPathVariableSSOProviderID).
			WithURLPathVars(map[string]string{api.URIPathVariableSSOProviderID: "1"}).
			OnHandlerFunc(resources.DeleteSSOProvider).
			Require().
			ResponseStatusCode(http.StatusNotFound)
	})
}

func TestManagementResource_SanitizeAndGetRoles(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		_, mockDB = apitest.NewAuthManagementResource(mockCtrl)
		testCtx   = context.Background()

		dbRoles = model.Roles{
			{Name: "God Role", Serial: model.Serial{ID: 1}},
			{Name: "Default Role", Serial: model.Serial{ID: 2}},
			{Name: "Valid Role", Serial: model.Serial{ID: 3}},
		}
		roleProvisionEnabledConfig  = model.SSOProviderAutoProvisionConfig{RoleProvision: true, DefaultRoleId: 2, Enabled: true}
		roleProvisionDisabledConfig = model.SSOProviderAutoProvisionConfig{RoleProvision: false, DefaultRoleId: 2, Enabled: true}
	)
	t.Run("role provision enabled - return valid role", func(t *testing.T) {
		mockDB.EXPECT().GetAllRoles(gomock.Any(), "", model.SQLFilter{}).Return(dbRoles, nil)
		roles, err := auth.SanitizeAndGetRoles(testCtx, roleProvisionEnabledConfig, []string{"ignored", "bh-valid-role"}, mockDB)
		require.Nil(t, err)
		require.Len(t, roles, 1)
		require.Equal(t, roles[0].ID, dbRoles[2].ID)
	})

	t.Run("role provision enabled - return default role when multiple valid roles", func(t *testing.T) {
		mockDB.EXPECT().GetAllRoles(gomock.Any(), "", model.SQLFilter{}).Return(dbRoles, nil)
		roles, err := auth.SanitizeAndGetRoles(testCtx, roleProvisionEnabledConfig, []string{"bh-valid-role", "ignored", "bh-god-role"}, mockDB)
		require.Nil(t, err)
		require.Len(t, roles, 1)
		require.Equal(t, roles[0].ID, roleProvisionEnabledConfig.DefaultRoleId)
	})

	t.Run("role provision enabled - return default role when no valid roles", func(t *testing.T) {
		mockDB.EXPECT().GetAllRoles(gomock.Any(), "", model.SQLFilter{}).Return(dbRoles, nil)
		roles, err := auth.SanitizeAndGetRoles(testCtx, roleProvisionEnabledConfig, []string{"bh-invalid-role", "ignored"}, mockDB)
		require.Nil(t, err)
		require.Len(t, roles, 1)
		require.Equal(t, roles[0].ID, roleProvisionEnabledConfig.DefaultRoleId)
	})

	t.Run("role provision disabled - return default role", func(t *testing.T) {
		mockDB.EXPECT().GetAllRoles(gomock.Any(), "", model.SQLFilter{}).Return(dbRoles, nil)
		roles, err := auth.SanitizeAndGetRoles(testCtx, roleProvisionDisabledConfig, []string{"bh-valid-role", "ignored", "bh-god-role"}, mockDB)
		require.Nil(t, err)
		require.Len(t, roles, 1)
		require.Equal(t, roles[0].ID, roleProvisionEnabledConfig.DefaultRoleId)
	})
}

func TestManagementResource_SSOLoginHandler(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
	}
	type testData struct {
		name            string
		ssoProviderSlug string
		setupMocks      func(*testing.T, *mocks.MockDatabase)
		expected        expected
	}

	tt := []testData{
		{
			name:            "Error: SSO Provider not found - Not Found",
			ssoProviderSlug: "non-existent-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "non-existent-provider").Return(model.SSOProvider{}, database.ErrNotFound)
			},
			expected: expected{
				responseCode: http.StatusNotFound,
			},
		},
		{
			name:            "Error: Database error - Internal Server Error",
			ssoProviderSlug: "provider-db-error",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider-db-error").Return(model.SSOProvider{}, errors.New("database error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
			},
		},
		{
			name:            "Error: Unsupported provider type - Not Implemented",
			ssoProviderSlug: "unsupported-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "unsupported-provider").Return(model.SSOProvider{
					Type: 999,
					Name: "Unsupported Provider",
					Slug: "unsupported-provider",
				}, nil)
			},
			expected: expected{
				responseCode: http.StatusNotImplemented,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

			testCase.setupMocks(t, mockDB)

			req, err := http.NewRequest("GET", "/api/v2/auth/sso/"+testCase.ssoProviderSlug, nil)
			require.NoError(t, err)

			vars := map[string]string{
				api.URIPathVariableSSOProviderSlug: testCase.ssoProviderSlug,
			}
			req = mux.SetURLVars(req, vars)

			response := httptest.NewRecorder()
			resources.SSOLoginHandler(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)
		})
	}
}

func TestManagementResource_SSOCallbackHandler(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
	}
	type testData struct {
		name            string
		ssoProviderSlug string
		setupMocks      func(*testing.T, *mocks.MockDatabase)
		expected        expected
	}

	tt := []testData{
		{
			name:            "Error: SSO Provider not found - Not Found",
			ssoProviderSlug: "non-existent-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "non-existent-provider").Return(model.SSOProvider{}, database.ErrNotFound)
			},
			expected: expected{
				responseCode: http.StatusNotFound,
			},
		},
		{
			name:            "Error: Database error - Internal Server Error",
			ssoProviderSlug: "provider-db-error",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider-db-error").Return(model.SSOProvider{}, errors.New("database error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
			},
		},
		{
			name:            "Error: Unsupported provider type - Not Implemented",
			ssoProviderSlug: "unsupported-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "unsupported-provider").Return(model.SSOProvider{
					Type: 999,
					Name: "Unsupported Provider",
					Slug: "unsupported-provider",
				}, nil)
			},
			expected: expected{
				responseCode: http.StatusNotImplemented,
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

			testCase.setupMocks(t, mockDB)

			req, err := http.NewRequest("POST", "/api/v2/auth/sso/"+testCase.ssoProviderSlug+"/callback", nil)
			require.NoError(t, err)

			vars := map[string]string{
				api.URIPathVariableSSOProviderSlug: testCase.ssoProviderSlug,
			}
			req = mux.SetURLVars(req, vars)

			response := httptest.NewRecorder()
			resources.SSOCallbackHandler(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)
		})
	}
}
