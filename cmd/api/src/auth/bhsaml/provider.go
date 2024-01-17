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

package bhsaml

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/stream"
	"github.com/specterops/bloodhound/src/config"
	bhCtx "github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

var (
	SAMLAPILoginRoot = "api/v2/login/saml"
)

const (
	SSODescriptorUseSigning    string = "signing"
	SSODescriptorUseEncryption string = "encryption"
)

func GetIDPSingleSignOnServiceURL(idp saml.IDPSSODescriptor, bindingType string) (string, error) {
	for _, singleSignOnService := range idp.SingleSignOnServices {
		if singleSignOnService.Binding == bindingType {
			return singleSignOnService.Location, nil
		}
	}

	return "", fmt.Errorf("no SSO service defined that supports the %s binding type", bindingType)
}

func GetIDPSSOSigningCertificateChain(idp saml.IDPSSODescriptor) []string {
	var certs []string

	for _, keyDescriptor := range idp.KeyDescriptors {
		if keyDescriptor.Use == SSODescriptorUseSigning {
			for _, cert := range keyDescriptor.KeyInfo.X509Data.X509Certificates {
				certs = append(certs, cert.Data)
			}
		}
	}

	return certs
}

func GetIDPSingleSignOnDescriptor(metadata *saml.EntityDescriptor, bindingType string) (saml.IDPSSODescriptor, error) {
	for _, idpSSODescriptor := range metadata.IDPSSODescriptors {
		for _, singleSignOnService := range idpSSODescriptor.SingleSignOnServices {
			if singleSignOnService.Binding == bindingType {
				return idpSSODescriptor, nil
			}
		}
	}

	return saml.IDPSSODescriptor{}, fmt.Errorf("no SSO service defined that supports the %s binding type", bindingType)
}

func FindSSODescriptorByURL(ssoURL string, idpMetadata *saml.EntityDescriptor) (saml.IDPSSODescriptor, error) {
	for _, idpSSODescriptor := range idpMetadata.IDPSSODescriptors {
		for _, singleSignOnService := range idpSSODescriptor.SingleSignOnServices {
			log.Debugf("[SAML] Scanning SSO descriptor with binding type %s and URL %s", singleSignOnService.Binding, singleSignOnService.Location)

			if singleSignOnService.Binding != saml.HTTPRedirectBinding && singleSignOnService.Binding != saml.HTTPPostBinding {
				continue
			}

			if singleSignOnService.Location == ssoURL {
				return idpSSODescriptor, nil
			}
		}
	}

	return saml.IDPSSODescriptor{}, fmt.Errorf("unable to find SSO URL %s in IDP metadata", ssoURL)
}

func ValidateIDPMetadata(idp model.SAMLProvider, metadata *saml.EntityDescriptor) error {
	if idpSSODescriptor, err := FindSSODescriptorByURL(idp.SingleSignOnURI, metadata); err != nil {
		return err
	} else {
		var (
			hasSigningKey    = false
			hasEncryptionKey = false
		)

		for _, keyDescriptor := range idpSSODescriptor.KeyDescriptors {
			if keyDescriptor.Use == SSODescriptorUseSigning {
				if _, err := crypto.X509ParseCert(keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data); err != nil {
					return fmt.Errorf("failed to parse signing certificate from IDP %s: %w", idp.IssuerURI, err)
				}

				hasSigningKey = true
			} else if keyDescriptor.Use == SSODescriptorUseEncryption {
				if _, err := crypto.X509ParseCert(keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data); err != nil {
					return fmt.Errorf("failed to parse encryption certificate from IDP %s: %w", idp.IssuerURI, err)
				}

				hasEncryptionKey = true
			}
		}

		if !hasSigningKey && !hasEncryptionKey {
			return fmt.Errorf("metadata for SSO URL %s is missing signing and encryption keys", idp.SingleSignOnURI)
		}
	}

	return nil
}

func fetchSAMLMetadataBytes(httpClient api.HTTPClient, ctx context.Context, metadataURL string) ([]byte, error) {
	if req, err := http.NewRequest("GET", metadataURL, nil); err != nil {
		return nil, err
	} else if resp, err := httpClient.Do(req.WithContext(ctx)); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		bodyReader := stream.NewLimitedReader(api.DefaultAPIPayloadReadLimitBytes, resp.Body)

		if resp.StatusCode < http.StatusOK && resp.StatusCode >= http.StatusMultipleChoices {
			if content, err := io.ReadAll(bodyReader); err != nil {
				return nil, fmt.Errorf("unexpected HTTP status code %d while validating IDP %s: %s", resp.StatusCode, metadataURL, "unable to read response body")
			} else {
				return nil, fmt.Errorf("unexpected HTTP status code %d while validating IDP %s: %s", resp.StatusCode, metadataURL, string(content))
			}
		}

		return io.ReadAll(bodyReader)
	}
}

