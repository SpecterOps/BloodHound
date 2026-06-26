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
)

type Permission struct {
	Authority string       `json:"authority"`
	Name      string       `json:"name"`
	ID        int32        `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

// ErrNoPermissionFound indicates that no permission with the given ID was found.
var ErrNoPermissionFound = errors.New("no permission was found")

type Role struct {
	ID          int32        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   sql.NullTime `json:"deleted_at"`
}

// ErrNoRoleFound indicates that no role with the given ID was found.
var ErrNoRoleFound = errors.New("no role was found")

type User struct {
	ID            uuid.UUID    `json:"id"`
	PrincipalName string       `json:"principal_name"`
	IsDisabled    bool         `json:"is_disabled"`
	EULAAccepted  bool         `json:"eula_accepted"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
}

// ErrNoUserFound indicates that no user with the given ID was found.
var ErrNoUserFound = errors.New("no user was found")

type Database interface {
	GetRole(ctx context.Context, id int32) (Role, error)
	GetUser(ctx context.Context, id uuid.UUID) (User, error)
	GetPermission(ctx context.Context, id int) (Permission, error)
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

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	return s.db.GetUser(ctx, id)
}

func (s *Service) GetPermission(ctx context.Context, id int) (Permission, error) {
	return s.db.GetPermission(ctx, id)
}
