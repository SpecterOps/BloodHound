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
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
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

// This retains support for the old saml login urls /api/{version}/login/saml/ that were added to their respective IDPs
func (s ManagementResource) SAMLLoginRedirect(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		bheCtx := ctx.FromRequest(request)
		redirectURL := api.URLJoinPath(*bheCtx.Host, fmt.Sprintf("/api/v2/sso/%s/login", ssoProvider.Slug))
		http.Redirect(response, request, redirectURL.String(), http.StatusFound)
	}
}

// This retains support for the old saml acs urls /api/{version}/login/saml/ that were added to their respective IDPs
func (s ManagementResource) SAMLCallbackRedirect(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		bheCtx := ctx.FromRequest(request)
		redirectURL := api.URLJoinPath(*bheCtx.Host, fmt.Sprintf("/api/v2/sso/%s/callback", ssoProvider.Slug))
		http.Redirect(response, request, redirectURL.String(), http.StatusTemporaryRedirect)
	}
}

func (s ManagementResource) ListSAMLSignOnEndpoints(response http.ResponseWriter, request *http.Request) {
	if samlProviders, err := s.db.GetAllSAMLProviders(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		var (
			samlSignOnEndpoints = make([]v2.SAMLSignOnEndpoint, len(samlProviders))
			requestContext      = ctx.Get(request.Context())
		)

		for idx, samlProvider := range samlProviders {
			samlProvider.FormatSAMLProviderURLs(*requestContext.Host)

			samlSignOnEndpoints[idx].Name = samlProvider.Name
			samlSignOnEndpoints[idx].InitiationURL = samlProvider.ServiceProviderInitiationURI
		}

		api.WriteBasicResponse(request.Context(), v2.ListSAMLSignOnEndpointsResponse{
			Endpoints: samlSignOnEndpoints,
		}, http.StatusOK, response)
	}
}

func (s ManagementResource) ListSAMLProviders(response http.ResponseWriter, request *http.Request) {
	if samlProviders, err := s.db.GetAllSAMLProviders(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		for _, samlProvider := range samlProviders {
			samlProvider.FormatSAMLProviderURLs(*ctx.Get(request.Context()).Host)
		}
		api.WriteBasicResponse(request.Context(), v2.ListSAMLProvidersResponse{SAMLProviders: samlProviders}, http.StatusOK, response)
	}
}

func (s ManagementResource) GetSAMLProvider(response http.ResponseWriter, request *http.Request) {
	pathVars := mux.Vars(request)

	if rawProviderID, hasID := pathVars[api.URIPathVariableSAMLProviderID]; !hasID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid, request), response)
	} else if providerID, err := strconv.ParseInt(rawProviderID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if provider, err := s.db.GetSAMLProvider(request.Context(), int32(providerID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), provider, http.StatusOK, response)
	}
}

