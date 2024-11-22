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

package model

import (
	"errors"
	"net/url"
	"path"

	"github.com/crewjam/saml"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/serde"
)

const (
	ObjectIDAttributeNameFormat = "urn:oasis:names:tc:SAML:2.0:attrname-format:uri"
	ObjectIDEmail               = "urn:oid:0.9.2342.19200300.100.1.3"
	XMLTypeString               = "xs:string"
	XMLSOAPClaimsEmailAddress   = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
)

var (
	ErrSAMLAssertion = errors.New("SAML assertion error")
)

// SAMLRootURIVersion is required for payloads to match the ACS / Callback url configured on IDPs
// While the DB column root_uri_version has a default of 2, it is also hardcoded in the db method CreateSAMLIdentityProvider
type SAMLRootURIVersion int

var (
	SAMLRootURIVersion1 SAMLRootURIVersion = 1 // "/v2/login/saml/{slug}/"
	SAMLRootURIVersion2 SAMLRootURIVersion = 2 // "/v2/sso/{slug}/"
)

type SAMLProvider struct {
	Name            string             `json:"name" gorm:"unique;index"`
	DisplayName     string             `json:"display_name"`
	IssuerURI       string             `json:"idp_issuer_uri"`
	SingleSignOnURI string             `json:"idp_sso_uri"`
	MetadataXML     []byte             `json:"-"`
	RootURIVersion  SAMLRootURIVersion `json:"root_uri_version"`

	// PrincipalAttributeMapping is an array of OID or XML Namespace element mapping strings that can be used to map a
	// SAML assertion to a user in the database.
	//
	// For example: ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress", "urn:oid:0.9.2342.19200300.100.1.3"]
	PrincipalAttributeMappings []string `json:"principal_attribute_mappings" gorm:"type:text[];column:ous"`

	// The below values generated values that point a client to SAML related resources hosted on the BloodHound instance
	// and should not be persisted to the database due to the fact that the URLs rely on the Host header that the user is
	// using to communicate to the API
	ServiceProviderIssuerURI     serde.URL `json:"sp_issuer_uri" gorm:"-"`
	ServiceProviderInitiationURI serde.URL `json:"sp_sso_uri" gorm:"-"`
	ServiceProviderMetadataURI   serde.URL `json:"sp_metadata_uri" gorm:"-"`
	ServiceProviderACSURI        serde.URL `json:"sp_acs_uri" gorm:"-"`

	SSOProviderID null.Int32 `json:"sso_provider_id"`

	Serial
}

type SAMLProviders []SAMLProvider

func (SAMLProvider) TableName() string {
	return "saml_providers"
}

func (s SAMLProvider) AuditData() AuditData {
	return AuditData{
		"saml_id":                      s.ID,
		"saml_name":                    s.Name,
		"principal_attribute_mappings": s.PrincipalAttributeMappings,
		"idp_url":                      s.IssuerURI,
		"root_uri_version":             s.RootURIVersion,
		"sso_provider_id":              s.SSOProviderID.Int32,
	}
}

// EmailAttributeNames returns the service provider's configuration principal attribute mappings. If unset, this
// function instead returns a default array of well-known values.
func (s SAMLProvider) emailAttributeNames() []string {
	if mappings := s.PrincipalAttributeMappings; len(mappings) > 0 {
		return mappings
	}

	return []string{ObjectIDEmail, XMLSOAPClaimsEmailAddress}
}

func assertionFindString(assertion *saml.Assertion, names ...string) (string, error) {
	for _, attributeStatement := range assertion.AttributeStatements {
		for _, attribute := range attributeStatement.Attributes {
			for _, validName := range names {
				if attribute.Name == validName && len(attribute.Values) > 0 {
					// Try to find an explicit XMLType of xs:string
					for _, value := range attribute.Values {
						if value.Type == XMLTypeString {
							return value.Value, nil
						}
					}
					log.Warnf("[SAML] Found attribute values for attribute %s however none of the values have an XML type of %s. Choosing the first value.", ObjectIDAttributeNameFormat, XMLTypeString)
					return attribute.Values[0].Value, nil
				}
			}
		}
	}

	return "", errors.New("attribute not found")
}

func (s SAMLProvider) GetSAMLUserPrincipalNameFromAssertion(assertion *saml.Assertion) (string, error) {
	for _, attrStmt := range assertion.AttributeStatements {
		for _, attr := range attrStmt.Attributes {
			for _, value := range attr.Values {
				log.Infof("[SAML] Assertion contains attribute: %s - %s=%v", attr.NameFormat, attr.Name, value)
			}
		}
	}

	// All SAML assertions must contain a eduPersonPrincipalName attribute. While this is not expected to be an email
	// address, it may be formatted as such.
	if principalName, err := assertionFindString(assertion, s.emailAttributeNames()...); err != nil {
		return "", ErrSAMLAssertion
	} else {
		return principalName, nil
	}
}

func (s *SAMLProvider) FormatSAMLProviderURLs(hostUrl url.URL) {
	root := hostUrl

	// To preserve existing IDP configurations, existing saml providers still use the old acs endpoint which redirects to the new callback handler
	switch s.RootURIVersion {
	case SAMLRootURIVersion1:
		root.Path = path.Join("/api/v1/login/saml", s.Name)
		s.ServiceProviderACSURI = serde.FromURL(*root.JoinPath("acs"))
	case SAMLRootURIVersion2:
		root.Path = path.Join("/api/v2/sso", s.Name)
		s.ServiceProviderACSURI = serde.FromURL(*root.JoinPath("callback"))
	}

	s.ServiceProviderIssuerURI = serde.FromURL(root)
	s.ServiceProviderInitiationURI = serde.FromURL(*root.JoinPath("login"))
	s.ServiceProviderMetadataURI = serde.FromURL(*root.JoinPath("metadata"))
}
