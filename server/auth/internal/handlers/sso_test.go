// Copyright 2026 Specter Ops, Inc.
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

package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/auth/internal/handlers"
	handlermocks "github.com/specterops/bloodhound/server/auth/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/auth/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlers_ListAuthProviders(t *testing.T) {
	var (
		queryFilters = params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "Example"}}}
		sortItems    = params.SortItems{{Field: "name", Direction: params.Descending}}
		host         = &url.URL{Scheme: "https", Host: "example.com"}
		providers    = []services.SSOProvider{
			{
				ID:   1,
				Name: "OIDC",
				Slug: "oidc",
				Type: services.ProviderOIDC,
				OIDCProvider: &services.OIDCProvider{
					ClientID:      "client",
					Issuer:        "issuer",
					SSOProviderID: 1,
				},
			},
			{
				ID:   2,
				Name: "SAML",
				Slug: "saml",
				Type: services.ProviderSAML,
				SAMLProvider: &services.SAMLProvider{
					Name:           "saml",
					RootURIVersion: 1,
				},
			},
			{
				ID:   3,
				Name: "SAML v2",
				Slug: "saml-v2",
				Type: services.ProviderSAML,
				SAMLProvider: &services.SAMLProvider{
					Name:           "saml-v2",
					RootURIVersion: 2,
				},
			},
		}
		unexpectedErr = errors.New("private database error")
	)

	testCases := []struct {
		name       string
		providers  []services.SSOProvider
		err        error
		wantStatus int
		assertBody func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "Returns OIDC and SAML providers with their login and callback URLs",
			providers:  providers,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var envelope struct {
					Data []struct {
						Type        string          `json:"type"`
						LoginURI    string          `json:"login_uri"`
						CallbackURI string          `json:"callback_uri"`
						Config      json.RawMessage `json:"config"`
						Details     struct {
							ClientID              string `json:"client_id"`
							ServiceProviderACSURI string `json:"sp_acs_uri"`
						} `json:"details"`
					} `json:"data"`
				}

				require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
				require.Len(t, envelope.Data, 3)

				responseProvider := envelope.Data[0]
				assert.Equal(t, "https://example.com/api/v2/sso/oidc/login", responseProvider.LoginURI)
				assert.Equal(t, "https://example.com/api/v2/sso/oidc/callback", responseProvider.CallbackURI)
				assert.Equal(t, "OIDC", responseProvider.Type)
				assert.Equal(t, "client", responseProvider.Details.ClientID)
				assert.NotEmpty(t, responseProvider.Config)
				assert.Equal(t, "https://example.com/api/v1/login/saml/saml/acs", envelope.Data[1].Details.ServiceProviderACSURI)
				assert.Equal(t, "https://example.com/api/v2/sso/saml-v2/callback", envelope.Data[2].Details.ServiceProviderACSURI)
			},
		},
		{
			name:       "Returns a null payload when no providers are configured",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.JSONEq(t, `{"data":null}`, recorder.Body.String())
			},
		},
		{
			name:       "Maps unexpected error to 500 without exposing database details",
			err:        unexpectedErr,
			wantStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.NotContains(t, recorder.Body.String(), unexpectedErr.Error())
			},
		},
		{
			name:       "Maps invalid query error to 400",
			err:        services.ErrInvalidProviderQuery,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				service  = handlermocks.NewMockAuth(t)
				handler  = handlers.NewHandlersContainer(service)
				request  = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/sso-providers", nil), &bhctx.Context{Host: host, Filters: queryFilters, Sort: sortItems})
				recorder = httptest.NewRecorder()
			)

			service.EXPECT().ListAuthProviders(mock.Anything, queryFilters, sortItems).Return(testCase.providers, testCase.err)

			handler.ListAuthProviders(recorder, request)

			assert.Equal(t, testCase.wantStatus, recorder.Code, recorder.Body.String())
			if testCase.assertBody != nil {
				testCase.assertBody(t, recorder)
			}
		})
	}
}
