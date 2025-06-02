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
	"fmt"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/model"
)

func getIDPSingleSignOnDescriptor(metadata *saml.EntityDescriptor, bindingType string) (saml.IDPSSODescriptor, error) {
	for _, idpSSODescriptor := range metadata.IDPSSODescriptors {
		for _, singleSignOnService := range idpSSODescriptor.SingleSignOnServices {
			if singleSignOnService.Binding == bindingType {
				return idpSSODescriptor, nil
			}
		}
	}

	return saml.IDPSSODescriptor{}, fmt.Errorf("no SSO service defined that supports the %s binding type", bindingType)
}

func GetIDPSingleSignOnServiceURL(metadata *saml.EntityDescriptor, bindingType string) (string, error) {
	if ssoDescriptor, err := getIDPSingleSignOnDescriptor(metadata, saml.HTTPPostBinding); err != nil {
		return "", err
	} else {
		for _, singleSignOnService := range ssoDescriptor.SingleSignOnServices {
			if singleSignOnService.Binding == bindingType {
				return singleSignOnService.Location, nil
			}
		}
	}
	return "", fmt.Errorf("no SSO service defined that supports the %s binding type", bindingType)
}

// GetAssertionConsumerServiceURL This may not be present, we return the first we find
func GetAssertionConsumerServiceURL(metadata *saml.EntityDescriptor, bindingType string) (string, error) {
	for _, spSSODescriptor := range metadata.SPSSODescriptors {
		for _, acs := range spSSODescriptor.AssertionConsumerServices {
			if acs.Binding == bindingType {
				return acs.Location, nil
			}
		}
	}

	return "", fmt.Errorf("no SAML ascertion consumer service url defined in metadata xml")
}

func NewServiceProvider(hostUrl url.URL, cfg config.Configuration, samlProvider model.SAMLProvider) (saml.ServiceProvider, error) {
	if spCert, spKey, err := crypto.X509ParsePair(cfg.SAML.ServiceProviderCertificate, cfg.SAML.ServiceProviderKey); err != nil {
		return saml.ServiceProvider{}, fmt.Errorf("failed to parse service provider %s's cert pair: %w", samlProvider.Name, err)
	} else if idpMetadata, err := samlsp.ParseMetadata(samlProvider.MetadataXML); err != nil {
		return saml.ServiceProvider{}, fmt.Errorf("failed to parse metadata XML for service provider %s: %w", samlProvider.Name, err)
	} else {
		samlProvider.FormatSAMLProviderURLs(hostUrl)

		return saml.ServiceProvider{
			EntityID:          samlProvider.ServiceProviderIssuerURI.String(),
			Key:               spKey,
			Certificate:       spCert,
			MetadataURL:       samlProvider.ServiceProviderMetadataURI.URL,
			AcsURL:            samlProvider.ServiceProviderACSURI.URL,
			IDPMetadata:       idpMetadata,
			AuthnNameIDFormat: saml.EmailAddressNameIDFormat,
			SignatureMethod:   dsig.RSASHA256SignatureMethod,
			AllowIDPInitiated: true,
		}, nil
	}
}
