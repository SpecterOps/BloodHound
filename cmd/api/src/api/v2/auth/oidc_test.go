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
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	apimocks "github.com/specterops/bloodhound/cmd/api/src/api/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	oidcmock "github.com/specterops/bloodhound/cmd/api/src/services/oidc/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	v2auth "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_CreateOIDCProvider(t *testing.T) {
	var (
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
	)
	defer mockCtrl.Finish()

	t.Run("successfully create a new OIDCProvider", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(0)).Return(model.Role{}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "Bloodhound gang", "https://localhost/v2auth", "bloodhound", model.SSOProviderConfig{}).Return(model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/v2auth",
		}, nil)

		test.Request(t).
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "Bloodhound gang",
				Issuer:   "https://localhost/v2auth",
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
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "Bloodhound gang2", "https://localhost/v2auth", "bloodhound", config).Return(model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/v2auth",
		}, nil)

		test.Request(t).
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "Bloodhound gang2",
				Issuer:   "https://localhost/v2auth",
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
			WithBody(v2auth.UpsertOIDCProviderRequest{
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
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "1234:not:a:url",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error invalid Issuer", func(t *testing.T) {
		request := v2auth.UpsertOIDCProviderRequest{
			Issuer: "12345:bloodhound",
		}
		test.Request(t).
			WithBody(request).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error issuer uses HTTP", func(t *testing.T) {
		test.Request(t).
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "http://test-issuer.com",
				ClientID: "bloodhound",
				Config:   &model.SSOProviderConfig{},
			}).
			OnHandlerFunc(resources.CreateOIDCProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest).
			ResponseJSONBody(api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: "Issuer: URL must use HTTPS"}},
			})
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().GetRole(gomock.Any(), int32(0)).Return(model.Role{}, nil)
		mockDB.EXPECT().CreateOIDCProvider(gomock.Any(), "test", "https://localhost/v2auth", "bloodhound", model.SSOProviderConfig{}).Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/v2auth",
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
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
		baseProvider         = model.SSOProvider{
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
			WithBody(v2auth.UpsertOIDCProviderRequest{
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
			WithBody(v2auth.UpsertOIDCProviderRequest{
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
			WithBody(v2auth.UpsertOIDCProviderRequest{
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
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "1234:not:a:url",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest)
	})

	t.Run("error issuer uses HTTP", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(v2auth.UpsertOIDCProviderRequest{Issuer: "http://test-issuer.com"}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusBadRequest).
			ResponseJSONBody(api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: "oidc provider issuer url invalid: URL must use HTTPS"}},
			})
	})

	t.Run("error creating oidc provider db entry", func(t *testing.T) {
		mockDB.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(baseProvider, nil)
		mockDB.EXPECT().UpdateOIDCProvider(gomock.Any(), gomock.Any()).Return(model.OIDCProvider{}, fmt.Errorf("error"))

		test.Request(t).
			WithURLPathVars(urlParams).
			WithBody(v2auth.UpsertOIDCProviderRequest{
				Name:     "test",
				Issuer:   "https://localhost/v2auth",
				ClientID: "bloodhound",
			}).
			OnHandlerFunc(resources.UpdateSSOProvider).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}

func TestManagementResource_OIDCLoginHandler(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockOIDC     *oidcmock.MockService
	}
	type expected struct {
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
			name: "Error: No OIDC Provider, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/login/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: OIDC Provider Creation Fails oidc.NewProvider, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
				mocks.mockOIDC.EXPECT().NewProvider(gomock.Any(), "https://test-issuer.com").Return(&oidc.Provider{}, errors.New("error"))
			}, expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/login/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Success: OIDC Login, Redirect to Provider - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
				providerConfig := oidc.ProviderConfig{AuthURL: "https://test-issuer.com/authorize"}
				mocks.mockOIDC.EXPECT().NewProvider(gomock.Any(), "https://test-issuer.com").Return(providerConfig.NewProvider(context.Background()), nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"https://test-issuer.com/authorize?access_type=offline&client_id=test-client-id&code_challenge=challenge&code_challenge_method=S256&redirect_uri=%2Fapi%2Fv2%2Fsso%2Fslug%2Flogin%2Fapi%2Fv2%2Fsso%2Ftest-provider%2Fcallback&response_mode=form_post&response_type=code&scope=openid+profile+email&state=state"}, "Set-Cookie": []string{"pkce=pkce; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; SameSite=None", "state=state; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; SameSite=None"}},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
				mockOIDC:     oidcmock.NewMockService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil, nil)
			resource.OIDC = mocks.mockOIDC

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso/{%s}/login", api.URIPathVariableSSOProviderSlug), resource.SSOLoginHandler).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, _ := test.ProcessResponse(t, response)

			// Cookies are regenerated in every response therefore the
			// the cookie attributes needed to be overwritten.
			header = test.ModifyCookieAttribute(header, "Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
			header = test.ModifyCookieAttribute(header, "state", "state")
			header = test.ModifyCookieAttribute(header, "pkce", "pkce")

			// Normalize the generated query parameters while retaining the authorization URL
			location, err := url.Parse(header.Get("Location"))
			require.NoError(t, err)

			query := location.Query()

			for name, value := range map[string]string{"code_challenge": "challenge", "state": "state"} {
				if query.Has(name) {
					query.Set(name, value)
					location.RawQuery = query.Encode()
					header.Set("Location", location.String())
				}
			}

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
		})
	}
}

