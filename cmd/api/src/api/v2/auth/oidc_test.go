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
	"fmt"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/api/v2/auth"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_CreateOIDCProvider(t *testing.T) {
	var (
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
	)
	defer mockCtrl.Finish()

	t.Run("successfully create a new OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(0)).Return(model.Role{}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "Bloodhound gang", "https://localhost/auth", "bloodhound", model.SSOProviderConfig{}).Return(model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/auth",
		}, nil)

		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Bloodhound gang",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
				Config:   &model.SSOProviderConfig{},
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusCreated)
	})

	t.Run("successfully create a new OIDCProvider with config values", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		mockDB.EXPECT().GetRole(gomock.Any(), int32(3)).Return(model.Role{Serial: model.Serial{ID: 3}}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "Bloodhound gang2", "https://localhost/auth", "bloodhound", config).Return(model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/auth",
		}, nil)

		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Bloodhound gang2",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
				Config:   &config,
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusCreated)
	})

	t.Run("error invalid role id", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(7)).Return(model.Role{Serial: model.Serial{ID: 7}}, fmt.Errorf("role id is invalid"))

		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Gotham Net 2",
				Issuer:   "https://gotham-2.net",
				ClientID: "gotham-net-2",
				Config: &model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 7,
						RoleProvision: true,
					},
				},
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error parsing body request", func(t *testing.T) {
		test.Request(t).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error validating request field", func(t *testing.T) {
		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "1234:not:a:url",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error invalid Issuer", func(t *testing.T) {
		request := auth.UpsertOIDCProviderRequest{
			Issuer: "12345:bloodhound",
		}
		test.Request(t).
			WithBody(request).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(0)).Return(model.Role{}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "test", "https://localhost/auth", "bloodhound", model.SSOProviderConfig{}).Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
				Config:   &model.SSOProviderConfig{},
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}

func TestManagementResource_UpdateOIDCProvider(t *testing.T) {
	var (
		mockCtrl          = gomock.NewController(t)
		resources, mockDB = apitest.NewAuthManagementResource(mockCtrl)
		baseProvider      = model.SSOProvider{
			Type: model.SessionAuthProviderOIDC,
			Name: "Gotham Net",
			OIDCProvider: &model.OIDCProvider{
				ClientID: "gotham-net",
				Issuer:   "https://gotham.net",
			},
		}
		urlParams = map[string]string{"sso_provider_id": "1"}
	)
	defer mockCtrl.Finish()

	t.Run("successfully update an OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().UpdateOIDCProvider(gomock.Any(), gomock.Any())

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Gotham Net 2",
				Issuer:   "https://gotham-2.net",
				ClientID: "gotham-net-2",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("successfully update an OIDCProvider with config values", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().GetRole(gomock.Any(), int32(3)).Return(model.Role{Serial: model.Serial{ID: 3}}, nil)
		mockDB.EXPECT().UpdateOIDCProvider(gomock.Any(), gomock.Any())

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Gotham Net 2",
				Issuer:   "https://gotham-2.net",
				ClientID: "gotham-net-2",
				Config: &model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 3,
						RoleProvision: true,
					},
				},
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("error invalid role id", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().GetRole(gomock.Any(), int32(7)).Return(model.Role{Serial: model.Serial{ID: 7}}, fmt.Errorf("role id is invalid"))

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "Gotham Net 2",
				Issuer:   "https://gotham-2.net",
				ClientID: "gotham-net-2",
				Config: &model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       true,
						DefaultRoleId: 7,
						RoleProvision: true,
					},
				},
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error not found while updating an unknown OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{}, database.ErrNotFound)

		test.Request(t).
			WithURLPathVars(urlParams).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusNotFound)
	})

	t.Run("error parsing body request", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)

		test.Request(t).
			WithURLPathVars(urlParams).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error validating request field", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "1234:not:a:url",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().UpdateOIDCProvider(gomock.Any(), gomock.Any()).Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/auth",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}