func FetchSAMLMetadata(httpClient api.HTTPClient, ctx context.Context, metadataURL string) (*saml.EntityDescriptor, error) {
	if data, err := fetchSAMLMetadataBytes(httpClient, ctx, metadataURL); err != nil {
		return nil, err
	} else {
		return samlsp.ParseMetadata(data)
	}
}

type ServiceProviderURLs struct {
	ServiceProviderRoot      url.URL
	SingleSignOnService      url.URL
	MetadataService          url.URL
	AssertionConsumerService url.URL
	SingleLogoutService      url.URL
}

func FormatServiceProviderURLs(hostURL url.URL, serviceProviderName string) ServiceProviderURLs {
	root := hostURL
	root.Path = path.Join("/", SAMLAPILoginRoot, serviceProviderName)

	return ServiceProviderURLs{
		ServiceProviderRoot:      root,
		SingleSignOnService:      api.URLJoinPath(root, "sso"),
		MetadataService:          api.URLJoinPath(root, "metadata"),
		AssertionConsumerService: api.URLJoinPath(root, "acs"),
		SingleLogoutService:      api.URLJoinPath(root, "slo"),
	}
}

func NewServiceProviderFactory(cfg config.Configuration, db database.Database) ServiceProviderFactory {
	return ServiceProviderFactory{
		db:  db,
		cfg: cfg,
	}
}

type ServiceProviderFactory struct {
	db  database.Database
	cfg config.Configuration
}

func (s ServiceProviderFactory) Lookup(serviceProviderName string, ctx context.Context) (ServiceProvider, error) {
	if spCert, spKey, err := crypto.X509ParsePair(s.cfg.SAML.ServiceProviderCertificate, s.cfg.SAML.ServiceProviderKey); err != nil {
		return ServiceProvider{}, fmt.Errorf("failed to parse service provider %s's cert pair: %w", serviceProviderName, err)
	} else if providerDetails, err := GetSAMLProviderByName(s.db, serviceProviderName, ctx); err != nil {
		return ServiceProvider{}, fmt.Errorf("database lookup failed up service provider %s: %w", serviceProviderName, err)
	} else if idpMetadata, err := samlsp.ParseMetadata(providerDetails.MetadataXML); err != nil {
		return ServiceProvider{}, fmt.Errorf("failed to parse metadata XML for service provider %s: %w", serviceProviderName, err)
	} else {
		return NewServiceProvider(providerDetails, FormatServiceProviderURLs(*bhCtx.Get(ctx).Host, serviceProviderName), samlsp.Options{
			EntityID:          providerDetails.ServiceProviderIssuerURI.String(),
			URL:               providerDetails.ServiceProviderIssuerURI.AsURL(),
			Key:               spKey,
			Certificate:       spCert,
			AllowIDPInitiated: true,
			SignRequest:       true,
			IDPMetadata:       idpMetadata,
		}), nil
	}
}

type ServiceProvider struct {
	URLs    ServiceProviderURLs
	Config  model.SAMLProvider
	Options samlsp.Options

	saml.ServiceProvider
}

func NewServiceProvider(providerDetails model.SAMLProvider, serviceProviderURLs ServiceProviderURLs, opts samlsp.Options) ServiceProvider {
	var (
		signatureMethod = dsig.RSASHA256SignatureMethod
		forceAuthn      *bool
	)

	if opts.ForceAuthn {
		forceAuthn = &opts.ForceAuthn
	}

	if !opts.SignRequest {
		signatureMethod = ""
	}

	return ServiceProvider{
		URLs:    serviceProviderURLs,
		Config:  providerDetails,
		Options: opts,

		ServiceProvider: saml.ServiceProvider{
			EntityID:          opts.EntityID,
			Key:               opts.Key,
			Certificate:       opts.Certificate,
			Intermediates:     opts.Intermediates,
			MetadataURL:       serviceProviderURLs.MetadataService,
			AcsURL:            serviceProviderURLs.AssertionConsumerService,
			SloURL:            serviceProviderURLs.SingleLogoutService,
			IDPMetadata:       opts.IDPMetadata,
			AuthnNameIDFormat: saml.EmailAddressNameIDFormat,
			ForceAuthn:        forceAuthn,
			SignatureMethod:   signatureMethod,
			AllowIDPInitiated: opts.AllowIDPInitiated,
		},
	}
}
