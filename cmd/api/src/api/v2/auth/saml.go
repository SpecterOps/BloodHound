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

package auth

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	saml2 "github.com/specterops/bloodhound/src/api/saml"
	"github.com/specterops/bloodhound/src/auth/bhsaml"
	"github.com/specterops/bloodhound/src/config"
	bhCtx "github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
)

const (
	// TODO: This might be better if generated at run-time. These values were taken from the crewjam SAML provider package
	defaultContentSecurityPolicy    = "default-src; script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; reflected-xss block; referrer no-referrer;"
	authInitiationContentBodyFormat = `<!DOCTYPE html>
<html>
<body>
%s
</body>
</html>
`
)

func serviceProviderFactory(ctx context.Context, cfg config.Configuration, samlProvider model.SAMLProvider) (bhsaml.ServiceProvider, error) {
	if spCert, spKey, err := crypto.X509ParsePair(cfg.SAML.ServiceProviderCertificate, cfg.SAML.ServiceProviderKey); err != nil {
		return bhsaml.ServiceProvider{}, fmt.Errorf("failed to parse service provider %s's cert pair: %w", samlProvider.Name, err)
	} else if idpMetadata, err := samlsp.ParseMetadata(samlProvider.MetadataXML); err != nil {
		return bhsaml.ServiceProvider{}, fmt.Errorf("failed to parse metadata XML for service provider %s: %w", samlProvider.Name, err)
	} else {
		// This is required to populate the samlProvider.ServiceProviderIssuerURI
		samlProvider = bhsaml.FormatSAMLProviderURLs(ctx, samlProvider)[0]
		return bhsaml.NewServiceProvider(samlProvider, bhsaml.FormatServiceProviderURLs(*bhCtx.Get(ctx).Host, samlProvider.Name), samlsp.Options{
			EntityID:          samlProvider.ServiceProviderIssuerURI.String(),
			URL:               samlProvider.ServiceProviderIssuerURI.AsURL(),
			Key:               spKey,
			Certificate:       spCert,
			AllowIDPInitiated: true,
			SignRequest:       true,
			IDPMetadata:       idpMetadata,
		}), nil
	}
}

func samlWriteAPIErrorResponse(request *http.Request, response http.ResponseWriter, statusCode int, message string) {
	api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(statusCode, message, request), response)
}

// Preserve old metadata endpoint
func (s ManagementResource) ServeMetadata(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableServiceProviderName]

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if serviceProvider, err := serviceProviderFactory(request.Context(), s.config, *ssoProvider.SAMLProvider); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		if content, err := xml.MarshalIndent(serviceProvider.Metadata(), "", "  "); err != nil {
			log.Errorf("[SAML] XML marshalling failure during service provider encoding for %s: %v", ssoProvider.SAMLProvider.IssuerURI, err)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		} else {
			response.Header().Set(headers.ContentType.String(), mediatypes.ApplicationSamlmetadataXml.String())
			if _, err := response.Write(content); err != nil {
				log.Errorf("[SAML] Failed to write response for serving metadata: %v", err)
			}
		}
	}
}

// HandleStartAuthFlow is called to start the SAML authentication process.
func (s ManagementResource) SAMLLoginHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if serviceProvider, err := serviceProviderFactory(request.Context(), s.config, *ssoProvider.SAMLProvider); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		providerResource := saml2.NewProviderResource(s.db, s.config, serviceProvider, samlWriteAPIErrorResponse)
		binding, bindingLocation := providerResource.BindingTypeAndLocation()
		// relayState is limited to 80 bytes but also must be integrity protected.
		// this means that we cannot use a JWT because it is way too long. Instead,
		// we set a signed cookie that encodes the original URL which we'll check
		// against the SAML response when we get it.
		if authReq, err := serviceProvider.MakeAuthenticationRequest(bindingLocation, binding, providerResource.ResponseBindingType); err != nil {
			log.Errorf("[SAML] Failed creating SAML authentication request: %v", err)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		} else if relayState, err := providerResource.RequestTracker.TrackRequest(response, request, authReq.ID); err != nil {
			log.Errorf("[SAML] Failed to create a valid relay state token for SAML provider %s: %v", serviceProvider.EntityID, err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		} else {
			switch binding {
			case saml.HTTPRedirectBinding:
				if redirectURL, err := authReq.Redirect(relayState, &serviceProvider.ServiceProvider); err != nil {
					log.Errorf("[SAML] Failed to format a redirect for SAML provider %s: %v", serviceProvider.EntityID, err)
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
				} else {
					response.Header().Add(headers.Location.String(), redirectURL.String())
					response.WriteHeader(http.StatusFound)
				}

			case saml.HTTPPostBinding:
				response.Header().Add(headers.ContentSecurityPolicy.String(), defaultContentSecurityPolicy)
				response.Header().Add(headers.ContentType.String(), mediatypes.TextHtml.String())
				response.WriteHeader(http.StatusOK)

				if _, err := response.Write([]byte(fmt.Sprintf(authInitiationContentBodyFormat, authReq.Post(relayState)))); err != nil {
					log.Errorf("[SAML] Failed to write response with HTTP POST binding: %v", err)
				}

			default:
				log.Errorf("[SAML] Unhandled binding type %s", binding)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			}
		}
	}
}

// HandleStartAuthFlow is called to start the SAML authentication process.
func (s ManagementResource) SAMLCallbackHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if serviceProvider, err := serviceProviderFactory(request.Context(), s.config, *ssoProvider.SAMLProvider); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		providerResource := saml2.NewProviderResource(s.db, s.config, serviceProvider, samlWriteAPIErrorResponse)
		if err := request.ParseForm(); err != nil {
			log.Errorf("[SAML] Failed to parse form POST: %v", err)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form POST is malformed", request), response)
		} else {
			if assertion, err := serviceProvider.ParseResponse(request, nil); err != nil {
				var typedErr *saml.InvalidResponseError
				switch {
				case errors.As(err, &typedErr):
					log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v - %s", serviceProvider.URLs.ServiceProviderRoot.String(), typedErr.PrivateErr, typedErr.Response)
				default:
					log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v", serviceProvider.URLs.ServiceProviderRoot.String(), err)
				}
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid, request), response)
			} else if principalName, err := providerResource.GetSAMLUserPrincipalNameFromAssertion(assertion); err != nil {
				log.Errorf("[SAML] Failed to lookup user for SAML provider %s: %v", serviceProvider.Config.Name, err)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "session assertion does not meet the requirements for user lookup", request), response)
			} else {
				s.authenticator.CreateSSOSession(request, response, principalName, ssoProvider)
			}
		}
	}
}
