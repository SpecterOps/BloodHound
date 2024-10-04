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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_ListAuthProviders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	resources, mockDB := apitest.NewAuthManagementResource(mockCtrl)
	defer mockCtrl.Finish()

	t.Run("successfully list auth providers without query parameters", func(t *testing.T) {
		ssoProviders := []model.SSOProvider{
			{
				Serial: model.Serial{ID: 1},
				Name:   "OIDC Provider 1",
				Slug:   "oidc-provider-1",
				Type:   model.SessionAuthProviderOIDC,
			},
			{
				Serial: model.Serial{ID: 2},
				Name:   "SAML Provider 1",
				Slug:   "saml-provider-1",
				Type:   model.SessionAuthProviderSAML,
			},
		}

		oidcProvider := model.OIDCProvider{
			SSOProviderID: 1,
			ClientID:      "client-id-1",
			Issuer:        "https://issuer1.com",
		}

		samlProvider := model.SAMLProvider{
			Serial:        model.Serial{ID: 2},
			Name:          "SAML Provider 1",
			DisplayName:   "SAML Provider One",
			IssuerURI:     "https://saml-issuer1.com",
			SSOProviderID: 2,
		}

		// default ordering and no filters
		mockDB.EXPECT().GetAllSSOProviders(
			gomock.Any(),
			"created_at",
			model.SQLFilter{SQLString: "", Params: nil},
		).Return(ssoProviders, nil)

		mockDB.EXPECT().GetOIDCProviderBySSOProviderID(gomock.Any(), 1).Return(oidcProvider, nil)
		mockDB.EXPECT().GetSAMLProviderBySSOProviderID(gomock.Any(), int32(2)).Return(samlProvider, nil)

		endpoint := "/api/v2/sso-providers"

		bhCtx := &ctx.Context{
			Host: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
		requestContext := context.WithValue(context.Background(), ctx.ValueKey, bhCtx)

		req, err := http.NewRequestWithContext(requestContext, "GET", endpoint, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "example.com"

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListAuthProviders).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("successfully list auth providers with sorting", func(t *testing.T) {
		ssoProviders := []model.SSOProvider{
			{
				Serial: model.Serial{ID: 2},
				Name:   "SAML Provider 1",
				Slug:   "saml-provider-1",
				Type:   model.SessionAuthProviderSAML,
			},
			{
				Serial: model.Serial{ID: 1},
				Name:   "OIDC Provider 1",
				Slug:   "oidc-provider-1",
				Type:   model.SessionAuthProviderOIDC,
			},
		}

		oidcProvider := model.OIDCProvider{
			SSOProviderID: 1,
			ClientID:      "client-id-1",
			Issuer:        "https://issuer1.com",
		}

		samlProvider := model.SAMLProvider{
			Serial:        model.Serial{ID: 2},
			Name:          "SAML Provider 1",
			DisplayName:   "SAML Provider One",
			IssuerURI:     "https://saml-issuer1.com",
			SSOProviderID: 2,
		}

		// sorting by name descending
		mockDB.EXPECT().GetAllSSOProviders(
			gomock.Any(),
			"name desc",
			model.SQLFilter{SQLString: "", Params: nil},
		).Return(ssoProviders, nil)

		mockDB.EXPECT().GetOIDCProviderBySSOProviderID(gomock.Any(), 1).Return(oidcProvider, nil)
		mockDB.EXPECT().GetSAMLProviderBySSOProviderID(gomock.Any(), int32(2)).Return(samlProvider, nil)

		endpoint := "/api/v2/sso-providers?sort_by=-name"

		bhCtx := &ctx.Context{
			Host: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
		requestContext := context.WithValue(context.Background(), ctx.ValueKey, bhCtx)

		req, err := http.NewRequestWithContext(requestContext, "GET", endpoint, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "example.com"

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/sso-providers", resources.ListAuthProviders).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("successfully list auth providers with filtering", func(t *testing.T) {
		ssoProviders := []model.SSOProvider{
			{
				Serial: model.Serial{ID: 1},
				Name:   "OIDC Provider 1",
				Slug:   "oidc-provider-1",
				Type:   model.SessionAuthProviderOIDC,
			},
		}

		oidcProvider := model.OIDCProvider{
			SSOProviderID: 1,
			ClientID:      "client-id-1",
			Issuer:        "https://issuer1.com",
		}

		// filtering by name
		mockDB.EXPECT().GetAllSSOProviders(
			gomock.Any(),
			"created_at",
			model.SQLFilter{
				SQLString: "name = ?",
				Params:    []interface{}{"OIDC Provider 1"},
			},
		).Return(ssoProviders, nil)

		mockDB.EXPECT().GetOIDCProviderBySSOProviderID(gomock.Any(), 1).Return(oidcProvider, nil)

		endpoint := "/api/v2/sso-providers?name=eq:OIDC Provider 1"

		bhCtx := &ctx.Context{
			Host: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
		requestContext := context.WithValue(context.Background(), ctx.ValueKey, bhCtx)

		req, err := http.NewRequestWithContext(requestContext, "GET", endpoint, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "example.com"

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/sso-providers", resources.ListAuthProviders).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("fail to list auth providers with invalid sort field", func(t *testing.T) {
		endpoint := "/api/v2/sso-providers?sort_by=invalid_field"

		bhCtx := &ctx.Context{
			Host: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
		requestContext := context.WithValue(context.Background(), ctx.ValueKey, bhCtx)

		req, err := http.NewRequestWithContext(requestContext, "GET", endpoint, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "example.com"

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/sso-providers", resources.ListAuthProviders).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("fail to list auth providers with invalid filter predicate", func(t *testing.T) {
		endpoint := "/api/v2/sso-providers?name=invalid_predicate:Provider"

		bhCtx := &ctx.Context{
			Host: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
		requestContext := context.WithValue(context.Background(), ctx.ValueKey, bhCtx)

		req, err := http.NewRequestWithContext(requestContext, "GET", endpoint, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "example.com"

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/sso-providers", resources.ListAuthProviders).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