func TestManagementResource_OIDCLoginHandler(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode    int
		locationHeader  bool
		redirectToLogin bool
		errorMessage    string
	}
	type testData struct {
		name        string
		ssoProvider model.SSOProvider
		setupMocks  func(*testing.T, *mocks.MockDatabase)
		expected    expected
		setupHost   func(*testing.T) *url.URL
	}

	mockHost := func(t *testing.T) *url.URL {
		host, err := url.Parse("https://example.com")
		require.NoError(t, err)
		return host
	}

	testProvider := func(includeOIDC bool) model.SSOProvider {
		provider := model.SSOProvider{
			Name: "Test Provider",
			Type: model.SessionAuthProviderOIDC,
			Slug: "test-provider",
			Config: model.SSOProviderConfig{
				AutoProvision: model.SSOProviderAutoProvisionConfig{
					Enabled:       true,
					RoleProvision: true,
					DefaultRoleId: 1,
				},
			},
			Serial: model.Serial{
				ID: 1,
			},
		}

		if includeOIDC {
			provider.OIDCProvider = &model.OIDCProvider{
				Issuer:   "https://test-issuer.com",
				ClientID: "test-client-id",
			}
		}

		return provider
	}

	tt := []testData{
		{
			name:        "Error: No OIDC Provider - Redirect to Login",
			ssoProvider: testProvider(false),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Your SSO connection failed due to misconfiguration, please contact your Administrator",
			},
			setupHost: mockHost,
		},
		{
			name:        "Error: OIDC Provider Creation Fails - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Your SSO connection failed due to misconfiguration, please contact your Administrator",
			},
			setupHost: mockHost,
		},
		{
			name:        "Success: OIDC Login - Redirect to Provider",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				locationHeader:  true,
				redirectToLogin: false,
			},
			setupHost: mockHost,
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.name == "Success: OIDC Login - Redirect to Provider" {
				t.Skip("Skipping test that requires external OIDC provider connectivity")
			}

			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)

			testCase.setupMocks(t, mockDB)

			host := testCase.setupHost(t)

			bhContext := &ctx.Context{
				Host: host,
			}

			goContext := ctx.Set(context.Background(), bhContext)

			req, err := http.NewRequest("GET", "/api/v2/auth/sso/oidc/login", nil)
			require.NoError(t, err)
			req = req.WithContext(goContext)

			response := httptest.NewRecorder()
			resources.OIDCLoginHandler(response, req, testCase.ssoProvider)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.expected.redirectToLogin {
				location := response.Header().Get("Location")
				require.NotEmpty(t, location)

				assert.Contains(t, location, "/ui/login")
				assert.Contains(t, location, url.QueryEscape(testCase.expected.errorMessage))
			}

			if testCase.expected.locationHeader {
				location := response.Header().Get("Location")
				require.NotEmpty(t, location)

				if !testCase.expected.redirectToLogin {
					assert.Contains(t, location, "response_type=code")
					assert.Contains(t, location, "client_id=test-client-id")
					assert.Contains(t, location, "state=")
					assert.Contains(t, location, "code_challenge=")
				}
			}

			if testCase.name == "Success: OIDC Login - Redirect to Provider" {
				cookies := response.Result().Cookies()
				var hasPKCECookie, hasStateCookie bool

				for _, cookie := range cookies {
					if cookie.Name == api.AuthPKCECookieName {
						hasPKCECookie = true
					}
					if cookie.Name == api.AuthStateCookieName {
						hasStateCookie = true
					}
				}

				assert.True(t, hasPKCECookie, "Should have PKCE cookie")
				assert.True(t, hasStateCookie, "Should have State cookie")
			}
		})
	}
}

