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
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/crewjam/saml"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/serde"
	samlmocks "github.com/specterops/bloodhound/cmd/api/src/services/saml/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"

	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2auth "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"go.uber.org/mock/gomock"
)

var validMetadataXML = `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://okta.com/saml">
		<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
			<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://okta.com/sso"/>
		</IDPSSODescriptor>
	</EntityDescriptor>`

func TestManagementResource_SAMLLoginRedirect(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Database Error SSO Provider not found - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider",
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success: Redirect to SSO provider login URL - Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider",
					},
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{
					Slug: "okta",
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/%7Bversion%7D/login/saml/provider/api/v2/sso/okta/login"}},
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

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/{version}/login/saml/{%s}", api.URIPathVariableSSOProviderSlug), resource.SAMLLoginRedirect).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusFound {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}

func TestManagementResource_SAMLCallbackRedirect(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Database Error SSO Provider not found - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider/acs",
					},
					Method: http.MethodPost,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success: Redirect to SSO provider login URL - Temporary Redirect",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider/acs",
					},
					Method: http.MethodPost,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{
					Slug: "okta",
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusTemporaryRedirect,
				responseHeader: http.Header{"Location": []string{"/api/%7Bversion%7D/login/saml/provider/acs/api/v2/sso/okta/callback"}},
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

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/{version}/login/saml/{%s}/acs", api.URIPathVariableSSOProviderSlug), resource.SAMLCallbackRedirect).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusTemporaryRedirect {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}

func TestManagementResource_ListSAMLSignOnEndpoints(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Database Error db.GetAllSAMLProviders - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/sso",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(model.SAMLProviders{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success: No endpoints provided, Empty list - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/sso",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(model.SAMLProviders{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"endpoints":[]}}`,
			},
		},
		{
			name: "Success: Listed Endpoints - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/sso",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
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

				oktaProvider.ServiceProviderInitiationURI = serde.URL{URL: url.URL{}}
				azureProvider.ServiceProviderInitiationURI = serde.URL{URL: url.URL{}}

				samlProviders := []model.SAMLProvider{oktaProvider, azureProvider}
				mock.mockDatabase.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(samlProviders, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"endpoints":[{"name":"Okta Provider","initiation_url":"Okta%20Provider/login"},{"name":"Azure Provider","initiation_url":"Azure%20Provider/login"}]}}`,
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

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(request.URL.Path, resource.ListSAMLSignOnEndpoints).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_ListSAMLProviders(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Database error db.GetAllSAMLProviders - Internal Server Error",
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(model.SAMLProviders{}, errors.New("error"))
			},
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml",
					},
					Method: http.MethodGet,
				}
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: No SAML providers provided, Empty list - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))

			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetAllSAMLProviders(gomock.Any()).Return([]model.SAMLProvider{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"saml_providers":[]}}`,
			},
		},
		{
			name: "Success: Multiple SAML providers provided - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))

			},
			setupMocks: func(t *testing.T, mock *mock) {
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
				// Okta
				oktaIssuerURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Okta%20Provider/issuer",
				}
				oktaInitURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Okta%20Provider/login",
				}
				oktaMetadataURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Okta%20Provider/metadata.xml",
				}
				oktaACSURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Okta%20Provider/acs",
				}
				// Azure
				azureIssuerURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Azure%20Provider/issuer",
				}
				azureInitURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Azure%20Provider/login",
				}
				azureMetadataURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Azure%20Provider/metadata.xml",
				}
				azureACSURL := &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/Azure%20Provider/acs",
				}
				oktaProvider.ServiceProviderIssuerURI = serde.URL{URL: *oktaIssuerURL}
				oktaProvider.ServiceProviderInitiationURI = serde.URL{URL: *oktaInitURL}
				oktaProvider.ServiceProviderMetadataURI = serde.URL{URL: *oktaMetadataURL}
				oktaProvider.ServiceProviderACSURI = serde.URL{URL: *oktaACSURL}

				azureProvider.ServiceProviderIssuerURI = serde.URL{URL: *azureIssuerURL}
				azureProvider.ServiceProviderInitiationURI = serde.URL{URL: *azureInitURL}
				azureProvider.ServiceProviderMetadataURI = serde.URL{URL: *azureMetadataURL}
				azureProvider.ServiceProviderACSURI = serde.URL{URL: *azureACSURL}

				samlProviders := []model.SAMLProvider{oktaProvider, azureProvider}
				mock.mockDatabase.EXPECT().GetAllSAMLProviders(gomock.Any()).Return(samlProviders, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
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
								"sp_issuer_uri": "https://example.com/Okta%2520Provider/issuer",
								"sp_sso_uri": "https://example.com/Okta%2520Provider/login",
								"sp_metadata_uri": "https://example.com/Okta%2520Provider/metadata.xml",
								"sp_acs_uri": "https://example.com/Okta%2520Provider/acs",
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
								"sp_issuer_uri": "https://example.com/Azure%2520Provider/issuer",
								"sp_sso_uri": "https://example.com/Azure%2520Provider/login",
								"sp_metadata_uri": "https://example.com/Azure%2520Provider/metadata.xml",
								"sp_acs_uri": "https://example.com/Azure%2520Provider/acs",
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
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(request.URL.Path, resource.ListSAMLProviders).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetSAMLProvider(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: Invalid provider ID format - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers/invalid",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name: "Error: Database error db.GetSAMLProvider - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers/1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSAMLProvider(gomock.Any(), int32(1)).Return(model.SAMLProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Provider found - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers/1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
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

				mock.mockDatabase.EXPECT().GetSAMLProvider(gomock.Any(), int32(1)).Return(oktaProvider, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
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
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/saml/providers/{%s}", api.URIPathVariableSAMLProviderID), resource.GetSAMLProvider).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_CreateSAMLProviderMultipart(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func(testName string) *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Missing Content Type Header, ParseMultipartForm - Bad Request",
			buildRequest: func(string) *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"request Content-Type isn't multipart/form-data"}]}`,
			},
		},
		{
			name: "Error: Empty multiform, ParseMultipartForm - Bad Request",
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", "bound"))

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"multipart: NextPart: EOF"}]}`,
			},
		},
		{
			name: "Error: missing name parameter, getProviderNameFromMultipartRequest - Bad Request",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("nope", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"form is missing \"name\" parameter"}]}`,
			},
		},
		{
			name: "Error: missing metadata parameter, getMetadataFromMultipartRequest - Bad Request",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"form is missing \"metadata\" parameter"}]}`,
			},
		},
		{
			name: "Error: missing sso provider config, getSSOProviderConfigFromMultipartRequest - Bad Request",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"form is missing \"config.auto_provision.enabled\" parameter"}]}`,
			},
		},
		{
			name: "Error: auth.GetIDPSingleSignOnServiceURL - Bad Request",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				// Binding causes error
				_, err = metadataFile.Write([]byte(`<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://okta.com/saml">
		<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
			<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-GET" Location="https://okta.com/sso"/>
		</IDPSSODescriptor>
	</EntityDescriptor>`))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision enabled form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision enabled form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)

			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"metadata does not have a SSO service that supports HTTP POST binding"}]}`,
			},
		},
		{
			name: "Error: Database error duplicate provider db.CreateSAMLIdentityProvider - Conflict",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)
				mock.mockDatabase.EXPECT().CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SAMLProvider{}, database.ErrDuplicateSSOProviderName)
			},
			expected: expected{
				responseCode:   http.StatusConflict,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":409,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"sso provider name must be unique"}]}`,
			},
		},
		{
			name: "Error: Database error db.CreateSAMLIdentityProvider- Bad Request",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)
				mock.mockDatabase.EXPECT().CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SAMLProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Provider Multipart Created - OK",
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/saml/providers",
					},
					Method: http.MethodPost,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error writing name field needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)
				mock.mockDatabase.EXPECT().CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SAMLProvider{
					Name:            "name",
					DisplayName:     "display",
					IssuerURI:       "uri",
					SingleSignOnURI: "uri",
					MetadataXML:     []byte{},
					RootURIVersion:  model.SAMLRootURIVersion1,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"created_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"display_name":"display","id":0,"idp_issuer_uri":"uri","idp_sso_uri":"uri","name":"name","principal_attribute_mappings":null,"root_uri_version":1,"sp_acs_uri":"","sp_issuer_uri":"","sp_metadata_uri":"","sp_sso_uri":"","sso_provider_id":null,"updated_at":"0001-01-01T00:00:00Z"}}`,
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

			request := testCase.buildRequest(t.Name())
			testCase.setupMocks(t, mocks)

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(request.URL.Path, resource.CreateSAMLProviderMultipart).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_UpdateSAMLProviderRequest(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		args         model.SSOProvider
		buildRequest func(testName string) *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}
	tt := []testData{
		{
			name: "Error: Nil ssoProvider.SAMLProvider - Not Found",
			args: model.SSOProvider{
				SAMLProvider: nil,
			},
			buildRequest: func(string) *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type: model.SessionAuthProviderSAML,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name: "Error: Missing Content Type Header, ParseMultipartForm - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(string) *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
			}, expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"request Content-Type isn't multipart/form-data"}]}`,
			},
		},
		{
			name: "Error: Empty multiform, ParseMultipartForm - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(name string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", "bound"))

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"multipart: NextPart: EOF"}]}`,
			},
		},
		{
			name: "Error: duplicate name parameters, getProviderNameFromMultipartRequest - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", "bound"))

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider #1")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}
				err = writer.WriteField("name", "okta provider #2")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"expected only one \"name\" parameter"}]}`,
			},
		},
		{
			name: "Error: duplicate metadata forms, getMetadataFromMultipartRequest - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				dupeMetadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = dupeMetadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"expected only one \"metadata\" parameter"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: duplicate sso provider config, getSSOProviderConfigFromMultipartRequest - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision enabled form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision enabled form field, needed for test %s: %v", testName, err)
				}

				dupeAutoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision enabled form file, needed for test %s: %v", testName, err)
				}
				_, err = dupeAutoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision enabled form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"expected only one \"config.auto_provision.enabled\" parameter"}]}`,
			},
		},
		{
			name: "Error: auth.GetIDPSingleSignOnServiceURL - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(`<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://okta.com/saml">
			<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
				<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-GET" Location="https://okta.com/sso"/>
			</IDPSSODescriptor>
		</EntityDescriptor>`))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision enabled form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision enabled form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)

			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"metadata does not have a SSO service that supports HTTP POST binding"}]}`,
			},
		},
		{
			name: "Error: Database error duplicate provider db.UpdateSAMLIdentityProvider - Conflict",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)
				mock.mockDatabase.EXPECT().UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).Return(model.SAMLProvider{}, database.ErrDuplicateSSOProviderName)
			},
			expected: expected{
				responseCode:   http.StatusConflict,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":409,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"sso provider name must be unique"}]}`,
			},
		},
		{
			name: "Error: Database error db.UpdateSAMLIdentityProvider - Bad Request",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error occurred while writing name field, needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)
				mock.mockDatabase.EXPECT().UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).Return(model.SAMLProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Success: Provider Multipart Updated - OK",
			args: model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Name:            "name",
					DisplayName:     "display",
					IssuerURI:       "uri",
					SingleSignOnURI: "uri",
					MetadataXML:     []byte{},
					RootURIVersion:  model.SAMLRootURIVersion1,
				},
			},
			buildRequest: func(testName string) *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1",
					},
					Method: http.MethodPatch,
					Header: http.Header{},
				}

				// Create in-memory multipart body
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				err := writer.WriteField("name", "okta provider")
				if err != nil {
					t.Fatalf("error writing name field needed for test %s: %v", testName, err)
				}

				metadataFile, err := writer.CreateFormFile("metadata", "metadata.xml")
				if err != nil {
					t.Fatalf("error occurred while creating metadata form file, needed for test %s: %v", testName, err)
				}
				_, err = metadataFile.Write([]byte(validMetadataXML))
				if err != nil {
					t.Fatalf("error occurred while writing to metadata form file, needed for test %s: %v", testName, err)
				}

				autoProvisionField, err := writer.CreateFormField("config.auto_provision.enabled")
				if err != nil {
					t.Fatalf("error occurred while writing creating config auto provision form file, needed for test %s: %v", testName, err)
				}
				_, err = autoProvisionField.Write([]byte("true"))
				if err != nil {
					t.Fatalf("error occurred while writing to auto provision form field, needed for test %s: %v", testName, err)
				}

				roleIDField, err := writer.CreateFormField("config.auto_provision.default_role_id")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision default role id form field, needed for test %s: %v", testName, err)
				}
				_, err = roleIDField.Write([]byte("1"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision default role id form field, needed for test %s: %v", testName, err)
				}

				roleProvisionField, err := writer.CreateFormField("config.auto_provision.role_provision")
				if err != nil {
					t.Fatalf("error occurred while creating config auto provision role provision form field, needed for test %s: %v", testName, err)
				}
				_, err = roleProvisionField.Write([]byte("false"))
				if err != nil {
					t.Fatalf("error occurred while writing to config auto provision role provision form field, needed for test %s: %v", testName, err)
				}

				// Close the writer to finalize the body
				writer.Close()

				request.Body = io.NopCloser(&body)
				request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{},
				}, nil)
				mock.mockDatabase.EXPECT().GetRole(gomock.Any(), int32(1)).Return(model.Role{}, nil)
				mock.mockDatabase.EXPECT().UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).Return(model.SAMLProvider{
					Name:            "name",
					DisplayName:     "display",
					IssuerURI:       "uri",
					SingleSignOnURI: "uri",
					MetadataXML:     []byte{},
					RootURIVersion:  model.SAMLRootURIVersion1,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"created_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false},"display_name":"display","id":0,"idp_issuer_uri":"uri","idp_sso_uri":"uri","name":"name","principal_attribute_mappings":null,"root_uri_version":1,"sp_acs_uri":"","sp_issuer_uri":"","sp_metadata_uri":"","sp_sso_uri":"","sso_provider_id":null,"updated_at":"0001-01-01T00:00:00Z"}}`,
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

			request := testCase.buildRequest(t.Name())
			testCase.setupMocks(t, mocks)

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso-providers/{%s}", api.URIPathVariableSSOProviderID), resource.UpdateSSOProvider).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_ServeMetadata(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockSAML     *samlmocks.MockService
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: Database error db.GetSSOProviderBySlug - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider/metadata",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: SAML Provider is nil - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider/metadata",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{
					Name:         "OIDC Provider",
					Slug:         "oidc-provider",
					Type:         model.SessionAuthProviderOIDC,
					SAMLProvider: nil,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name: "Error: NewServiceProvider Unable to parse SAML cert and provider key - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider/metadata",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{
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
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"failed to parse metadata XML for service provider Okta Provider: EOF"}]}`,
			},
		},
		{
			name: "Success: Metadata Served - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/{version}/login/saml/provider/metadata",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider").Return(model.SSOProvider{
					Name: "Okta",
					Slug: "okta",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name:            "Okta Provider",
						DisplayName:     "Okta SSO",
						IssuerURI:       "https://okta.com/saml",
						SingleSignOnURI: "https://okta.com/sso",
						MetadataXML:     []byte(validMetadataXML),
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/samlmetadata+xml"}},
				responseBody:   string("<EntityDescriptor xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\" validUntil=\"XXX\" entityID=\"Okta%20Provider\">\n  <SPSSODescriptor xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\" validUntil=\"XXX\" protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\" AuthnRequestsSigned=\"true\" WantAssertionsSigned=\"true\">\n    <KeyDescriptor use=\"encryption\">\n      <KeyInfo xmlns=\"http://www.w3.org/2000/09/xmldsig#\">\n        <X509Data xmlns=\"http://www.w3.org/2000/09/xmldsig#\">\n          <X509Certificate xmlns=\"http://www.w3.org/2000/09/xmldsig#\">MIIDrzCCApegAwIBAgIUG+lLMPpkTfwVe5fVbBhWe/cIxrgwDQYJKoZIhvcNAQELBQAwfzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUxGjAYBgNVBAoMEVNwZWN0ZXIgT3BzLCBJbmMuMQwwCgYDVQQDDANCb2IxHzAdBgkqhkiG9w0BCQEWEHNwYW1AZXhhbXBsZS5jb20wIBcNMjUwNjAyMTY1ODMzWhgPMzAyNDEwMDMxNjU4MzNaMH8xCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApXYXNoaW5ndG9uMRAwDgYDVQQHDAdTZWF0dGxlMRowGAYDVQQKDBFTcGVjdGVyIE9wcywgSW5jLjEMMAoGA1UEAwwDQm9iMR8wHQYJKoZIhvcNAQkBFhBzcGFtQGV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5KMZPTuMDMIS3N24Y2l6FLkbSZnG/Xa16MiSRtuvK6qExkVZ0Jr/7kDJ0F9eOWe6SgRx7/o/olnKC3vs5HwduSoQzLiFH7T8FRGEOlmXbCtfFeWQzL0PwvoX7YQDP4d1IlEdznMWieXM73hKL/UonrydrsJt6zeyhS0suAZz+UMpocsc9xtFtdaZBVSnSjqrK+PoT2V2WYUUdnrBh9xnJ1aIX78PUoU8fF8YviJ9Anqe+sAsLX0nOVclj4LNGX+LhFoQcR4DkxolCNmLozD091Xvodhrst0kYKtgpJl8wFJ1ULeyvCvXKR3fNKseM7DkTBd5c3fZORZpKRrs8UNc1wIDAQABoyEwHzAdBgNVHQ4EFgQUlFcqpLK4byN5z7dGyNpLjV+/sMIwDQYJKoZIhvcNAQELBQADggEBALRZSG5BlQcMZN9kZCBVqUunCQ96cgCZUp77b0QEPXsSn+KEj+br224DLmKiZ1z7wHNQefjEtAxcajS5ZTapOAekMo9Tnc5MUvAGh4cFImcn1b17u2IcMHAXOQNi6exBWToDKVAYwCpxQ2x9ELHs0mU1HC/HrkqLIAbQjoBL0ZK55euoLJQ35s7M9uNLv4sBmXLKdScZiDsQgvTwSyFGVDrzOnkq7IsAfHZ23vwndY11/x7dQIqcdNmN/rTSlx0Gz0SArtOvIjIeMtvmv8KAPCjU+AvTGmNXZWbvr5lFJOf46ORopwpJtTijhzvihLvBrLm6UdrLuosbvXuJGqk2EEM=</X509Certificate>\n        </X509Data>\n      </KeyInfo>\n      <EncryptionMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#aes128-cbc\"></EncryptionMethod>\n      <EncryptionMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#aes192-cbc\"></EncryptionMethod>\n      <EncryptionMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#aes256-cbc\"></EncryptionMethod>\n      <EncryptionMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p\"></EncryptionMethod>\n    </KeyDescriptor>\n    <KeyDescriptor use=\"signing\">\n      <KeyInfo xmlns=\"http://www.w3.org/2000/09/xmldsig#\">\n        <X509Data xmlns=\"http://www.w3.org/2000/09/xmldsig#\">\n          <X509Certificate xmlns=\"http://www.w3.org/2000/09/xmldsig#\">MIIDrzCCApegAwIBAgIUG+lLMPpkTfwVe5fVbBhWe/cIxrgwDQYJKoZIhvcNAQELBQAwfzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUxGjAYBgNVBAoMEVNwZWN0ZXIgT3BzLCBJbmMuMQwwCgYDVQQDDANCb2IxHzAdBgkqhkiG9w0BCQEWEHNwYW1AZXhhbXBsZS5jb20wIBcNMjUwNjAyMTY1ODMzWhgPMzAyNDEwMDMxNjU4MzNaMH8xCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApXYXNoaW5ndG9uMRAwDgYDVQQHDAdTZWF0dGxlMRowGAYDVQQKDBFTcGVjdGVyIE9wcywgSW5jLjEMMAoGA1UEAwwDQm9iMR8wHQYJKoZIhvcNAQkBFhBzcGFtQGV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5KMZPTuMDMIS3N24Y2l6FLkbSZnG/Xa16MiSRtuvK6qExkVZ0Jr/7kDJ0F9eOWe6SgRx7/o/olnKC3vs5HwduSoQzLiFH7T8FRGEOlmXbCtfFeWQzL0PwvoX7YQDP4d1IlEdznMWieXM73hKL/UonrydrsJt6zeyhS0suAZz+UMpocsc9xtFtdaZBVSnSjqrK+PoT2V2WYUUdnrBh9xnJ1aIX78PUoU8fF8YviJ9Anqe+sAsLX0nOVclj4LNGX+LhFoQcR4DkxolCNmLozD091Xvodhrst0kYKtgpJl8wFJ1ULeyvCvXKR3fNKseM7DkTBd5c3fZORZpKRrs8UNc1wIDAQABoyEwHzAdBgNVHQ4EFgQUlFcqpLK4byN5z7dGyNpLjV+/sMIwDQYJKoZIhvcNAQELBQADggEBALRZSG5BlQcMZN9kZCBVqUunCQ96cgCZUp77b0QEPXsSn+KEj+br224DLmKiZ1z7wHNQefjEtAxcajS5ZTapOAekMo9Tnc5MUvAGh4cFImcn1b17u2IcMHAXOQNi6exBWToDKVAYwCpxQ2x9ELHs0mU1HC/HrkqLIAbQjoBL0ZK55euoLJQ35s7M9uNLv4sBmXLKdScZiDsQgvTwSyFGVDrzOnkq7IsAfHZ23vwndY11/x7dQIqcdNmN/rTSlx0Gz0SArtOvIjIeMtvmv8KAPCjU+AvTGmNXZWbvr5lFJOf46ORopwpJtTijhzvihLvBrLm6UdrLuosbvXuJGqk2EEM=</X509Certificate>\n        </X509Data>\n      </KeyInfo>\n    </KeyDescriptor>\n    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>\n    <AssertionConsumerService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"\" index=\"1\"></AssertionConsumerService>\n    <AssertionConsumerService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact\" Location=\"\" index=\"2\"></AssertionConsumerService>\n  </SPSSODescriptor>\n</EntityDescriptor>"),
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
				mockSAML:     samlmocks.NewMockService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2auth.NewManagementResource(config.Configuration{
				SAML: config.SAMLConfiguration{
					ServiceProviderCertificate:        ValidCert,
					ServiceProviderKey:                ValidKey,
					ServiceProviderCertificateCAChain: "",
				},
			}, mocks.mockDatabase, auth.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)
			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/{version}/login/saml/{%s}/metadata", api.URIPathVariableSSOProviderSlug), resources.ServeMetadata).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusOK {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {

				// find all validUntil fields and replace the time.Time value with
				// a persistent value.
				// matches 'validUntil=' followed by non-space/non-semicolon chars
				regex := regexp.MustCompile(`validUntil=[^ ;]*`)
				body := regex.ReplaceAllString(body, "validUntil=\"XXX\"")

				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}

func TestManagementResource_ServeSigningCertificate(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type expected struct {
		responseCode   int
		responseHeader http.Header
		responseBody   string
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: invalid provider ID - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/id/signing-certificate",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name: "Error: Database error db.GetSSOProviderById - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1/signing-certificate",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":500,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}]}`,
			},
		},
		{
			name: "Error: ssoProvider.SAMLProvider is nil - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1/signing-certificate",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Name:         "OIDC Provider",
					Slug:         "oidc-provider",
					Type:         model.SessionAuthProviderOIDC,
					SAMLProvider: nil,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"http_status":404,"timestamp":"0001-01-01T00:00:00Z","request_id":"","errors":[{"context":"","message":"resource not found"}]}`,
			},
		},
		{
			name: "Success: Served - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso-providers/1/signing-certificate",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderById(gomock.Any(), int32(1)).Return(model.SSOProvider{
					Name: "OIDC Provider",
					Slug: "oidc-provider",
					Type: model.SessionAuthProviderOIDC,
					SAMLProvider: &model.SAMLProvider{
						Name:            "name",
						DisplayName:     "display",
						IssuerURI:       "uri",
						SingleSignOnURI: "uri",
						MetadataXML:     []byte{},
						RootURIVersion:  model.SAMLRootURIVersion1,
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Disposition": []string{"attachment; filename=\"oidc-provider-signing-certificate.pem\""}, "Content-Type": []string{"text/plain; charset=utf-8"}},
				responseBody:   "-----BEGIN CERTIFICATE-----\n\n-----END CERTIFICATE-----",
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

			resource := v2auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, auth.Authorizer{}, nil, nil, nil)

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso-providers/{%s}/signing-certificate", api.URIPathVariableSSOProviderID), resource.ServeSigningCertificate).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusOK {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}

func TestManagementResource_SAMLLoginHandler(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockSAML     *samlmocks.MockService
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
			name: "Error: Nil SAML Provider, Redirect to Login with Error Message - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name:         "Test Provider",
					Slug:         "test-provider",
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: nil,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/login/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: auth.NewServiceProvider error, Redirect to Login with Error Message - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name: "Test Provider",
					Slug: "test-provider",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name: "name",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/login/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: SAML.MakeAuthenticationRequest, redirect error - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name: "Test Provider",
					Slug: "test-provider",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name:        "name",
						MetadataXML: []byte(validMetadataXML),
					},
				}, nil)
				mock.mockSAML.EXPECT().MakeAuthenticationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&saml.AuthnRequest{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/login/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: saml.HTTPRedirectBinding, redirect error - Found",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name: "Test Provider",
					Slug: "test-provider",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name: "name",
						MetadataXML: []byte(`<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://okta.com/saml">
		<IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
			<SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://okta.com/sso"/>
		</IDPSSODescriptor>
	</EntityDescriptor>`),
					},
				}, nil)
				mock.mockSAML.EXPECT().MakeAuthenticationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&saml.AuthnRequest{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"?SAMLRequest=fMuxqsJAEIXhVwnT3%2BtoORghYBPQRsXCbgkDBpKZdc8s%2BPiSVFbCXxw4fHukecrS1XjaRV9VEc17ngyyHC3VYuIJI8TSrJAY5NqdT7L7Z0mAlhjd6Ivk3yYXDx98oqY%2FtkTNXQtGt2X2QNXeEMmiJWbe%2Fq3dmGXtQZvDJwAA%2F%2F8%3D&SigAlg=http%3A%2F%2Fwww.w3.org%2F2001%2F04%2Fxmldsig-more%23rsa-sha256&Signature=y1tzz0uKcHIGTzUzyfo6wkJKJ7%2FLhD7vH6mmCV7W0eKlL58z6w3M%2BWCoGaBtXldzx4tSTB2RWEqCpYTw9gM%2BjoA9dBPLlzBxN0Sz97XxzgA9chdd4gTXyjcMHntNmsRqkrzcnLJmKJppL3LhIjmxt%2BDhya8MU0URHiZWGj%2BYxjFr0PQm5wOHHSjZH8J51r9lYPth4vO76XlYI64WefD1eH3RhRtskXC%2F7FQJ1KHpE6X1cbWjrGsPT7TdojDA8dJvV0nf9VUiO0CSgWFpIq%2BZZoYJDqsUiwvX0iR6z%2F3K4oNsbgp9NQ1lJD57tuNQVBx3YYvA6R52FQ64hSb2LjtpRQ%3D%3D"}},
			},
		},
		{
			name: "Success: saml.HTTPPostBinding, OK",
			buildRequest: func() *http.Request {
				request := http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/login",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name: "Test Provider",
					Slug: "test-provider",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name:        "name",
						MetadataXML: []byte(validMetadataXML),
					},
				}, nil)
				mock.mockSAML.EXPECT().MakeAuthenticationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&saml.AuthnRequest{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Security-Policy": []string{"default-src; script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; reflected-xss block; referrer no-referrer;"}, "Content-Type": []string{"text/html"}},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
				mockSAML:     samlmocks.NewMockService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2auth.NewManagementResource(config.Configuration{
				SAML: config.SAMLConfiguration{
					ServiceProviderCertificate:        ValidCert,
					ServiceProviderKey:                ValidKey,
					ServiceProviderCertificateCAChain: "",
				},
			}, mocks.mockDatabase, auth.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)
			resources.SAML = mocks.mockSAML
			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso/{%s}/login", api.URIPathVariableSSOProviderSlug), resources.SSOLoginHandler).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, _ := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
		})
	}
}

func TestManagementResource_SAMLCallbackHandler(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockSAML     *samlmocks.MockService
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
			name: "Error: Nil SAML Provider, Redirect to Login with Error Message - Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name:         "Test Provider",
					Slug:         "test-provider",
					Type:         model.SessionAuthProviderSAML,
					SAMLProvider: nil,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: auth.NewServiceProvider error, Redirect to Login with Error Message - Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
					Name: "Test Provider",
					Slug: "test-provider",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Name: "name",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Your+SSO+connection+failed+due+to+misconfiguration%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: parseResponse typed error, Failed to parse ACS response for provider - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{}, &saml.InvalidResponseError{
					PrivateErr: errors.New("error"),
				})
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+SSO+response%3A+Failed+to+parse+ACS+response+Authentication+failed"}},
			},
		},
		{
			name: "Error: parseResponse, Failed to parse ACS response for provider - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+SSO+response%3A+Failed+to+parse+ACS+response+error"}},
			},
		},
		{
			name: "Error: SAMLProvider.GetSAMLUserPrincipalNameFromAssertion - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Invalid+assertion%3A+no+valid+email+address+found"}},
			},
		},
		{
			name: "Error: CreateSSOSession Invalid User - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{
					AttributeStatements: []saml.AttributeStatement{{
						Attributes: []saml.Attribute{{
							FriendlyName: "uid",
							Name:         model.XMLSOAPClaimsEmailAddress,
							NameFormat:   model.ObjectIDAttributeNameFormat,
							Values: []saml.AttributeValue{{
								Type:  model.XMLTypeString,
								Value: "username",
							}},
						}},
					}},
				}, nil)
				mock.mockDatabase.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2)
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=Your+user+is+not+allowed%2C+please+contact+your+Administrator"}},
			},
		},
		{
			name: "Error: error creating user session - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.SessionTTLHours).Return(appcfg.Parameter{}, nil)
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					Serial: model.Serial{
						ID: int32(1),
					},
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{
					AttributeStatements: []saml.AttributeStatement{{
						Attributes: []saml.Attribute{{
							FriendlyName: "uid",
							Name:         model.XMLSOAPClaimsEmailAddress,
							NameFormat:   model.ObjectIDAttributeNameFormat,
							Values: []saml.AttributeValue{{
								Type:  model.XMLTypeString,
								Value: "username",
							}},
						}},
					}},
				}, nil)
				mock.mockDatabase.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2)
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: int32(1),
							Valid: true,
						},
					},
					SSOProvider: &model.SSOProvider{
						Name:   "name",
						Slug:   "slug",
						Serial: model.Serial{ID: int32(1)},
						SAMLProvider: &model.SAMLProvider{
							Serial:        model.Serial{ID: 1234},
							SSOProviderID: null.Int32From(int32(1)),
						},
					},
				}, nil)
				mock.mockDatabase.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).Return(model.UserSession{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui/login?error=We%E2%80%99re+having+trouble+connecting.+Please+check+your+internet+and+try+again."}},
			},
		},
		{
			name: "Success: jit saml user created + session created w/ auto provision config enabled - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.SessionTTLHours).Return(appcfg.Parameter{}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					Serial: model.Serial{
						ID: int32(1),
					},
					Config: model.SSOProviderConfig{
						AutoProvision: model.SSOProviderAutoProvisionConfig{
							Enabled:       true,
							RoleProvision: true,
							DefaultRoleId: int32(1),
						},
					},
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{
					AttributeStatements: []saml.AttributeStatement{{
						Attributes: []saml.Attribute{{
							FriendlyName: "uid",
							Name:         model.XMLSOAPClaimsEmailAddress,
							NameFormat:   model.ObjectIDAttributeNameFormat,
							Values: []saml.AttributeValue{{
								Type:  model.XMLTypeString,
								Value: "username",
							}},
						}},
					}},
				}, nil)
				mock.mockDatabase.EXPECT().GetAllRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Roles{
					{
						Permissions: model.Permissions{model.NewPermission("auth", "ManageUsers")},
					},
				}, nil)
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{}, database.ErrNotFound)
				mock.mockDatabase.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				mock.mockDatabase.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).AnyTimes()
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: int32(1),
							Valid: true,
						},
					},
					SSOProvider: &model.SSOProvider{
						Name:   "name",
						Slug:   "slug",
						Serial: model.Serial{ID: int32(1)},
						SAMLProvider: &model.SAMLProvider{
							Serial:        model.Serial{ID: 1234},
							SSOProviderID: null.Int32From(int32(1)),
						},
					},
				}, nil)
				mock.mockDatabase.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).Return(model.UserSession{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui"}, "Set-Cookie": []string{"token=token; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Success: jit saml user updated + session created w/ auto provision config enabled - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.SessionTTLHours).Return(appcfg.Parameter{}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					Serial: model.Serial{
						ID: int32(1),
					},
					Config: model.SSOProviderConfig{
						AutoProvision: model.SSOProviderAutoProvisionConfig{
							Enabled:       true,
							RoleProvision: true,
							DefaultRoleId: int32(1),
						},
					},
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{
					AttributeStatements: []saml.AttributeStatement{{
						Attributes: []saml.Attribute{{
							FriendlyName: "uid",
							Name:         model.XMLSOAPClaimsEmailAddress,
							NameFormat:   model.ObjectIDAttributeNameFormat,
							Values: []saml.AttributeValue{{
								Type:  model.XMLTypeString,
								Value: "username",
							}},
						}},
					}},
				}, nil)
				mock.mockDatabase.EXPECT().GetAllRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Roles{
					{
						Permissions: model.Permissions{model.NewPermission("auth", "ManageUsers")},
					},
				}, nil)
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{}, nil)
				mock.mockDatabase.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockDatabase.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).AnyTimes()
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: int32(1),
							Valid: true,
						},
					},
					SSOProvider: &model.SSOProvider{
						Name:   "name",
						Slug:   "slug",
						Serial: model.Serial{ID: int32(1)},
						SAMLProvider: &model.SAMLProvider{
							Serial:        model.Serial{ID: 1234},
							SSOProviderID: null.Int32From(int32(1)),
						},
					},
				}, nil)
				mock.mockDatabase.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).Return(model.UserSession{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui"}, "Set-Cookie": []string{"token=token; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
		{
			name: "Success: session created - Redirect to Login with Error Message",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/slug/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.SessionTTLHours).Return(appcfg.Parameter{}, nil).Times(2)
				mock.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "slug").Return(model.SSOProvider{
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
					Serial: model.Serial{
						ID: int32(1),
					},
				}, nil)
				mock.mockSAML.EXPECT().ParseResponse(gomock.Any(), gomock.Any(), nil).Return(&saml.Assertion{
					AttributeStatements: []saml.AttributeStatement{{
						Attributes: []saml.Attribute{{
							FriendlyName: "uid",
							Name:         model.XMLSOAPClaimsEmailAddress,
							NameFormat:   model.ObjectIDAttributeNameFormat,
							Values: []saml.AttributeValue{{
								Type:  model.XMLTypeString,
								Value: "username",
							}},
						}},
					}},
				}, nil)
				mock.mockDatabase.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2)
				mock.mockDatabase.EXPECT().LookupUser(gomock.Any(), "username").Return(model.User{
					SSOProviderID: null.Int32{
						NullInt32: sql.NullInt32{
							Int32: int32(1),
							Valid: true,
						},
					},
					SSOProvider: &model.SSOProvider{
						Name:   "name",
						Slug:   "slug",
						Serial: model.Serial{ID: int32(1)},
						SAMLProvider: &model.SAMLProvider{
							Serial:        model.Serial{ID: 1234},
							SSOProviderID: null.Int32From(int32(1)),
						},
					},
				}, nil)
				mock.mockDatabase.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).Return(model.UserSession{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusFound,
				responseHeader: http.Header{"Location": []string{"/api/v2/sso/slug/callback/ui"}, "Set-Cookie": []string{"token=token; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
				mockSAML:     samlmocks.NewMockService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2auth.NewManagementResource(config.Configuration{
				SAML: config.SAMLConfiguration{
					ServiceProviderCertificate:        ValidCert,
					ServiceProviderKey:                ValidKey,
					ServiceProviderCertificateCAChain: "",
				},
			}, mocks.mockDatabase, auth.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)
			resources.SAML = mocks.mockSAML
			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso/{%s}/callback", api.URIPathVariableSSOProviderSlug), http.HandlerFunc(resources.SSOCallbackHandler)).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, _ := test.ProcessResponse(t, response)

			// Cookies are regenerated in every response therefore the
			// the cookie attributes needed to be overwritten.
			header = test.ModifyCookieAttribute(header, "Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
			header = test.ModifyCookieAttribute(header, "token", "token")

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
		})
	}
}
