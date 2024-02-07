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
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth/bhsaml"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

const (
	ErrAttributeNotFound = errors.Error("attribute not found")

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

const (
	ErrorUserDisabled                 = errors.Error("user disabled")
	ErrorUserNotFound                 = errors.Error("User not found")
	ErrorSAMLAssertion                = errors.Error("SAML assertion error")
	ErrorUserNotAuthorizedForProvider = errors.Error("User not authorized for this provider")

	APIVersion1 = 1
	APIVersion2 = 2
)

type WriteAPIErrorResponse func(request *http.Request, response http.ResponseWriter, statusCode int, message string)

type RootResource struct {
	cfg                   config.Configuration
	writeAPIErrorResponse WriteAPIErrorResponse
	db                    database.Database
	spLock                *sync.Mutex
	spFactory             bhsaml.ServiceProviderFactory
	samlProviders         map[string]ProviderResource
}

func NewSAMLRootResource(cfg config.Configuration, db database.Database, writeAPIErrorResponse WriteAPIErrorResponse) *RootResource {
	return &RootResource{
		cfg:                   cfg,
		writeAPIErrorResponse: writeAPIErrorResponse,
		db:                    db,
		spLock:                &sync.Mutex{},
		spFactory:             bhsaml.NewServiceProviderFactory(cfg, db),
		samlProviders:         make(map[string]ProviderResource),
	}
}

func (s *RootResource) initInstance(idpName string, ctx context.Context) (ProviderResource, error) {
	if serviceProvider, err := s.spFactory.Lookup(idpName, ctx); err != nil {
		log.Errorf("[SAML] Failed initializing SAML SP instance %s: %v", idpName, err)
		return ProviderResource{}, err
	} else {
		providerResource := NewProviderResource(s.db, s.cfg, serviceProvider, s.writeAPIErrorResponse)

		s.spLock.Lock()
		defer s.spLock.Unlock()

		// Cache the provider for future invocations
		s.samlProviders[idpName] = providerResource

		return providerResource, nil
	}
}

func (s *RootResource) getInstance(organization string) (ProviderResource, bool) {
	s.spLock.Lock()
	defer s.spLock.Unlock()

	instance, found := s.samlProviders[organization]
	return instance, found
}

func (s *RootResource) clearInstance(organization string) {
	s.spLock.Lock()
	defer s.spLock.Unlock()

	delete(s.samlProviders, organization)
}

func (s *RootResource) fetchInstance(organization string, ctx context.Context) (ProviderResource, error) {
	if instance, hasInstance := s.getInstance(organization); !hasInstance {
		// Create a new instance if we don't have one at the ready
		return s.initInstance(organization, ctx)
	} else if _, err := s.db.GetSAMLProvider(instance.serviceProvider.Config.ID); err != nil {
		// In the case where the provider is no longer in the database we must clean up the existing ref and recreate it
		if errors.Is(err, database.ErrNotFound) {
			s.clearInstance(organization)
			if rv, err := s.initInstance(organization, ctx); err != nil {
				return rv, api.FormatDatabaseError(err)
			} else {
				return rv, nil
			}
		} else {
			return ProviderResource{}, api.FormatDatabaseError(err)
		}
	} else {
		// Instance is still valid, return it
		return instance, nil
	}
}

func (s *RootResource) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	pathVars := mux.Vars(request)

	if providerName, hasProviderName := pathVars[api.URIPathVariableServiceProviderName]; !hasProviderName {
		s.writeAPIErrorResponse(request, response, http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid)
	} else if samlInstance, err := s.fetchInstance(providerName, request.Context()); err != nil {
		s.writeAPIErrorResponse(request, response, http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid)
	} else {
		samlInstance.ServeHTTP(response, request)
	}
}

type ProviderResource struct {
	db                    database.Database
	cfg                   config.Configuration
	authenticator         api.Authenticator
	serviceProvider       bhsaml.ServiceProvider
	requestTracker        samlsp.RequestTracker
	bindingType           string
	responseBindingType   string
	writeAPIErrorResponse WriteAPIErrorResponse
}

func NewProviderResource(db database.Database, cfg config.Configuration, serviceProvider bhsaml.ServiceProvider, writeAPIErrorResponse WriteAPIErrorResponse) ProviderResource {
	return ProviderResource{
		db:                    db,
		cfg:                   cfg,
		authenticator:         api.NewAuthenticator(cfg, db, database.NewContextInitializer(db)),
		serviceProvider:       serviceProvider,
		requestTracker:        bhsaml.NewCookieRequestTracker(serviceProvider),
		writeAPIErrorResponse: writeAPIErrorResponse,

		// This is intentionally left empty - see SAML binding types
		bindingType:         "",
		responseBindingType: saml.HTTPPostBinding,
	}
}

