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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/src/utils/validation"
	"golang.org/x/oauth2"
)

const oidcTokenExchangeBodyLimit = 1 << 20

var (
	ErrOIDCProviderMissing  = errors.New("oidc provider missing")
	ErrOIDCIssuerURLInvalid = errors.New("oidc provider issuer url invalid")
	ErrRoleIDInvalid        = errors.New("role id invalid")
	ErrEmailMissing         = errors.New("email missing")
)

type oidcClaims struct {
	Name              string `json:"name"`
	FamilyName        string `json:"family_name"`
	DisplayName       string `json:"given_name"`
	Email             string `json:"email"` // Not always present
	Verified          bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"` // Present in Entra claims, may be an email

	Roles []string `json:"roles"`
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
		api.RedirectToLoginURL(response, request, "Your SSO connection failed due to misconfiguration, please contact your Administrator")
	} else if state, err := config.GenerateRandomBase64String(77); err != nil {
		slog.WarnContext(request.Context(), fmt.Sprintf("[OIDC] Failed to generate state: %v", err))
		// Technical issues scenario
		api.RedirectToLoginURL(response, request, "We're having trouble connecting. Please check your internet and try again.")
	} else if provider, err := oidc.NewProvider(request.Context(), ssoProvider.OIDCProvider.Issuer); err != nil {
		slog.WarnContext(request.Context(), fmt.Sprintf("[OIDC] Failed to create OIDC provider: %v", err))
		// SSO misconfiguration or technical issue
		// Treat this as a misconfiguration scenario
		api.RedirectToLoginURL(response, request, "Your SSO connection failed due to misconfiguration, please contact your Administrator")
	} else {
		conf := &oauth2.Config{
			ClientID:    ssoProvider.OIDCProvider.ClientID,
			Endpoint:    provider.Endpoint(),
			RedirectURL: getRedirectURL(hostURL, ssoProvider),
			Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
		}

		// use PKCE to protect against CSRF attacks
		// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
		verifier := oauth2.GenerateVerifier()

		// Store PKCE on web browser in secure cookie for retrieval in callback
		api.SetSecureBrowserCookie(request, response, api.AuthPKCECookieName, verifier, time.Now().UTC().Add(time.Minute*7), true, http.SameSiteNoneMode)

		// Store State on web browser in secure cookie for retrieval in callback
		api.SetSecureBrowserCookie(request, response, api.AuthStateCookieName, state, time.Now().UTC().Add(time.Minute*7), true, http.SameSiteNoneMode)

		// Redirect user to consent page to ask for permission for the scopes specified above and specify POST callback
		redirectURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("response_mode", "form_post"))
		response.Header().Add(headers.Location.String(), redirectURL)
		response.WriteHeader(http.StatusFound)
	}
}

func (s ManagementResource) OIDCCallbackHandler(response http.ResponseWriter, request *http.Request, ssoProvider model.SSOProvider) {
	var (
		state = request.FormValue(api.FormParameterState)
		code  = request.FormValue(api.FormParameterCode)
	)

	// No matter what happens, wipe the auth cookies
	api.DeleteBrowserCookie(request, response, api.AuthStateCookieName)
	api.DeleteBrowserCookie(request, response, api.AuthPKCECookieName)

	if ssoProvider.OIDCProvider == nil {
		// SSO misconfiguration scenario
		api.RedirectToLoginURL(response, request, "Your SSO connection failed due to misconfiguration, please contact your Administrator")
	} else if code == "" {
		// Missing authorization code implies a credentials or form issue
		slog.WarnContext(request.Context(), "[OIDC] auth code is missing")
		api.RedirectToLoginURL(response, request, "Invalid SSO Provider response: `code` parameter is missing")
	} else if state == "" {
		// Missing state parameter - treat as technical issue
		slog.WarnContext(request.Context(), "[OIDC] state parameter is missing")
		api.RedirectToLoginURL(response, request, "Invalid SSO Provider response: `state` parameter is missing")
	} else if pkceVerifier, err := request.Cookie(api.AuthPKCECookieName); err != nil {
		// Missing PKCE verifier - likely a technical or config issue
		slog.WarnContext(request.Context(), "[OIDC] pkce cookie is missing")
		api.RedirectToLoginURL(response, request, "Invalid request: `pkce` is missing")
	} else if stateCookie, err := request.Cookie(api.AuthStateCookieName); err != nil {
		// Missing state - likely a technical or config issue
		slog.WarnContext(request.Context(), "[OIDC] state cookie is missing")
		api.RedirectToLoginURL(response, request, "Invalid request: `state` is missing")
	} else if stateCookie.Value != state {
		// State mismatch
		slog.WarnContext(request.Context(), "[OIDC] state does not match")
		api.RedirectToLoginURL(response, request, "Invalid: `state` do not match")
	} else if provider, err := oidc.NewProvider(request.Context(), ssoProvider.OIDCProvider.Issuer); err != nil {
		// SSO misconfiguration scenario
		slog.WarnContext(request.Context(), fmt.Sprintf("[OIDC] Failed to create OIDC provider: %v", err))
		api.RedirectToLoginURL(response, request, "Your SSO connection failed due to misconfiguration, please contact your Administrator")
	} else if claims, err := getOIDCClaims(request.Context(), provider, ssoProvider, pkceVerifier, code); err != nil {
		slog.WarnContext(request.Context(), fmt.Sprintf("[OIDC] %v", err))
		api.RedirectToLoginURL(response, request, fmt.Sprintf("Exchange failed: %s", err.Error()))
	} else if email, err := getEmailFromOIDCClaims(claims); errors.Is(err, ErrEmailMissing) { // Note email claims are not always present so we will check different claim keys for possible email
		slog.WarnContext(request.Context(), "[OIDC] Claims did not contain any valid email address")
		api.RedirectToLoginURL(response, request, "Claims invalid: no valid email address found")
	} else {
		if ssoProvider.Config.AutoProvision.Enabled {
			if err := jitOIDCUserCreation(request.Context(), ssoProvider, email, claims, s.db); err != nil {
				// It is safe to let this request drop into the CreateSSOSession function below to ensure proper audit logging
				slog.WarnContext(request.Context(), fmt.Sprintf("[OIDC] Error during JIT User Creation: %v", err))
			}
		}

		s.authenticator.CreateSSOSession(request, response, email, ssoProvider)
	}
}

