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
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/crewjam/saml"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	bhceauth "github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	samlmocks "github.com/specterops/bloodhound/cmd/api/src/services/saml/mocks"

	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	ValidCert = `-----BEGIN CERTIFICATE-----
MIIDrzCCApegAwIBAgIUG+lLMPpkTfwVe5fVbBhWe/cIxrgwDQYJKoZIhvcNAQEL
BQAwfzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcM
B1NlYXR0bGUxGjAYBgNVBAoMEVNwZWN0ZXIgT3BzLCBJbmMuMQwwCgYDVQQDDANC
b2IxHzAdBgkqhkiG9w0BCQEWEHNwYW1AZXhhbXBsZS5jb20wIBcNMjUwNjAyMTY1
ODMzWhgPMzAyNDEwMDMxNjU4MzNaMH8xCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApX
YXNoaW5ndG9uMRAwDgYDVQQHDAdTZWF0dGxlMRowGAYDVQQKDBFTcGVjdGVyIE9w
cywgSW5jLjEMMAoGA1UEAwwDQm9iMR8wHQYJKoZIhvcNAQkBFhBzcGFtQGV4YW1w
bGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5KMZPTuMDMIS
3N24Y2l6FLkbSZnG/Xa16MiSRtuvK6qExkVZ0Jr/7kDJ0F9eOWe6SgRx7/o/olnK
C3vs5HwduSoQzLiFH7T8FRGEOlmXbCtfFeWQzL0PwvoX7YQDP4d1IlEdznMWieXM
73hKL/UonrydrsJt6zeyhS0suAZz+UMpocsc9xtFtdaZBVSnSjqrK+PoT2V2WYUU
dnrBh9xnJ1aIX78PUoU8fF8YviJ9Anqe+sAsLX0nOVclj4LNGX+LhFoQcR4Dkxol
CNmLozD091Xvodhrst0kYKtgpJl8wFJ1ULeyvCvXKR3fNKseM7DkTBd5c3fZORZp
KRrs8UNc1wIDAQABoyEwHzAdBgNVHQ4EFgQUlFcqpLK4byN5z7dGyNpLjV+/sMIw
DQYJKoZIhvcNAQELBQADggEBALRZSG5BlQcMZN9kZCBVqUunCQ96cgCZUp77b0QE
PXsSn+KEj+br224DLmKiZ1z7wHNQefjEtAxcajS5ZTapOAekMo9Tnc5MUvAGh4cF
Imcn1b17u2IcMHAXOQNi6exBWToDKVAYwCpxQ2x9ELHs0mU1HC/HrkqLIAbQjoBL
0ZK55euoLJQ35s7M9uNLv4sBmXLKdScZiDsQgvTwSyFGVDrzOnkq7IsAfHZ23vwn
dY11/x7dQIqcdNmN/rTSlx0Gz0SArtOvIjIeMtvmv8KAPCjU+AvTGmNXZWbvr5lF
JOf46ORopwpJtTijhzvihLvBrLm6UdrLuosbvXuJGqk2EEM=
-----END CERTIFICATE-----
`
	ValidKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDkoxk9O4wMwhLc
