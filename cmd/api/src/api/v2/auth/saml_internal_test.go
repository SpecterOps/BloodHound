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
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	apimocks "github.com/specterops/bloodhound/cmd/api/src/api/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/serde"
	"github.com/specterops/bloodhound/packages/go/headers"
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
		testAuthenticator = api.NewAuthenticator(config.Configuration{}, mockDB, apimocks.NewMockAuthExtensions(mockCtrl))

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
		context.WithValue(context.TODO(), bhctx.ValueKey, &bhctx.Context{Host: &hostUrl.URL}),
		http.MethodPost,
		"http://localhost",
		nil,
	)

	t.Run("successfully create sso session", func(t *testing.T) {
		var (
			response              = httptest.NewRecorder()
			expectedCookiePattern = `token=.*; Path=/; Expires=.*; Secure; SameSite=Strict`
		)

		mockDB.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Times(2).Do(func(_ context.Context, log model.AuditLog) {
			require.Equal(t, model.AuditLogActionLoginAttempt, log.Action)
			require.Equal(t, username, log.Fields["username"])
			require.Equal(t, auth.ProviderTypeSAML, log.Fields["auth_type"])
		})
		mockDB.EXPECT().LookupUser(gomock.Any(), username).Return(user, nil)
		mockDB.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).Return(model.UserSession{}, nil)
		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.SessionTTLHours).Return(appcfg.Parameter{}, nil).AnyTimes()

		principalName, err := gothamSAML.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
		require.Nil(t, err)

		before := time.Now().UTC()
		testAuthenticator.CreateSSOSession(httpRequest, response, principalName, gothamSSO)
		after := time.Now().UTC()

		cookieHeader := response.Header().Get(headers.SetCookie.String())
		cookies := (&http.Response{Header: http.Header{"Set-Cookie": {cookieHeader}}}).Cookies()
		expectedExpiry := before.Add(appcfg.DefaultSessionTTLHours * time.Hour)
		tolerance := after.Sub(before) + time.Second

		require.Regexp(t, expectedCookiePattern, cookieHeader)
		require.WithinDuration(t, expectedExpiry, cookies[0].Expires, tolerance)
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

func TestAuth_SAMLLogging(t *testing.T) {
	var (
		testCases = []struct {
			name                   string
			principalAttributeName string
			principalValue         string
			groupMembership        string
			employeeID             string
			expectedPrincipalName  string
			expectedError          error
		}{
			{
				name:                   "principal found with group and employee attributes",
				principalAttributeName: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
				principalValue:         "alice123@example.invalid",
				groupMembership:        "Finance-Admins",
				employeeID:             "E10427",
				expectedPrincipalName:  "alice123@example.invalid",
			},
			{
				name:                   "principal missing with group and employee attributes",
				principalAttributeName: "uid",
				principalValue:         "bob123@example.invalid",
				groupMembership:        "Payroll-Readers",
				employeeID:             "E20853",
				expectedError:          model.ErrSAMLAssertion,
			},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				logBuffer      bytes.Buffer
				previousLogger = slog.Default()
				samlProvider   = model.SAMLProvider{}
				testAssertion  = &saml.Assertion{
					AttributeStatements: []saml.AttributeStatement{{
						Attributes: []saml.Attribute{
							{
								Name:       testCase.principalAttributeName,
								NameFormat: model.ObjectIDAttributeNameFormat,
								Values: []saml.AttributeValue{{
									Type:  model.XMLTypeString,
									Value: testCase.principalValue,
								}},
							},
							{
								Name:       "groups",
								NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
								Values: []saml.AttributeValue{{
									Type:  model.XMLTypeString,
									Value: testCase.groupMembership,
								}},
							},
							{
								Name:       "employeeNumber",
								NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
								Values: []saml.AttributeValue{{
									Type:  model.XMLTypeString,
									Value: testCase.employeeID,
								}},
							},
						},
					}},
				}
			)

			t.Cleanup(func() { slog.SetDefault(previousLogger) })
			// enable debug logging so this catches raw values logged at any standard level
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})))

			principalName, err := samlProvider.GetSAMLUserPrincipalNameFromAssertion(testAssertion)
			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.expectedPrincipalName, principalName)
			require.Contains(t, logBuffer.String(), `"name":"`+testCase.principalAttributeName+`"`)
			require.Contains(t, logBuffer.String(), `"name":"groups"`)
			require.Contains(t, logBuffer.String(), `"name":"employeeNumber"`)
			require.Contains(t, logBuffer.String(), `"name_format":"`+model.ObjectIDAttributeNameFormat+`"`)
			require.Contains(t, logBuffer.String(), `"name_format":"urn:oasis:names:tc:SAML:2.0:attrname-format:basic"`)
			require.NotContains(t, logBuffer.String(), testCase.principalValue)
			require.NotContains(t, logBuffer.String(), testCase.groupMembership)
			require.NotContains(t, logBuffer.String(), testCase.employeeID)
		})
	}
}
