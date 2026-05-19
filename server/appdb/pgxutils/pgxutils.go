package pgxutils

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// PgxQuerier is a minimal interface that matches the pgxpool.Pool and pgx.Conn QueryRow methods.
type PgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