func TestManagementResource_OIDCCallbackHandler(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockOIDC     *oidcmock.MockService
	}
	type expected struct {
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
			name: "Error: No OIDC Provider, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: nil,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Error: Empty Authorization Code, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+SSO+Provider+response%3A+%60code%60+parameter+is+missing"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Error: Empty State, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
					Form:   url.Values{},
				}

				request.Form.Add("code", "test")

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+SSO+Provider+response%3A+%60state%60+parameter+is+missing"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Error: Auth PKCE Cookie ErrNoCookie, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
					Form:   url.Values{},
				}

				request.Form.Add("code", "test")
				request.Form.Add("state", "test")

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+request%3A+%60pkce%60+is+missing"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Error: Auth State Cookie ErrNoCookie, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
					Form:   url.Values{},
					Header: http.Header{},
				}

				request.Form.Add("code", "test")
				request.Form.Add("state", "test")

				pkceCookie := &http.Cookie{
					Name:     api.AuthPKCECookieName,
					Value:    "valid-pkce",
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				request.AddCookie(pkceCookie)

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+request%3A+%60state%60+is+missing"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Error: State Mismatch, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
					Form:   url.Values{},
					Header: http.Header{},
				}

				request.Form.Add("code", "test")
				request.Form.Add("state", "test")

				pkceCookie := &http.Cookie{
					Name:     api.AuthPKCECookieName,
					Value:    "valid-pkce",
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				request.AddCookie(pkceCookie)

				stateCookie := &http.Cookie{
					Name:     api.AuthStateCookieName,
					Value:    "state",
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				request.AddCookie(stateCookie)

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid%3A+%60state%60+do+not+match"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Error: oidc.NewProvider Error, Redirect to Login - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
					Form:   url.Values{},
					Header: http.Header{},
				}

				request.Form.Add("code", "test")
				request.Form.Add("state", "state")

				pkceCookie := &http.Cookie{
					Name:     api.AuthPKCECookieName,
					Value:    "valid-pkce",
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				request.AddCookie(pkceCookie)

				stateCookie := &http.Cookie{
					Name:     api.AuthStateCookieName,
					Value:    "state",
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				request.AddCookie(stateCookie)

				bhContext := &bhctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), bhctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					OIDCProvider: &model.OIDCProvider{
						Issuer:   "https://test-issuer.com",
						ClientID: "test-client-id",
					},
				}, nil)
				mocks.mockOIDC.EXPECT().NewProvider(gomock.Any(), "https://test-issuer.com").Return(&oidc.Provider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}, "Set-Cookie": []string{"state=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT", "pkce=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
				mockOIDC:     oidcmock.NewMockService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil, nil)
			resource.OIDC = mocks.mockOIDC
			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso/{%s}/callback", api.URIPathVariableSSOProviderSlug), http.HandlerFunc(resource.SSOCallbackHandler)).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, _ := test.ProcessResponse(t, response)

			updatedHeader := test.ModifyCookieAttribute(header, "Expires", "Thu, 01 Jan 1970 00:00:00 GMT")

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, updatedHeader)
		})
	}
}

