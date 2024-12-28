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
	} else if autoProvisionEnabled, hasAutoProvisionEnabled := request.MultipartForm.Value["config.auto_provision.enabled"]; !hasAutoProvisionEnabled || len(autoProvisionEnabled) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"config.auto_provision.enabled\" parameter or \"config.auto_provision.enabled\" parameter has more than one value", request), response)
	} else if isAutoProvisionEnabled, err := strconv.ParseBool(autoProvisionEnabled[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.enabled\" parameter could not be converted to bool", request), response)
	} else if defaultRoleId, hasDefaultRoleId := request.MultipartForm.Value["config.auto_provision.default_role_id"]; !hasDefaultRoleId || len(defaultRoleId) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"config.auto_provision.default_role_id\" parameter or \"config.auto_provision.default_role_id\" has more than one value", request), response)
	} else if defaultRoleIdInt, err := strconv.Atoi(defaultRoleId[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.default_role_id\" parameter could not be converted to int", request), response)
	} else if defaultRole, err := s.db.GetRole(request.Context(), int32(defaultRoleIdInt)); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.default_role_id\" parameter is invalid", request), response)
	} else if roleProvision, hasRoleProvisioned := request.MultipartForm.Value["config.auto_provision.role_provision"]; !hasRoleProvisioned || len(roleProvision) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"config.auto_provision.role_provision\" parameter or \"config.auto_provision.role_provision\" has more than one value", request), response)
	} else if isRoleProvisioned, err := strconv.ParseBool(roleProvision[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.role_provision\" parameter could not be converted to bool", request), response)
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
				config = model.SSOProviderConfig{
					AutoProvision: model.SSOProviderAutoProvisionConfig{
						Enabled:       isAutoProvisionEnabled,
						DefaultRoleId: defaultRole.ID,
						RoleProvision: isRoleProvisioned,
					},
				}
			}

			samlIdentityProvider.Name = providerNames[0]
			samlIdentityProvider.DisplayName = providerNames[0]
			samlIdentityProvider.MetadataXML = metadataXML
			samlIdentityProvider.IssuerURI = metadata.EntityID
			samlIdentityProvider.SingleSignOnURI = ssoURL

			if newSAMLProvider, err := s.db.CreateSAMLIdentityProvider(request.Context(), samlIdentityProvider, config); err != nil {
				if errors.Is(err, database.ErrDuplicateSSOProviderName) {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseSSOProviderDuplicateName, request), response)
				} else {
					api.HandleDatabaseError(request, response, err)
				}
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
	} else if autoProvisionEnabled, hasAutoProvisionEnabled := request.MultipartForm.Value["config.auto_provision.enabled"]; len(autoProvisionEnabled) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"config.auto_provision.enabled\" parameter ", request), response)
	} else {
		if hasAutoProvisionEnabled {
			if isAutoProvisionEnabled, err := strconv.ParseBool(autoProvisionEnabled[0]); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.enabled\" parameter could not be converted to bool", request), response)
				return
			} else {
				if isAutoProvisionEnabled {
					if defaultRoleId, hasDefaultRoleId := request.MultipartForm.Value["config.auto_provision.default_role_id"]; !hasDefaultRoleId || len(defaultRoleId) > 1 {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"config.auto_provision.default_role_id\" parameter or \"config.auto_provision.default_role_id\" has more than one value", request), response)
						return
					} else if defaultRoleIdInt, err := strconv.Atoi(defaultRoleId[0]); err != nil {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.default_role_id\" parameter could not be converted to int", request), response)
						return
					} else if defaultRole, err := s.db.GetRole(request.Context(), int32(defaultRoleIdInt)); err != nil {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.default_role_id\" parameter is invalid", request), response)
						return
					} else if roleProvision, hasRoleProvisioned := request.MultipartForm.Value["config.auto_provision.role_provision"]; !hasRoleProvisioned || len(roleProvision) > 1 {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"config.auto_provision.role_provision\" parameter or \"config.auto_provision.role_provision\" has more than one value", request), response)
						return
					} else if isRoleProvisioned, err := strconv.ParseBool(roleProvision[0]); err != nil {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "\"config.auto_provision.role_provision\" parameter could not be converted to bool", request), response)
						return
					} else {
						ssoProvider.Config.AutoProvision.Enabled = isAutoProvisionEnabled
						ssoProvider.Config.AutoProvision.DefaultRoleId = defaultRole.ID
						ssoProvider.Config.AutoProvision.RoleProvision = isRoleProvisioned
					}
				} else {
					ssoProvider.Config.AutoProvision = model.SSOProviderAutoProvisionConfig{}
				}
			}
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
			if errors.Is(err, database.ErrDuplicateSSOProviderName) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseSSOProviderDuplicateName, request), response)
			} else {
				api.HandleDatabaseError(request, response, err)
			}
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
		// SAML misconfiguration scenario
		redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")

	} else if serviceProvider, err := auth.NewServiceProvider(*ctx.Get(request.Context()).Host, s.config, *ssoProvider.SAMLProvider); err != nil {
		log.Errorf("[SAML] Service provider creation failed: %v", err)
		// Technical issues scenario
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
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
			// SAML misconfiguration or technical issue
			// Since this likely indicates a configuration problem, we treat it as a misconfiguration scenario
			redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
		} else {
			switch binding {
			case saml.HTTPRedirectBinding:
				if redirectURL, err := authReq.Redirect("", &serviceProvider); err != nil {
					log.Errorf("[SAML] Failed to format a redirect for SAML provider %s: %v", serviceProvider.EntityID, err)
					// Likely a technical or configuration issue
					redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
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
					// Technical issues scenario
					redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
				}

			default:
				log.Errorf("[SAML] Unhandled binding type %s", binding)
				// Treating unknown binding as a misconfiguration
				redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
			}
		}
	}
}

