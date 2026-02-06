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

package api

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../LICENSE.header -destination=./mocks/authenticator.go -package=mocks . Authenticator,AuthExtensions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/crypto"
	"github.com/specterops/bloodhound/packages/go/headers"
)

var (
	ErrInvalidAuth                  = errors.New("invalid authentication")
	ErrNoUserSecret                 = errors.New("user does not have a secret auth provider registered")
	ErrUserDisabled                 = errors.New("user disabled")
	ErrUserNotAuthorizedForProvider = errors.New("user not authorized for this provider")
	ErrInvalidAuthProvider          = errors.New("invalid auth provider")
	ErrApiKeysDisabled              = errors.New("use of API keys has been disabled")
)

type LoginRequest struct {
	LoginMethod string `json:"login_method"`
	Username    string `json:"username"`
	Secret      string `json:"secret,omitempty"`
	OTP         string `json:"otp,omitempty"`
}

type LoginDetails struct {
	User         model.User
	SessionToken string
}

type LoginResponse struct {
	UserID       string `json:"user_id"`
	AuthExpired  bool   `json:"auth_expired"`
	SessionToken string `json:"session_token"`
}

func parseRequestDate(rawDate string) (time.Time, error) {
	if requestDate, err := time.Parse(time.RFC3339, rawDate); err != nil {
		if requestDate, err := time.Parse(time.RFC3339Nano, rawDate); err == nil {
			return requestDate, nil
		}

		return requestDate, fmt.Errorf("malformed request date: %w", err)
	} else {
		return requestDate, nil
	}
}

type AuthExtensions interface {
	InitContextFromToken(ctx context.Context, authToken model.AuthToken) (auth.Context, error)
	InitContextFromClaims(ctx context.Context, claims *jwt.RegisteredClaims) (auth.Context, error)
	ParseClaimsAndVerifySignature(ctx context.Context, jwtToken string) (*jwt.RegisteredClaims, error)
}

type authExtensions struct {
	cfg config.Configuration
	db  database.Database
}

func NewAuthExtensions(cfg config.Configuration, db database.Database) AuthExtensions {
	return authExtensions{
		cfg: cfg,
		db:  db,
	}
}

func (s authExtensions) InitContextFromToken(ctx context.Context, authToken model.AuthToken) (auth.Context, error) {
	if authToken.UserID.Valid {
		if user, err := s.db.GetUser(ctx, authToken.UserID.UUID); err != nil {
			return auth.Context{}, err
		} else {
			return auth.Context{
				Owner: user,
			}, nil
		}
	}

	return auth.Context{}, database.ErrNotFound
}

func (s authExtensions) InitContextFromClaims(_ context.Context, _ *jwt.RegisteredClaims) (auth.Context, error) {
	return auth.Context{}, nil
}

func (s authExtensions) ParseClaimsAndVerifySignature(ctx context.Context, jwtToken string) (*jwt.RegisteredClaims, error) {
	claims := jwt.RegisteredClaims{}
	if token, err := jwt.ParseWithClaims(jwtToken, &claims, s.jwtKeyFunc); err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return &claims, ErrInvalidAuth
		}
		return &claims, err
	} else if !token.Valid {
		return &claims, ErrInvalidAuth
	}

	return &claims, nil
}

func (s authExtensions) jwtKeyFunc(_ *jwt.Token) (any, error) {
	return s.cfg.Crypto.JWT.SigningKeyBytes()
}

type Authenticator interface {
	LoginWithSecret(ctx context.Context, loginRequest LoginRequest) (LoginDetails, error)
	Logout(ctx context.Context, userSession model.UserSession)
	ValidateSecret(ctx context.Context, secret string, authSecret model.AuthSecret) error
	ValidateRequestSignature(tokenID uuid.UUID, request *http.Request, serverTime time.Time) (auth.Context, int, error)
	CreateSession(ctx context.Context, user model.User, authProvider any) (string, error)
	CreateSSOSession(request *http.Request, response http.ResponseWriter, principalNameOrEmail string, ssoProvider model.SSOProvider)
	ValidateBearerToken(ctx context.Context, jwtToken string) (auth.Context, error)
	ValidateSession(ctx context.Context, jwtTokenString string) (auth.Context, error)
}

