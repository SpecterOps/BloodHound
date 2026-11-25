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

package saml_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	bhceSAML "github.com/specterops/bloodhound/cmd/api/src/services/saml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseResponse verifies how a real crewjam ServiceProvider validates all SAMLResponse/assertion signing
// combinations before ParseResponse's validations take over
func TestParseResponse(t *testing.T) {
	tests := []struct {
		name         string
		buildRequest func(t *testing.T, idp testIdP) *http.Request
		wantAccepted bool
		errContains  string
	}{
		{
			name: "unsigned Response, unsigned assertion is rejected",
			buildRequest: func(t *testing.T, idp testIdP) *http.Request {
				encoded, _ := idp.buildUnsignedSAMLResponse(t, defaultResponseID,
					assertionSpec{id: "_assertion_id_a", signed: false})
				return newSAMLRequest(t, encoded)
			},
			// crewjam returns an 'Authentication failed' error, so verify only that it is rejected
			wantAccepted: false,
		},
		{
			name: "unsigned Response, signed assertion is accepted",
			buildRequest: func(t *testing.T, idp testIdP) *http.Request {
				encoded, _ := idp.buildUnsignedSAMLResponse(t, defaultResponseID)
				return newSAMLRequest(t, encoded)
			},
			wantAccepted: true,
		},
		{
			name: "signed Response, unsigned assertion is accepted",
			buildRequest: func(t *testing.T, idp testIdP) *http.Request {
				return newSAMLRequest(t, idp.buildSignedSAMLResponse(t, "_assertion_id_a", false))
			},
			wantAccepted: true,
		},
		{
			name: "signed Response, signed assertion is accepted",
			buildRequest: func(t *testing.T, idp testIdP) *http.Request {
				return newSAMLRequest(t, idp.buildSignedSAMLResponse(t, "_assertion_id_a", true))
			},
			wantAccepted: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				idp            = newTestIdP(t)
				req            = testCase.buildRequest(t, idp)
				validated, err = (&bhceSAML.Client{}).ParseResponse(idp.newServiceProvider(), req, nil)
			)

			if testCase.wantAccepted {
				require.NoError(t, err)
				require.NotNil(t, validated)
			} else {
				require.Error(t, err)
				assert.Nil(t, validated)
				if testCase.errContains != "" {
					assert.Contains(t, err.Error(), testCase.errContains)
				}
			}
		})
	}
}

// TestParseResponse_MultipleAssertions verifies the multiple-assertion rejection: the assertions pass crewjam's own validation,
// but a SAMLResponse carrying more than one is still rejected (regardless of signing).
func TestParseResponse_MultipleAssertions(t *testing.T) {
	tests := []struct {
		name       string
		assertions []assertionSpec
	}{
		{
			name: "two signed assertions are rejected",
			assertions: []assertionSpec{
				{id: "_assertion_id_a", signed: true},
				{id: "_assertion_id_b", signed: true},
			},
		},
		{
			name: "one signed and one unsigned assertion are rejected",
			assertions: []assertionSpec{
				{id: "_assertion_id_a", signed: true},
				{id: "_assertion_id_b", signed: false},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				idp        = newTestIdP(t)
				encoded, _ = idp.buildUnsignedSAMLResponse(t, defaultResponseID, testCase.assertions...)
				req        = newSAMLRequest(t, encoded)
			)

			validated, err := (&bhceSAML.Client{}).ParseResponse(idp.newServiceProvider(), req, nil)

			require.Error(t, err)
			assert.Nil(t, validated)
			assert.Contains(t, err.Error(), "multiple assertions")
		})
	}
}

