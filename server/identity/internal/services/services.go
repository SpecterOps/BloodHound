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

type Database interface {
	GetRole(ctx context.Context, id int32) (Role, error)
	GetPermission(ctx context.Context, id int) (Permission, error)
	ListRoles(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Role, error)
	ListUsers(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]User, error)
	ListPermissions(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]Permission, error)
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
