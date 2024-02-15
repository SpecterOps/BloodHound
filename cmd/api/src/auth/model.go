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

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
)

const (
	ProviderTypeSecret = "secret"

	HMAC_SHA2_256 = "hmac-sha2-256"
)

type SessionData struct {
	jwt.StandardClaims
}

func (s SessionData) SessionID() (int64, error) {
	return strconv.ParseInt(s.Id, 10, 64)
}

func (s SessionData) UserID() (uuid.UUID, error) {
	return uuid.FromString(s.Subject)
}

type PermissionOverrides struct {
	Enabled     bool
	Permissions model.Permissions
}

type SimpleIdentity struct {
	ID    uuid.UUID
	Name  string
	Email string
	Key   string
}

type IdentityResolver interface {
	GetIdentity(ctx Context) (SimpleIdentity, error)
}

type idResolver struct{}

func NewIdentityResolver() IdentityResolver {
	return idResolver{}
}

func (s idResolver) GetIdentity(ctx Context) (SimpleIdentity, error) {
	if user, ok := GetUserFromAuthCtx(ctx); !ok {
		return SimpleIdentity{}, errors.New("error retrieving user from auth context")
	} else {
		return SimpleIdentity{
			ID:    user.ID,
			Name:  user.PrincipalName,
			Email: user.EmailAddress.String,
			Key:   "user_id",
		}, nil
	}
}

type AuditLogger interface {
	AppendAuditLog(ctx context.Context, entry model.AuditEntry) error
}

type Authorizer interface {
	HasPermission(ctx Context, requiredPermission model.Permission, grantedPermissions model.Permissions) bool
	AllowsPermission(ctx Context, requiredPermission model.Permission) bool
	AllowsAllPermissions(ctx Context, requiredPermissions model.Permissions) bool
	AllowsAtLeastOnePermission(ctx Context, requiredPermissions model.Permissions) bool
	AuditLogUnauthorizedAccess(request *http.Request)
}

type authorizer struct {
	auditLogger AuditLogger
}

func NewAuthorizer(auditLogger AuditLogger) Authorizer {
	return authorizer{auditLogger: auditLogger}
}

func (s authorizer) HasPermission(ctx Context, requiredPermission model.Permission, grantedPermissions model.Permissions) bool {
	if ctx.PermissionOverrides.Enabled {
		return ctx.PermissionOverrides.Permissions.Has(requiredPermission)
	}

	return grantedPermissions.Has(requiredPermission)
}

func (s authorizer) AllowsPermission(ctx Context, requiredPermission model.Permission) bool {
	if user, isUser := GetUserFromAuthCtx(ctx); isUser {
		return s.HasPermission(ctx, requiredPermission, user.Roles.Permissions())
	}

	return false
}

func (s authorizer) AllowsAllPermissions(ctx Context, requiredPermissions model.Permissions) bool {
	if user, isUser := GetUserFromAuthCtx(ctx); isUser {
		grantedPermissions := user.Roles.Permissions()
		for _, permission := range requiredPermissions {
			if !s.HasPermission(ctx, permission, grantedPermissions) {
				return false
			}
		}
	}

	return true
}

func (s authorizer) AllowsAtLeastOnePermission(ctx Context, requiredPermissions model.Permissions) bool {
	if user, isUser := GetUserFromAuthCtx(ctx); isUser {
		grantedPermissions := user.Roles.Permissions()
		for _, permission := range requiredPermissions {
			if s.HasPermission(ctx, permission, grantedPermissions) {
				return true
			}
		}
	}

	return false
}

func (s authorizer) AuditLogUnauthorizedAccess(request *http.Request) {
	// Ignore read logs as they are less likely to occur from malicious access
	if request.Method != "GET" {
		if err := s.auditLogger.AppendAuditLog(
			request.Context(),
			model.AuditEntry{
				Action: "UnauthorizedAccessAttempt",
				Model:  model.AuditData{"endpoint": request.Method + " " + request.URL.Path},
				Status: model.AuditStatusFailure,
			},
		); err != nil {
			log.Errorf("error creating audit log for unauthorized access: %s", err.Error())
		}
	}
}

type Context struct {
	PermissionOverrides PermissionOverrides
	Owner               any
	Session             model.UserSession
}

func (s Context) Authenticated() bool {
	return s.Owner != nil
}

func GetUserFromAuthCtx(ctx Context) (model.User, bool) {
	switch typed := ctx.Owner.(type) {
	case model.User:
		return typed, true
	default:
		return model.User{}, false
	}
}

// NewUserAuthToken creates a new User model.AuthToken using the details provided
//
// This isn't an ideal location for this function but it was determined to be the best place "for now".
// See https://specterops.atlassian.net/browse/BED-3367
func NewUserAuthToken(ownerId string, tokenName string, hmacMethod string) (model.AuthToken, error) {
	var (
		tokenBytes = make([]byte, 40)
	)

	ownerUuid, err := uuid.FromString(ownerId)
	if err != nil {
		return model.AuthToken{}, err
	}

	authToken := model.AuthToken{
		UserID:     uuid.NullUUID{UUID: ownerUuid, Valid: true},
		HmacMethod: hmacMethod,
		LastAccess: time.Now().UTC(),
		Name:       null.StringFrom(tokenName),
	}

	if hmacMethod != HMAC_SHA2_256 {
		return authToken, fmt.Errorf("HMAC method %s is not supported", hmacMethod)
	}

	if id, err := uuid.NewV4(); err != nil {
		return authToken, err
	} else {
		authToken.ID = id
	}

	if _, err := rand.Read(tokenBytes); err != nil {
		return authToken, nil
	}

	authToken.Key = base64.StdEncoding.EncodeToString(tokenBytes)
	return authToken, nil
}