// TestParseResponse_EmptyIdentifiers verifies ParseResponse's checks for an empty Response ID and empty assertion ID"
func TestParseResponse_EmptyIdentifiers(t *testing.T) {
	t.Run("empty Response ID is rejected", func(t *testing.T) {
		var (
			idp = newTestIdP(t)
			// setting up an unsigned SAMLResponse (thus its Response ID is never checked by crewjam),
			// so it reaches ParseResponse's empty Response ID guard
			encoded, _ = idp.buildUnsignedSAMLResponse(t, "",
				assertionSpec{id: "_assertion_id_a", signed: true})
			req = newSAMLRequest(t, encoded)
		)

		validated, err := (&bhceSAML.Client{}).ParseResponse(idp.newServiceProvider(), req, nil)

		require.Error(t, err)
		assert.Nil(t, validated)
		assert.Contains(t, err.Error(), "SAMLResponse ID is empty")
	})

	t.Run("empty assertion ID is rejected", func(t *testing.T) {
		var (
			idp = newTestIdP(t)
			// setting up a signed SAMLResponse containing an unsigned assertion whose ID is empty
			// (assertion ID not checked by crewjam), so it reaches ParseResponse's empty assertion ID guard
			req = newSAMLRequest(t, idp.buildSignedSAMLResponse(t, "", false))
		)

		validated, err := (&bhceSAML.Client{}).ParseResponse(idp.newServiceProvider(), req, nil)

		require.Error(t, err)
		assert.Nil(t, validated)
		assert.Contains(t, err.Error(), "assertion ID is empty")
	})
}

// TestParseResponse_ResponseFields verifies the SAMLResponse fields ParseResponse extracts again from the request/raw XML, which
// crewjam validates but discards.
func TestParseResponse_ResponseFields(t *testing.T) {
	var (
		idp                   = newTestIdP(t)
		encoded, issueInstant = idp.buildUnsignedSAMLResponse(t, defaultResponseID)
		req                   = newSAMLRequest(t, encoded)
	)

	validated, err := (&bhceSAML.Client{}).ParseResponse(idp.newServiceProvider(), req, nil)

	require.NoError(t, err)
	require.NotNil(t, validated)
	// check that og Response ID matches extracted
	assert.Equal(t, defaultResponseID, validated.ResponseID)
	// rounded to the millisecond so IssueInstant matches exactly after crewjam's round trip
	assert.True(t, issueInstant.Equal(validated.ResponseIssueInstant),
		"expected ResponseIssueInstant %s to equal %s", validated.ResponseIssueInstant, issueInstant)
}

// TestParseResponse_AssertionID verifies that ParseResponse reads a plaintext assertion's unnamespaced ID,
// rejects a namespace declaration in its place, and rejects duplicate unnamespaced ID attributes.
func TestParseResponse_AssertionID(t *testing.T) {
	const (
		assertionID              = "_real_assertion_id"
		misleadingNamespaceValue = "urn:example:misleading-assertion-id"
	)

	testCases := []struct {
		name                 string
		removeUnnamespacedID bool
		prependDuplicateID   bool
		wantAssertionID      string
		wantError            string
	}{
		{
			name:            "unnamespaced ID wins over namespace declaration",
			wantAssertionID: assertionID,
		},
		{
			name:                 "namespace declaration alone is rejected",
			removeUnnamespacedID: true,
			wantError:            "assertion is missing an unnamespaced ID attribute",
		},
		{
			name:               "duplicate unnamespaced IDs are rejected",
			prependDuplicateID: true,
			wantError:          "assertion contains multiple unnamespaced ID attributes",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				idp = newTestIdP(t)
				now = time.Now().Round(time.Millisecond).UTC()
			)
			acsURL, err := url.Parse(idp.acsURL)
			require.NoError(t, err)

			responseBuilder := saml.IdpAuthnRequest{
				IDP:             idp.idp,
				Now:             now,
				SPSSODescriptor: &saml.SPSSODescriptor{},
				ACSEndpoint:     &saml.IndexedEndpoint{Location: acsURL.String()},
				Assertion:       idp.buildAssertion(assertionID, now),
			}
			// supplying AssertionEl leaves it unsigned while MakeResponse signs the enclosing response
			responseBuilder.AssertionEl = responseBuilder.Assertion.Element()
			responseBuilder.AssertionEl.CreateAttr("xmlns:ID", misleadingNamespaceValue)
			if testCase.removeUnnamespacedID {
				require.NotNil(t, responseBuilder.AssertionEl.RemoveAttr("ID"))
			}
			require.NoError(t, responseBuilder.MakeResponse())

			document := etree.NewDocument()
			document.SetRoot(responseBuilder.ResponseEl)
			rawXML, err := document.WriteToBytes()
			require.NoError(t, err)
			if testCase.prependDuplicateID {
				const assertionStart = "<saml:Assertion "
				require.Contains(t, string(rawXML), assertionStart)
				// etree retains the last ID for crewjam validation; raw XML still contains both IDs
				rawXML = []byte(strings.Replace(string(rawXML), assertionStart,
					assertionStart+`ID="_shadow_assertion_id" `, 1))
			}
			request := newSAMLRequest(t, base64.StdEncoding.EncodeToString(rawXML))

			validated, err := (&bhceSAML.Client{}).ParseResponse(idp.newServiceProvider(), request, nil)

			if testCase.wantError != "" {
				require.Error(t, err)
				assert.Nil(t, validated)
				assert.Contains(t, err.Error(), testCase.wantError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, validated)
				assert.Equal(t, testCase.wantAssertionID, validated.AssertionID)
			}
		})
	}
}

