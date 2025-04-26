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
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/serde"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
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

func TestManagementResource_SAMLLoginRedirect(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		redirectURL  string
	}
	type testData struct {
		name            string
		ssoProviderSlug string
		setupMocks      func(*testing.T, *mocks.MockDatabase)
		setupContext    func(*testing.T) context.Context
		expected        expected
	}

	tt := []testData{
		{
			name:            "Error: SSO Provider not found - Internal Server Error",
			ssoProviderSlug: "nonexistent-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "nonexistent-provider").Return(model.SSOProvider{}, sql.ErrNoRows)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
			},
		},
		{
			name:            "Error: Database error - Internal Server Error",
			ssoProviderSlug: "okta",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "okta").Return(model.SSOProvider{}, errors.New("database error"))
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
			},
		},
		{
			name:            "Success: Redirect to SSO provider login URL - Found",
			ssoProviderSlug: "okta",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				ssoProvider := model.SSOProvider{
					Name: "Okta",
					Slug: "okta",
					Type: model.SessionAuthProviderSAML,
				}
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "okta").Return(ssoProvider, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusFound,
				redirectURL:  "https://example.com/api/v2/sso/okta/login",
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

			endpointURL := fmt.Sprintf("/api/v2/sso/%s/redirect", testCase.ssoProviderSlug)
			req, err := http.NewRequest("GET", endpointURL, nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			vars := map[string]string{
				api.URIPathVariableSSOProviderSlug: testCase.ssoProviderSlug,
			}
			req = mux.SetURLVars(req, vars)

			response := httptest.NewRecorder()
			resources.SAMLLoginRedirect(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Success: Redirect to SSO provider login URL - Found" {
				redirectURL := response.Header().Get("Location")
				assert.Equal(t, testCase.expected.redirectURL, redirectURL)
			} else {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.Contains(t, responseBodyWithDefaultTimestamp, `"http_status":500`)
			}
		})
	}
}

func TestManagementResource_SAMLCallbackRedirect(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		redirectURL  string
	}
	type testData struct {
		name            string
		ssoProviderSlug string
		setupMocks      func(*testing.T, *mocks.MockDatabase)
		setupContext    func(*testing.T) context.Context
		expected        expected
	}

	tt := []testData{
		{
			name:            "Error: SSO Provider not found - Internal Server Error",
			ssoProviderSlug: "nonexistent-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "nonexistent-provider").Return(model.SSOProvider{}, sql.ErrNoRows)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
			},
		},
		{
			name:            "Error: Database error - Internal Server Error",
			ssoProviderSlug: "okta",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "okta").Return(model.SSOProvider{}, errors.New("database error"))
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
			},
		},
		{
			name:            "Success: Redirect to SSO provider callback URL - Temporary Redirect",
			ssoProviderSlug: "okta",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				ssoProvider := model.SSOProvider{
					Name: "Okta",
					Slug: "okta",
					Type: model.SessionAuthProviderSAML,
				}
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "okta").Return(ssoProvider, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusTemporaryRedirect,
				redirectURL:  "https://example.com/api/v2/sso/okta/callback",
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

			endpointURL := fmt.Sprintf("/api/v2/sso/%s/redirect/callback", testCase.ssoProviderSlug)
			req, err := http.NewRequest("GET", endpointURL, nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			vars := map[string]string{
				api.URIPathVariableSSOProviderSlug: testCase.ssoProviderSlug,
			}
			req = mux.SetURLVars(req, vars)

			response := httptest.NewRecorder()
			resources.SAMLCallbackRedirect(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Success: Redirect to SSO provider callback URL - Temporary Redirect" {
				redirectURL := response.Header().Get("Location")
				assert.Equal(t, testCase.expected.redirectURL, redirectURL)
			} else {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.Contains(t, responseBodyWithDefaultTimestamp, `"http_status":500`)
			}
		})
	}
}