func TestManagementResource_OIDCCallbackHandler(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode    int
		redirectToLogin bool
		errorMessage    string
	}
	type testData struct {
		name         string
		ssoProvider  model.SSOProvider
		setupMocks   func(*testing.T, *mocks.MockDatabase)
		expected     expected
		setupHost    func(*testing.T) *url.URL
		setupCookies func(*testing.T, *http.Request)
		formParams   map[string]string
	}

	mockHost := func(t *testing.T) *url.URL {
		host, err := url.Parse("https://example.com")
		require.NoError(t, err)
		return host
	}

	testProvider := func(includeOIDC bool) model.SSOProvider {
		provider := model.SSOProvider{
			Name: "Test Provider",
			Type: model.SessionAuthProviderOIDC,
			Slug: "test-provider",
			Config: model.SSOProviderConfig{
				AutoProvision: model.SSOProviderAutoProvisionConfig{
					Enabled:       true,
					RoleProvision: true,
					DefaultRoleId: 1,
				},
			},
			Serial: model.Serial{
				ID: 1,
			},
		}

		if includeOIDC {
			provider.OIDCProvider = &model.OIDCProvider{
				Issuer:   "https://test-issuer.com",
				ClientID: "test-client-id",
			}
		}

		return provider
	}

	validCookies := func(t *testing.T, req *http.Request) {
		stateCookie := &http.Cookie{
			Name:     api.AuthStateCookieName,
			Value:    "valid-state",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
		pkceCookie := &http.Cookie{
			Name:     api.AuthPKCECookieName,
			Value:    "valid-pkce",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
		req.AddCookie(stateCookie)
		req.AddCookie(pkceCookie)
	}

	noCookies := func(t *testing.T, req *http.Request) {
		// No cookies added
	}

	onlyStateCookie := func(t *testing.T, req *http.Request) {
		stateCookie := &http.Cookie{
			Name:     api.AuthStateCookieName,
			Value:    "valid-state",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
		req.AddCookie(stateCookie)
	}

	onlyPKCECookie := func(t *testing.T, req *http.Request) {
		pkceCookie := &http.Cookie{
			Name:     api.AuthPKCECookieName,
			Value:    "valid-pkce",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
		req.AddCookie(pkceCookie)
	}

	invalidStateCookie := func(t *testing.T, req *http.Request) {
		stateCookie := &http.Cookie{
			Name:     api.AuthStateCookieName,
			Value:    "invalid-state",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
		pkceCookie := &http.Cookie{
			Name:     api.AuthPKCECookieName,
			Value:    "valid-pkce",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
		req.AddCookie(stateCookie)
		req.AddCookie(pkceCookie)
	}

	tt := []testData{
		{
			name:        "Error: No OIDC Provider - Redirect to Login",
			ssoProvider: testProvider(false),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Your SSO connection failed due to misconfiguration, please contact your Administrator",
			},
			setupHost:    mockHost,
			setupCookies: noCookies,
			formParams: map[string]string{
				api.FormParameterCode:  "auth-code",
				api.FormParameterState: "valid-state",
			},
		},
		{
			name:        "Error: Missing Code Parameter - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Invalid SSO Provider response: `code` parameter is missing",
			},
			setupHost:    mockHost,
			setupCookies: noCookies,
			formParams: map[string]string{
				api.FormParameterState: "valid-state",
			},
		},
		{
			name:        "Error: Missing State Parameter - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Invalid SSO Provider response: `state` parameter is missing",
			},
			setupHost:    mockHost,
			setupCookies: noCookies,
			formParams: map[string]string{
				api.FormParameterCode: "auth-code",
			},
		},
		{
			name:        "Error: Missing PKCE Cookie - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Invalid request: `pkce` is missing",
			},
			setupHost:    mockHost,
			setupCookies: onlyStateCookie,
			formParams: map[string]string{
				api.FormParameterCode:  "auth-code",
				api.FormParameterState: "valid-state",
			},
		},
		{
			name:        "Error: Missing State Cookie - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Invalid request: `state` is missing",
			},
			setupHost:    mockHost,
			setupCookies: onlyPKCECookie,
			formParams: map[string]string{
				api.FormParameterCode:  "auth-code",
				api.FormParameterState: "valid-state",
			},
		},
		{
			name:        "Error: State Mismatch - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Invalid: `state` do not match",
			},
			setupHost:    mockHost,
			setupCookies: invalidStateCookie,
			formParams: map[string]string{
				api.FormParameterCode:  "auth-code",
				api.FormParameterState: "valid-state",
			},
		},
		{
			name:        "Error: OIDC Provider Creation Fails - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks:  func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Your SSO connection failed due to misconfiguration, please contact your Administrator",
			},
			setupHost:    mockHost,
			setupCookies: validCookies,
			formParams: map[string]string{
				api.FormParameterCode:  "auth-code",
				api.FormParameterState: "valid-state",
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

			host := testCase.setupHost(t)

			req := httptest.NewRequest("POST", "/api/v2/auth/sso/oidc/callback", nil)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			req.Form = url.Values{}
			for key, value := range testCase.formParams {
				req.Form.Add(key, value)
			}

			testCase.setupCookies(t, req)

			bhContext := &ctx.Context{
				Host: host,
			}

			goContext := ctx.Set(context.Background(), bhContext)
			req = req.WithContext(goContext)

			response := httptest.NewRecorder()
			resources.OIDCCallbackHandler(response, req, testCase.ssoProvider)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.expected.redirectToLogin {
				location := response.Header().Get("Location")
				require.NotEmpty(t, location)

				assert.Contains(t, location, "/ui/login")
				assert.Contains(t, location, url.QueryEscape(testCase.expected.errorMessage))
			}

			cookies := response.Result().Cookies()
			for _, cookie := range cookies {
				if cookie.Name == api.AuthPKCECookieName || cookie.Name == api.AuthStateCookieName {
					assert.LessOrEqual(t, cookie.MaxAge, 0, "Auth cookies should be deleted")
				}
			}
		})
	}
}
