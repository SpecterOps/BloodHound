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

package appdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

const (
	tablePermissions      = "permissions"
	tableRoles            = "roles"
	tableRolesPermissions = "roles_permissions"
)

type queryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type pgxQuerier interface {
	queryExecer
}

type permission struct {
	Authority string    `db:"authority"`
	Name      string    `db:"name"`
	ID        int32     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type role struct {
	ID          int32     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Store struct {
	db pgxQuerier
}

func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

func toPermission(row permission) services.Permission {
	return services.Permission{
		Authority: row.Authority,
		Name:      row.Name,
		ID:        row.ID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toRole(row role, permissionRows []permission) services.Role {
	var permissions []services.Permission
	for _, permissionRow := range permissionRows {
		permissions = append(permissions, toPermission(permissionRow))
	}
	return services.Role{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Permissions: permissions,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (s *Store) GetRole(ctx context.Context, id int32) (services.Role, error) {
	var (
		roleSB   = sqlbuilder.PostgreSQL.NewSelectBuilder()
		roleRows pgx.Rows
		roleRow  role
		err      error
	)

	roleSB.Select("*").From(tableRoles).Where(roleSB.Equal("id", id)).Limit(1)
	roleQuery, roleArgs := roleSB.Build()

	roleRows, err = s.db.Query(ctx, roleQuery, roleArgs...)
	if err != nil {
		return services.Role{}, err
	}
	roleRow, err = pgx.CollectOneRow(roleRows, pgx.RowToStructByName[role])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Role{}, services.ErrNoRoleFound
	}
	if err != nil {
		return services.Role{}, fmt.Errorf("finding role: %s", err)
	}

	// Fetch permissions associated with this role via the join table.
	permQuery := fmt.Sprintf(
		"SELECT p.id, p.authority, p.name, p.created_at, p.updated_at FROM %s p JOIN %s rp ON rp.permission_id = p.id WHERE rp.role_id = $1",
		tablePermissions,
		tableRolesPermissions,
	)
	permRows, permErr := s.db.Query(ctx, permQuery, id)
	if permErr != nil {
		return services.Role{}, fmt.Errorf("querying permissions for role: %s", permErr)
	}
	permissionRows, permErr := pgx.CollectRows(permRows, pgx.RowToStructByName[permission])
	if permErr != nil {
		return services.Role{}, fmt.Errorf("collecting permissions for role: %s", permErr)
	}

	return toRole(roleRow, permissionRows), nil
}

func (s *Store) GetPermission(ctx context.Context, id int) (services.Permission, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("*")
	sb.From(tablePermissions)
	sb.Where(sb.Equal("id", id))
	sb.Limit(1)

	query, args := sb.Build()

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return services.Permission{}, err
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[permission])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Permission{}, services.ErrNoPermissionFound
	}
	if err != nil {
		return services.Permission{}, fmt.Errorf("finding permission: %s", err)
	}

	return toPermission(row), nil
}