func (s ProviderResource) getTrackedRequestIDs(request *http.Request) []string {
	var (
		trackedRequests = s.requestTracker.GetTrackedRequests(request)
		requestIDs      = make([]string, len(trackedRequests))
	)

	for idx, trackedRequest := range trackedRequests {
		requestIDs[idx] = trackedRequest.SAMLRequestID
	}

	return requestIDs
}

func assertionFindString(assertion *saml.Assertion, names ...string) (string, error) {
	for _, attributeStatement := range assertion.AttributeStatements {
		for _, attribute := range attributeStatement.Attributes {
			for _, validName := range names {
				if attribute.Name == validName && len(attribute.Values) > 0 {
					// Try to find an explicit XMLType of xs:string
					for _, value := range attribute.Values {
						if value.Type == bhsaml.XMLTypeString {
							return value.Value, nil
						}
					}

					log.Errorf("[SAML] Found attribute values for attribute %s however none of the values have an XML type of %s. Choosing the first value.", bhsaml.ObjectIDAttributeNameFormat, bhsaml.XMLTypeString)
					return attribute.Values[0].Value, nil
				}
			}
		}
	}

	return "", ErrAttributeNotFound
}

// emailAttributeNames returns the service provider's configuration principal attribute mappings. If unset, this
// function instead returns a default array of well-known values.
func (s ProviderResource) emailAttributeNames() []string {
	if mappings := s.serviceProvider.Config.PrincipalAttributeMappings; len(mappings) > 0 {
		return mappings
	}

	return []string{bhsaml.ObjectIDEmail, bhsaml.XMLSOAPClaimsEmailAddress}
}

func (s ProviderResource) lookupSAMLUser(assertion *saml.Assertion) (model.User, error) {
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
		return model.User{}, ErrorSAMLAssertion
	} else {
		if user, err := s.db.LookupUser(principalName); err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return model.User{}, api.FormatDatabaseError(err)
			} else {
				return model.User{}, ErrorUserNotFound
			}
		} else if !user.SAMLProviderID.Valid || s.serviceProvider.Config.ID != user.SAMLProvider.ID {
			return model.User{}, ErrorUserNotAuthorizedForProvider
		} else {
			return user, nil
		}
	}
}

func sameSiteValue(host url.URL) http.SameSite {
	if host.Scheme == "https" {
		return http.SameSiteStrictMode
	} else {
		return http.SameSiteDefaultMode
	}
}

// NOTE: Set-Cookie should generally have the Domain field blank to ensure the cookie is only included with requests against the host, excluding subdomains; however,
// most browsers will ignore Set-Cookie headers from localhost responses if the Domain field is not set explicitly.
func domainValue(host url.URL) string {
	if strings.Contains(host.Hostname(), "localhost") {
		return host.Hostname()
	} else {
		return ""
	}
}

func (s ProviderResource) createSessionFromAssertion(request *http.Request, response http.ResponseWriter, expires time.Time, assertion *saml.Assertion) {
	hostURL := *ctx.FromRequest(request).Host

	if user, err := s.lookupSAMLUser(assertion); err != nil {
		log.Errorf("[SAML] Failed to lookup user for SAML provider %s: %v", s.serviceProvider.Config.Name, err)

		switch err {
		case ErrorSAMLAssertion:
			s.writeAPIErrorResponse(request, response, http.StatusBadRequest, "session assertion does not meet the requirements for user lookup")
		case ErrorUserNotFound, ErrorUserNotAuthorizedForProvider:
			// This is a tiny bit more descriptive for the end user without leaking any sensitive information
			s.writeAPIErrorResponse(request, response, http.StatusForbidden, "user is not allowed")
		default:
			s.writeAPIErrorResponse(request, response, http.StatusInternalServerError, "session creation failure")
		}
	} else if sessionJWT, err := s.authenticator.CreateSession(user, s.serviceProvider.Config); err != nil {
		if locationURL := api.URLJoinPath(hostURL, api.UserDisabledPath); err == ErrorUserDisabled {
			response.Header().Add(headers.Location.String(), locationURL.String())
			response.WriteHeader(http.StatusFound)
		} else {
			log.Errorf("[SAML] Failed to create user session for SAML provider %s: %v", s.serviceProvider.Config.Name, err)
			s.writeAPIErrorResponse(request, response, http.StatusInternalServerError, "session creation failure")
		}
	} else {
		locationURL := api.URLJoinPath(hostURL, api.UserInterfacePath)

		// Set the token cookie
		http.SetCookie(response, &http.Cookie{
			Name:     api.AuthTokenCookieName,
			Value:    sessionJWT,
			Expires:  expires,
			Secure:   hostURL.Scheme == "https",
			SameSite: sameSiteValue(hostURL),
			Path:     "/",
			Domain:   domainValue(hostURL),
		})

		// Redirect back to the UI landing page
		response.Header().Add(headers.Location.String(), locationURL.String())
		response.WriteHeader(http.StatusFound)
	}
}