// --- TEST SETUP ---

const defaultResponseID = "_response_id"

// newServiceProvider builds a SP that trusts the IdP's signing certificate and accepts IdP-initiated
// responses, so it will validate the signed assertions.
func (s testIdP) newServiceProvider() saml.ServiceProvider {
	var acsURL, _ = url.Parse(s.acsURL)

	return saml.ServiceProvider{
		EntityID:          s.spEntityID,
		AcsURL:            *acsURL,
		IDPMetadata:       s.idp.Metadata(),
		AllowIDPInitiated: true,
	}
}

// newSAMLRequest wraps a base64 SAMLResponse in an *http.Request (via PostForm), the way ParseResponse expects it
func newSAMLRequest(t *testing.T, samlResponse string) *http.Request {
	t.Helper()

	req, reqErr := http.NewRequest(http.MethodPost, "https://sp.test/callback", strings.NewReader(""))
	require.NoError(t, reqErr)

	req.PostForm = url.Values{"SAMLResponse": []string{samlResponse}}
	return req
}

type assertionSpec struct {
	id     string
	signed bool
}

// buildUnsignedSAMLResponse creates an unsigned samlp:Response with the given SAMLResponse ID and assertions
// then base64-encodes it as an IdP would over the HTTP-POST binding.
func (s testIdP) buildUnsignedSAMLResponse(t *testing.T, responseID string, assertions ...assertionSpec) (string, time.Time) {
	t.Helper()

	var issueInstant = time.Now().Round(time.Millisecond).UTC()

	if len(assertions) == 0 {
		// default base case is to have one SAMLResponse with one single assertion
		assertions = []assertionSpec{{id: "_assertion_id_a", signed: true}}
	}

	response := saml.Response{
		ID:           responseID,
		Version:      "2.0",
		IssueInstant: issueInstant,
		Issuer:       &saml.Issuer{Value: s.entityID},
		Status: saml.Status{
			StatusCode: saml.StatusCode{Value: saml.StatusSuccess},
		},
	}

	responseEl := response.Element()
	for _, spec := range assertions {
		responseEl.AddChild(s.buildAssertionEl(t, spec))
	}

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)

	rawXML, writeErr := doc.WriteToBytes()
	require.NoError(t, writeErr)

	return base64.StdEncoding.EncodeToString(rawXML), issueInstant
}

