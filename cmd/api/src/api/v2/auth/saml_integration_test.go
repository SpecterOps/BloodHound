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

//go:build integration

package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2auth "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestManagementResource_SAMLCallbackHandler_AssertionReplay checks that the callback
// rejects reuse of a signed assertion even with a new response ID and injected ID namespace
func TestManagementResource_SAMLCallbackHandler_AssertionReplay(t *testing.T) {
	const (
		providerID               int32 = 1
		principalName                  = "user@sp.test"
		firstAssertionID               = "_first_assertion"
		freshAssertionID               = "_fresh_assertion"
		initialResponseID              = "_initial_response"
		controlResponseID              = "_control_response"
		modifiedResponseID             = "_modified_response"
		freshResponseID                = "_fresh_response"
		injectedIDNamespaceValue       = "_6789"
	)
	var (
		now         = time.Now().UTC()
		hostURL     = url.URL{Scheme: "https", Host: "sp.test"}
		metadataURL = url.URL{Scheme: "https", Host: "idp.test", Path: "/metadata"}
		certificate = x509.Certificate{
			SerialNumber: big.NewInt(1),
			NotBefore:    now.Add(-time.Hour),
			NotAfter:     now.Add(time.Hour),
		}
	)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	certificateDER, err := x509.CreateCertificate(rand.Reader, &certificate, &certificate, &key.PublicKey, key)
	require.NoError(t, err)
	signingCertificate, err := x509.ParseCertificate(certificateDER)
	require.NoError(t, err)

	idp := &saml.IdentityProvider{
		Key:             key,
		Certificate:     signingCertificate,
		MetadataURL:     metadataURL,
		SSOURL:          url.URL{Scheme: "https", Host: "idp.test", Path: "/sso"},
		SignatureMethod: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
	}
	metadataXML, err := xml.Marshal(idp.Metadata())
	require.NoError(t, err)
	ssoProviderRecord := model.SSOProvider{
		Serial: model.Serial{ID: providerID},
		Name:   "Test SAML Provider",
		Type:   model.SessionAuthProviderSAML,
		SAMLProvider: &model.SAMLProvider{
			Name:           "test-idp",
			RootURIVersion: model.SAMLRootURIVersion2,
			IssuerURI:      metadataURL.String(),
			MetadataXML:    metadataXML,
		},
	}
	ssoProviderRecord.SAMLProvider.FormatSAMLProviderURLs(hostURL)
	callbackURL := ssoProviderRecord.SAMLProvider.ServiceProviderACSURI.String()

	newSignedAssertion := func(assertionID string) *etree.Element {
		t.Helper()

		signingRequest := &saml.IdpAuthnRequest{
			IDP:             idp,
			Now:             now,
			SPSSODescriptor: &saml.SPSSODescriptor{},
			ACSEndpoint:     &saml.IndexedEndpoint{Location: callbackURL},
			Assertion: &saml.Assertion{
				ID:           assertionID,
				Version:      "2.0",
				IssueInstant: now,
				Issuer:       saml.Issuer{Value: metadataURL.String()},
				Subject: &saml.Subject{
					NameID: &saml.NameID{Value: principalName},
					SubjectConfirmations: []saml.SubjectConfirmation{{
						Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
						SubjectConfirmationData: &saml.SubjectConfirmationData{
							Recipient:    callbackURL,
							NotOnOrAfter: now.Add(time.Hour),
						},
					}},
				},
				Conditions: &saml.Conditions{
					NotBefore:    now.Add(-time.Minute),
					NotOnOrAfter: now.Add(time.Hour),
					AudienceRestrictions: []saml.AudienceRestriction{{
						Audience: saml.Audience{Value: ssoProviderRecord.SAMLProvider.ServiceProviderIssuerURI.String()},
					}},
				},
				AttributeStatements: []saml.AttributeStatement{{
					Attributes: []saml.Attribute{{
						Name: model.XMLSOAPClaimsEmailAddress,
						Values: []saml.AttributeValue{{
							Type:  model.XMLTypeString,
							Value: principalName,
						}},
					}},
				}},
			},
		}
		require.NoError(t, signingRequest.MakeAssertionEl())
		return signingRequest.AssertionEl
	}

	// response wraps the assertion, keeping it unsigned to test the unsigned response with signed assertions case
	buildUnsignedResponse := func(responseID string, signedAssertion *etree.Element, withInjectedIDNamespace bool) string {
		t.Helper()

		response := saml.Response{
			ID:           responseID,
			Version:      "2.0",
			IssueInstant: now,
			Destination:  callbackURL,
			Issuer:       &saml.Issuer{Value: metadataURL.String()},
			Status:       saml.Status{StatusCode: saml.StatusCode{Value: saml.StatusSuccess}},
		}

		responseElement := response.Element()
		if withInjectedIDNamespace {
			// the unsigned <ID:m> extension element declares xmlns:ID with a value that may be mistaken for the signed assertion's ID
			extensions := etree.NewElement("samlp:Extensions")
			idPrefixedExtensionElement := extensions.CreateElement("ID:m")
			idPrefixedExtensionElement.CreateAttr("xmlns:ID", injectedIDNamespaceValue)
			responseElement.InsertChildAt(1, extensions)
		}
		responseElement.AddChild(signedAssertion.Copy())

		document := etree.NewDocument()
		document.SetRoot(responseElement)
		rawXML, err := document.WriteToBytes()
		require.NoError(t, err)
		return base64.StdEncoding.EncodeToString(rawXML)
	}

	// this signed assertion is reused in the initial login and both replay responses
	signedAssertion := newSignedAssertion(firstAssertionID)
	initialResponse := buildUnsignedResponse(initialResponseID, signedAssertion, false)
	controlResponse := buildUnsignedResponse(controlResponseID, signedAssertion, false)
	modifiedResponse := buildUnsignedResponse(modifiedResponseID, signedAssertion, true)
	freshResponse := buildUnsignedResponse(freshResponseID, newSignedAssertion(freshAssertionID), false)

	var (
		controller         = gomock.NewController(t)
		mockDatabase       = mocks.NewMockDatabase(controller)
		consumedResponses  = map[string]struct{}{}
		consumedAssertions = map[string]struct{}{}
		sessionWrites      int
		replayRejected     bool
		submittedResponse  string
		submittedAssertion string
	)

	mockDatabase.EXPECT().
		CreateSAMLConsumedIdentifiers(gomock.Any(), providerID, metadataURL.String(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int32, _, responseID, assertionID string, _ time.Time) error {
			submittedResponse, submittedAssertion = responseID, assertionID
			// consumed SAMLResponse and assertion IDs are stored in these in-memory maps across subtests, simulating
			// the saml_consumed_identifiers table used to detect replays
			_, responseConsumed := consumedResponses[responseID]
			_, assertionConsumed := consumedAssertions[assertionID]
			replayRejected = responseConsumed || assertionConsumed
			if replayRejected {
				return database.ErrSAMLIdentifierAlreadyConsumed
			}

			consumedResponses[responseID] = struct{}{}
			consumedAssertions[assertionID] = struct{}{}
			return nil
		}).AnyTimes()
	mockDatabase.EXPECT().LookupUser(gomock.Any(), principalName).Return(model.User{
		PrincipalName: principalName,
		SSOProviderID: null.Int32From(providerID),
	}, nil).AnyTimes()
	mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.SessionTTLHours).Return(appcfg.Parameter{
		Key: appcfg.SessionTTLHours,
		Value: types.JSONBObject{Object: appcfg.SessionTTLHoursParameter{
			Hours: appcfg.DefaultSessionTTLHours,
		}},
	}, nil).AnyTimes()
	mockDatabase.EXPECT().CreateAuditLog(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockDatabase.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, session model.UserSession) (model.UserSession, error) {
			sessionWrites++
			session.ID = int64(sessionWrites)
			session.CreatedAt = now
			return session, nil
		}).AnyTimes()

	cfg := config.Configuration{SAML: config.SAMLConfiguration{
		ServiceProviderCertificate: ValidCert,
		ServiceProviderKey:         ValidKey,
	}}
	cfg.Crypto.JWT.SetSigningKeyBytes([]byte("test-session-signing-key-with-32-bytes"))
	resource := v2auth.NewManagementResource(cfg, mockDatabase, auth.NewAuthorizer(mockDatabase), api.NewAuthenticator(cfg, mockDatabase, nil), nil, nil, nil)

	testCases := []struct {
		name        string
		response    string
		responseID  string
		assertionID string
		wantSession bool
	}{
		{"initial_login", initialResponse, initialResponseID, firstAssertionID, true},
		{"response_id_only_replay", controlResponse, controlResponseID, firstAssertionID, false},
		{"namespace_carrier_replay", modifiedResponse, modifiedResponseID, firstAssertionID, false},
		{"fresh_login", freshResponse, freshResponseID, freshAssertionID, true},
	}

	// These requests share the replay store and must run in order.
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			form := url.Values{"SAMLResponse": {testCase.response}}
			request := httptest.NewRequest(http.MethodPost, callbackURL, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request = request.WithContext((&bhctx.Context{Host: &hostURL}).ConstructGoContext())
			recorder := httptest.NewRecorder()
			previousSessionWrites := sessionWrites
			submittedResponse, submittedAssertion = "", ""
			replayRejected = false

			resource.SAMLCallbackHandler(recorder, request, ssoProviderRecord)

			var tokenCookie string
			for _, cookie := range recorder.Result().Cookies() {
				if cookie.Name == api.AuthTokenCookieName {
					tokenCookie = cookie.Value
				}
			}
			if testCase.wantSession {
				assert.Equal(t, 1, sessionWrites-previousSessionWrites, "valid login must create a session")
				assert.NotEmpty(t, tokenCookie)
			} else {
				assert.Equal(t, 0, sessionWrites-previousSessionWrites, "replay must not create a session")
				assert.Empty(t, tokenCookie)
				assert.NotContains(t, consumedResponses, testCase.responseID)
			}
			assert.Equal(t, testCase.responseID, submittedResponse, "response ID submitted to replay persistence")
			assert.Equal(t, testCase.assertionID, submittedAssertion, "signed assertion ID submitted to replay persistence")
			assert.Equal(t, !testCase.wantSession, replayRejected)
			assert.Equal(t, http.StatusFound, recorder.Code)
		})
	}
}
