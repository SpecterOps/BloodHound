package appdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

const (
	tablePermissions = "permissions"
)

type queryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type pgxQuerier interface {
	queryExecer
}

type permission struct {
	Authority string       `db:"authority"`
	Name      string       `db:"name"`
	ID        int32        `db:"id"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
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
		DeletedAt: row.DeletedAt,
	}
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
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[permission])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Permission{}, services.ErrNoPermissionFound
	}
	if err != nil {
		return services.Permission{}, fmt.Errorf("finding permission: %s", err)
	}

	return toPermission(row), nil
}
