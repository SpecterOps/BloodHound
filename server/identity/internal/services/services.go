// Copyright 2026 Specter Ops, Inc.
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

package services

//go:generate go tool mockery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/params"
)

type Permission struct {
	Authority string
	Name      string
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

// ErrNoPermissionFound indicates that no permission with the given ID was found.
var ErrNoPermissionFound = errors.New("no permission was found")

type Role struct {
	ID          int32
	Name        string
	Description string
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

// ErrNoRoleFound indicates that no role with the given ID was found.
var ErrNoRoleFound = errors.New("no role was found")

// AuthSecret is the credential material associated with a user.
type AuthSecret struct {
	ID            int32
	DigestMethod  string
	ExpiresAt     time.Time
	TOTPActivated bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
}

// EnvironmentAccessControl is a single environment-targeted access control entry
// scoping a user to a specific environment.
type EnvironmentAccessControl struct {
	ID            int64
	UserID        string
	EnvironmentID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
}

// User is the identity domain representation of a BloodHound user.
type User struct {
	ID                               uuid.UUID
	SSOProviderID                    sql.NullInt32
	FirstName                        sql.NullString
	LastName                         sql.NullString
	EmailAddress                     sql.NullString
	PrincipalName                    string
	LastLogin                        time.Time
	IsDisabled                       bool
	AllEnvironments                  bool
	EULAAccepted                     bool
	Roles                            []Role
	EnvironmentTargetedAccessControl []EnvironmentAccessControl
	AuthSecret                       *AuthSecret
	CreatedAt                        time.Time
	UpdatedAt                        time.Time
	DeletedAt                        sql.NullTime
}

type AuthToken struct {
	UserID     uuid.NullUUID
	Name       sql.NullString
	HmacMethod string
	LastAccess time.Time
	ID         uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  sql.NullTime
	CreatedBy  uuid.NullUUID
	DeletedAt  sql.NullTime
}

type AuthTokenRequester struct {
	UserID         uuid.UUID
	IsUser         bool
	CanManageUsers bool
}

const authTokenUserIDField = "user_id"

var (
	ErrInvalidAuthTokenUserID            = errors.New("invalid UUID")
	ErrAuthTokenUserIDRequiresAdmin      = errors.New("filtering tokens by another user's user_id requires admin privileges")
	ErrAuthTokenUserIDOperatorNotAllowed = errors.New("non-admin users may only apply the eq operator when filtering by user_id")
)

type Database interface {
	GetRole(ctx context.Context, id int32) (Role, error)
	GetPermission(ctx context.Context, id int) (Permission, error)
	ListRoles(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Role, error)
	ListUsers(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]User, error)
	ListPermissions(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Permission, error)
	ListAuthTokens(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]AuthToken, error)
}

type Service struct {
	db Database
}

func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}

func (s *Service) GetRole(ctx context.Context, id int32) (Role, error) {
	return s.db.GetRole(ctx, id)
}

func (s *Service) GetPermission(ctx context.Context, id int) (Permission, error) {
	return s.db.GetPermission(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Role, error) {
	return s.db.ListRoles(ctx, queryFilters, sortItems)
}

func (s *Service) ListUsers(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]User, error) {
	return s.db.ListUsers(ctx, queryFilters, sortItems)
}

func (s *Service) ListPermissions(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Permission, error) {
	return s.db.ListPermissions(ctx, queryFilters, sortItems)
}

func (s *Service) ListAuthTokens(ctx context.Context, requester AuthTokenRequester, queryFilters params.Filters, sortItems params.SortItems) ([]AuthToken, error) {
	scopedFilters, err := scopeAuthTokenFilters(requester, queryFilters)
	if err != nil {
		return nil, err
	}

	return s.db.ListAuthTokens(ctx, scopedFilters, sortItems)
}

func scopeAuthTokenFilters(requester AuthTokenRequester, queryFilters params.Filters) (params.Filters, error) {
	var (
		scopedFilters   = maps.Clone(queryFilters)
		userIDFilters   = queryFilters[authTokenUserIDField]
		validatedUserID uuid.UUID
		err             error
	)

	if scopedFilters == nil {
		scopedFilters = params.Filters{}
	}

	if len(userIDFilters) > 0 {
		if validatedUserID, err = uuid.FromString(userIDFilters[0].Value); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidAuthTokenUserID, err)
		}
	}

	if !requester.IsUser || requester.CanManageUsers {
		return scopedFilters, nil
	}

	if len(userIDFilters) > 0 {
		if validatedUserID != requester.UserID {
			return nil, ErrAuthTokenUserIDRequiresAdmin
		} else if userIDFilters[0].Operator != params.Equals {
			return nil, ErrAuthTokenUserIDOperatorNotAllowed
		}
	}

	scopedFilters[authTokenUserIDField] = []params.Filter{{
		Field:       authTokenUserIDField,
		Operator:    params.Equals,
		Value:       requester.UserID.String(),
		SetOperator: params.FilterAnd,
	}}

	return scopedFilters, nil
}
