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
	"github.com/specterops/bloodhound/packages/go/params"
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
	permissions := make([]services.Permission, 0, len(permissionRows))
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

	permissionRows, err := s.getRolePermissions(ctx, roleRow.ID)
	if err != nil {
		return services.Role{}, err
	}

	return toRole(roleRow, permissionRows), nil
}

// getRolePermissions retrieves the permissions associated with a role via the
// join table. It is shared by GetRole and ListRoles so both issue the same
// per-role permissions query.
func (s *Store) getRolePermissions(ctx context.Context, roleID int32) ([]permission, error) {
	var (
		sb   = sqlbuilder.PostgreSQL.NewSelectBuilder()
		rows pgx.Rows
		err  error
	)

	sb.Select("p.id", "p.authority", "p.name", "p.created_at", "p.updated_at")
	sb.From(tablePermissions + " p")
	sb.Join(tableRolesPermissions+" rp", "rp.permission_id = p.id")
	sb.Where(sb.Equal("rp.role_id", roleID))

	query, args := sb.Build()

	rows, err = s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying permissions for role: %s", err)
	}
	permissionRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[permission])
	if err != nil {
		return nil, fmt.Errorf("collecting permissions for role: %s", err)
	}

	return permissionRows, nil
}

// roleColumns maps the API-facing role field names to their underlying database
// columns. It is the single source of truth for which role fields may be sorted
// or filtered on at the persistence layer, guarding against SQL injection by
// only ever emitting known column names into the query.
var roleColumns = map[string]string{
	"id":          "id",
	"name":        "name",
	"description": "description",
	"created_at":  "created_at",
	"updated_at":  "updated_at",
	"deleted_at":  "deleted_at",
}

// applyRoleFilters translates the validated query filters into WHERE expressions
// on the supplied builder. Filters targeting the same field are combined using
// that field's SetOperator; distinct fields are combined with AND.
func applyRoleFilters(sb *sqlbuilder.SelectBuilder, queryFilters params.Filters) error {
	for field, fieldFilters := range queryFilters {
		column, isKnown := roleColumns[field]
		if !isKnown {
			return fmt.Errorf("role filter references unknown field %q", field)
		}

		expressions := make([]string, 0, len(fieldFilters))
		setOperator := params.FilterAnd
		for _, filter := range fieldFilters {
			setOperator = filter.SetOperator
			switch filter.Operator {
			case params.Equals:
				expressions = append(expressions, sb.Equal(column, filter.Value))
			case params.NotEquals:
				expressions = append(expressions, sb.NotEqual(column, filter.Value))
			case params.GreaterThan:
				expressions = append(expressions, sb.GreaterThan(column, filter.Value))
			case params.GreaterThanOrEquals:
				expressions = append(expressions, sb.GreaterEqualThan(column, filter.Value))
			case params.LessThan:
				expressions = append(expressions, sb.LessThan(column, filter.Value))
			case params.LessThanOrEquals:
				expressions = append(expressions, sb.LessEqualThan(column, filter.Value))
			case params.ApproximatelyEquals:
				expressions = append(expressions, sb.ILike(column, "%"+filter.Value+"%"))
			default:
				return fmt.Errorf("role filter uses unsupported operator %q", filter.Operator)
			}
		}

		if setOperator == params.FilterOr {
			sb.Where(sb.Or(expressions...))
		} else {
			sb.Where(sb.And(expressions...))
		}
	}

	return nil
}

// buildRoleOrderBy translates the validated sort items into ORDER BY terms,
// mapping each field to its database column.
func buildRoleOrderBy(sortItems params.SortItems) ([]string, error) {
	orderBy := make([]string, 0, len(sortItems))
	for _, sortItem := range sortItems {
		column, isKnown := roleColumns[sortItem.Field]
		if !isKnown {
			return nil, fmt.Errorf("role sort references unknown field %q", sortItem.Field)
		}

		if sortItem.Direction == params.Descending {
			orderBy = append(orderBy, column+" DESC")
		} else {
			orderBy = append(orderBy, column+" ASC")
		}
	}

	return orderBy, nil
}

// ListRoles retrieves all roles matching the supplied filters, ordered by the
// supplied sort items, and preloads each role's permissions. It mirrors the
// legacy GetAllRoles behavior, which returns every matching role without
// pagination.
func (s *Store) ListRoles(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]services.Role, error) {
	var (
		roleSB      = sqlbuilder.PostgreSQL.NewSelectBuilder()
		roleRows    pgx.Rows
		listedRoles []role
		result      []services.Role
		orderBy     []string
		err         error
	)

	roleSB.Select("*").From(tableRoles)

	if err = applyRoleFilters(roleSB, queryFilters); err != nil {
		return nil, err
	}

	orderBy, err = buildRoleOrderBy(sortItems)
	if err != nil {
		return nil, err
	}
	if len(orderBy) > 0 {
		roleSB.OrderBy(orderBy...)
	}

	roleQuery, roleArgs := roleSB.Build()

	roleRows, err = s.db.Query(ctx, roleQuery, roleArgs...)
	if err != nil {
		return nil, err
	}
	listedRoles, err = pgx.CollectRows(roleRows, pgx.RowToStructByName[role])
	if err != nil {
		return nil, fmt.Errorf("collecting roles: %s", err)
	}

	result = make([]services.Role, 0, len(listedRoles))
	for _, roleRow := range listedRoles {
		permissionRows, err := s.getRolePermissions(ctx, roleRow.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, toRole(roleRow, permissionRows))
	}

	return result, nil
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