type AuthenticatorBase struct {
	cfg             config.Configuration
	db              database.Database
	authExtensions  AuthExtensions
	secretDigester  crypto.SecretDigester
	concurrencyLock chan struct{}
}

func NewAuthenticator(cfg config.Configuration, db database.Database, authExtensions AuthExtensions) Authenticator {
	return AuthenticatorBase{
		cfg:             cfg,
		db:              db,
		authExtensions:  authExtensions,
		secretDigester:  cfg.Crypto.Argon2.NewDigester(),
		concurrencyLock: make(chan struct{}, 1),
	}
}

func (s AuthenticatorBase) auditLogin(requestContext context.Context, commitID uuid.UUID, status model.AuditLogEntryStatus, user model.User, fields types.JSONUntypedObject) {
	bhCtx := ctx.Get(requestContext)
	auditLog := model.AuditLog{
		Action:          model.AuditLogActionLoginAttempt,
		Fields:          fields,
		RequestID:       bhCtx.RequestID,
		SourceIpAddress: bhCtx.RequestIP,
		Status:          status,
		CommitID:        commitID,
	}

	if user.PrincipalName != "" {
		auditLog.ActorID = user.ID.String()
		auditLog.ActorName = user.PrincipalName
		auditLog.ActorEmail = user.EmailAddress.ValueOrZero()
	}

	err := s.db.CreateAuditLog(requestContext, auditLog)
	if err != nil {
		slog.WarnContext(requestContext, "Failed to write login audit log", attr.Error(err))
	}
}

func (s AuthenticatorBase) validateSecretLogin(ctx context.Context, loginRequest LoginRequest) (model.User, string, error) {
	if user, err := s.db.LookupUser(ctx, loginRequest.Username); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, "", ErrInvalidAuth
		}

		return model.User{}, "", FormatDatabaseError(err)
	} else if user.AuthSecret == nil {
		return user, "", ErrNoUserSecret
	} else if err := s.ValidateSecret(ctx, loginRequest.Secret, *user.AuthSecret); err != nil {
		return user, "", err
	} else if err = auth.ValidateTOTPSecret(loginRequest.OTP, *user.AuthSecret); err != nil {
		return user, "", err
	} else if sessionToken, err := s.CreateSession(ctx, user, *user.AuthSecret); err != nil {
		return user, "", err
	} else {
		return user, sessionToken, nil
	}
}

func (s AuthenticatorBase) LoginWithSecret(ctx context.Context, loginRequest LoginRequest) (LoginDetails, error) {
	auditLogFields := types.JSONUntypedObject{"username": loginRequest.Username, "auth_type": auth.ProviderTypeSecret}

	if commitID, err := uuid.NewV4(); err != nil {
		slog.ErrorContext(ctx, "Error generating commit ID for login", attr.Error(err))
		return LoginDetails{}, err
	} else {
		s.auditLogin(ctx, commitID, model.AuditLogStatusIntent, model.User{}, auditLogFields)

		if user, sessionToken, err := s.validateSecretLogin(ctx, loginRequest); err != nil {
			auditLogFields["error"] = err
			s.auditLogin(ctx, commitID, model.AuditLogStatusFailure, user, auditLogFields)
			return LoginDetails{}, err
		} else {
			s.auditLogin(ctx, commitID, model.AuditLogStatusSuccess, user, auditLogFields)
			return LoginDetails{
				User:         user,
				SessionToken: sessionToken,
			}, nil
		}
	}
}

func (s AuthenticatorBase) Logout(ctx context.Context, userSession model.UserSession) {
	s.db.EndUserSession(ctx, userSession)
}