func TestManagementResource_ListSAMLSignOnEndpoints(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		responseBody string
	}
	type testData struct {
		name         string
		setupMocks   func(*testing.T, *mocks.MockDatabase)
		setupContext func(*testing.T) context.Context
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Database error - Internal Server Error",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(nil, errors.New("database error"))
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: No SAML providers - Empty list",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAllSAMLProviders(gomock.Any()).Return([]model.SAMLProvider{}, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{"data":{"endpoints":[]}}`,
			},
		},
		{
			name: "Success: Multiple SAML providers",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				oktaProvider := model.SAMLProvider{
					Name:            "Okta Provider",
					DisplayName:     "Okta SSO",
					IssuerURI:       "https://okta.com/issuer",
					SingleSignOnURI: "https://okta.com/sso",
					SSOProviderID:   null.Int32From(1),
				}

				azureProvider := model.SAMLProvider{
					Name:            "Azure Provider",
					DisplayName:     "Azure SSO",
					IssuerURI:       "https://azure.com/issuer",
					SingleSignOnURI: "https://azure.com/sso",
					SSOProviderID:   null.Int32From(2),
				}

				loginOktaURL, _ := url.Parse("https://example.com/Okta%20Provider/login")
				loginAzureURL, _ := url.Parse("https://example.com/Azure%20Provider/login")

				oktaProvider.ServiceProviderInitiationURI = serde.URL{URL: *loginOktaURL}
				azureProvider.ServiceProviderInitiationURI = serde.URL{URL: *loginAzureURL}

				samlProviders := []model.SAMLProvider{oktaProvider, azureProvider}
				mockDB.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(samlProviders, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{"data":{"endpoints":[{"name":"Okta Provider","initiation_url":"https://example.com/Okta%20Provider/login"},{"name":"Azure Provider","initiation_url":"https://example.com/Azure%20Provider/login"}]}}`,
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

			req, err := http.NewRequest("GET", "/api/v2/saml/signon-endpoints", nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			response := httptest.NewRecorder()
			resources.ListSAMLSignOnEndpoints(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Error: Database error - Internal Server Error" {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.JSONEq(t, testCase.expected.responseBody, responseBodyWithDefaultTimestamp)
			} else {
				assert.JSONEq(t, testCase.expected.responseBody, response.Body.String())
			}
		})
	}
}

func TestManagementResource_ListSAMLProviders(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		responseBody string
	}
	type testData struct {
		name         string
		setupMocks   func(*testing.T, *mocks.MockDatabase)
		setupContext func(*testing.T) context.Context
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Database error - Internal Server Error",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(nil, errors.New("database error"))
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: No SAML providers - Empty list",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAllSAMLProviders(gomock.Any()).Return([]model.SAMLProvider{}, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{"data":{"saml_providers":[]}}`,
			},
		},
		{
			name: "Success: Multiple SAML providers",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				oktaProvider := model.SAMLProvider{
					Name:                       "Okta Provider",
					DisplayName:                "Okta SSO",
					IssuerURI:                  "https://okta.com/issuer",
					SingleSignOnURI:            "https://okta.com/sso",
					SSOProviderID:              null.Int32From(1),
					RootURIVersion:             model.SAMLRootURIVersion1,
					PrincipalAttributeMappings: []string{"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"},
					Serial: model.Serial{
						ID: 1,
					},
				}

				azureProvider := model.SAMLProvider{
					Name:                       "Azure Provider",
					DisplayName:                "Azure SSO",
					IssuerURI:                  "https://azure.com/issuer",
					SingleSignOnURI:            "https://azure.com/sso",
					SSOProviderID:              null.Int32From(2),
					RootURIVersion:             model.SAMLRootURIVersion2,
					PrincipalAttributeMappings: []string{"urn:oid:0.9.2342.19200300.100.1.3"},
					Serial: model.Serial{
						ID: 2,
					},
				}

				oktaIssuerURL, _ := url.Parse("https://example.com/Okta%20Provider/issuer")
				oktaInitURL, _ := url.Parse("https://example.com/Okta%20Provider/login")
				oktaMetadataURL, _ := url.Parse("https://example.com/Okta%20Provider/metadata.xml")
				oktaACSURL, _ := url.Parse("https://example.com/Okta%20Provider/acs")

				azureIssuerURL, _ := url.Parse("https://example.com/Azure%20Provider/issuer")
				azureInitURL, _ := url.Parse("https://example.com/Azure%20Provider/login")
				azureMetadataURL, _ := url.Parse("https://example.com/Azure%20Provider/metadata.xml")
				azureACSURL, _ := url.Parse("https://example.com/Azure%20Provider/acs")

				oktaProvider.ServiceProviderIssuerURI = serde.URL{URL: *oktaIssuerURL}
				oktaProvider.ServiceProviderInitiationURI = serde.URL{URL: *oktaInitURL}
				oktaProvider.ServiceProviderMetadataURI = serde.URL{URL: *oktaMetadataURL}
				oktaProvider.ServiceProviderACSURI = serde.URL{URL: *oktaACSURL}

				azureProvider.ServiceProviderIssuerURI = serde.URL{URL: *azureIssuerURL}
				azureProvider.ServiceProviderInitiationURI = serde.URL{URL: *azureInitURL}
				azureProvider.ServiceProviderMetadataURI = serde.URL{URL: *azureMetadataURL}
				azureProvider.ServiceProviderACSURI = serde.URL{URL: *azureACSURL}

				samlProviders := []model.SAMLProvider{oktaProvider, azureProvider}
				mockDB.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(samlProviders, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{
					"data": {
						"saml_providers": [
							{
								"name": "Okta Provider",
								"display_name": "Okta SSO",
								"idp_issuer_uri": "https://okta.com/issuer",
								"idp_sso_uri": "https://okta.com/sso",
								"root_uri_version": 1,
								"principal_attribute_mappings": ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"],
								"sp_issuer_uri": "https://example.com/Okta%20Provider/issuer",
								"sp_sso_uri": "https://example.com/Okta%20Provider/login",
								"sp_metadata_uri": "https://example.com/Okta%20Provider/metadata.xml",
								"sp_acs_uri": "https://example.com/Okta%20Provider/acs",
								"sso_provider_id": 1,
								"id": 1,
								"created_at": "0001-01-01T00:00:00Z",
								"updated_at": "0001-01-01T00:00:00Z",
								"deleted_at": {
									"Time": "0001-01-01T00:00:00Z",
									"Valid": false
								}
							},
							{
								"name": "Azure Provider",
								"display_name": "Azure SSO",
								"idp_issuer_uri": "https://azure.com/issuer",
								"idp_sso_uri": "https://azure.com/sso",
								"root_uri_version": 2,
								"principal_attribute_mappings": ["urn:oid:0.9.2342.19200300.100.1.3"],
								"sp_issuer_uri": "https://example.com/Azure%20Provider/issuer",
								"sp_sso_uri": "https://example.com/Azure%20Provider/login",
								"sp_metadata_uri": "https://example.com/Azure%20Provider/metadata.xml",
								"sp_acs_uri": "https://example.com/Azure%20Provider/acs",
								"sso_provider_id": 2,
								"id": 2,
								"created_at": "0001-01-01T00:00:00Z",
								"updated_at": "0001-01-01T00:00:00Z",
								"deleted_at": {
									"Time": "0001-01-01T00:00:00Z",
									"Valid": false
								}
							}
						]
					}
				}`,
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

			req, err := http.NewRequest("GET", "/api/v2/saml/providers", nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			response := httptest.NewRecorder()
			resources.ListSAMLProviders(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Error: Database error - Internal Server Error" {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.JSONEq(t, testCase.expected.responseBody, responseBodyWithDefaultTimestamp)
			} else {
				assert.JSONEq(t, testCase.expected.responseBody, response.Body.String())
			}
		})
	}
}

func TestManagementResource_GetSAMLProvider(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		responseBody string
	}
	type testData struct {
		name         string
		providerID   string
		setupMocks   func(*testing.T, *mocks.MockDatabase)
		setupContext func(*testing.T) context.Context
		expected     expected
	}

	tt := []testData{
		{
			name:       "Error: Missing provider ID - Unauthorized",
			providerID: "",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusUnauthorized,
				responseBody: `{"http_status":401,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"authentication is invalid"}]}`,
			},
		},
		{
			name:       "Error: Invalid provider ID format - Not Found",
			providerID: "invalid",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusNotFound,
				responseBody: `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name:       "Error: Provider not found - Internal Server Error",
			providerID: "1",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSAMLProvider(gomock.Any(), int32(1)).Return(model.SAMLProvider{}, sql.ErrNoRows)
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name:       "Error: Database error - Internal Server Error",
			providerID: "1",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSAMLProvider(gomock.Any(), int32(1)).Return(model.SAMLProvider{}, errors.New("database error"))
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name:       "Success: Provider found - OK",
			providerID: "1",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				oktaProvider := model.SAMLProvider{
					Name:                       "Okta Provider",
					DisplayName:                "Okta SSO",
					IssuerURI:                  "https://okta.com/issuer",
					SingleSignOnURI:            "https://okta.com/sso",
					SSOProviderID:              null.Int32From(1),
					RootURIVersion:             model.SAMLRootURIVersion1,
					PrincipalAttributeMappings: []string{"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"},
					Serial: model.Serial{
						ID: 1,
					},
				}

				mockDB.EXPECT().GetSAMLProvider(gomock.Any(), int32(1)).Return(oktaProvider, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{
					"data": {
						"name": "Okta Provider",
						"display_name": "Okta SSO",
						"idp_issuer_uri": "https://okta.com/issuer",
						"idp_sso_uri": "https://okta.com/sso",
						"root_uri_version": 1,
						"principal_attribute_mappings": ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"],
						"sp_issuer_uri": "",
						"sp_sso_uri": "",
						"sp_metadata_uri": "",
						"sp_acs_uri": "",
						"sso_provider_id": 1,
						"id": 1,
						"created_at": "0001-01-01T00:00:00Z",
						"updated_at": "0001-01-01T00:00:00Z",
						"deleted_at": {
							"Time": "0001-01-01T00:00:00Z",
							"Valid": false
						}
					}
				}`,
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

			endpointURL := fmt.Sprintf("/api/v2/saml/providers/%s", testCase.providerID)
			req, err := http.NewRequest("GET", endpointURL, nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			if testCase.providerID != "" {
				vars := map[string]string{
					api.URIPathVariableSAMLProviderID: testCase.providerID,
				}
				req = mux.SetURLVars(req, vars)
			}

			response := httptest.NewRecorder()
			resources.GetSAMLProvider(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Success: Provider found - OK" {
				assert.JSONEq(t, testCase.expected.responseBody, response.Body.String())
			} else {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.JSONEq(t, testCase.expected.responseBody, responseBodyWithDefaultTimestamp)
			}
		})
	}
}

func TestManagementResource_CreateSAMLProviderMultipart(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		responseBody string
	}
	type testData struct {
		name       string
		setupForm  func() (*bytes.Buffer, *multipart.Writer)
		setupMocks func(*testing.T, *mocks.MockDatabase)
		expected   expected
	}

	validMetadataXML := `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://okta.com/saml">
		<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
			<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://okta.com/sso"/>
		</IDPSSODescriptor>
	</EntityDescriptor>`

	tt := []testData{
		{
			name: "Error: ParseMultipartForm fails",
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"form is missing \"name\" parameter"}]}`,
			},
		},
		{
			name: "Error: Missing provider name",
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatal(err)
				}
				metadataFile.Write([]byte(validMetadataXML))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()
			},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"form is missing \"name\" parameter"}]}`,
			},
		},
		{
			name: "Error: Missing metadata file",
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("Okta Provider"))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()
			},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"form is missing \"metadata\" parameter"}]}`,
			},
		},
		{
			name: "Error: Duplicate provider name",
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("Existing Provider"))

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatal(err)
				}
				metadataFile.Write([]byte(validMetadataXML))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				config := model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 1,
						RoleProvision: false,
					},
				}

				mockDB.EXPECT().
					CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Eq(config)).
					Return(model.SAMLProvider{}, database.ErrDuplicateSSOProviderName)
			},
			expected: expected{
				responseCode: http.StatusConflict,
				responseBody: `{
            "http_status":409,
            "timestamp":"0001-01-01T00:00:00Z",
            "request_id":"",
            "errors":[
                { "context":"", "message":"sso provider name must be unique" }
            ]
        }`,
			},
		},
		{
			name: "Error: Database error",
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("New Provider"))

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatal(err)
				}
				metadataFile.Write([]byte(validMetadataXML))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				config := model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 1,
						RoleProvision: false,
					},
				}

				mockDB.EXPECT().
					CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Eq(config)).
					Return(model.SAMLProvider{}, errors.New("database error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Valid provider created",
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("New Provider"))

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatal(err)
				}
				metadataFile.Write([]byte(validMetadataXML))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				config := model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 1,
						RoleProvision: false,
					},
				}

				newProvider := model.SAMLProvider{
					Name:            "New Provider",
					DisplayName:     "New Provider",
					IssuerURI:       "https://okta.com/saml",
					SingleSignOnURI: "https://okta.com/sso",
					MetadataXML:     []byte(validMetadataXML),
					Serial: model.Serial{
						ID: 1,
					},
				}

				mockDB.EXPECT().
					CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Eq(config)).
					Return(newProvider, nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{
        "data": {
            "name": "New Provider",
            "display_name": "New Provider",
            "idp_issuer_uri": "https://okta.com/saml",
            "idp_sso_uri": "https://okta.com/sso",
            "root_uri_version": 0,
            "principal_attribute_mappings": null,
            "sp_issuer_uri": "",
            "sp_sso_uri": "",
            "sp_metadata_uri": "",
            "sp_acs_uri": "",
            "sso_provider_id": null,
            "id": 1,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "deleted_at": {
                "Time": "0001-01-01T00:00:00Z",
                "Valid": false
            }
        }
    }`,
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

			body, writer := testCase.setupForm()
			req, err := http.NewRequest("POST", "/api/v2/saml/providers", body)
			require.NoError(t, err)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			response := httptest.NewRecorder()
			resources.CreateSAMLProviderMultipart(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Success: Valid provider created" {
				assert.JSONEq(t, testCase.expected.responseBody, response.Body.String())
			} else {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.JSONEq(t, testCase.expected.responseBody, responseBodyWithDefaultTimestamp)
			}
		})
	}
}

func TestManagementResource_UpdateSAMLProviderRequest(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		responseBody string
	}
	type testData struct {
		name        string
		ssoProvider model.SSOProvider
		setupForm   func() (*bytes.Buffer, *multipart.Writer)
		setupMocks  func(*testing.T, *mocks.MockDatabase)
		expected    expected
	}

	validMetadataXML := `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://okta.com/saml">
		<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
			<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://okta.com/sso"/>
		</IDPSSODescriptor>
	</EntityDescriptor>`

	tt := []testData{
		{
			name: "Error: SAMLProvider is nil",
			ssoProvider: model.SSOProvider{
				SAMLProvider: nil,
			},
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode: http.StatusNotFound,
				responseBody: `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name: "Error: ParseMultipartForm fails",
			ssoProvider: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Name:        "Existing Provider",
					DisplayName: "Existing Provider",
				},
			},
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				return body, nil
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode: http.StatusBadRequest,
				responseBody: `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"request Content-Type isn't multipart/form-data"}]}`,
			},
		},
		{
			name: "Error: Duplicate provider name",
			ssoProvider: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Name:        "Existing Provider",
					DisplayName: "Existing Provider",
				},
			},
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("Updated Provider"))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				mockDB.EXPECT().
					UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).
					Return(model.SAMLProvider{}, database.ErrDuplicateSSOProviderName)
			},
			expected: expected{
				responseCode: http.StatusConflict,
				responseBody: `{"http_status":409,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"sso provider name must be unique"}]}`,
			},
		},
		{
			name: "Error: Database error",
			ssoProvider: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Name:        "Existing Provider",
					DisplayName: "Existing Provider",
				},
			},
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("Updated Provider"))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				mockDB.EXPECT().
					UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).
					Return(model.SAMLProvider{}, errors.New("database error"))
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Provider updated with new name",
			ssoProvider: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Name:        "Existing Provider",
					DisplayName: "Existing Provider",
					Serial: model.Serial{
						ID: 1,
					},
				},
			},
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("Updated Provider"))

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatal(err)
				}
				metadataFile.Write([]byte(validMetadataXML))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				updatedProvider := model.SAMLProvider{
					Name:            "Updated Provider",
					DisplayName:     "Updated Provider",
					IssuerURI:       "https://okta.com/saml",
					SingleSignOnURI: "https://okta.com/sso",
					MetadataXML:     []byte(validMetadataXML),
					Serial: model.Serial{
						ID: 1,
					},
				}

				mockDB.EXPECT().
					UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).
					Return(updatedProvider, nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{
					"data": {
						"name": "Updated Provider",
						"display_name": "Updated Provider",
						"idp_issuer_uri": "https://okta.com/saml",
						"idp_sso_uri": "https://okta.com/sso",
						"root_uri_version": 0,
						"principal_attribute_mappings": null,
						"sp_issuer_uri": "",
						"sp_sso_uri": "",
						"sp_metadata_uri": "",
						"sp_acs_uri": "",
						"sso_provider_id": null,
						"id": 1,
						"created_at": "0001-01-01T00:00:00Z",
						"updated_at": "0001-01-01T00:00:00Z",
						"deleted_at": {
							"Time": "0001-01-01T00:00:00Z",
							"Valid": false
						}
					}
				}`,
			},
		},
		{
			name: "Success: Provider updated without metadata",
			ssoProvider: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Name:        "Existing Provider",
					DisplayName: "Existing Provider",
					IssuerURI:   "https://existing.com/saml",
					Serial: model.Serial{
						ID: 1,
					},
				},
			},
			setupForm: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				nameField, err := writer.CreateFormField("name")
				if err != nil {
					t.Fatal(err)
				}
				nameField.Write([]byte("Updated Provider"))

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatal(err)
				}
				autoProvisionField.Write([]byte("true"))

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatal(err)
				}
				roleIDField.Write([]byte("1"))

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatal(err)
				}
				roleProvisionField.Write([]byte("false"))

				writer.Close()
				return body, writer
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{
					Name: "Admin",
					Serial: model.Serial{
						ID: 1,
					},
				}, nil).AnyTimes()

				updatedProvider := model.SAMLProvider{
					Name:        "Updated Provider",
					DisplayName: "Updated Provider",
					IssuerURI:   "https://existing.com/saml",
					Serial: model.Serial{
						ID: 1,
					},
				}

				mockDB.EXPECT().
					UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).
					Return(updatedProvider, nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{
					"data": {
						"name": "Updated Provider",
						"display_name": "Updated Provider",
						"idp_issuer_uri": "https://existing.com/saml",
						"idp_sso_uri": "",
						"root_uri_version": 0,
						"principal_attribute_mappings": null,
						"sp_issuer_uri": "",
						"sp_sso_uri": "",
						"sp_metadata_uri": "",
						"sp_acs_uri": "",
						"sso_provider_id": null,
						"id": 1,
						"created_at": "0001-01-01T00:00:00Z",
						"updated_at": "0001-01-01T00:00:00Z",
						"deleted_at": {
							"Time": "0001-01-01T00:00:00Z",
							"Valid": false
						}
					}
				}`,
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

			body, writer := testCase.setupForm()
			req, err := http.NewRequest("PUT", "/api/v2/saml/providers/1", body)
			require.NoError(t, err)

			if writer != nil {
				req.Header.Set("Content-Type", writer.FormDataContentType())
			}

			response := httptest.NewRecorder()
			resources.UpdateSAMLProviderRequest(response, req, testCase.ssoProvider)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.name == "Success: Provider updated with new name" || testCase.name == "Success: Provider updated without metadata" {
				assert.JSONEq(t, testCase.expected.responseBody, response.Body.String())
			} else {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.JSONEq(t, testCase.expected.responseBody, responseBodyWithDefaultTimestamp)
			}
		})
	}
}