func (s ProviderResource) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// This is a little lazy but no need for a full regex oriented mux for just four URLs
	switch request.URL.Path {
	// TODO: Right now both the SP root and the SSO endpoints serve a valid SAML auth initiation flow - in the future
	//		 we will deprecate using the SP root URL as a valid SAML auth initiation flow endpoint.
	case s.serviceProvider.URLs.ServiceProviderRoot.Path, s.serviceProvider.URLs.SingleSignOnService.Path:
		s.serveStartAuthFlow(response, request)

	case s.serviceProvider.URLs.MetadataService.Path:
		s.serveMetadata(response, request)

	case s.serviceProvider.URLs.AssertionConsumerService.Path:
		s.serveAssertionConsumerService(response, request)

	case s.serviceProvider.SloURL.Path:
		// TODO: Not implemented yet

	default:
		s.writeAPIErrorResponse(request, response, http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound)
	}
}

func (s ProviderResource) bindingTypeAndLocation() (string, string) {
	var binding, bindingLocation string

	if s.bindingType != "" {
		binding = s.bindingType
		bindingLocation = s.serviceProvider.GetSSOBindingLocation(binding)
	} else {
		binding = saml.HTTPRedirectBinding
		bindingLocation = s.serviceProvider.GetSSOBindingLocation(binding)

		if bindingLocation == "" {
			binding = saml.HTTPPostBinding
			bindingLocation = s.serviceProvider.GetSSOBindingLocation(binding)
		}
	}

	return binding, bindingLocation
}

// HandleStartAuthFlow is called to start the SAML authentication process.
func (s ProviderResource) serveStartAuthFlow(response http.ResponseWriter, request *http.Request) {
	binding, bindingLocation := s.bindingTypeAndLocation()

	// relayState is limited to 80 bytes but also must be integrity protected.
	// this means that we cannot use a JWT because it is way too long. Instead,
	// we set a signed cookie that encodes the original URL which we'll check
	// against the SAML response when we get it.
	if authReq, err := s.serviceProvider.MakeAuthenticationRequest(bindingLocation, binding, s.responseBindingType); err != nil {
		log.Errorf("[SAML] Failed creating SAML authentication request: %v", err)
		s.writeAPIErrorResponse(request, response, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError)
	} else if relayState, err := s.requestTracker.TrackRequest(response, request, authReq.ID); err != nil {
		log.Errorf("[SAML] Failed to create a valid relay state token for SAML provider %s: %v", s.serviceProvider.EntityID, err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
	} else {
		switch binding {
		case saml.HTTPRedirectBinding:
			if redirectURL, err := authReq.Redirect(relayState, &s.serviceProvider.ServiceProvider); err != nil {
				log.Errorf("[SAML] Failed to format a redirect for SAML provider %s: %v", s.serviceProvider.EntityID, err)
				s.writeAPIErrorResponse(request, response, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError)
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
			s.writeAPIErrorResponse(request, response, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError)
		}
	}
}

func (s ProviderResource) serveMetadata(response http.ResponseWriter, request *http.Request) {
	if content, err := xml.MarshalIndent(s.serviceProvider.Metadata(), "", "  "); err != nil {
		log.Errorf("[SAML] XML marshalling failure during service provider encoding for %s: %v", s.serviceProvider.URLs.ServiceProviderRoot, err)
		s.writeAPIErrorResponse(request, response, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError)
	} else {
		response.Header().Set(headers.ContentType.String(), mediatypes.ApplicationSamlmetadataXml.String())
		if _, err := response.Write(content); err != nil {
			log.Errorf("[SAML] Failed to write response for serving metadata: %v", err)
		}
	}
}

func (s ProviderResource) serveAssertionConsumerService(response http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		log.Errorf("[SAML] Failed to parse form POST: %v", err)
		s.writeAPIErrorResponse(request, response, http.StatusBadRequest, "form POST is malformed")
	} else {
		possibleRequestIDs := s.getTrackedRequestIDs(request)

		if s.serviceProvider.AllowIDPInitiated {
			possibleRequestIDs = append(possibleRequestIDs, "")
		}

		if assertion, err := s.serviceProvider.ParseResponse(request, possibleRequestIDs); err != nil {
			switch typedErr := err.(type) {
			case *saml.InvalidResponseError:
				log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v - %s", s.serviceProvider.URLs.ServiceProviderRoot.String(), typedErr.PrivateErr, typedErr.Response)

			default:
				log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v", s.serviceProvider.URLs.ServiceProviderRoot.String(), err)
			}

			s.writeAPIErrorResponse(request, response, http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid)
		} else {
			s.createSessionFromAssertion(request, response, time.Now().UTC().Add(s.cfg.AuthSessionTTL()), assertion)
		}
	}
}