// buildSignedSAMLResponse creates a signed samlp:Response containing a single assertion via crewjam's MakeResponse.
// signedAssertion controls whether the inner assertion is individually signed too. A random Response ID is assigned.
func (s testIdP) buildSignedSAMLResponse(t *testing.T, assertionID string, signedAssertion bool) string {
	t.Helper()

	var (
		acsURL, _    = url.Parse(s.acsURL)
		authnRequest = saml.IdpAuthnRequest{
			IDP:             s.idp,
			Now:             time.Now().Round(time.Millisecond).UTC(),
			SPSSODescriptor: &saml.SPSSODescriptor{},
			ACSEndpoint:     &saml.IndexedEndpoint{Location: acsURL.String()},
			Assertion:       s.buildAssertion(assertionID, time.Now()),
		}
	)

	// initialize the XML assertion child so that crewjam's MakeResponse doesn't sign the assertion
	// (so we can test the unsigned assertion case)
	if !signedAssertion {
		authnRequest.AssertionEl = authnRequest.Assertion.Element()
	}

	require.NoError(t, authnRequest.MakeResponse())

	doc := etree.NewDocument()
	doc.SetRoot(authnRequest.ResponseEl)

	rawXML, writeErr := doc.WriteToBytes()
	require.NoError(t, writeErr)

	return base64.StdEncoding.EncodeToString(rawXML)
}

// buildAssertionEl builds a single assertion xml element
func (s testIdP) buildAssertionEl(t *testing.T, spec assertionSpec) *etree.Element {
	t.Helper()

	var authnRequest = saml.IdpAuthnRequest{
		IDP:             s.idp,
		Now:             time.Now(),
		SPSSODescriptor: &saml.SPSSODescriptor{}, // no encryption cert -- signed assertions stay plaintext
		Assertion:       s.buildAssertion(spec.id, time.Now()),
	}

	if !spec.signed {
		return authnRequest.Assertion.Element()
	}
	require.NoError(t, authnRequest.MakeAssertionEl())

	return authnRequest.AssertionEl
}

// buildAssertion builds a minimal SAML assertion the SP will accept: fresh timestamps/windows and matching
// issuer/recipient/audience.
func (s testIdP) buildAssertion(id string, now time.Time) *saml.Assertion {
	return &saml.Assertion{
		ID:           id,
		IssueInstant: now,
		Version:      "2.0",
		Issuer:       saml.Issuer{Value: s.entityID},
		Subject: &saml.Subject{
			NameID: &saml.NameID{Value: "user@sp.test"},
			SubjectConfirmations: []saml.SubjectConfirmation{{
				Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
				SubjectConfirmationData: &saml.SubjectConfirmationData{
					Recipient:    s.acsURL,
					NotOnOrAfter: now.Add(time.Hour),
				},
			}},
		},
		Conditions: &saml.Conditions{
			NotBefore:    now.Add(-time.Minute),
			NotOnOrAfter: now.Add(time.Hour),
			AudienceRestrictions: []saml.AudienceRestriction{{
				Audience: saml.Audience{Value: s.spEntityID},
			}},
		},
	}
}

// testIdP wraps a crewjam IdP so we can create assertions signed by crewjam itself.
type testIdP struct {
	idp        *saml.IdentityProvider
	entityID   string
	acsURL     string
	spEntityID string
}

// newTestIdP generates a fresh RSA key and self-signed certificate and builds a crewjam IdP whose
// Metadata() includes that certificate, so the real ServiceProvider.ParseResponse can validate the assertions we sign.
func newTestIdP(t *testing.T) testIdP {
	t.Helper()

	var (
		key, keyErr = rsa.GenerateKey(rand.Reader, 2048)
		metadataURL = url.URL{Scheme: "https", Host: "idp.test", Path: "/metadata"}
		ssoURL      = url.URL{Scheme: "https", Host: "idp.test", Path: "/sso"}
	)
	require.NoError(t, keyErr)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "idp.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	certDER, certErr := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, certErr)

	cert, parseErr := x509.ParseCertificate(certDER)
	require.NoError(t, parseErr)

	return testIdP{
		idp: &saml.IdentityProvider{
			Key:             key,
			Certificate:     cert,
			MetadataURL:     metadataURL,
			SSOURL:          ssoURL,
			SignatureMethod: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
		},
		entityID:   metadataURL.String(),
		acsURL:     "https://sp.test/callback",
		spEntityID: "https://sp.test/metadata",
	}
}