func (s ManagementResource) CreateSAMLProviderMultipart(response http.ResponseWriter, request *http.Request) {
	var samlIdentityProvider model.SAMLProvider

	if err := request.ParseMultipartForm(api.DefaultAPIPayloadReadLimitBytes); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if providerNames, hasProviderName := request.MultipartForm.Value["name"]; !hasProviderName {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"name\" parameter", request), response)
	} else if numProviderNames := len(providerNames); numProviderNames == 0 || numProviderNames > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"name\" parameter", request), response)
	} else if metadataXMLFileHandles, hasMetadataXML := request.MultipartForm.File["metadata"]; !hasMetadataXML {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"metadata\" parameter", request), response)
	} else if numHeaders := len(metadataXMLFileHandles); numHeaders == 0 || numHeaders > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"metadata\" parameter", request), response)
	} else if metadataXMLReader, err := metadataXMLFileHandles[0].Open(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if autoProvisionEnabled, hasAutoProvisionEnabled := request.MultipartForm.Value["config.auto_provision.enabled"]; !hasAutoProvisionEnabled || len(autoProvisionEnabled) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"enabled\" parameter", request), response)
	} else if isAutoProvisionEnabled, err := strconv.ParseBool(autoProvisionEnabled[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"enabled\" parameter could not be converted to bool", request), response)
	} else if defaultRole, hasDefaultRole := request.MultipartForm.Value["config.auto_provision.default_role"]; !hasDefaultRole {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"default_role\" parameter", request), response)
	} else if defaultRoleValue, err := strconv.Atoi(defaultRole[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id must be a valid number", request), response)
	} else if defaultRoleValue > 5 || defaultRoleValue < 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id is invalid", request), response)
	} else if isRoleProvisioned, err := strconv.ParseBool(request.MultipartForm.Value["config.auto_provision.role_provision"][0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"role_provision\" parameter could not be converted to bool", request), response)
	} else {
		defer metadataXMLReader.Close()

		if metadataXML, err := io.ReadAll(metadataXMLReader); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else if metadata, err := samlsp.ParseMetadata(metadataXML); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else if ssoURL, err := auth.GetIDPSingleSignOnServiceURL(metadata, saml.HTTPPostBinding); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "metadata does not have a SSO service that supports HTTP POST binding", request), response)
		} else {
			var config model.SSOProviderConfig

			if isAutoProvisionEnabled {
				// Role IDs range from 1 to 5, and so a value of 0 indicates that the int32 DefaultRole value is unset
				if defaultRoleValue == 0 {
					defaultRoleValue = 3
				}

				config = model.SSOProviderConfig{
					AutoProvision: model.AutoProvision{
						Enabled:       isAutoProvisionEnabled,
						DefaultRole:   int32(defaultRoleValue),
						RoleProvision: isRoleProvisioned,
					},
				}
			} else {
				config = model.SSOProviderConfig{
					AutoProvision: model.AutoProvision{
						Enabled:       isAutoProvisionEnabled,
						DefaultRole:   0,
						RoleProvision: false,
					},
				}
			}

			samlIdentityProvider.Name = providerNames[0]
			samlIdentityProvider.DisplayName = providerNames[0]
			samlIdentityProvider.MetadataXML = metadataXML
			samlIdentityProvider.IssuerURI = metadata.EntityID
			samlIdentityProvider.SingleSignOnURI = ssoURL

			if newSAMLProvider, err := s.db.CreateSAMLIdentityProvider(request.Context(), samlIdentityProvider, config); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				api.WriteBasicResponse(request.Context(), newSAMLProvider, http.StatusOK, response)
			}
		}
	}
}

func (s ManagementResource) DeleteSAMLProvider(response http.ResponseWriter, request *http.Request) {
	var (
		identityProvider model.SAMLProvider
		rawProviderID    = mux.Vars(request)[api.URIPathVariableSAMLProviderID]
		requestContext   = ctx.FromRequest(request)
	)

	if providerID, err := strconv.ParseInt(rawProviderID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if identityProvider, err = s.db.GetSAMLProvider(request.Context(), int32(providerID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if user, isUser := auth.GetUserFromAuthCtx(requestContext.AuthCtx); isUser && user.SSOProviderID == identityProvider.SSOProviderID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "user may not delete their own SAML auth provider", request), response)
	} else if providerUsers, err := s.db.GetSSOProviderUsers(request.Context(), int(identityProvider.SSOProviderID.Int32)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := s.db.DeleteSSOProvider(request.Context(), int(identityProvider.SSOProviderID.Int32)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), v2.DeleteSAMLProviderResponse{
			AffectedUsers: providerUsers,
		}, http.StatusOK, response)
	}
}