func (s AuthenticatorBase) ValidateSecret(ctx context.Context, secret string, authSecret model.AuthSecret) error {
	select {
	case s.concurrencyLock <- struct{}{}:
		defer func() {
			<-s.concurrencyLock
		}()

		return ValidateSecret(s.secretDigester, secret, authSecret)
	case <-ctx.Done():
		return context.Background().Err()
	}
}

func ValidateSecret(secretDigester crypto.SecretDigester, secret string, authSecret model.AuthSecret) error {
	if secretDigester.Method() != authSecret.DigestMethod {
		return fmt.Errorf("secret provider for user contains an unsupported digest method: %s", authSecret.DigestMethod)
	}

	if digest, err := secretDigester.ParseDigest(authSecret.Digest); err != nil {
		return err
	} else if digest.Validate(secret) {
		return nil
	}

	return ErrInvalidAuth
}

func validateRequestTime(serverTime, requestDate time.Time) error {
	const (
		// Max TTL for a request signature
		maxClockSkew = time.Hour
	)

	// Adjust the time to the correct timezone
	serverAdjustedTime := serverTime.In(requestDate.Location())

	if serverAdjustedTime.Before(requestDate) {
		if delta := requestDate.Sub(serverAdjustedTime); delta > maxClockSkew {
			return fmt.Errorf("signature too far ahead by %.2f minutes. Server time is %s. Client time is %s", (delta - maxClockSkew).Minutes(), serverAdjustedTime.Format(time.RFC3339), requestDate.Format(time.RFC3339))
		}
	} else {
		if delta := serverAdjustedTime.Sub(requestDate); delta > maxClockSkew {
			return fmt.Errorf("signature too far behind by %.2f minutes. Server time is %s. Client time is %s", (delta - maxClockSkew).Minutes(), serverAdjustedTime.Format(time.RFC3339), requestDate.Format(time.RFC3339))
		}
	}

	return nil
}

func handleAuthDBError(err error) (auth.Context, int, error) {
	if errors.Is(err, database.ErrNotFound) {
		return auth.Context{}, http.StatusUnauthorized, FormatDatabaseError(err)
	} else {
		return auth.Context{}, http.StatusInternalServerError, FormatDatabaseError(err)
	}
}

// ThresholdLargePayload represents the request payload size in bytes before signed requests are validated in a manner that doesn't run the risk of exhausting system memory.
// This value was derived assuming the host system operates on a SSD with maximum read/write throughput of ~500MiBps and targets an optimal validation time of ~0.1s for "large" payloads.
// e.g. - 500 MiBps * 0.1s = 50MiB
const ThresholdLargePayload int64 = 50 << 20