type oidcTestTransport func(*http.Request) (*http.Response, error)

func (s oidcTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return s(request)
}

func TestManagementResource_OIDCRequiresHTTPS(t *testing.T) {
	const (
		issuer         = "https://idp.example"
		clientID       = "test-client"
		userEmail      = "user@example.com"
		discoveryPath  = "/.well-known/openid-configuration"
		authorizePath  = "/authorize"
		tokenPath      = "/token"
		tokenRedirect  = "/token/redirected"
		keysPath       = "/keys"
		issuerError    = "Your SSO provider's issuer URL must be a valid HTTPS URL. Ask your administrator to update it"
		downgradeError = "OIDC redirect must use HTTPS"
	)

	signingKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	idToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   issuer,
		"aud":   clientID,
		"sub":   "test-user",
		"email": userEmail,
		"exp":   time.Now().Add(time.Hour).Unix(),
	}).SignedString(signingKey)
	require.NoError(t, err)

	invalidIDToken := idToken[:strings.LastIndex(idToken, ".")+1] + "AA"

	tests := []struct {
		name             string
		handler          func(v2auth.ManagementResource, http.ResponseWriter, *http.Request, model.SSOProvider)
		httpIssuer       bool
		httpEndpoint     string
		redirectPath     string
		redirectLocation string
		invalidSignature bool
		expectedRequests []string
		expectedError    string
	}{
		{
			name:          "login rejects stored HTTP issuer",
			handler:       v2auth.ManagementResource.OIDCLoginHandler,
			httpIssuer:    true,
			expectedError: issuerError,
		},
		{
			name:          "callback rejects stored HTTP issuer",
			handler:       v2auth.ManagementResource.OIDCCallbackHandler,
			httpIssuer:    true,
			expectedError: issuerError,
		},
		{
			name:             "login rejects HTTP authorization endpoint",
			handler:          v2auth.ManagementResource.OIDCLoginHandler,
			httpEndpoint:     authorizePath,
			expectedRequests: []string{discoveryPath},
			expectedError:    "authorization endpoint must be a valid HTTPS URL",
		},
		{
			name:             "callback rejects HTTP token endpoint",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			httpEndpoint:     tokenPath,
			expectedRequests: []string{discoveryPath},
			expectedError:    "invalid OIDC token endpoint: URL must use HTTPS",
		},
		{
			name:             "callback rejects HTTP signing-key endpoint",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			httpEndpoint:     keysPath,
			expectedRequests: []string{discoveryPath},
			expectedError:    "invalid OIDC signing-key URL: URL must use HTTPS",
		},
		{
			name:             "discovery rejects HTTP redirect",
			handler:          v2auth.ManagementResource.OIDCLoginHandler,
			redirectPath:     discoveryPath,
			redirectLocation: "http://idp.example/discovery",
			expectedRequests: []string{discoveryPath},
			expectedError:    "Your SSO connection failed due to misconfiguration",
		},
		{
			name:             "token exchange rejects HTTP redirect",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			redirectPath:     tokenPath,
			redirectLocation: "http://idp.example" + tokenRedirect,
			expectedRequests: []string{discoveryPath, tokenPath},
			expectedError:    downgradeError,
		},
		{
			name:             "signing-key retrieval rejects HTTP redirect",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			redirectPath:     keysPath,
			redirectLocation: "http://idp.example/keys/redirected",
			expectedRequests: []string{discoveryPath, tokenPath, keysPath},
			expectedError:    downgradeError,
		},
		{
			name:             "invalid signature prevents session creation",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			invalidSignature: true,
			expectedRequests: []string{discoveryPath, tokenPath, keysPath},
			expectedError:    "failed to verify signature",
		},
		{
			name:             "valid HTTPS callback reaches session creation",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			expectedRequests: []string{discoveryPath, tokenPath, keysPath},
		},
		{
			name:             "token exchange follows HTTPS redirect",
			handler:          v2auth.ManagementResource.OIDCCallbackHandler,
			redirectPath:     tokenPath,
			redirectLocation: issuer + tokenRedirect,
			expectedRequests: []string{discoveryPath, tokenPath, tokenRedirect, keysPath},
		},
	}

	// Keep these subtests sequential while they replace the default transport.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				requests []string
				provider = model.SSOProvider{OIDCProvider: &model.OIDCProvider{Issuer: issuer, ClientID: clientID}}
				ctrl     = gomock.NewController(t)
			)

			if testCase.httpIssuer {
				provider.OIDCProvider.Issuer = "http://idp.example"
			}

			endpointURL := func(path string) string {
				if testCase.httpEndpoint == path {
					return "http://idp.example" + path
				}
				return issuer + path
			}

			originalTransport := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = originalTransport })

			http.DefaultTransport = oidcTestTransport(func(request *http.Request) (*http.Response, error) {
				if request.URL.Scheme != "https" {
					requests = append(requests, request.URL.String())
					return nil, errors.New("unexpected plaintext OIDC request")
				}

				requests = append(requests, request.URL.Path)
				if request.URL.Path == testCase.redirectPath {
					return &http.Response{
						StatusCode: http.StatusTemporaryRedirect,
						Header:     http.Header{"Location": []string{testCase.redirectLocation}},
						Body:       http.NoBody,
					}, nil
				}

				var body any

				switch request.URL.Path {
				case discoveryPath:
					body = oidc.ProviderConfig{
						IssuerURL: issuer,
						AuthURL:   endpointURL(authorizePath),
						TokenURL:  endpointURL(tokenPath),
						JWKSURL:   endpointURL(keysPath),
					}
				case tokenPath, tokenRedirect:
					if request.Method != http.MethodPost {
						return nil, fmt.Errorf("unexpected token request method: %s", request.Method)
					}

					responseIDToken := idToken
					if testCase.invalidSignature {
						responseIDToken = invalidIDToken
					}

					body = map[string]string{"access_token": "test-access-token", "id_token": responseIDToken}
				case keysPath:
					body = map[string]any{"keys": []map[string]string{{
						"kty": "RSA",
						"n":   base64.RawURLEncoding.EncodeToString(signingKey.N.Bytes()),
						"e":   "AQAB",
					}}}
				default:
					return nil, fmt.Errorf("unexpected OIDC request to %s", request.URL)
				}

				responseBody, err := json.Marshal(body)
				if err != nil {
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				}, nil
			})

			mockAuthenticator := apimocks.NewMockAuthenticator(ctrl)
			if testCase.expectedError == "" {
				mockAuthenticator.EXPECT().CreateSSOSession(gomock.Any(), gomock.Any(), userEmail, provider)
			}

			resource := v2auth.NewManagementResource(config.Configuration{}, nil, auth.Authorizer{}, mockAuthenticator, nil, nil, nil)
			request := httptest.NewRequest(http.MethodGet, "https://bloodhound.example/sso?code=test&state=test", nil)
			request = request.WithContext(context.WithValue(request.Context(), bhctx.ValueKey, &bhctx.Context{Host: request.URL}))
			request.AddCookie(&http.Cookie{Name: api.AuthStateCookieName, Value: "test"})
			request.AddCookie(&http.Cookie{Name: api.AuthPKCECookieName, Value: "test"})
			response := httptest.NewRecorder()

			testCase.handler(resource, response, request, provider)

			assert.Equal(t, testCase.expectedRequests, requests)
			if testCase.expectedError != "" {
				assert.Equal(t, http.StatusFound, response.Code)
				location, err := url.Parse(response.Header().Get("Location"))
				require.NoError(t, err)
				assert.Contains(t, location.Query().Get("error"), testCase.expectedError)
			}
		})
	}
}