// UpdateSAMLProviderRequest updates an SAML provider entry, support for partial payloads
func (s ManagementResource) UpdateSAMLProviderRequest(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if err := request.ParseMultipartForm(api.DefaultAPIPayloadReadLimitBytes); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if providerNames, hasProviderName := request.MultipartForm.Value["name"]; len(providerNames) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"name\" parameter", request), response)
	} else if metadataXMLFileHandles, hasMetadataXML := request.MultipartForm.File["metadata"]; len(metadataXMLFileHandles) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"metadata\" parameter", request), response)
	} else if autoProvisionEnabled, hasAutoProvisionEnabled := request.MultipartForm.Value["config.auto_provision.enabled"]; !hasAutoProvisionEnabled || len(autoProvisionEnabled) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"enabled\" parameter", request), response)
	} else if isAutoProvisionEnabled, err := strconv.ParseBool(autoProvisionEnabled[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"enabled\" parameter could not be converted to bool", request), response)
	} else if defaultRole, hasDefaultRole := request.MultipartForm.Value["config.auto_provision.default_role"]; !hasDefaultRole {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"default_role\" parameter", request), response)
	} else if defaultRoleValue, err := strconv.Atoi(defaultRole[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id must be a valid number", request), response)
	} else if defaultRoleValue > 5 || defaultRoleValue < 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id is invalid", request), response)
	} else if isRoleProvisioned, err := strconv.ParseBool(request.MultipartForm.Value["config.auto_provision.role_provision"][0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"role_provision\" parameter could not be converted to bool", request), response)
	} else {
		if isAutoProvisionEnabled {
			// Role IDs range from 1 to 5, and so a value of 0 indicates that the int32 DefaultRole value is unset
			if defaultRoleValue == 0 {
				defaultRoleValue = 3
			}

			ssoProvider.Config.AutoProvision.Enabled = isAutoProvisionEnabled
			ssoProvider.Config.AutoProvision.DefaultRole = int32(defaultRoleValue)
			ssoProvider.Config.AutoProvision.RoleProvision = isRoleProvisioned
		} else {
			ssoProvider.Config.AutoProvision.Enabled = isAutoProvisionEnabled
			ssoProvider.Config.AutoProvision.DefaultRole = 0
			ssoProvider.Config.AutoProvision.RoleProvision = false
		}

		if hasProviderName {
			ssoProvider.Name = providerNames[0]

			ssoProvider.SAMLProvider.Name = providerNames[0]
			ssoProvider.SAMLProvider.DisplayName = providerNames[0]
		}

		if hasMetadataXML {
			if metadataXMLReader, err := metadataXMLFileHandles[0].Open(); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			} else {
				defer metadataXMLReader.Close()

				if metadataXML, err := io.ReadAll(metadataXMLReader); err != nil {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
					return
				} else if metadata, err := samlsp.ParseMetadata(metadataXML); err != nil {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
					return
				} else if ssoURL, err := auth.GetIDPSingleSignOnServiceURL(metadata, saml.HTTPPostBinding); err != nil {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "metadata does not have a SSO service that supports HTTP POST binding", request), response)
					return
				} else {
					ssoProvider.SAMLProvider.MetadataXML = metadataXML
					ssoProvider.SAMLProvider.IssuerURI = metadata.EntityID
					ssoProvider.SAMLProvider.SingleSignOnURI = ssoURL

					// It's possible to update the ACS url which will be reflected in the metadataXML, we need to guarantee it is set to only what we expect if it is present
					if acsUrl, err := auth.GetAssertionConsumerServiceURL(metadata, saml.HTTPPostBinding); err == nil {
						if !strings.Contains(acsUrl, model.SAMLRootURIVersionMap[ssoProvider.SAMLProvider.RootURIVersion]) {
							var validUri bool
							for rootUriVersion, path := range model.SAMLRootURIVersionMap {
								if strings.Contains(acsUrl, path) {
									ssoProvider.SAMLProvider.RootURIVersion = rootUriVersion
									validUri = true
									break
								}
							}
							if !validUri {
								api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "metadata does not have a valid ACS location", request), response)
								return
							}
						}
					}
				}
			}
		}

		if newSAMLProvider, err := s.db.UpdateSAMLIdentityProvider(request.Context(), ssoProvider); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), newSAMLProvider, http.StatusOK, response)
		}
	}
}