func (s AuthenticatorBase) ValidateRequestSignature(tokenID uuid.UUID, request *http.Request, serverTime time.Time) (auth.Context, int, error) {
	if requestDateHeader := request.Header.Get(headers.RequestDate.String()); requestDateHeader == "" {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("no request date header")
	} else if requestDate, err := parseRequestDate(requestDateHeader); err != nil {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("malformed request date: %w", err)
	} else if signatureHeader := request.Header.Get(headers.Signature.String()); signatureHeader == "" {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("no signature header")
	} else if apiKeysEnabled := appcfg.GetAPITokensParameter(context.Background(), s.db); !apiKeysEnabled {
		return auth.Context{}, http.StatusUnauthorized, ErrApiKeysDisabled
	} else if signatureBytes, err := base64.StdEncoding.DecodeString(signatureHeader); err != nil {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("malformed signature header: %w", err)
	} else if authToken, err := s.db.GetAuthToken(request.Context(), tokenID); err != nil {
		return handleAuthDBError(err)
	} else if authContext, err := s.authExtensions.InitContextFromToken(request.Context(), authToken); err != nil {
		return handleAuthDBError(err)
	} else if user, isUser := auth.GetUserFromAuthCtx(authContext); isUser && user.IsDisabled {
		return authContext, http.StatusForbidden, ErrUserDisabled
	} else if err := validateRequestTime(serverTime, requestDate); err != nil {
		return auth.Context{}, http.StatusUnauthorized, err
	} else {
		var (
			readCloser io.ReadCloser
			teeReader  io.Reader
		)

		if request.Body != nil {
			if request.ContentLength > ThresholdLargePayload || request.ContentLength == -1 {
				// Request payload is "large" or the size is unknown; tee byte stream to disk for subsequent reads to avoid exhausting system memory
				if tempFile, err := NewSelfDestructingTempFile(s.cfg.TempDirectory(), "bh-request-"); err != nil {
					return auth.Context{}, http.StatusInternalServerError, fmt.Errorf("unable to validate request signature: %w", err)
				} else {
					readCloser = tempFile
					teeReader = io.TeeReader(request.Body, tempFile)
				}
			} else {
				// Request payload is "small"; tee byte stream to buffer for subsequent reads
				var buf bytes.Buffer
				teeReader = io.TeeReader(request.Body, &buf)
				readCloser = io.NopCloser(&buf)
			}
		}

		if digestNow, err := NewRequestSignature(request.Context(), sha256.New, authToken.Key, requestDate.Format(time.RFC3339), request.Method, request.RequestURI, teeReader); err != nil {
			if readCloser != nil {
				readCloser.Close()
			}
			return authContext, http.StatusInternalServerError, err
		} else {
			if subtle.ConstantTimeCompare(signatureBytes, digestNow) != 1 {
				if readCloser != nil {
					readCloser.Close()
				}
				return authContext, http.StatusUnauthorized, fmt.Errorf("digest validation failed: signature digest mismatch")
			}

			authToken.LastAccess = time.Now().UTC()

			if err := s.db.UpdateAuthToken(request.Context(), authToken); err != nil {
				slog.ErrorContext(request.Context(), "Error updating last access on AuthToken", attr.Error(err))
			}

			if sdtf, ok := readCloser.(*SelfDestructingTempFile); ok {
				sdtf.Seek(0, io.SeekStart)
			}

			request.Body = readCloser

			return authContext, http.StatusOK, nil
		}
	}
}

func DeleteBrowserCookie(request *http.Request, response http.ResponseWriter, name string) {
	SetSecureBrowserCookie(request, response, name, "", time.Now().UTC(), false, 0)
}

func SetSecureBrowserCookie(request *http.Request, response http.ResponseWriter, name, value string, expires time.Time, httpOnly bool, sameSite http.SameSite) {
	var (
		hostURL = *ctx.FromRequest(request).Host
		isHttps = hostURL.Scheme == "https"
	)

	// If sameSite is not explicitly set, we want to rely on the host scheme
	if sameSite == 0 && isHttps {
		sameSite = http.SameSiteStrictMode
	}

	// NOTE: Browsers will not set localhost cookies with sameSite set to None. This is a local network SSO workaround
	if strings.Contains(hostURL.Hostname(), "localhost") && sameSite == http.SameSiteNoneMode {
		sameSite = http.SameSiteDefaultMode
	}

	// Set the token cookie
	http.SetCookie(response, &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expires,
		Secure:   isHttps,
		HttpOnly: httpOnly,
		SameSite: sameSite,
		Path:     "/",
	})
}