// HandleStartAuthFlow is called to start the SAML authentication process.
func (s ManagementResource) SAMLCallbackHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	if ssoProvider.SAMLProvider == nil {
		// SAML misconfiguration
		redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
	} else if serviceProvider, err := auth.NewServiceProvider(*ctx.Get(request.Context()).Host, s.config, *ssoProvider.SAMLProvider); err != nil {
		log.Errorf("[SAML] Service provider creation failed: %v", err)
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
	} else {
		if err := request.ParseForm(); err != nil {
			log.Errorf("[SAML] Failed to parse form POST: %v", err)
			// Technical issues or invalid form data
			// This is not covered by acceptance criteria directly; treat as technical issue
			redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
		} else {
			if assertion, err := serviceProvider.ParseResponse(request, nil); err != nil {
				var typedErr *saml.InvalidResponseError
				switch {
				case errors.As(err, &typedErr):
					log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v - %s", ssoProvider.SAMLProvider.IssuerURI, typedErr.PrivateErr, typedErr.Response)
				default:
					log.Errorf("[SAML] Failed to parse ACS response for provider %s: %v", ssoProvider.SAMLProvider.IssuerURI, err)
				}
				// SAML credentials issue scenario (authentication failed)
				redirectToLoginPage(response, request, "Your SSO was unable to authenticate your user, please contact your Administrator")
			} else if principalName, err := ssoProvider.SAMLProvider.GetSAMLUserPrincipalNameFromAssertion(assertion); err != nil {
				log.Errorf("[SAML] Failed to lookup user for SAML provider %s: %v", ssoProvider.Name, err)
				// SAML credentials issue scenario again
				redirectToLoginPage(response, request, "Your SSO was unable to authenticate your user, please contact your Administrator")
			} else {
				if ssoProvider.Config.AutoProvision.Enabled {
					if user, err := s.db.LookupUser(request.Context(), principalName); err != nil {
						if errors.Is(err, database.ErrNotFound) {
							if role, err := s.db.GetRole(request.Context(), ssoProvider.Config.AutoProvision.DefaultRoleId); err != nil {
								api.HandleDatabaseError(request, response, err)
								return
							} else {
								user = model.User{
									EmailAddress:  null.StringFrom(principalName),
									PrincipalName: principalName,
									Roles:         model.Roles{role},
									SSOProviderID: null.Int32From(ssoProvider.ID),
									EULAAccepted:  true, // EULA Acceptance does not pertain to Bloodhound Community Edition; this flag is used for Bloodhound Enterprise users
									FirstName:     null.StringFrom(principalName),
									LastName:      null.StringFrom("Last name not found"),
								}

								if givenName, err := ssoProvider.SAMLProvider.GetSAMLUserGivenNameFromAssertion(assertion); err == nil {
									user.FirstName = null.StringFrom(givenName)
								}

								if surname, err := ssoProvider.SAMLProvider.GetSAMLUserSurnameFromAssertion(assertion); err == nil {
									user.LastName = null.StringFrom(surname)
								}

								if _, err := s.db.CreateUser(request.Context(), user); err != nil {
									// It is safe to let this request drop into the CreateSSOSession function below to ensure proper audit logging
									log.Errorf("[SAML] JIT User Create Error: %v", err)
								}
							}
						}
					}
				}

				s.authenticator.CreateSSOSession(request, response, principalName, ssoProvider)
			}
		}
	}
}