// Preserve old metadata endpoint for saml providers
func (s ManagementResource) ServeMetadata(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if serviceProvider, err := auth.NewServiceProvider(*ctx.Get(request.Context()).Host, s.config, *ssoProvider.SAMLProvider); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		// Note: This is the samlsp metadata tied to authenticate flow and will not be the same as the XML metadata used to import the SAML provider initially
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

// Provide the saml provider certifcate
func (s ManagementResource) ServeSigningCertificate(response http.ResponseWriter, request *http.Request) {
	rawProviderID := mux.Vars(request)[api.URIPathVariableSSOProviderID]

	if ssoProviderID, err := strconv.ParseInt(rawProviderID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if ssoProvider, err := s.db.GetSSOProviderById(request.Context(), int32(ssoProviderID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else {
		// Note this is the public cert not necessarily the IDP cert
		response.Header().Set(headers.ContentDisposition.String(), fmt.Sprintf("attachment; filename=\"%s-signing-certificate.pem\"", ssoProvider.Slug))
		if _, err := response.Write([]byte(crypto.FormatCert(s.config.SAML.ServiceProviderCertificate))); err != nil {
			log.Errorf("[SAML] Failed to write response for serving signing certificate: %v", err)
		}
	}
}

// HandleStartAuthFlow is called to start the SAML authentication process.
func (s ManagementResource) SAMLLoginHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	if ssoProvider.SAMLProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if serviceProvider, err := auth.NewServiceProvider(*ctx.Get(request.Context()).Host, s.config, *ssoProvider.SAMLProvider); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		var (
			binding         = saml.HTTPRedirectBinding
			bindingLocation = serviceProvider.GetSSOBindingLocation(binding)
		)
		if bindingLocation == "" {
			binding = saml.HTTPPostBinding
			bindingLocation = serviceProvider.GetSSOBindingLocation(binding)
		}

		// TODO: add actual relay state support - BED-5071
		if authReq, err := serviceProvider.MakeAuthenticationRequest(bindingLocation, binding, saml.HTTPPostBinding); err != nil {
			log.Errorf("[SAML] Failed creating SAML authentication request: %v", err)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		} else {
			switch binding {
			case saml.HTTPRedirectBinding:
				if redirectURL, err := authReq.Redirect("", &serviceProvider); err != nil {
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

				if _, err := response.Write([]byte(fmt.Sprintf(authInitiationContentBodyFormat, authReq.Post("")))); err != nil {
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
	} else if serviceProvider, err := auth.NewServiceProvider(*ctx.Get(request.Context()).Host, s.config, *ssoProvider.SAMLProvider); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		if err := request.ParseForm(); err != nil {
			log.Errorf("[SAML] Failed to parse form POST: %v", err)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form POST is malformed", request), response)
		} else {
			if assertion, err := serviceProvider.ParseResponse(request, nil); err != nil {
				var typedErr *saml.InvalidResponseError
				switch {
				case errors.As(err, &typedErr):
					log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v - %s", ssoProvider.SAMLProvider.IssuerURI, typedErr.PrivateErr, typedErr.Response)
				default:
					log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v", ssoProvider.SAMLProvider.IssuerURI, err)
				}
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid, request), response)
			} else if principalName, err := ssoProvider.SAMLProvider.GetSAMLUserPrincipalNameFromAssertion(assertion); err != nil {
				log.Errorf("[SAML] Failed to lookup user for SAML provider %s: %v", ssoProvider.Name, err)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "session assertion does not meet the requirements for user lookup", request), response)
			} else if ssoProvider.Config.AutoProvision.Enabled {
				if user, err := s.db.LookupUser(request.Context(), principalName); err != nil {
					if errors.Is(err, database.ErrNotFound) {
						user.EmailAddress = null.StringFrom(principalName)
						user.PrincipalName = principalName
						user.Roles = model.Roles{
							{
								Serial: model.Serial{
									ID: ssoProvider.Config.AutoProvision.DefaultRole,
								},
							},
						}
						user.SSOProviderID = null.Int32From(ssoProvider.ID)

						// Need to find a work around since BHE cannot auto accept EULA as true
						user.EULAAccepted = true

						if givenName, err := ssoProvider.SAMLProvider.GetSAMLUserGivenNameFromAssertion(assertion); err != nil {
							user.FirstName = null.StringFrom(principalName)
						} else {
							user.FirstName = null.StringFrom(givenName)
						}

						if surname, err := ssoProvider.SAMLProvider.GetSAMLUserSurNameFromAssertion(assertion); err != nil {
							user.LastName = null.StringFrom("Last name not found")
						} else {
							user.LastName = null.StringFrom(surname)
						}

						if _, err := s.db.CreateUser(request.Context(), user); err != nil {
							api.HandleDatabaseError(request, response, err)
						}

						s.authenticator.CreateSSOSession(request, response, principalName, ssoProvider)
					}
				}
			} else {
				s.authenticator.CreateSSOSession(request, response, principalName, ssoProvider)
			}
		}
	}
}