func (s AuthenticatorBase) CreateSSOSession(request *http.Request, response http.ResponseWriter, principalNameOrEmail string, ssoProvider model.SSOProvider) {
	var (
		hostURL    = *ctx.FromRequest(request).Host
		requestCtx = request.Context()
		err        error

		authProvider any
		user         model.User

		commitID        uuid.UUID
		auditLogFields  = types.JSONUntypedObject{"username": principalNameOrEmail, "sso_provider_id": ssoProvider.ID}
		auditLogOutcome = model.AuditLogStatusFailure
	)

	switch ssoProvider.Type {
	case model.SessionAuthProviderSAML:
		auditLogFields["auth_type"] = auth.ProviderTypeSAML
		if ssoProvider.SAMLProvider != nil {
			auditLogFields["saml_provider_id"] = ssoProvider.SAMLProvider.ID
			authProvider = *ssoProvider.SAMLProvider
		}
	case model.SessionAuthProviderOIDC:
		auditLogFields["auth_type"] = auth.ProviderTypeOIDC
		if ssoProvider.OIDCProvider != nil {
			auditLogFields["oidc_provider_id"] = ssoProvider.OIDCProvider.ID
			authProvider = *ssoProvider.OIDCProvider
		}
	}

	// Generate commit ID for audit logging
	if commitID, err = uuid.NewV4(); err != nil {
		slog.WarnContext(request.Context(), "[SSO] Error generating commit ID for login", attr.Error(err))
		RedirectToLoginURL(response, request, "We’re having trouble connecting. Please check your internet and try again.")
		return
	}

	// Log the intent to authenticate
	s.auditLogin(requestCtx, commitID, model.AuditLogStatusIntent, user, auditLogFields)

	// Log authentication success or failure
	defer func() {
		s.auditLogin(requestCtx, commitID, auditLogOutcome, user, auditLogFields)
	}()

	if user, err = s.db.LookupUser(requestCtx, principalNameOrEmail); err != nil {
		auditLogFields["error"] = err
		if !errors.Is(err, database.ErrNotFound) {
			slog.WarnContext(request.Context(), "[SSO] Error looking up user", attr.Error(err))
			RedirectToLoginURL(response, request, "We’re having trouble connecting. Please check your internet and try again.")
		} else {
			RedirectToLoginURL(response, request, "Your user is not allowed, please contact your Administrator")
		}
	} else {
		if !user.SSOProviderID.Valid || ssoProvider.ID != user.SSOProviderID.Int32 {
			auditLogFields["error"] = ErrUserNotAuthorizedForProvider
			RedirectToLoginURL(response, request, "Your user is not allowed, please contact your Administrator")
			return
		}

		if sessionJWT, err := s.CreateSession(requestCtx, user, authProvider); err != nil {
			auditLogFields["error"] = err
			if locationURL := URLJoinPath(hostURL, UserDisabledPath); errors.Is(err, ErrUserDisabled) {
				response.Header().Add(headers.Location.String(), locationURL.String())
				response.WriteHeader(http.StatusFound)
			} else {
				slog.WarnContext(request.Context(), "[SSO] Error creating session", attr.Error(err))
				RedirectToLoginURL(response, request, "We’re having trouble connecting. Please check your internet and try again.")
			}
		} else {
			auditLogOutcome = model.AuditLogStatusSuccess

			locationURL := URLJoinPath(hostURL, UserInterfacePath)

			// Set the token cookie, httpOnly must be false for the UI to pick up and store token
			SetSecureBrowserCookie(request, response, AuthTokenCookieName, sessionJWT, time.Now().UTC().Add(appcfg.GetSessionTTLHours(request.Context(), s.db)), false, 0)

			// Redirect back to the UI landing page
			response.Header().Add(headers.Location.String(), locationURL.String())
			response.WriteHeader(http.StatusFound)
		}
	}
}

func (s AuthenticatorBase) CreateSession(ctx context.Context, user model.User, authProvider any) (string, error) {
	if user.IsDisabled {
		return "", ErrUserDisabled
	}

	slog.InfoContext(
		ctx,
		"Creating session for user",
		slog.String("user_id", user.ID.String()),
		slog.String("principal_name", user.PrincipalName),
	)

	userSession := model.UserSession{
		User:      user,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(appcfg.GetSessionTTLHours(ctx, s.db)),
	}

	switch typedAuthProvider := authProvider.(type) {
	case model.AuthSecret:
		userSession.AuthProviderType = model.SessionAuthProviderSecret
		userSession.AuthProviderID = typedAuthProvider.ID
	case model.SAMLProvider:
		userSession.AuthProviderType = model.SessionAuthProviderSAML
		userSession.AuthProviderID = typedAuthProvider.ID
	case model.OIDCProvider:
		userSession.AuthProviderType = model.SessionAuthProviderOIDC
		userSession.AuthProviderID = typedAuthProvider.ID
	default:
		return "", ErrInvalidAuthProvider
	}

	if newSession, err := s.db.CreateUserSession(ctx, userSession); err != nil {
		return "", FormatDatabaseError(err)
	} else if signingKeyBytes, err := s.cfg.Crypto.JWT.SigningKeyBytes(); err != nil {
		return "", err
	} else {
		var (
			jwtClaims = jwt.RegisteredClaims{
				Issuer:    s.cfg.GetRootURLHost(),
				ID:        strconv.FormatInt(newSession.ID, 10),
				Subject:   user.ID.String(),
				IssuedAt:  jwt.NewNumericDate(newSession.CreatedAt.UTC()),
				ExpiresAt: jwt.NewNumericDate(newSession.ExpiresAt.UTC()),
			}

			token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
		)

		return token.SignedString(signingKeyBytes)
	}
}

