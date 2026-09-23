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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

const (
	tablePermissions      = "permissions"
	tableRoles            = "roles"
	tableRolesPermissions = "roles_permissions"
	tableUsers            = "users"
	tableUsersRoles       = "users_roles"
	tableETAC             = "environment_targeted_access_control"
	tableAuthSecrets      = "auth_secrets"
)

var (
	validPermissionColumns = []string{"id", "authority", "name", "created_at", "updated_at"}
	validRoleColumns       = []string{"id", "name", "description", "created_at", "updated_at"}
)

var validUserColumns = []string{
	"id",
	"sso_provider_id",
	"first_name",
	"last_name",
	"email_address",
	"principal_name",
	"last_login",
	"is_disabled",
	"all_environments",
	"eula_accepted",
	"created_at",
	"updated_at",
}

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

	roleSB.Select(validRoleColumns...).From(tableRoles).Where(roleSB.Equal("id", id)).Limit(1)
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
		return nil, fmt.Errorf("querying permissions for role: %w", err)
	}
	permissionRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[permission])
	if err != nil {
		return nil, fmt.Errorf("collecting permissions for role: %s", err)
	}

	return permissionRows, nil
}

// rolePermission carries the owning role id alongside a permission so that a
// single batched query can be grouped back to individual roles in memory.
type rolePermission struct {
	permission
	RoleID int32 `db:"role_id"`
}

// getPermissionsForRoles retrieves the permissions for every supplied role id in
// a single query and groups them by role id, avoiding the N+1 round trips that a
// per-role query would incur when listing many roles. Roles with no permissions
// are simply absent from the returned map.
func (s *Store) getPermissionsForRoles(ctx context.Context, roleIDs []int32) (map[int32][]permission, error) {
	var (
		sb            = sqlbuilder.PostgreSQL.NewSelectBuilder()
		rows          pgx.Rows
		permissionsBy = make(map[int32][]permission, len(roleIDs))
		err           error
	)

	sb.Select("rp.role_id", "p.id", "p.authority", "p.name", "p.created_at", "p.updated_at")
	sb.From(tablePermissions + " p")
	sb.Join(tableRolesPermissions+" rp", "rp.permission_id = p.id")
	sb.Where(sb.In("rp.role_id", sqlbuilder.List(roleIDs)))

	query, args := sb.Build()

	rows, err = s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying permissions for roles: %w", err)
	}
	joinedRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[rolePermission])
	if err != nil {
		return nil, fmt.Errorf("collecting permissions for roles: %s", err)
	}

	for _, joinedRow := range joinedRows {
		permissionsBy[joinedRow.RoleID] = append(permissionsBy[joinedRow.RoleID], joinedRow.permission)
	}

	return permissionsBy, nil
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
}

// nullFilterValue is the sentinel filter value the API translates into a SQL
// NULL comparison. It mirrors the legacy filter builder, which renders eq:null /
// neq:null as IS NULL / IS NOT NULL rather than binding the literal string
// "null" through =/<>.
const nullFilterValue = "null"

// buildFilterComparison translates a single validated filter into a SQL WHERE
// expression on the supplied builder. Equality filters whose value is the null
// sentinel become IS NULL / IS NOT NULL, matching the legacy contract; every
// other operator binds the value as a parameter. The boolean is false when the
// operator is unsupported, letting callers surface a field-specific error.
func buildFilterComparison(sb *sqlbuilder.SelectBuilder, column string, filter params.Filter) (string, bool) {
	switch filter.Operator {
	case params.Equals:
		if filter.Value == nullFilterValue {
			return sb.IsNull(column), true
		}
		return sb.Equal(column, filter.Value), true
	case params.NotEquals:
		if filter.Value == nullFilterValue {
			return sb.IsNotNull(column), true
		}
		return sb.NotEqual(column, filter.Value), true
	case params.GreaterThan:
		return sb.GreaterThan(column, filter.Value), true
	case params.GreaterThanOrEquals:
		return sb.GreaterEqualThan(column, filter.Value), true
	case params.LessThan:
		return sb.LessThan(column, filter.Value), true
	case params.LessThanOrEquals:
		return sb.LessEqualThan(column, filter.Value), true
	default:
		return "", false
	}
}

