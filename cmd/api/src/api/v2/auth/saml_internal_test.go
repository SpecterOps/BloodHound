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

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/serde"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuth_CreateSSOSession(t *testing.T) {
	var (
		gothamSAML = model.SAMLProvider{Serial: model.Serial{ID: 1}}
		gothamSSO  = model.SSOProvider{SAMLProvider: &gothamSAML, Serial: model.Serial{ID: 1}, Type: model.SessionAuthProviderSAML}
		username   = "harls"
		user       = model.User{
			PrincipalName: username,
			SSOProviderID: null.Int32From(1),
		}

		mockCtrl          = gomock.NewController(t)
		mockDB            = dbmocks.NewMockDatabase(mockCtrl)
		testAuthenticator = api.NewAuthenticator(config.Configuration{}, mockDB, dbmocks.NewMockAuthContextInitializer(mockCtrl))

		hostUrl = serde.MustParseURL("https://example.com")

		testAssertion = &saml.Assertion{
			AttributeStatements: []saml.AttributeStatement{{
				Attributes: []saml.Attribute{{
					FriendlyName: "uid",
					Name:         model.XMLSOAPClaimsEmailAddress,
					NameFormat:   model.ObjectIDAttributeNameFormat,
					Values: []saml.AttributeValue{{
						Type:  model.XMLTypeString,
						Value: username,
					}},
				}},
			}},
		}
	)
	defer mockCtrl.Finish()

	httpRequest, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), ctx.ValueKey, &ctx.Context{Host: &hostUrl.URL}),
		http.MethodPost,
		"http://localhost",
		nil,
	)

	t.Run("successfully create sso session", func(t *testing.T) {
		var (
			response              = httptest.NewRecorder()
			expires               = time.Now().UTC()
			expectedCookieContent = fmt.Sprintf("token=.*; Path=/; Expires=%s; Secure; SameSite=Strict", expires.Format(http.TimeFormat))
		)

		mockDB.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2).Do(func(_ context.Context, log model.AuditLog) {
			require.Equal(t, model.AuditLogActionLoginAttempt, log.Action)
			require.Equal(t, username, log.Fields["username"])
			require.Equal(t, auth.ProviderTypeSAML, log.Fields["auth_type"])
		})
		mockDB.EXPECT().LookupUser(gomock.Any(), username).Return(user, nil)
		mockDB.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).Return(model.UserSession{}, nil)

		principalName, err := gothamSAML.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
		require.Nil(t, err)

		testAuthenticator.CreateSSOSession(httpRequest, response, principalName, gothamSSO)

		require.Regexp(t, expectedCookieContent, response.Header().Get(headers.SetCookie.String()))
		require.Equal(t, "https://example.com/ui", response.Header().Get(headers.Location.String()))
		require.Equal(t, http.StatusFound, response.Code)
	})

	t.Run("Forbidden 403 if user isn't in db", func(t *testing.T) {
		response := httptest.NewRecorder()

		mockDB.EXPECT().LookupUser(gomock.Any(), username).Return(model.User{}, database.ErrNotFound)
		mockDB.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2).Do(func(_ context.Context, log model.AuditLog) {
			require.Equal(t, model.AuditLogActionLoginAttempt, log.Action)
			require.Equal(t, username, log.Fields["username"])
			require.Equal(t, auth.ProviderTypeSAML, log.Fields["auth_type"])
			if log.Status == model.AuditLogStatusFailure {
				require.Equal(t, database.ErrNotFound, log.Fields["error"])
			}
		})
		principalName, err := gothamSAML.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
		require.Nil(t, err)

		testAuthenticator.CreateSSOSession(httpRequest, response, principalName, gothamSSO)

		require.Equal(t, http.StatusFound, response.Code)
		location, err := response.Result().Location()
		require.Nil(t, err)
		require.Equal(t, location.Query(), url.Values{"error": {"Your user is not allowed, please contact your Administrator"}})
	})

	t.Run("Forbidden 403 if user isn't associated with a SAML Provider", func(t *testing.T) {
		response := httptest.NewRecorder()

		mockDB.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2).Do(func(_ context.Context, log model.AuditLog) {
			require.Equal(t, model.AuditLogActionLoginAttempt, log.Action)
			require.Equal(t, username, log.Fields["username"])
			require.Equal(t, auth.ProviderTypeSAML, log.Fields["auth_type"])
			if log.Status == model.AuditLogStatusFailure {
				require.Equal(t, api.ErrUserNotAuthorizedForProvider, log.Fields["error"])
			}
		})

		mockDB.EXPECT().LookupUser(gomock.Any(), username).Return(model.User{}, nil)

		principalName, err := gothamSAML.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
		require.Nil(t, err)

		testAuthenticator.CreateSSOSession(httpRequest, response, principalName, gothamSSO)

		require.Equal(t, http.StatusFound, response.Code)
		location, err := response.Result().Location()
		require.Nil(t, err)
		require.Equal(t, location.Query(), url.Values{"error": {"Your user is not allowed, please contact your Administrator"}})
	})

	t.Run("Forbidden 403 if user isn't associated with specified SAML Provider", func(t *testing.T) {
		response := httptest.NewRecorder()

		mockDB.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2).Do(func(_ context.Context, log model.AuditLog) {
			require.Equal(t, model.AuditLogActionLoginAttempt, log.Action)
			require.Equal(t, username, log.Fields["username"])
			require.Equal(t, auth.ProviderTypeSAML, log.Fields["auth_type"])
			if log.Status == model.AuditLogStatusFailure {
				require.Equal(t, api.ErrUserNotAuthorizedForProvider.Error(), log.Fields["error"].(error).Error())
			}
		})
		mockDB.EXPECT().LookupUser(gomock.Any(), username).Return(model.User{
			SSOProviderID: null.Int32From(2),
			SSOProvider: &model.SSOProvider{
				SAMLProvider: &model.SAMLProvider{
					Serial: model.Serial{
						ID: 2,
					},
				},
			},
		}, nil)

		principalName, err := gothamSAML.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
		require.Nil(t, err)

		testAuthenticator.CreateSSOSession(httpRequest, response, principalName, gothamSSO)

		require.Equal(t, http.StatusFound, response.Code)
		location, err := response.Result().Location()
		require.Nil(t, err)
		require.Equal(t, location.Query(), url.Values{"error": {"Your user is not allowed, please contact your Administrator"}})
	})

	t.Run("Correctly fails with SAML assertion error if assertion is invalid", func(t *testing.T) {
		testAssertion.AttributeStatements[0].Attributes[0].Values = nil

		_, err := gothamSAML.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
		require.ErrorIs(t, err, model.ErrSAMLAssertion)
	})
}