func (s AuthenticatorBase) ValidateBearerToken(ctx context.Context, jwtToken string) (auth.Context, error) {
	if claims, err := s.authExtensions.ParseClaimsAndVerifySignature(ctx, jwtToken); err != nil {
		return auth.Context{}, err
	} else if authContext, err := s.authExtensions.InitContextFromClaims(ctx, claims); err != nil {
		slog.ErrorContext(ctx, "Error initializing auth context from claims", attr.Error(err))
		return auth.Context{}, err
	} else if authContext.Owner == nil {
		// The above logic is currently used to determine if the token is created from BloodHound. If nil, it was created by BloodHound.
		slog.DebugContext(ctx, "No owner claim found for token, defaulting to BloodHound provided token")
		if authContext, err = s.ValidateSession(ctx, claims.ID); err != nil {
			return auth.Context{}, err
		} else {
			return authContext, nil
		}
	} else {
		return authContext, nil
	}
}

func (s AuthenticatorBase) ValidateSession(ctx context.Context, claimsID string) (auth.Context, error) {

	if sessionID, err := strconv.ParseInt(claimsID, 10, 64); err != nil {
		slog.InfoContext(
			ctx,
			"Sessions ID is invalid",
			slog.String("claims_id", claimsID),
			attr.Error(err),
		)
		return auth.Context{}, ErrInvalidAuth
	} else if session, err := s.db.GetUserSession(ctx, sessionID); err != nil {
		slog.InfoContext(
			ctx,
			"Unable to find session",
			slog.String("claims_id", claimsID),
			attr.Error(err),
		)
		return auth.Context{}, ErrInvalidAuth
	} else if session.Expired() {
		slog.InfoContext(ctx, "Session is expired", slog.String("claims_id", claimsID))
		return auth.Context{}, ErrInvalidAuth
	} else {
		authContext := auth.Context{
			Owner:   session.User,
			Session: session,
		}

		if session.AuthProviderType == model.SessionAuthProviderSecret && session.User.AuthSecret == nil {
			slog.DebugContext(ctx, "No auth secret found for user", slog.String("user_id", session.UserID.String()))
			return auth.Context{}, ErrNoUserSecret
		} else if session.AuthProviderType == model.SessionAuthProviderSecret && session.User.AuthSecret.Expired() {
			var (
				authManageSelfPermission = auth.Permissions().AuthManageSelf
				permissions              model.Permissions
			)

			if session.User.Roles.Permissions().Has(authManageSelfPermission) {
				permissions = append(permissions, authManageSelfPermission)
			}

			authContext.PermissionOverrides = auth.PermissionOverrides{
				Enabled:     true,
				Permissions: permissions,
			}

			// EULA Acceptance does not pertain to Bloodhound Community Edition; this flag is used for Bloodhound Enterprise users.
			// This value is automatically set to true for Bloodhound Community Edition in the patchEULAAcceptance and CreateUser functions.
		} else if !session.User.EULAAccepted {
			authContext.PermissionOverrides = auth.PermissionOverrides{
				Enabled: true,
				Permissions: model.Permissions{
					auth.Permissions().AuthAcceptEULA,
				},
			}
		}

		return authContext, nil
	}
}
