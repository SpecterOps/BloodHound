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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../LICENSE.header -destination=./mocks/authenticator.go -package=mocks . Authenticator

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
)

const (
	ErrInvalidAuth  = errors.Error("invalid authentication")
	ErrNoUserSecret = errors.Error("user does not have a secret auth provider registered")
	ErrUserDisabled = errors.Error("user disabled")
)

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

type Authenticator interface {
	LoginWithSecret(ctx context.Context, loginRequest LoginRequest) (LoginDetails, error)
	Logout(userSession model.UserSession)
	ValidateSecret(ctx context.Context, secret string, authSecret model.AuthSecret) error
	ValidateRequestSignature(tokenID uuid.UUID, request *http.Request, serverTime time.Time) (auth.Context, int, error)
	CreateSession(user model.User, authProvider any) (string, error)
	ValidateSession(jwtTokenString string) (auth.Context, error)
}

type authenticator struct {
	cfg             config.Configuration
	db              database.Database
	secretDigester  crypto.SecretDigester
	concurrencyLock chan struct{}
	ctxInitializer  database.AuthContextInitializer
}

func NewAuthenticator(cfg config.Configuration, db database.Database, ctxInitializer database.AuthContextInitializer) Authenticator {
	return authenticator{
		cfg:             cfg,
		db:              db,
		secretDigester:  cfg.Crypto.Argon2.NewDigester(),
		concurrencyLock: make(chan struct{}, 1),
		ctxInitializer:  ctxInitializer,
	}
}

