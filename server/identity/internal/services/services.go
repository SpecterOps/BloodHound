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
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/appcfg"
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

// CreateAuthTokenRequest is the input for creating an auth token. An empty UserID
// targets the requester.
type CreateAuthTokenRequest struct {
	Name   string
	UserID string
}

// CreatedAuthToken is an auth token together with its secret key. The key is only
// ever returned to the client at creation time.
type CreatedAuthToken struct {
	AuthToken
	Key string
}

const (
	authTokenUserIDField = "user_id"
	authTokenHMACMethod  = "hmac-sha2-256"
	authTokenKeyLength   = 40
)

var (
	ErrInvalidAuthTokenUserID            = errors.New("invalid UUID")
	ErrAuthTokenUserIDRequiresAdmin      = errors.New("filtering tokens by another user's user_id requires admin privileges")
	ErrAuthTokenUserIDOperatorNotAllowed = errors.New("non-admin users may only apply the eq operator when filtering by user_id")
	ErrAPITokensDisabled                 = errors.New("API key creation is disabled")
	ErrAuthTokenTargetRequiresAdmin      = errors.New("missing permission to create tokens for other users")
	ErrAuthTokenUserNotFound             = errors.New("user_id does not match an existing user")
	ErrAuthTokenKeyGeneration            = errors.New("failed to generate auth token key")
)

// UserMinimal is the minimal, non-sensitive representation of an active BloodHound user.
type UserMinimal struct {
	ID           uuid.UUID
	EmailAddress sql.NullString
	FirstName    sql.NullString
	LastName     sql.NullString
}

// AuthTokenReader lists tokens using the supplied requester-scoped filters.
type AuthTokenReader interface {
	ListAuthTokens(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]AuthToken, error)
}

// TokenConfigReader reads the application settings that govern auth token creation.
type TokenConfigReader interface {
	GetConfigAPITokens(ctx context.Context) (appcfg.APITokensParameter, error)
	GetConfigAPITokenExpiration(ctx context.Context) (appcfg.APITokenExpirationParameter, error)
}

type Database interface {
	GetRole(ctx context.Context, id int32) (Role, error)
	GetPermission(ctx context.Context, id int) (Permission, error)
	ListRoles(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Role, error)
	ListUsers(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]User, error)
	ListActiveUsersMinimal(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]UserMinimal, error)
	ListPermissions(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Permission, error)
	ListAuthTokens(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]AuthToken, error)
	CreateAuthToken(ctx context.Context, authToken CreatedAuthToken) (CreatedAuthToken, error)
}

type Service struct {
	db                Database
	authTokenReader   AuthTokenReader
	tokenConfigReader TokenConfigReader
	randomSource      io.Reader
	now               func() time.Time
}

type Option func(*Service)

func WithAuthTokenReader(reader AuthTokenReader) Option {
	if reader == nil {
		panic("identity: auth token reader must not be nil")
	}
	return func(service *Service) {
		service.authTokenReader = reader
	}
}

// WithRandomSource replaces crypto/rand as the source of auth token key bytes.
func WithRandomSource(randomSource io.Reader) Option {
	return func(service *Service) {
		service.randomSource = randomSource
	}
}

// WithClock replaces time.Now as the source of auth token timestamps.
func WithClock(now func() time.Time) Option {
	return func(service *Service) {
		service.now = now
	}
}

func NewService(databaseInterface Database, tokenConfigReader TokenConfigReader, options ...Option) *Service {
	if tokenConfigReader == nil {
		panic("identity: token config reader must not be nil")
	}

	var service = &Service{
		db:                databaseInterface,
		authTokenReader:   databaseInterface,
		tokenConfigReader: tokenConfigReader,
		randomSource:      rand.Reader,
		now:               time.Now,
	}
	for _, option := range options {
		option(service)
	}
	return service
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

// ListActiveUsersMinimal returns active, non-support users. When no sort is requested,
// results are sorted by email_address ascending.
func (s *Service) ListActiveUsersMinimal(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]UserMinimal, error) {
	var (
		defaultSortField = "email_address"
		orderedSort      = slices.Clone(sortItems)
	)

	if !slices.ContainsFunc(orderedSort, func(sortItem params.SortItem) bool { return sortItem.Field == defaultSortField }) {
		orderedSort = append(orderedSort, params.SortItem{Field: defaultSortField, Direction: params.Ascending})
	}

	return s.db.ListActiveUsersMinimal(ctx, queryFilters, orderedSort)
}

func (s *Service) ListPermissions(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Permission, error) {
	return s.db.ListPermissions(ctx, queryFilters, sortItems)
}

func (s *Service) ListAuthTokens(ctx context.Context, requester AuthTokenRequester, queryFilters params.Filters, sortItems params.SortItems) ([]AuthToken, error) {
	scopedFilters, err := scopeAuthTokenFilters(requester, queryFilters)
	if err != nil {
		return nil, err
	}

	return s.authTokenReader.ListAuthTokens(ctx, scopedFilters, sortItems)
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

// CreateAuthToken creates an auth token for the requester, or for another user when
// the requester can manage users.
func (s *Service) CreateAuthToken(ctx context.Context, requester AuthTokenRequester, request CreateAuthTokenRequest) (CreatedAuthToken, error) {
	var (
		targetUserID = request.UserID
		now          = s.now().UTC()
		keyBytes     = make([]byte, authTokenKeyLength)
	)

	// On a read failure the adapter logs and returns the parameter default, which
	// is the value legacy behavior falls back to.
	apiTokensParameter, _ := s.tokenConfigReader.GetConfigAPITokens(ctx)
	if !apiTokensParameter.Enabled {
		return CreatedAuthToken{}, ErrAPITokensDisabled
	}

	if targetUserID == "" {
		targetUserID = requester.UserID.String()
	}

	ownerID, err := uuid.FromString(targetUserID)
	if err != nil {
		return CreatedAuthToken{}, fmt.Errorf("%w: %w", ErrInvalidAuthTokenUserID, err)
	} else if ownerID != requester.UserID && !requester.CanManageUsers {
		return CreatedAuthToken{}, ErrAuthTokenTargetRequiresAdmin
	}

	tokenID, err := uuid.NewV4()
	if err != nil {
		return CreatedAuthToken{}, fmt.Errorf("generating auth token id: %w", err)
	}

	if _, err := io.ReadFull(s.randomSource, keyBytes); err != nil {
		return CreatedAuthToken{}, fmt.Errorf("%w: %w", ErrAuthTokenKeyGeneration, err)
	}

	authToken := CreatedAuthToken{
		AuthToken: AuthToken{
			ID:         tokenID,
			UserID:     uuid.NullUUID{UUID: ownerID, Valid: true},
			Name:       sql.NullString{String: request.Name, Valid: true},
			HmacMethod: authTokenHMACMethod,
			LastAccess: now,
			CreatedBy:  uuid.NullUUID{UUID: requester.UserID, Valid: true},
		},
		Key: base64.StdEncoding.EncodeToString(keyBytes),
	}

	apiTokenExpirationParameter, _ := s.tokenConfigReader.GetConfigAPITokenExpiration(ctx)
	if apiTokenExpirationParameter.Enabled {
		authToken.ExpiresAt = sql.NullTime{Time: now.AddDate(0, 0, apiTokenExpirationParameter.ExpirationPeriod), Valid: true}
	}

	return s.db.CreateAuthToken(ctx, authToken)
}
