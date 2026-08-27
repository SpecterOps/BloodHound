// Copyright 2025 Specter Ops, Inc.
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
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/crewjam/saml"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/saml.go -package=mocks . Service

// Service serves as a lightweight wrapper around the SAML package.
type Service interface {
	MakeAuthenticationRequest(serviceProvider saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*saml.AuthnRequest, error)
	ParseResponse(serviceProvider saml.ServiceProvider, req *http.Request, possibleRequestIDs []string) (*ValidatedResponse, error)
}

type Client struct{}

// ValidatedResponse represents a successfully parsed and validated SAML login: the verified assertion
// and the SAMLResponse's own ID and IssueInstant
type ValidatedResponse struct {
	Assertion            *saml.Assertion
	ResponseID           string
	ResponseIssueInstant time.Time
}

type assertionCounter struct {
	XMLName             xml.Name
	Assertions          []struct{} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	EncryptedAssertions []struct{} `xml:"urn:oasis:names:tc:SAML:2.0:assertion EncryptedAssertion"`
}

// MakeAuthenticationRequest abstracts creating an SAML authentication request using
// the HTTP-Redirect binding. It returns a URL that we will redirect the user to in order to start the auth process.
func (s *Client) MakeAuthenticationRequest(serviceProvider saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*saml.AuthnRequest, error) {
	return serviceProvider.MakeAuthenticationRequest(idpURL, binding, resultBinding)
}

// ParseResponse wraps the parsing and validation of the IdP's SAMLResponse in req and returns the verified assertion
// together with the SAMLResponse's own details in a ValidatedResponse.
func (s *Client) ParseResponse(serviceProvider saml.ServiceProvider, req *http.Request, possibleRequestIDs []string) (*ValidatedResponse, error) {
	var (
		fullResponse ValidatedResponse
		samlResponse saml.Response
		issuer       string
	)

	// Rejecting explicitly since BloodHound doesn't support HTTP-Artifact binding
	if req.Form.Get("SAMLart") != "" {
		return nil, errors.New("saml: HTTP-Artifact binding is not supported")
	}

	assertion, err := serviceProvider.ParseResponse(req, possibleRequestIDs)
	if err != nil {
		return nil, err
	}
	rawXMLSAMLResponse, err := base64.StdEncoding.DecodeString(req.PostForm.Get("SAMLResponse"))
	if err != nil {
		return nil, fmt.Errorf("saml: failed to decode SAMLResponse: %w", err)
	}
	if err := xml.Unmarshal(rawXMLSAMLResponse, &samlResponse); err != nil {
		return nil, fmt.Errorf("saml: failed to unmarshal SAMLResponse: %w", err)
	}
	if samlResponse.Issuer != nil {
		issuer = samlResponse.Issuer.Value
	}

	// It is rare but possible for a SAMLResponse to have more than one signed assertion (with an unsigned SAMLResponse);
	// crewjam validates them all, but returns only the first. Since discarded assertions are replayable,
	// reject a SAMLResponse with multiple assertions
	if assertionCount, err := countAssertions(rawXMLSAMLResponse); err != nil {
		return nil, fmt.Errorf("saml: failed to count assertions: %w", err)
	} else if assertionCount > 1 {
		return nil, fmt.Errorf("saml: SAMLResponse contains multiple assertions (issuer: %q)", issuer)
	}

	if samlResponse.ID == "" {
		return nil, fmt.Errorf("saml: SAMLResponse ID is empty (issuer: %q)", issuer)
	}
	if assertion == nil {
		return nil, fmt.Errorf("saml: assertion is nil (issuer: %q)", issuer)
	}
	if assertion.ID == "" {
		return nil, fmt.Errorf("saml: assertion ID is empty (issuer: %q)", issuer)
	}

	fullResponse = ValidatedResponse{
		Assertion:            assertion,
		ResponseID:           samlResponse.ID,
		ResponseIssueInstant: samlResponse.IssueInstant,
	}

	return &fullResponse, nil
}

// countAssertions returns how many SAML assertions (plaintext plus encrypted, in the SAML assertion namespace)
// appear in the raw response XML.
func countAssertions(rawResponseXML []byte) (int, error) {
	var counter assertionCounter
	if err := xml.Unmarshal(rawResponseXML, &counter); err != nil {
		return 0, err
	}

	return len(counter.Assertions) + len(counter.EncryptedAssertions), nil
}

// CalculateSAMLTimeExpiry returns the time when a consumed SAML identifier (SAMLResponse or assertion) can be safely deleted,
// since by then any replay attempt would be rejected as expired. It uses whichever identifier's IssueInstant is most recent
// plus the allowed delay (crewjam's MaxIssueDelay).
func CalculateSAMLTimeExpiry(responseIssueInstant, assertionIssueInstant time.Time) time.Time {
	var (
		responseExpiration  = responseIssueInstant.Add(saml.MaxIssueDelay)
		assertionExpiration = assertionIssueInstant.Add(saml.MaxIssueDelay)
		latest              = responseExpiration
	)

	if assertionExpiration.After(latest) {
		latest = assertionExpiration
	}
	return latest
}
