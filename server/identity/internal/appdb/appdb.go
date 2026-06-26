package appdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

const (
	tablePermissions      = "permissions"
	tableRoles            = "roles"
	tableRolesPermissions = "roles_permissions"
	tableUsers            = "users"
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

type user struct {
	ID            string    `db:"id"`
	PrincipalName string    `db:"principal_name"`
	IsDisabled    bool      `db:"is_disabled"`
	EULAAccepted  bool      `db:"eula_accepted"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
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

func toUser(row user) (services.User, error) {
	var (
		userID uuid.UUID
		err    error
	)
	userID, err = uuid.FromString(row.ID)
	if err != nil {
		return services.User{}, fmt.Errorf("parsing user id: %w", err)
	}
	return services.User{
		ID:            userID,
		PrincipalName: row.PrincipalName,
		IsDisabled:    row.IsDisabled,
		EULAAccepted:  row.EULAAccepted,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
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

func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (services.User, error) {
	var (
		userSB   = sqlbuilder.PostgreSQL.NewSelectBuilder()
		userRows pgx.Rows
		userRow  user
		err      error
	)

	userSB.Select("id", "principal_name", "is_disabled", "eula_accepted", "created_at", "updated_at").
		From(tableUsers).
		Where(userSB.Equal("id", id.String())).
		Limit(1)
	userQuery, userArgs := userSB.Build()

	userRows, err = s.db.Query(ctx, userQuery, userArgs...)
	if err != nil {
		return services.User{}, err
	}
	userRow, err = pgx.CollectOneRow(userRows, pgx.RowToStructByName[user])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.User{}, services.ErrNoUserFound
	}
	if err != nil {
		return services.User{}, fmt.Errorf("finding user: %s", err)
	}

	return toUser(userRow)
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