// OIDC Token exchange adapted from golang.org/x/oauth2
func exchangeCodeForToken(reqCtx context.Context, ssoProvider model.SSOProvider, tokenUrl string, pkceVerifier *http.Cookie, code string) (*oauth2.Token, error) {
	var (
		hostUrl = *ctx.Get(reqCtx).Host
		payload = url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {ssoProvider.OIDCProvider.ClientID},
			"redirect_uri":  {getRedirectURL(hostUrl, ssoProvider)},
			"code":          {code},
			"code_verifier": {pkceVerifier.Value},
		}
	)

	if req, err := http.NewRequest("POST", tokenUrl, strings.NewReader(payload.Encode())); err != nil {
		return nil, fmt.Errorf("failed to init exchange request %v", err)
	} else {
		// Set custom headers
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationXWwwFormUrlencoded.String())
		req.Header.Set(headers.Origin.String(), hostUrl.String())

		r, err := http.DefaultClient.Do(req.WithContext(reqCtx))
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		if body, err := io.ReadAll(io.LimitReader(r.Body, oidcTokenExchangeBodyLimit)); err != nil {
			return nil, fmt.Errorf("cannot fetch token: %v", err)
		} else {
			var token *oauth2.Token
			content, _, _ := mime.ParseMediaType(r.Header.Get(headers.ContentType.String()))
			switch content {
			case mediatypes.ApplicationXWwwFormUrlencoded.String(), "text/plain":
				// some endpoints return a query string
				if vals, err := url.ParseQuery(string(body)); err != nil {
					return nil, fmt.Errorf("cannot parse token response: %v", err)
				} else {
					token = &oauth2.Token{
						AccessToken:  vals.Get("access_token"),
						TokenType:    vals.Get("token_type"),
						RefreshToken: vals.Get("refresh_token"),
					}
					token = token.WithExtra(vals)

					expires, _ := strconv.Atoi(vals.Get("expires_in"))
					if expires != 0 {
						token.Expiry = time.Now().Add(time.Duration(expires) * time.Second)
					}
				}
			default:
				if err = json.Unmarshal(body, &token); err != nil {
					return nil, fmt.Errorf("cannot parse token json: %v", err)
				}

				optionalFields := make(map[string]interface{})
				_ = json.Unmarshal(body, &optionalFields) // no error checks for optional fields
				token = token.WithExtra(optionalFields)
			}

			if token.AccessToken == "" {
				return nil, fmt.Errorf("server response missing access_token %v %v %v", token.Extra("error"), token.Extra("error_description"), token.Extra("error_uri"))
			}

			return token, nil
		}
	}
}

func getOIDCClaims(reqCtx context.Context, provider *oidc.Provider, ssoProvider model.SSOProvider, pkceVerifier *http.Cookie, code string) (oidcClaims, error) {
	var (
		oidcVerifier = provider.Verifier(&oidc.Config{ClientID: ssoProvider.OIDCProvider.ClientID})
		claims       = oidcClaims{}
	)

	if token, err := exchangeCodeForToken(reqCtx, ssoProvider, provider.Endpoint().TokenURL, pkceVerifier, code); err != nil {
		return claims, fmt.Errorf("token exchange: %v", err)
	} else if token == nil {
		return claims, fmt.Errorf("token is nil somehow... abort")
	} else if rawIDToken, ok := token.Extra("id_token").(string); !ok { // Extract the ID Token from OAuth2 token
		return claims, fmt.Errorf("token missing key id_token: %v", err)
	} else if idToken, err := oidcVerifier.Verify(reqCtx, rawIDToken); err != nil {
		return claims, fmt.Errorf("id token verification: %v", err)
	} else if err := idToken.Claims(&claims); err != nil {
		return claims, fmt.Errorf("parse claims: %v", err)
	} else {
		return claims, nil
	}
}

func getEmailFromOIDCClaims(claims oidcClaims) (string, error) {
	if claims.Email != "" {
		return claims.Email, nil
	} else if utils.IsValidEmail(claims.PreferredUsername) {
		return claims.PreferredUsername, nil
	}

	return "", ErrEmailMissing
}

func jitOIDCUserCreation(ctx context.Context, ssoProvider model.SSOProvider, email string, claims oidcClaims, u jitUserCreator) error {
	if roles, err := SanitizeAndGetRoles(ctx, ssoProvider.Config.AutoProvision, claims.Roles, u); err != nil {
		return fmt.Errorf("sanitize roles: %v", err)
	} else if len(roles) != 1 {
		return fmt.Errorf("invalid roles")
	} else if _, err := u.LookupUser(ctx, email); err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("lookup user: %v", err)
	} else if errors.Is(err, database.ErrNotFound) {
		var user = model.User{
			EmailAddress:  null.StringFrom(email),
			PrincipalName: email,
			Roles:         roles,
			SSOProviderID: null.Int32From(ssoProvider.ID),
			EULAAccepted:  true, // EULA Acceptance does not pertain to Bloodhound Community Edition; this flag is used for Bloodhound Enterprise users
			FirstName:     null.StringFrom(email),
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
