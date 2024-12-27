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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/log"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/validation"
	"golang.org/x/oauth2"
)

// UpsertOIDCProviderRequest represents the body of create & update provider endpoints
type UpsertOIDCProviderRequest struct {
	Name     string                   `json:"name" validate:"required"`
	Issuer   string                   `json:"issuer"  validate:"url"`
	ClientID string                   `json:"client_id" validate:"required"`
	Config   *model.SSOProviderConfig `json:"config,omitempty"`
}

// UpdateOIDCProviderRequest updates an OIDC provider, support for only partial payloads
func (s ManagementResource) UpdateOIDCProviderRequest(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	var upsertReq UpsertOIDCProviderRequest

	if ssoProvider.OIDCProvider == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&upsertReq, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		if upsertReq.Name != "" {
			ssoProvider.Name = upsertReq.Name
		}

		if upsertReq.ClientID != "" {
			ssoProvider.OIDCProvider.ClientID = upsertReq.ClientID
		}

		if upsertReq.Issuer != "" {
			if err := validation.ValidUrl(upsertReq.Issuer); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "issuer url is invalid", request), response)
				return
			}

			ssoProvider.OIDCProvider.Issuer = upsertReq.Issuer
		}

		if upsertReq.Config != nil {
			if _, err := s.db.GetRole(request.Context(), int32(upsertReq.Config.AutoProvision.DefaultRole)); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id is invalid", request), response)
				return
			}

			if !upsertReq.Config.AutoProvision.Enabled {
				ssoProvider.Config.AutoProvision = model.SSOProviderAutoProvisionConfig{}
			} else {
				ssoProvider.Config.AutoProvision = upsertReq.Config.AutoProvision
			}
		}

		if oidcProvider, err := s.db.UpdateOIDCProvider(request.Context(), ssoProvider); err != nil {
			if errors.Is(err, database.ErrDuplicateSSOProviderName) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseSSOProviderDuplicateName, request), response)
			} else {
				api.HandleDatabaseError(request, response, err)
			}
		} else {
			api.WriteBasicResponse(request.Context(), oidcProvider, http.StatusOK, response)
		}
	}
}

// CreateOIDCProvider creates an OIDC provider entry given a valid request
func (s ManagementResource) CreateOIDCProvider(response http.ResponseWriter, request *http.Request) {
	var upsertReq UpsertOIDCProviderRequest

	if err := api.ReadJSONRequestPayloadLimited(&upsertReq, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if validated := validation.Validate(upsertReq); validated != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, validated.Error(), request), response)
	} else if upsertReq.Config == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "config is required", request), response)
	} else if _, err := s.db.GetRole(request.Context(), int32(upsertReq.Config.AutoProvision.DefaultRole)); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id is invalid", request), response)
	} else {
		if !upsertReq.Config.AutoProvision.Enabled {
			upsertReq.Config.AutoProvision = model.SSOProviderAutoProvisionConfig{}
		}

		if oidcProvider, err := s.db.CreateOIDCProvider(request.Context(), upsertReq.Name, upsertReq.Issuer, upsertReq.ClientID, *upsertReq.Config); err != nil {
			if errors.Is(err, database.ErrDuplicateSSOProviderName) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseSSOProviderDuplicateName, request), response)
			} else {
				api.HandleDatabaseError(request, response, err)
			}
		} else {
			api.WriteBasicResponse(request.Context(), oidcProvider, http.StatusCreated, response)
		}
	}
}

func getRedirectURL(request *http.Request, provider model.SSOProvider) string {
	hostUrl := *ctx.FromRequest(request).Host
	return fmt.Sprintf("%s/api/v2/sso/%s/callback", hostUrl.String(), provider.Slug)
}

func (s ManagementResource) OIDCLoginHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	if ssoProvider.OIDCProvider == nil {
		// SSO misconfiguration scenario
		redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
	} else if state, err := config.GenerateRandomBase64String(77); err != nil {
		log.Errorf("[OIDC] Failed to generate state: %v", err)
		// Technical issues scenario
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
	} else if provider, err := oidc.NewProvider(request.Context(), ssoProvider.OIDCProvider.Issuer); err != nil {
		log.Errorf("[OIDC] Failed to create OIDC provider: %v", err)
		// SSO misconfiguration or technical issue
		// Treat this as a misconfiguration scenario
		redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
	} else {
		conf := &oauth2.Config{
			ClientID:    ssoProvider.OIDCProvider.ClientID,
			Endpoint:    provider.Endpoint(),
			RedirectURL: getRedirectURL(request, ssoProvider),
			Scopes:      []string{"openid", "profile", "email", "email_verified", "name", "given_name", "family_name"},
		}

		// use PKCE to protect against CSRF attacks
		// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
		verifier := oauth2.GenerateVerifier()

		// Store PKCE on web browser in secure cookie for retrieval in callback
		api.SetSecureBrowserCookie(request, response, api.AuthPKCECookieName, verifier, time.Now().UTC().Add(time.Minute*7), true)

		// Store State on web browser in secure cookie for retrieval in callback
		api.SetSecureBrowserCookie(request, response, api.AuthStateCookieName, state, time.Now().UTC().Add(time.Minute*7), true)

		// Redirect user to consent page to ask for permission for the scopes specified above.
		redirectURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

		response.Header().Add(headers.Location.String(), redirectURL)
		response.WriteHeader(http.StatusFound)
	}
}

