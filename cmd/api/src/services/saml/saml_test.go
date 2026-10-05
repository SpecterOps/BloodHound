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

package saml

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCountAssertions verifies that countAssertions counts every <Assertion> and <EncryptedAssertion>
// in the SAML assertion namespace, counting from the raw XML because crewjam's saml.Response only keeps one assertion.
func TestCountAssertions(t *testing.T) {
	const assertionNS = `xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"`

	testCases := []struct {
		name      string
		xml       string
		wantCount int
	}{
		{
			name: "single plaintext assertion",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ` + assertionNS + `>
				<saml:Assertion ID="a1"/>
			</samlp:Response>`,
			wantCount: 1,
		},
		{
			name: "two plaintext assertions",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ` + assertionNS + `>
				<saml:Assertion ID="a1"/>
				<saml:Assertion ID="a2"/>
			</samlp:Response>`,
			wantCount: 2,
		},
		{
			name: "one plaintext and one encrypted assertion",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ` + assertionNS + `>
				<saml:Assertion ID="a1"/>
				<saml:EncryptedAssertion/>
			</samlp:Response>`,
			wantCount: 2,
		},
		{
			name: "two encrypted assertions",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ` + assertionNS + `>
				<saml:EncryptedAssertion/>
				<saml:EncryptedAssertion/>
			</samlp:Response>`,
			wantCount: 2,
		},
		{
			name: "no assertions",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ` + assertionNS + `>
			</samlp:Response>`,
			wantCount: 0,
		},
		{
			name: "assertion in wrong namespace is not counted",
			xml: `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ` + assertionNS + ` xmlns:evil="urn:example:evil">
				<saml:Assertion ID="a1"/>
				<evil:Assertion ID="spoofed"/>
			</samlp:Response>`,
			wantCount: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := countAssertions([]byte(testCase.xml))

			require.NoError(t, err)
			assert.Equal(t, testCase.wantCount, count, "total assertion count")
		})
	}
}

// TestCalculateSAMLTimeExpiry verifies that CalculateSAMLTimeExpiry returns whichever IssueInstant + MaxIssueDelay is later.
func TestCalculateSAMLTimeExpiry(t *testing.T) {
	base := time.Date(2026, 8, 18, 23, 3, 0, 0, time.UTC)

	testCases := []struct {
		name                  string
		responseIssueInstant  time.Time
		assertionIssueInstant time.Time
		want                  time.Time
	}{
		{
			name:                  "response later than assertion uses response deadline",
			responseIssueInstant:  base.Add(40 * time.Second),
			assertionIssueInstant: base,
			want:                  base.Add(40 * time.Second).Add(saml.MaxIssueDelay),
		},
		{
			name:                  "assertion later than response uses assertion deadline",
			responseIssueInstant:  base,
			assertionIssueInstant: base.Add(40 * time.Second),
			want:                  base.Add(40 * time.Second).Add(saml.MaxIssueDelay),
		},
		{
			name:                  "equal instants use that instant plus delay",
			responseIssueInstant:  base,
			assertionIssueInstant: base,
			want:                  base.Add(saml.MaxIssueDelay),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := CalculateSAMLTimeExpiry(testCase.responseIssueInstant, testCase.assertionIssueInstant)

			assert.True(t, testCase.want.Equal(got), "expected %s, got %s", testCase.want, got)
		})
	}
}

// TestParseResponse_RejectsArtifactBinding verifies that ParseResponse rejects HTTP-Artifact
// requests (SAMLart in either the POST form or as a query parameter), since we don't support HTTP-Artifact binding.
func TestParseResponse_RejectsArtifactBinding(t *testing.T) {
	testCases := []struct {
		name    string
		request func() *http.Request
	}{
		{
			name: "SAMLart in POST form",
			request: func() *http.Request {
				form := url.Values{"SAMLart": {"artifact-id"}}
				req := httptest.NewRequest(http.MethodPost, "/api/v2/sso/slug/callback", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
		},
		{
			name: "SAMLart as a query parameter",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/sso/slug/callback?SAMLart=artifact-id", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				client Client
				req    = testCase.request()
			)
			require.NoError(t, req.ParseForm())

			validatedResponse, err := client.ParseResponse(saml.ServiceProvider{}, req, nil)

			require.ErrorContains(t, err, "HTTP-Artifact binding is not supported")
			assert.Nil(t, validatedResponse)
		})
	}
}
