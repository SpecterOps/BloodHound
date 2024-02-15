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

package saml

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/headers"

	"github.com/specterops/bloodhound/src/api"
	apimocks "github.com/specterops/bloodhound/src/api/mocks"
	"github.com/specterops/bloodhound/src/auth/bhsaml"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	dbmocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/serde"

	"github.com/crewjam/saml"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProviderResource_createSessionFromAssertion(t *testing.T) {
	const (
		badUsername  = "bad"
		goodUsername = "good"
		goodJWT      = "fake"
	)

	var (
		goodUser = model.User{
			PrincipalName: goodUsername,
			SAMLProvider: &model.SAMLProvider{
				Serial: model.Serial{
					ID: 1,
				},
			},

			SAMLProviderID: null.Int32From(1),
		}

		mockCtrl          = gomock.NewController(t)
		mockDB            = dbmocks.NewMockDatabase(mockCtrl)
		mockAuthenticator = apimocks.NewMockAuthenticator(mockCtrl)
		resource          = ProviderResource{
			db:            mockDB,
			authenticator: mockAuthenticator,
			serviceProvider: bhsaml.ServiceProvider{
				Config: model.SAMLProvider{
					Serial: model.Serial{
						ID: 1,
					},
				},
			},
			cfg: config.Configuration{
				RootURL: serde.MustParseURL("https://example.com"),
			},
			writeAPIErrorResponse: func(request *http.Request, response http.ResponseWriter, statusCode int, message string) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(statusCode, message, request), response)
			},
		}
	)

	defer mockCtrl.Finish()

	var (
		expires               = time.Now().UTC().Add(time.Hour)
		response              = httptest.NewRecorder()
		expectedCookieContent = fmt.Sprintf("token=fake; Path=/; Expires=%s; Secure; SameSite=Strict", expires.Format(http.TimeFormat))

		testAssertion = &saml.Assertion{
			AttributeStatements: []saml.AttributeStatement{
				{
					Attributes: []saml.Attribute{
						{
							FriendlyName: "uid",
							Name:         bhsaml.XMLSOAPClaimsEmailAddress,
							NameFormat:   bhsaml.ObjectIDAttributeNameFormat,
							Values: []saml.AttributeValue{
								{
									Type:  bhsaml.XMLTypeString,
									Value: goodUsername,
								},
							},
						},
					},
				},
			},
		}
	)

	httpRequest, _ := http.NewRequestWithContext(context.WithValue(context.TODO(), ctx.ValueKey, &ctx.Context{Host: &resource.cfg.RootURL.URL}), http.MethodPost, "http://localhost", nil)

	// Test happy path
	mockDB.EXPECT().LookupUser(goodUsername).Return(goodUser, nil)
	mockAuthenticator.EXPECT().CreateSession(goodUser, gomock.Any()).Return(goodJWT, nil)

	resource.createSessionFromAssertion(httpRequest, response, expires, testAssertion)

	require.Equal(t, expectedCookieContent, response.Header().Get(headers.SetCookie.String()))
	require.Equal(t, "https://example.com/ui", response.Header().Get(headers.Location.String()))
	require.Equal(t, http.StatusFound, response.Code)

	// Change the assertion statement attribute to the bad username to assert we get a 403
	testAssertion.AttributeStatements[0].Attributes[0].Values[0].Value = badUsername

	mockDB.EXPECT().LookupUser(badUsername).Return(model.User{}, database.ErrNotFound)

	response = httptest.NewRecorder()

	resource.createSessionFromAssertion(httpRequest, response, expires, testAssertion)
	require.Equal(t, http.StatusForbidden, response.Code)

	// Change the db return to a user that isn't associated with a SAML Provider
	mockDB.EXPECT().LookupUser(badUsername).Return(model.User{}, nil)

	response = httptest.NewRecorder()

	resource.createSessionFromAssertion(httpRequest, response, expires, testAssertion)
	require.Equal(t, http.StatusForbidden, response.Code)

	// Change the db return to a user that isn't associated with this SAML Provider
	mockDB.EXPECT().LookupUser(badUsername).Return(model.User{
		SAMLProviderID: null.Int32From(2),
		SAMLProvider: &model.SAMLProvider{
			Serial: model.Serial{
				ID: 2,
			},
		},
	}, nil)

	response = httptest.NewRecorder()

	resource.createSessionFromAssertion(httpRequest, response, expires, testAssertion)
	require.Equal(t, http.StatusForbidden, response.Code)

	// Remove the assertion statement attribute for the username
	testAssertion.AttributeStatements[0].Attributes[0].Values = nil

	response = httptest.NewRecorder()

	resource.createSessionFromAssertion(httpRequest, response, expires, testAssertion)
	require.Equal(t, http.StatusBadRequest, response.Code)
}
