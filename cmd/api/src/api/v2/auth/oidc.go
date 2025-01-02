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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/validation"
	"golang.org/x/oauth2"
)

var (
	ErrOIDCProviderMissing  = errors.New("oidc provider missing")
	ErrOIDCIssuerURLInvalid = errors.New("oidc provider issuer url invalid")
	ErrRoleIDInvalid        = errors.New("role id invalid")
	ErrEmailClaimMissing    = errors.New("")
)

type oidcClaims struct {
	Name        string `json:"name"`
	FamilyName  string `json:"family_name"`
	DisplayName string `json:"given_name"`
	Email       string `json:"email"`
	Verified    bool   `json:"email_verified"`
}

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

	if err := api.ReadJSONRequestPayloadLimited(&upsertReq, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if ssoProvider, err := updateOIDCProvider(request.Context(), ssoProvider, upsertReq, s.db); errors.Is(err, ErrOIDCProviderMissing) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if errors.Is(err, ErrOIDCIssuerURLInvalid) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "issuer url is invalid", request), response)
	} else if errors.Is(err, ErrRoleIDInvalid) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id is invalid", request), response)
	} else if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if oidcProvider, err := s.db.UpdateOIDCProvider(request.Context(), ssoProvider); errors.Is(err, database.ErrDuplicateSSOProviderName) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseSSOProviderDuplicateName, request), response)
	} else if err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), oidcProvider, http.StatusOK, response)
	}
}

func updateOIDCProvider(ctx context.Context, ssoProvider model.SSOProvider, upsertReq UpsertOIDCProviderRequest, r getRoler) (model.SSOProvider, error) {
	if ssoProvider.OIDCProvider == nil {
		return ssoProvider, ErrOIDCProviderMissing
	}

	if upsertReq.Name != "" {
		ssoProvider.Name = upsertReq.Name
	}

	if upsertReq.ClientID != "" {
		ssoProvider.OIDCProvider.ClientID = upsertReq.ClientID
	}

	if upsertReq.Issuer != "" {
		if _, err := url.ParseRequestURI(upsertReq.Issuer); err != nil {
			return ssoProvider, ErrOIDCIssuerURLInvalid
		}

		ssoProvider.OIDCProvider.Issuer = upsertReq.Issuer
	}

	// Need to ensure that if no config is specified, we don't accidentally wipe the existing configuration
	if upsertReq.Config != nil {
		if !upsertReq.Config.AutoProvision.Enabled {
			ssoProvider.Config.AutoProvision = model.SSOProviderAutoProvisionConfig{}
		} else if _, err := r.GetRole(ctx, upsertReq.Config.AutoProvision.DefaultRoleId); err != nil {
			return ssoProvider, ErrRoleIDInvalid
		} else {
			ssoProvider.Config.AutoProvision = upsertReq.Config.AutoProvision
		}
	}

	return ssoProvider, nil
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
	} else if _, err := s.db.GetRole(request.Context(), upsertReq.Config.AutoProvision.DefaultRoleId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "role id is invalid", request), response)
	} else if oidcProvider, err := s.db.CreateOIDCProvider(request.Context(), upsertReq.Name, upsertReq.Issuer, upsertReq.ClientID, *upsertReq.Config); errors.Is(err, database.ErrDuplicateSSOProviderName) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseSSOProviderDuplicateName, request), response)
	} else if err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), oidcProvider, http.StatusCreated, response)
	}
}

func getRedirectURL(hostUrl url.URL, provider model.SSOProvider) string {
	return fmt.Sprintf("%s/api/v2/sso/%s/callback", hostUrl.String(), provider.Slug)
}

func (s ManagementResource) OIDCLoginHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	var hostURL = *ctx.Get(request.Context()).Host

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
			RedirectURL: getRedirectURL(hostURL, ssoProvider),
			Scopes:      []string{"openid", "profile", "email"},
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
	} else if claims, err := getOIDCClaims(request.Context(), provider, ssoProvider, pkceVerifier, code[0]); err != nil {
		log.Errorf("[OIDC] %v", err)
		redirectToLoginPage(response, request, "Your SSO was unable to authenticate your user, please contact your Administrator")
	} else {
		if ssoProvider.Config.AutoProvision.Enabled {
			if err := jitOIDCUserCreation(request.Context(), ssoProvider, claims, s.db); err != nil {
				// It is safe to let this request drop into the CreateSSOSession function below to ensure proper audit logging
				log.Errorf("[OIDC] Error during JIT User Creation: %v", err)
			}
		}

		s.authenticator.CreateSSOSession(request, response, claims.Email, ssoProvider)
	}
}

func getOIDCClaims(reqCtx context.Context, provider *oidc.Provider, ssoProvider model.SSOProvider, pkceVerifier *http.Cookie, code string) (oidcClaims, error) {
	var (
		hostURL      = *ctx.Get(reqCtx).Host
		oidcVerifier = provider.Verifier(&oidc.Config{ClientID: ssoProvider.OIDCProvider.ClientID})
		oauth2Conf   = &oauth2.Config{
			ClientID:    ssoProvider.OIDCProvider.ClientID,
			Endpoint:    provider.Endpoint(),
			RedirectURL: getRedirectURL(hostURL, ssoProvider), // Required as verification check
		}
		claims = oidcClaims{}
	)

	if token, err := oauth2Conf.Exchange(reqCtx, code, oauth2.VerifierOption(pkceVerifier.Value)); err != nil {
		return claims, fmt.Errorf("id token exchange: %v", err)
	} else if rawIDToken, ok := token.Extra("id_token").(string); !ok { // Extract the ID Token from OAuth2 token
		return claims, fmt.Errorf("id token missing key id_token: %v", err)
	} else if idToken, err := oidcVerifier.Verify(reqCtx, rawIDToken); err != nil {
		return claims, fmt.Errorf("id token verification: %v", err)
	} else if err := idToken.Claims(&claims); err != nil {
		return claims, fmt.Errorf("parse claims: %v", err)
	} else if claims.Email == "" {
		return claims, ErrEmailClaimMissing
	} else {
		return claims, nil
	}
}

func jitOIDCUserCreation(ctx context.Context, ssoProvider model.SSOProvider, claims oidcClaims, u jitUserCreator) error {
	if role, err := u.GetRole(ctx, ssoProvider.Config.AutoProvision.DefaultRoleId); err != nil {
		return fmt.Errorf("get role: %v", err)
	} else if _, err := u.LookupUser(ctx, claims.Email); err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("lookup user: %v", err)
	} else if errors.Is(err, database.ErrNotFound) {
		var user = model.User{
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

		if _, err := u.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("create user: %v", err)
		}
	}

	return nil
}