func (s ManagementResource) OIDCCallbackHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	var (
		queryParams = request.URL.Query()
		state       = queryParams[api.QueryParameterState]
		code        = queryParams[api.QueryParameterCode]
	)

	if ssoProvider.OIDCProvider == nil {
		// SSO misconfiguration scenario
		redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
	} else if len(code) == 0 {
		// Missing authorization code implies a credentials or form issue
		// Not explicitly covered, treat as technical issue
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
	} else if pkceVerifier, err := request.Cookie(api.AuthPKCECookieName); err != nil {
		// Missing PKCE verifier - likely a technical or config issue
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
	} else if len(state) == 0 {
		// Missing state parameter - treat as technical issue
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
	} else if stateCookie, err := request.Cookie(api.AuthStateCookieName); err != nil || stateCookie.Value != state[0] {
		// Invalid state - treat as technical issue or misconfiguration
		redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
	} else if provider, err := oidc.NewProvider(request.Context(), ssoProvider.OIDCProvider.Issuer); err != nil {
		log.Errorf("[OIDC] Failed to create OIDC provider: %v", err)
		// SSO misconfiguration scenario
		redirectToLoginPage(response, request, "Your SSO Connection failed, please contact your Administrator")
	} else {
		var (
			oidcVerifier = provider.Verifier(&oidc.Config{ClientID: ssoProvider.OIDCProvider.ClientID})
			oauth2Conf   = &oauth2.Config{
				ClientID:    ssoProvider.OIDCProvider.ClientID,
				Endpoint:    provider.Endpoint(),
				RedirectURL: getRedirectURL(request, ssoProvider), // Required as verification check
			}
		)

		if token, err := oauth2Conf.Exchange(request.Context(), code[0], oauth2.VerifierOption(pkceVerifier.Value)); err != nil {
			log.Errorf("[OIDC] Token exchange failed: %v", err)
			// SAML credentials issue equivalent for OIDC authentication
			redirectToLoginPage(response, request, "Your SSO was unable to authenticate your user, please contact your Administrator")
		} else if rawIDToken, ok := token.Extra("id_token").(string); !ok { // Extract the ID Token from OAuth2 token
			// Missing ID token - credentials issue
			redirectToLoginPage(response, request, "Your SSO was unable to authenticate your user, please contact your Administrator")
		} else if idToken, err := oidcVerifier.Verify(request.Context(), rawIDToken); err != nil {
			log.Errorf("[OIDC] ID token verification failed: %v", err)
			// Credentials issue scenario
			redirectToLoginPage(response, request, "Your SSO was unable to authenticate your user, please contact your Administrator")
		} else {
			// Extract custom claims
			var claims struct {
				Name        string `json:"name"`
				FamilyName  string `json:"family_name"`
				DisplayName string `json:"given_name"`
				Email       string `json:"email"`
				Verified    bool   `json:"email_verified"`
			}
			if err := idToken.Claims(&claims); err != nil {
				log.Errorf("[OIDC] Failed to parse claims: %v", err)
				// Technical or credentials issue
				// Not explicitly covered; treat as a technical issue
				redirectToLoginPage(response, request, "We’re having trouble connecting. Please check your internet and try again.")
			} else if claims.Email == "" {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "email claim is missing", request), response)
				return
			} else {
				if ssoProvider.Config.AutoProvision.Enabled {
					if user, err := s.db.LookupUser(request.Context(), claims.Email); err != nil {
						if errors.Is(err, database.ErrNotFound) {
							if role, err := s.db.GetRole(request.Context(), ssoProvider.Config.AutoProvision.DefaultRole); err != nil {
								api.HandleDatabaseError(request, response, err)
								return
							} else {
								user = model.User{
									EmailAddress:  null.StringFrom(claims.Email),
									PrincipalName: claims.Email,
									Roles:         model.Roles{role},
									SSOProviderID: null.Int32From(ssoProvider.ID),
									EULAAccepted:  true, // EULA Acceptance does not pertain to Bloodhound Community Edition; this flag is used for Bloodhound Enterprise users
									FirstName:     null.StringFrom(claims.Email),
									LastName:      null.StringFrom("Last name not found"),
								}

								if claims.DisplayName != "" {
									user.FirstName = null.StringFrom(claims.DisplayName)
								}

								if claims.FamilyName != "" {
									user.LastName = null.StringFrom(claims.FamilyName)
								}

								if _, err := s.db.CreateUser(request.Context(), user); err != nil {
									// It is safe to let this request drop into the CreateSSOSession function below to ensure proper audit logging
									log.Errorf("[OIDC] JIT User Create Error: %v", err)
								}
							}
						}
					}
				}

				s.authenticator.CreateSSOSession(request, response, claims.Email, ssoProvider)
			}
		}
	}
}