func (s authenticator) auditLogin(requestContext context.Context, commitID uuid.UUID, user model.User, loginRequest LoginRequest, status string, loginError error) {
	bhCtx := ctx.Get(requestContext)
	auditLog := model.AuditLog{
		Action:          "LoginAttempt",
		Fields:          types.JSONUntypedObject{"username": loginRequest.Username},
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

	if status == string(model.AuditStatusFailure) {
		auditLog.Fields["error"] = loginError
	}

	s.db.CreateAuditLog(auditLog)
}

func (s authenticator) validateSecretLogin(ctx context.Context, loginRequest LoginRequest) (model.User, string, error) {
	if user, err := s.db.LookupUser(loginRequest.Username); err != nil {
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
	} else if sessionToken, err := s.CreateSession(user, *user.AuthSecret); err != nil {
		return user, "", err
	} else {
		return user, sessionToken, nil
	}
}

func (s authenticator) LoginWithSecret(ctx context.Context, loginRequest LoginRequest) (LoginDetails, error) {
	var (
		commitID     uuid.UUID
		err          error
		sessionToken string
		user         model.User
	)

	commitID, err = uuid.NewV4()
	if err != nil {
		log.Errorf("error generating commit ID for login: %s", err)
		return LoginDetails{}, err
	}

	s.auditLogin(ctx, commitID, user, loginRequest, string(model.AuditStatusIntent), err)

	user, sessionToken, err = s.validateSecretLogin(ctx, loginRequest)

	if err != nil {
		s.auditLogin(ctx, commitID, user, loginRequest, string(model.AuditStatusFailure), err)
		return LoginDetails{}, err
	} else {
		s.auditLogin(ctx, commitID, user, loginRequest, string(model.AuditStatusSuccess), err)
		return LoginDetails{
			User:         user,
			SessionToken: sessionToken,
		}, nil
	}
}

func (s authenticator) Logout(userSession model.UserSession) {
	s.db.EndUserSession(userSession)
}

func (s authenticator) ValidateSecret(ctx context.Context, secret string, authSecret model.AuthSecret) error {
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

func (s authenticator) ValidateRequestSignature(tokenID uuid.UUID, request *http.Request, serverTime time.Time) (auth.Context, int, error) {
	if requestDateHeader := request.Header.Get(headers.RequestDate.String()); requestDateHeader == "" {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("no request date header")
	} else if requestDate, err := parseRequestDate(requestDateHeader); err != nil {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("malformed request date: %w", err)
	} else if signatureHeader := request.Header.Get(headers.Signature.String()); signatureHeader == "" {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("no signature header")
	} else if signatureBytes, err := base64.StdEncoding.DecodeString(signatureHeader); err != nil {
		return auth.Context{}, http.StatusBadRequest, fmt.Errorf("malformed signature header: %w", err)
	} else if authToken, err := s.db.GetAuthToken(tokenID); err != nil {
		return handleAuthDBError(err)
	} else if authContext, err := s.ctxInitializer.InitContextFromToken(authToken); err != nil {
		return handleAuthDBError(err)
	} else if user, isUser := auth.GetUserFromAuthCtx(authContext); isUser && user.IsDisabled {
		return authContext, http.StatusForbidden, errors.Error("user disabled")
	} else if err := validateRequestTime(serverTime, requestDate); err != nil {
		return auth.Context{}, http.StatusUnauthorized, err
	} else {
		// Read the body of the request to compute the actual signature. This forces the body into memory so there may
		// be a scaling pain-point here.
		if readBody, digestNow, err := GenerateRequestSignature(authToken.Key, requestDate.Format(time.RFC3339), auth.HMAC_SHA2_256, request.Method, request.RequestURI, request.Body); err != nil {
			return authContext, http.StatusInternalServerError, err
		} else {
			request.Body = &readerDelegatedCloser{
				source: readBody,
				closer: request.Body,
			}

			if subtle.ConstantTimeCompare(signatureBytes, digestNow) != 1 {
				return authContext, http.StatusUnauthorized, fmt.Errorf("digest validation failed: signature digest mismatch")
			}

			authToken.LastAccess = time.Now().UTC()

			if err := s.db.UpdateAuthToken(authToken); err != nil {
				log.Errorf("Error updating last access on AuthToken: %v", err)
			}

			return authContext, http.StatusOK, nil
		}
	}
}

func (s authenticator) CreateSession(user model.User, authProvider any) (string, error) {
	if user.IsDisabled {
		return "", ErrUserDisabled
	}

	log.Infof("Creating session for user: %s(%s)", user.ID, user.PrincipalName)

	userSession := model.UserSession{
		User:      user,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(s.cfg.AuthSessionTTL()),
	}

	switch typedAuthProvider := authProvider.(type) {
	case model.AuthSecret:
		userSession.AuthProviderType = model.SessionAuthProviderSecret
		userSession.AuthProviderID = typedAuthProvider.ID

	case model.SAMLProvider:
		userSession.AuthProviderType = model.SessionAuthProviderSAML
		userSession.AuthProviderID = typedAuthProvider.ID
	}

	if newSession, err := s.db.CreateUserSession(userSession); err != nil {
		return "", FormatDatabaseError(err)
	} else if signingKeyBytes, err := s.cfg.Crypto.JWT.SigningKeyBytes(); err != nil {
		return "", err
	} else {
		var (
			jwtClaims = &auth.SessionData{
				StandardClaims: jwt.StandardClaims{
					Id:        strconv.FormatInt(newSession.ID, 10),
					Subject:   user.ID.String(),
					IssuedAt:  newSession.CreatedAt.UTC().Unix(),
					ExpiresAt: newSession.ExpiresAt.UTC().Unix(),
				},
			}

			token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
		)

		return token.SignedString(signingKeyBytes)
	}
}

func (s authenticator) jwtSigningKey(token *jwt.Token) (any, error) {
	return s.cfg.Crypto.JWT.SigningKeyBytes()
}

func (s authenticator) ValidateSession(jwtTokenString string) (auth.Context, error) {
	claims := auth.SessionData{}

	if token, err := jwt.ParseWithClaims(jwtTokenString, &claims, s.jwtSigningKey); err != nil {
		if err == jwt.ErrSignatureInvalid {
			return auth.Context{}, ErrInvalidAuth
		}

		return auth.Context{}, err
	} else if !token.Valid {
		log.Infof("Token invalid")
		return auth.Context{}, ErrInvalidAuth
	} else if sessionID, err := claims.SessionID(); err != nil {
		log.Infof("Session ID %s invalid: %v", claims.Id, err)
		return auth.Context{}, ErrInvalidAuth
	} else if session, err := s.db.GetUserSession(sessionID); err != nil {
		log.Infof("Unable to find session %d", sessionID)
		return auth.Context{}, ErrInvalidAuth
	} else if session.Expired() {
		log.Infof("Session %s is expired", sessionID)
		return auth.Context{}, ErrInvalidAuth
	} else {
		authContext := auth.Context{
			Owner:   session.User,
			Session: session,
		}

		if session.AuthProviderType == model.SessionAuthProviderSecret && session.User.AuthSecret.Expired() {
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
