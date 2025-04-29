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
	"regexp"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/database/mocks"

	"github.com/pkg/errors"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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

func TestManagementResource_ServeMetadata(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode int
		contentType  string
		responseBody string
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
			name:            "Error: Provider not found",
			ssoProviderSlug: "nonexistent-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "nonexistent-provider").Return(model.SSOProvider{}, sql.ErrNoRows)
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusInternalServerError,
				contentType:  "application/json",
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name:            "Error: Provider is not a SAML provider",
			ssoProviderSlug: "oidc-provider",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "oidc-provider").Return(model.SSOProvider{
					Name:         "OIDC Provider",
					Slug:         "oidc-provider",
					Type:         model.SessionAuthProviderOIDC,
					SAMLProvider: nil,
				}, nil)
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			expected: expected{
				responseCode: http.StatusNotFound,
				contentType:  "application/json",
				responseBody: `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name:            "Error: Service provider creation fails",
			ssoProviderSlug: "okta",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetSSOProviderBySlug(gomock.Any(), "okta").Return(model.SSOProvider{
					Name: "Okta",
					Slug: "okta",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name:            "Okta Provider",
						DisplayName:     "Okta SSO",
						IssuerURI:       "https://okta.com/saml",
						SingleSignOnURI: "https://okta.com/sso",
					},
				}, nil)
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
				contentType:  "application/json",
				responseBody: `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"failed to parse service provider Okta Provider's cert pair: x509: malformed certificate"}]}`,
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

			endpointURL := fmt.Sprintf("/api/v2/sso/%s/metadata.xml", testCase.ssoProviderSlug)
			req, err := http.NewRequest("GET", endpointURL, nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			vars := map[string]string{
				api.URIPathVariableSSOProviderSlug: testCase.ssoProviderSlug,
			}
			req = mux.SetURLVars(req, vars)

			response := httptest.NewRecorder()
			resources.ServeMetadata(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.expected.responseCode != http.StatusOK {
				responseBodyWithDefaultTimestamp, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
				require.NoError(t, err)
				assert.JSONEq(t, testCase.expected.responseBody, responseBodyWithDefaultTimestamp)
			}
		})
	}
}

func TestManagementResource_ServeSigningCertificate(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode        int
		contentType         string
		contentDisposition  string
		responseBodyPattern string
	}
	type testData struct {
		name       string
		providerID string
		setupMocks func(*testing.T, *mocks.MockDatabase)
		expected   expected
	}

	tt := []testData{
		{
			name:       "Error: Invalid provider ID format - Not Found",
			providerID: "invalid",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			expected: expected{
				responseCode:        http.StatusNotFound,
				contentType:         "application/json",
				responseBodyPattern: `{"http_status":404,"timestamp":".*","request_id":"","errors":\[{"context":"","message":"resource not found"}\]}`,
			},
		},
		{
			name:       "Error: Provider not found - Database Error",
			providerID: "1",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().
					GetSSOProviderById(gomock.Any(), int32(1)).
					Return(model.SSOProvider{}, sql.ErrNoRows)
			},
			expected: expected{
				responseCode:        http.StatusInternalServerError,
				contentType:         "application/json",
				responseBodyPattern: `{"http_status":500,"timestamp":".*","request_id":"","errors":\[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}\]}`,
			},
		},
		{
			name:       "Error: Provider is not a SAML provider - Not Found",
			providerID: "1",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().
					GetSSOProviderById(gomock.Any(), int32(1)).
					Return(model.SSOProvider{
						Name:         "OIDC Provider",
						Slug:         "oidc-provider",
						Type:         model.SessionAuthProviderOIDC,
						SAMLProvider: nil,
					}, nil)
			},
			expected: expected{
				responseCode:        http.StatusNotFound,
				contentType:         "application/json",
				responseBodyPattern: `{"http_status":404,"timestamp":".*","request_id":"","errors":\[{"context":"","message":"resource not found"}\]}`,
			},
		},
		{
			name:       "Success: SAML provider certificate served",
			providerID: "1",
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().
					GetSSOProviderById(gomock.Any(), int32(1)).
					Return(model.SSOProvider{
						Name: "Okta Provider",
						Slug: "okta-provider",
						Type: model.SessionAuthProviderSAML,
						SAMLProvider: &model.SAMLProvider{
							Name:        "Okta Provider",
							DisplayName: "Okta SSO",
						},
					}, nil)
			},
			expected: expected{
				responseCode:        http.StatusOK,
				contentDisposition:  `attachment; filename="okta-provider-signing-certificate.pem"`,
				responseBodyPattern: ".*BEGIN CERTIFICATE.*",
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

			endpointURL := fmt.Sprintf("/api/v2/saml/providers/%s/signing-certificate", testCase.providerID)
			req, err := http.NewRequest("GET", endpointURL, nil)
			require.NoError(t, err)

			vars := map[string]string{
				api.URIPathVariableSSOProviderID: testCase.providerID,
			}
			req = mux.SetURLVars(req, vars)

			response := httptest.NewRecorder()
			resources.ServeSigningCertificate(response, req)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.expected.contentType != "" {
				assert.Contains(t, response.Header().Get("Content-Type"), testCase.expected.contentType)
			}

			if testCase.expected.contentDisposition != "" {
				assert.Equal(t, testCase.expected.contentDisposition, response.Header().Get("Content-Disposition"))
			}

			if testCase.expected.responseBodyPattern != "" {
				responseBody := response.Body.String()
				matched, err := regexp.MatchString(testCase.expected.responseBodyPattern, responseBody)
				require.NoError(t, err)
				assert.True(t, matched, "Response body does not match expected pattern. Got: %s", responseBody)
			}
		})
	}
}

func TestManagementResource_SAMLLoginHandler(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode    int
		headers         map[string]string
		responsePattern string
	}
	type testData struct {
		name         string
		ssoProvider  model.SSOProvider
		setupMocks   func(*testing.T, *mocks.MockDatabase)
		setupContext func(*testing.T) context.Context
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Nil SAML Provider - Redirect to Login with Error Message",
			ssoProvider: model.SSOProvider{
				Name:         "Test Provider",
				Slug:         "test-provider",
				Type:         model.SessionAuthProviderSAML,
				SAMLProvider: nil,
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusFound,
				headers: map[string]string{
					"Location": "https://example.com/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator",
				},
			},
		},
		{
			name: "Error: Service Provider Creation Fails - Redirect to Login with Error Message",
			ssoProvider: model.SSOProvider{
				Name: "Test Provider",
				Slug: "test-provider",
				Type: model.SessionAuthProviderSAML,
				SAMLProvider: &model.SAMLProvider{
					Name:            "Test SAML Provider",
					DisplayName:     "Test SAML SSO",
					IssuerURI:       "invalid-uri",
					SingleSignOnURI: "",
				},
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusFound,
				headers: map[string]string{
					"Location": "https://example.com/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator",
				},
			},
		},
		{
			name: "Error: Authentication Request Creation Fails - Redirect to Login with Error Message",
			ssoProvider: model.SSOProvider{
				Name: "Valid Provider",
				Slug: "valid-provider",
				Type: model.SessionAuthProviderSAML,
				SAMLProvider: &model.SAMLProvider{
					Name:            "Valid SAML Provider",
					DisplayName:     "Valid SAML SSO",
					IssuerURI:       "https://valid-provider.com/saml",
					SingleSignOnURI: "https://valid-provider.com/sso",
					MetadataXML: []byte(`<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://valid-provider.com/saml">
						<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
							<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://valid-provider.com/sso"/>
						</IDPSSODescriptor>
					</EntityDescriptor>`),
				},
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusFound,
				headers: map[string]string{
					"Location": "https://example.com/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator",
				},
			},
		},
		{
			name: "Error: Post Binding Also Fails - Redirect to Login with Error Message",
			ssoProvider: model.SSOProvider{
				Name: "POST Provider",
				Slug: "post-provider",
				Type: model.SessionAuthProviderSAML,
				SAMLProvider: &model.SAMLProvider{
					Name:            "POST SAML Provider",
					DisplayName:     "POST SAML SSO",
					IssuerURI:       "https://post-provider.com/saml",
					SingleSignOnURI: "https://post-provider.com/sso",
					MetadataXML: []byte(`<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://post-provider.com/saml">
						<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
							<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://post-provider.com/sso"/>
						</IDPSSODescriptor>
					</EntityDescriptor>`),
				},
			},
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {},
			setupContext: func(t *testing.T) context.Context {
				hostURL, _ := url.Parse("https://example.com")
				userContext := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
				bhCtx := ctx.Get(userContext)
				bhCtx.Host = hostURL
				return userContext
			},
			expected: expected{
				responseCode: http.StatusFound,
				headers: map[string]string{
					"Location": "https://example.com/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator",
				},
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

			req, err := http.NewRequest("GET", "/api/v2/sso/provider/login", nil)
			require.NoError(t, err)

			reqCtx := testCase.setupContext(t)
			req = req.WithContext(reqCtx)

			response := httptest.NewRecorder()
			resources.SAMLLoginHandler(response, req, testCase.ssoProvider)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			for name, expectedValue := range testCase.expected.headers {
				value := response.Header().Get(name)
				assert.Equal(t, expectedValue, value, "Header %s does not match", name)
			}

			if testCase.expected.responsePattern != "" {
				responseBody := response.Body.String()
				matched, err := regexp.MatchString(testCase.expected.responsePattern, responseBody)
				require.NoError(t, err)
				assert.True(t, matched, "Response body does not match expected pattern. Got: %s", responseBody)
			}
		})
	}
}

func TestManagementResource_SAMLCallbackHandler(t *testing.T) {
	t.Parallel()

	type expected struct {
		responseCode    int
		redirectToLogin bool
		errorMessage    string
	}
	type testData struct {
		name        string
		ssoProvider model.SSOProvider
		setupMocks  func(*testing.T, *mocks.MockDatabase)
		expected    expected
		setupHost   func(*testing.T) *url.URL
		formValues  url.Values
	}

	mockHost := func(t *testing.T) *url.URL {
		host, err := url.Parse("https://example.com")
		require.NoError(t, err)
		return host
	}

	testProvider := func(includeSAML bool) model.SSOProvider {
		provider := model.SSOProvider{
			Name: "Test Provider",
			Type: model.SessionAuthProviderSAML,
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

		if includeSAML {
			provider.SAMLProvider = &model.SAMLProvider{
				Name:            "Test SAML Provider",
				DisplayName:     "Test SAML Provider",
				IssuerURI:       "https://test-issuer.com",
				SingleSignOnURI: "https://test-issuer.com/sso",
				PrincipalAttributeMappings: []string{
					"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
				},
				SSOProviderID: null.Int32From(1),
			}
		}

		return provider
	}

	validFormValues := url.Values{
		"SAMLResponse": []string{"valid-saml-response"},
	}

	tt := []testData{
		{
			name:        "Error: No SAML Provider - Redirect to Login",
			ssoProvider: testProvider(false),
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
			},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Your SSO connection failed due to misconfiguration, please contact your Administrator",
			},
			setupHost:  mockHost,
			formValues: validFormValues,
		},
		{
			name:        "Error: Service Provider Creation Failed - Redirect to Login",
			ssoProvider: testProvider(true),
			setupMocks: func(t *testing.T, mockDB *mocks.MockDatabase) {
			},
			expected: expected{
				responseCode:    http.StatusFound,
				redirectToLogin: true,
				errorMessage:    "Your SSO connection failed due to misconfiguration, please contact your Administrator",
			},
			setupHost:  mockHost,
			formValues: validFormValues,
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

			var req *http.Request
			var err error
			if testCase.formValues != nil {
				req, err = http.NewRequest("POST", "/api/v2/saml/callback", strings.NewReader(testCase.formValues.Encode()))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, err = http.NewRequest("POST", "/api/v2/saml/callback", nil)
				require.NoError(t, err)
			}

			bhContext := &ctx.Context{
				Host: host,
			}
			goContext := ctx.Set(context.Background(), bhContext)
			req = req.WithContext(goContext)

			response := httptest.NewRecorder()
			resources.SAMLCallbackHandler(response, req, testCase.ssoProvider)

			assert.Equal(t, testCase.expected.responseCode, response.Code)

			if testCase.expected.redirectToLogin {
				location := response.Header().Get("Location")
				require.NotEmpty(t, location)
				assert.Contains(t, location, "/login")
				assert.Contains(t, location, url.QueryEscape(testCase.expected.errorMessage))
			}
		})
	}
}