3bhjaXoUuRtJmcb9drXoyJJG268rqoTGRVnQmv/uQMnQX145Z7pKBHHv+j+iWcoL
e+zkfB25KhDMuIUftPwVEYQ6WZdsK18V5ZDMvQ/C+hfthAM/h3UiUR3OcxaJ5czv
eEov9SievJ2uwm3rN7KFLSy4BnP5Qymhyxz3G0W11pkFVKdKOqsr4+hPZXZZhRR2
esGH3GcnVohfvw9ShTx8Xxi+In0Cep76wCwtfSc5VyWPgs0Zf4uEWhBxHgOTGiUI
2YujMPT3Ve+h2Guy3SRgq2CkmXzAUnVQt7K8K9cpHd80qx4zsORMF3lzd9k5Fmkp
GuzxQ1zXAgMBAAECggEARNBmD0z12P0sijddgOZFLSmNcfiLsMvi8l4z0IncTisz
bS2AW83bC82KMGITzPlQU2jFFjJeprGZox04boiAtbNYfRVoU+O4H2s3PgyrC45+
PuvqSgT5UnjNbNpX0+4kLiD19KYk+Xol1UmCIq8J+8TPPMMeLDaGT5kKJZUjoLip
YBXM4wlmK45nO/nffoo09wDkgSQ27/p6h53/HnbwCcC/cRJsZ7EGejZ7EH82gORK
F8PyplbQKm6spjwsgcU2G193ZNowmLFPacuhKCpDbB44HBFzXC7B9xYD3YsY2HmA
0C27WI98xlKxSmjKtdHgZD740V5i96vWQjdea3A6CQKBgQD/CvcxGCElZ52TE5rl
TkX3cKVIUkE02l/P7lZRMm+XFCh4qw9d+kky6etSdCphIWfMuyh2tGnC0FNSaQ6n
Q/avBg+6O09izxZVaupp2PVfGmA67Jg2SYYE7ABFXFCb6qj3N9O4Bn4LHuaXIGIW
FP1TR3pIjTQ/vMDIe5obDuNmRQKBgQDlfsNzu0hYOvSNNre6HSib/bjcx4tUZ10Y
K5K7rjxiN2EuUwZykP8nPNwhdn6hbRZ36uuv43RBwSqrEuxlgdPGR5QpRk7jYrd1
F0GSXiSKgITXoRnaDEUynSylKIA9Qn6vlpRtdhzSxzcQ4WvxNHwRX5NGTzssE5nU
kJ41MqoGawKBgQDgyiZzZAQa9seA0V/Nyf6LCAL1ymHklrCqETSNHnoSW9cL/CFw
QGBx+pDJvM95irr1TORuM7ef2IQH98bNkG6Fdz83cn0W5tWVdcWkg3BJYXL9nHjQ
KF9ySRw4BhSaR+qi8tattTM01AiDnSw2sEtTMoXKGoK5xsDYM3Dxdl7hTQKBgQCj
d0SO9dKVDgFNaLE7fzOC0RnRIM1MpId6BOdyiav3JY0yKu9HwaIM99uwdi/CmepM
JmgUk8YmZAoZatQ5hV0sOaX+NFdSvekBHTyWnjoW8W4uDVFVsDHF2JCJX6zgdbG5
Ll+xDFWBiWbevkJdv82zrkk/5oW2YovLDeuy5tCW2wKBgQDSjm+4IDV/7OQE1DRI
jO+N7YByHDnepPp4BPOfGxkSDJQPwsmsoj5B0J0tsXkKpLFUZMxJZEbSCnYBj9K7
q6U+h0JTO081nY4wb43VkFIGuE2HeBNtghNx5duE0fR0Ao0ja4Vf/sYbainUThaS
SMRMGd2G85dGQI0qXDJBWXabpA==
-----END PRIVATE KEY-----
`
)

func TestManagementResource_ListAuthProviders(t *testing.T) {
	const endpoint = "/api/v2/sso-providers"

	var (
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
		reqCtx               = &ctx.Context{Host: &url.URL{}}

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
			SQLString: "name = 'OIDC Provider 1'",
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
		ssoDeleteURL         = "/api/v2/sso-providers/%s"
		mockCtrl             = gomock.NewController(t)
		resources, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
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
			WithContext(&ctx.Context{AuthCtx: bhceauth.Context{
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
		mockCtrl     = gomock.NewController(t)
		_, mockDB, _ = apitest.NewAuthManagementResource(mockCtrl)
		testCtx      = context.Background()

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

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockSAML     *samlmocks.MockService
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
			name: "Error: SSO Provider not found - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/non-existent-provider/login",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "non-existent-provider").Return(model.SSOProvider{}, database.ErrNotFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"resource not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: Database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/provider-db-error/login",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider-db-error").Return(model.SSOProvider{}, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: Unsupported provider type - Not Implemented",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/unsupported-provider/login",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "unsupported-provider").Return(model.SSOProvider{
					Type: 999,
					Name: "Unsupported Provider",
					Slug: "unsupported-provider",
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotImplemented,
				responseBody:   `{"errors":[{"context":"","message":"All good things to those who wait. Not implemented."}],"http_status":501,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: SAML provider type - OK",
			buildRequest: func() *http.Request {
				ssoProviderSlug := "saml-provider"

				req, err := http.NewRequest("GET", fmt.Sprintf("/api/v2/sso/{%s}/login", ssoProviderSlug), nil)
				require.NoError(t, err)

				vars := map[string]string{
					api.URIPathVariableSSOProviderSlug: "saml-provider",
				}
				req = mux.SetURLVars(req, vars)

				req = req.WithContext(ctx.Set(req.Context(), &ctx.Context{Host: &url.URL{Host: "loremipsum"}}))

				return req
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				ssoProvider := model.SSOProvider{
					Type: 1,
					Name: "SAML Provider",
					Slug: "saml-provider",
					SAMLProvider: &model.SAMLProvider{
						Name:        "SAML Provider",
						DisplayName: "SAML Provider",
						MetadataXML: []byte(validMetadataXML),
					},
				}

				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), gomock.Any()).Return(ssoProvider, nil)
				mocks.mockSAML.EXPECT().MakeAuthenticationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&saml.AuthnRequest{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   "<!DOCTYPE html>\n<html>\n<body>\n<form method=\"post\" action=\"\" id=\"SAMLRequestForm\"><input type=\"hidden\" name=\"SAMLRequest\" value=\"value\" /><input type=\"hidden\" name=\"RelayState\" value=\"\" /><input id=\"SAMLSubmitButton\" type=\"submit\" value=\"Submit\" /></form><script>document.getElementById('SAMLSubmitButton').style.visibility=\"hidden\";document.getElementById('SAMLRequestForm').submit();</script>\n</body>\n</html>\n",
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

			resources := auth.NewManagementResource(config.Configuration{
				SAML: config.SAMLConfiguration{
					ServiceProviderCertificate:        ValidCert,
					ServiceProviderKey:                ValidKey,
					ServiceProviderCertificateCAChain: "",
				},
			}, mocks.mockDatabase, bhceauth.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)
			resources.SAML = mocks.mockSAML
			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso/{%s}/login", api.URIPathVariableSSOProviderSlug), resources.SSOLoginHandler).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusOK {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				re := regexp.MustCompile(`(<input[^>]+name=["']SAMLRequest["'][^>]*value=["'])[^\"]*(")`)
				// Value in html response is regenerated every time therefore it
				// needed to be overwritten to test the contract.
				updatedBody := re.ReplaceAllString(body, fmt.Sprintf(`${1}%s$2`, "value"))
				assert.Equal(t, testCase.expected.responseBody, updatedBody)
			}
		})
	}
}

func TestManagementResource_SSOCallbackHandler(t *testing.T) {
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
			name: "Error: SSO Provider not found - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/non-existent-provider/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "non-existent-provider").Return(model.SSOProvider{}, database.ErrNotFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"resource not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`},
		},
		{
			name: "Error: Database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/provider-db-error/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "provider-db-error").Return(model.SSOProvider{}, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: Unsupported provider type - Not Implemented",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/sso/unsupported-provider/callback",
					},
					Method: http.MethodGet,
				}

				bhContext := &ctx.Context{
					Host: request.URL,
				}
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhContext))
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetSSOProviderBySlug(gomock.Any(), "unsupported-provider").Return(model.SSOProvider{
					Type: 999,
					Name: "Unsupported Provider",
					Slug: "unsupported-provider",
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusNotImplemented,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"All good things to those who wait. Not implemented."}],"http_status":501,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
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

			resource := auth.NewManagementResource(config.Configuration{}, mocks.mockDatabase, bhceauth.NewAuthorizer(mocks.mockDatabase), api.NewAuthenticator(config.Configuration{}, mocks.mockDatabase, nil), nil, nil)
			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/sso/{%s}/callback", api.URIPathVariableSSOProviderSlug), http.HandlerFunc(resource.SSOCallbackHandler)).Methods(request.Method)
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