// permissionColumns maps the API-facing permission fields to their database columns.
var permissionColumns = map[string]string{
	"authority":  "authority",
	"name":       "name",
	"id":         "id",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// applyPermissionFilters translates validated permission filters into WHERE expressions.
func applyPermissionFilters(sb *sqlbuilder.SelectBuilder, queryFilters params.Filters) error {
	for field, fieldFilters := range queryFilters {
		column, isKnown := permissionColumns[field]
		if !isKnown {
			return fmt.Errorf("permission filter references unknown field %q", field)
		}

		setOperator := params.FilterAnd
		if len(fieldFilters) > 0 {
			setOperator = fieldFilters[0].SetOperator
		}

		expressions := make([]string, 0, len(fieldFilters))
		for _, filter := range fieldFilters {
			if filter.Value == nullFilterValue {
				switch filter.Operator {
				case params.Equals:
					expressions = append(expressions, sb.IsNull(column))
				case params.NotEquals:
					expressions = append(expressions, sb.IsNotNull(column))
				default:
					return fmt.Errorf("permission filter uses unsupported null operator %q", filter.Operator)
				}

				continue
			}

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
			default:
				return fmt.Errorf("permission filter uses unsupported operator %q", filter.Operator)
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

// buildPermissionOrderBy translates validated sort items into ORDER BY terms.
func buildPermissionOrderBy(sortItems params.SortItems) ([]string, error) {
	orderBy := make([]string, 0, len(sortItems))
	for _, sortItem := range sortItems {
		column, isKnown := permissionColumns[sortItem.Field]
		if !isKnown {
			return nil, fmt.Errorf("permission sort references unknown field %q", sortItem.Field)
		}

		if sortItem.Direction == params.Descending {
			orderBy = append(orderBy, column+" DESC")
		} else {
			orderBy = append(orderBy, column+" ASC")
		}
	}

	return orderBy, nil
}

// ListPermissions retrieves every permission matching the supplied filters and sort order.
func (s *Store) ListPermissions(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]services.Permission, error) {
	var (
		permissionSB      = sqlbuilder.PostgreSQL.NewSelectBuilder()
		permissionRows    pgx.Rows
		listedPermissions []permission
		result            []services.Permission
		orderBy           []string
		err               error
	)

	permissionSB.Select(validPermissionColumns...).From(tablePermissions)

	if err = applyPermissionFilters(permissionSB, queryFilters); err != nil {
		return nil, err
	}

	orderBy, err = buildPermissionOrderBy(sortItems)
	if err != nil {
		return nil, err
	}
	if len(orderBy) > 0 {
		permissionSB.OrderBy(orderBy...)
	}

	permissionQuery, permissionArgs := permissionSB.Build()
	permissionRows, err = s.db.Query(ctx, permissionQuery, permissionArgs...)
	if err != nil {
		return nil, err
	}
	listedPermissions, err = pgx.CollectRows(permissionRows, pgx.RowToStructByName[permission])
	if err != nil {
		return nil, fmt.Errorf("collecting permissions: %s", err)
	}

	result = make([]services.Permission, 0, len(listedPermissions))
	for _, listedPermission := range listedPermissions {
		result = append(result, toPermission(listedPermission))
	}

	return result, nil
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

		// A field's filters share a single set operator, so it is derived once
		// from the first filter rather than reassigned per iteration; otherwise
		// only the last filter would decide whether the group uses AND or OR.
		setOperator := params.FilterAnd
		if len(fieldFilters) > 0 {
			setOperator = fieldFilters[0].SetOperator
		}

		expressions := make([]string, 0, len(fieldFilters))
		for _, filter := range fieldFilters {
			expression, ok := buildFilterComparison(sb, column, filter)
			if !ok {
				return fmt.Errorf("role filter uses unsupported operator %q", filter.Operator)
			}
			expressions = append(expressions, expression)
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
		roleSB            = sqlbuilder.PostgreSQL.NewSelectBuilder()
		roleRows          pgx.Rows
		listedRoles       []role
		result            []services.Role
		orderBy           []string
		roleIDs           []int32
		permissionsByRole map[int32][]permission
		err               error
	)

	roleSB.Select(validRoleColumns...).From(tableRoles)

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

	if len(listedRoles) == 0 {
		return []services.Role{}, nil
	}

	roleIDs = make([]int32, 0, len(listedRoles))
	for _, roleRow := range listedRoles {
		roleIDs = append(roleIDs, roleRow.ID)
	}

	permissionsByRole, err = s.getPermissionsForRoles(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	result = make([]services.Role, 0, len(listedRoles))
	for _, roleRow := range listedRoles {
		result = append(result, toRole(roleRow, permissionsByRole[roleRow.ID]))
	}

	return result, nil
}

func (s *Store) GetPermission(ctx context.Context, id int) (services.Permission, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("id", "authority", "name", "created_at", "updated_at")
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

// userRow is the flat projection of the users table selected by ListUsers. The
// nullable scalar columns use stdlib nullable types so pgx can scan NULLs that
// gorm would otherwise coerce to zero values.
type userRow struct {
	ID              uuid.UUID      `db:"id"`
	SSOProviderID   sql.NullInt32  `db:"sso_provider_id"`
	FirstName       sql.NullString `db:"first_name"`
	LastName        sql.NullString `db:"last_name"`
	EmailAddress    sql.NullString `db:"email_address"`
	PrincipalName   string         `db:"principal_name"`
	LastLogin       time.Time      `db:"last_login"`
	IsDisabled      bool           `db:"is_disabled"`
	AllEnvironments bool           `db:"all_environments"`
	EULAAccepted    bool           `db:"eula_accepted"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

// userRole carries the owning user id alongside a role so that a single batched
// query joining users_roles can be grouped back to individual users in memory.
type userRole struct {
	role
	UserID uuid.UUID `db:"user_id"`
}

// etacRow is the flat projection of an environment_targeted_access_control row.
// updated_at has no default and may be NULL for rows not written through gorm.
type etacRow struct {
	ID            int64        `db:"id"`
	UserID        uuid.UUID    `db:"user_id"`
	EnvironmentID string       `db:"environment_id"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     sql.NullTime `db:"updated_at"`
}

// authSecretRow is the flat projection of an auth_secrets row. Secret material
// such as the digest and TOTP secret is intentionally left unselected.
type authSecretRow struct {
	ID            int32        `db:"id"`
	UserID        uuid.UUID    `db:"user_id"`
	DigestMethod  string       `db:"digest_method"`
	ExpiresAt     sql.NullTime `db:"expires_at"`
	TOTPActivated bool         `db:"totp_activated"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
}

func toEnvironmentAccessControl(row etacRow) services.EnvironmentAccessControl {
	return services.EnvironmentAccessControl{
		ID:            row.ID,
		UserID:        row.UserID.String(),
		EnvironmentID: row.EnvironmentID,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func toAuthSecret(row authSecretRow) services.AuthSecret {
	return services.AuthSecret{
		ID:            row.ID,
		DigestMethod:  row.DigestMethod,
		ExpiresAt:     row.ExpiresAt.Time,
		TOTPActivated: row.TOTPActivated,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func toUser(row userRow, roles []services.Role, environmentAccessControl []services.EnvironmentAccessControl, authSecret *services.AuthSecret) services.User {
	return services.User{
		ID:                               row.ID,
		SSOProviderID:                    row.SSOProviderID,
		FirstName:                        row.FirstName,
		LastName:                         row.LastName,
		EmailAddress:                     row.EmailAddress,
		PrincipalName:                    row.PrincipalName,
		LastLogin:                        row.LastLogin,
		IsDisabled:                       row.IsDisabled,
		AllEnvironments:                  row.AllEnvironments,
		EULAAccepted:                     row.EULAAccepted,
		Roles:                            roles,
		EnvironmentTargetedAccessControl: environmentAccessControl,
		AuthSecret:                       authSecret,
		CreatedAt:                        row.CreatedAt,
		UpdatedAt:                        row.UpdatedAt,
	}
}

// userColumns maps the API-facing user field names to their underlying database
// columns. It is the single source of truth for which user fields may be sorted
// or filtered on at the persistence layer, guarding against SQL injection by
// only ever emitting known column names into the query.
var userColumns = map[string]string{
	"id":             "id",
	"first_name":     "first_name",
	"last_name":      "last_name",
	"email_address":  "email_address",
	"principal_name": "principal_name",
	"last_login":     "last_login",
	"created_at":     "created_at",
	"updated_at":     "updated_at",
}

// applyUserFilters translates the validated query filters into WHERE expressions
// on the supplied builder. Filters targeting the same field are combined using
// that field's SetOperator; distinct fields are combined with AND.
func applyUserFilters(sb *sqlbuilder.SelectBuilder, queryFilters params.Filters) error {
	for field, fieldFilters := range queryFilters {
		column, isKnown := userColumns[field]
		if !isKnown {
			return fmt.Errorf("user filter references unknown field %q", field)
		}

		// A field's filters share a single set operator, so it is derived once
		// from the first filter rather than reassigned per iteration; otherwise
		// only the last filter would decide whether the group uses AND or OR.
		setOperator := params.FilterAnd
		if len(fieldFilters) > 0 {
			setOperator = fieldFilters[0].SetOperator
		}

		expressions := make([]string, 0, len(fieldFilters))
		for _, filter := range fieldFilters {
			expression, ok := buildFilterComparison(sb, column, filter)
			if !ok {
				return fmt.Errorf("user filter uses unsupported operator %q", filter.Operator)
			}
			expressions = append(expressions, expression)
		}

		if setOperator == params.FilterOr {
			sb.Where(sb.Or(expressions...))
		} else {
			sb.Where(sb.And(expressions...))
		}
	}

	return nil
}

// buildUserOrderBy translates the validated sort items into ORDER BY terms,
// mapping each field to its database column.
func buildUserOrderBy(sortItems params.SortItems) ([]string, error) {
	orderBy := make([]string, 0, len(sortItems))
	for _, sortItem := range sortItems {
		column, isKnown := userColumns[sortItem.Field]
		if !isKnown {
			return nil, fmt.Errorf("user sort references unknown field %q", sortItem.Field)
		}

		if sortItem.Direction == params.Descending {
			orderBy = append(orderBy, column+" DESC")
		} else {
			orderBy = append(orderBy, column+" ASC")
		}
	}

	return orderBy, nil
}

// getRolesForUsers retrieves the roles (with their permissions) for every
// supplied user id in a pair of batched queries and groups them by user id,
// avoiding the N+1 round trips that a per-user query would incur when listing
// many users. Users with no roles are simply absent from the returned map.
func (s *Store) getRolesForUsers(ctx context.Context, userIDs []string) (map[uuid.UUID][]services.Role, error) {
	var (
		sb                = sqlbuilder.PostgreSQL.NewSelectBuilder()
		rows              pgx.Rows
		userRoleRows      []userRole
		rolesByUser       = make(map[uuid.UUID][]services.Role, len(userIDs))
		roleIDsSeen       = make(map[int32]struct{})
		roleIDs           []int32
		permissionsByRole map[int32][]permission
		err               error
	)

	sb.Select("ur.user_id", "r.id", "r.name", "r.description", "r.created_at", "r.updated_at")
	sb.From(tableRoles + " r")
	sb.Join(tableUsersRoles+" ur", "ur.role_id = r.id")
	sb.Where(sb.In("ur.user_id", sqlbuilder.List(userIDs)))

	query, args := sb.Build()

	rows, err = s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying roles for users: %w", err)
	}
	userRoleRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[userRole])
	if err != nil {
		return nil, fmt.Errorf("collecting roles for users: %s", err)
	}

	for _, userRoleRow := range userRoleRows {
		if _, seen := roleIDsSeen[userRoleRow.ID]; !seen {
			roleIDsSeen[userRoleRow.ID] = struct{}{}
			roleIDs = append(roleIDs, userRoleRow.ID)
		}
	}

	if len(roleIDs) > 0 {
		if permissionsByRole, err = s.getPermissionsForRoles(ctx, roleIDs); err != nil {
			return nil, err
		}
	}

	for _, userRoleRow := range userRoleRows {
		rolesByUser[userRoleRow.UserID] = append(rolesByUser[userRoleRow.UserID], toRole(userRoleRow.role, permissionsByRole[userRoleRow.ID]))
	}

	return rolesByUser, nil
}

// getEnvironmentAccessControlForUsers retrieves the environment-targeted access
// control entries for every supplied user id in a single query and groups them
// by user id. Users with no entries are absent from the returned map.
func (s *Store) getEnvironmentAccessControlForUsers(ctx context.Context, userIDs []string) (map[uuid.UUID][]services.EnvironmentAccessControl, error) {
	var (
		sb         = sqlbuilder.PostgreSQL.NewSelectBuilder()
		rows       pgx.Rows
		etacRows   []etacRow
		etacByUser = make(map[uuid.UUID][]services.EnvironmentAccessControl, len(userIDs))
		err        error
	)

	sb.Select("id", "user_id", "environment_id", "created_at", "updated_at")
	sb.From(tableETAC)
	sb.Where(sb.In("user_id", sqlbuilder.List(userIDs)))

	query, args := sb.Build()

	rows, err = s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying environment access control for users: %w", err)
	}
	etacRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[etacRow])
	if err != nil {
		return nil, fmt.Errorf("collecting environment access control for users: %s", err)
	}

	for _, row := range etacRows {
		etacByUser[row.UserID] = append(etacByUser[row.UserID], toEnvironmentAccessControl(row))
	}

	return etacByUser, nil
}

// getAuthSecretsForUsers retrieves the auth secret for every supplied user id in
// a single query and groups them by user id. SSO users without a secret are
// absent from the returned map, mirroring the legacy nil AuthSecret.
func (s *Store) getAuthSecretsForUsers(ctx context.Context, userIDs []string) (map[uuid.UUID]*services.AuthSecret, error) {
	var (
		sb             = sqlbuilder.PostgreSQL.NewSelectBuilder()
		rows           pgx.Rows
		authSecretRows []authSecretRow
		secretByUser   = make(map[uuid.UUID]*services.AuthSecret, len(userIDs))
		err            error
	)

	sb.Select("id", "user_id", "digest_method", "expires_at", "totp_activated", "created_at", "updated_at")
	sb.From(tableAuthSecrets)
	sb.Where(sb.In("user_id", sqlbuilder.List(userIDs)))

	query, args := sb.Build()

	rows, err = s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying auth secrets for users: %w", err)
	}
	authSecretRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[authSecretRow])
	if err != nil {
		return nil, fmt.Errorf("collecting auth secrets for users: %s", err)
	}

	for _, row := range authSecretRows {
		secret := toAuthSecret(row)
		secretByUser[row.UserID] = &secret
	}

	return secretByUser, nil
}

// ListUsers retrieves all users matching the supplied filters, ordered by the
// supplied sort items, excluding support accounts, and preloads each user's
// roles (with permissions), environment-targeted access control entries and auth
// secret. It mirrors the legacy GetAllUsers behavior, which returns every
// matching user without pagination.
func (s *Store) ListUsers(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems) ([]services.User, error) {
	var (
		userSB       = sqlbuilder.PostgreSQL.NewSelectBuilder()
		userRows     pgx.Rows
		listedUsers  []userRow
		result       []services.User
		orderBy      []string
		userIDs      []string
		rolesByUser  map[uuid.UUID][]services.Role
		etacByUser   map[uuid.UUID][]services.EnvironmentAccessControl
		secretByUser map[uuid.UUID]*services.AuthSecret
		err          error
	)

	userSB.Select(validUserColumns...).From(tableUsers)

	// Support accounts are never surfaced by the list endpoint; the exclusion is
	// combined with any caller-supplied filters using AND.
	userSB.Where(userSB.Equal("support_account", false))

	if err = applyUserFilters(userSB, queryFilters); err != nil {
		return nil, err
	}

	orderBy, err = buildUserOrderBy(sortItems)
	if err != nil {
		return nil, err
	}
	if len(orderBy) > 0 {
		userSB.OrderBy(orderBy...)
	}

	userQuery, userArgs := userSB.Build()

	userRows, err = s.db.Query(ctx, userQuery, userArgs...)
	if err != nil {
		return nil, err
	}
	listedUsers, err = pgx.CollectRows(userRows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, fmt.Errorf("collecting users: %s", err)
	}

	if len(listedUsers) == 0 {
		return []services.User{}, nil
	}

	userIDs = make([]string, 0, len(listedUsers))
	for _, listedUser := range listedUsers {
		userIDs = append(userIDs, listedUser.ID.String())
	}

	if rolesByUser, err = s.getRolesForUsers(ctx, userIDs); err != nil {
		return nil, err
	}
	if etacByUser, err = s.getEnvironmentAccessControlForUsers(ctx, userIDs); err != nil {
		return nil, err
	}
	if secretByUser, err = s.getAuthSecretsForUsers(ctx, userIDs); err != nil {
		return nil, err
	}

	result = make([]services.User, 0, len(listedUsers))
	for _, listedUser := range listedUsers {
		result = append(result, toUser(listedUser, rolesByUser[listedUser.ID], etacByUser[listedUser.ID], secretByUser[listedUser.ID]))
	}

	return result, nil
}
