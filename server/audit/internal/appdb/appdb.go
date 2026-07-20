package appdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const (
	tableAuditLogs             = "audit_logs"
	errorCodeNotNullConstraint = "23502"
	errorCodeCheckConstraint   = "23514"
)

// pgxQuerier is the minimal pgx surface the audit Store relies on. It is
// satisfied by both *pgxpool.Pool and pgx.Tx.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Store performs audit persistence directly against PostgreSQL. It implements
// services.Database.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// InsertAuditLog writes a single audit row (intent, success, or failure). The
// caller (service layer) is responsible for setting Status, CommitID, Source,
// and for redacting Fields before this is called. id is auto-assigned by the
// sequence; created_at is set explicitly here.
func (s *Store) InsertAuditLog(ctx context.Context, record services.AuditRecord) error {
	var (
		fieldsArg     any = record.Fields // map[string]any -> pgx JSONBCodec -> json.Marshal
		insertBuilder     = sqlbuilder.PostgreSQL.NewInsertBuilder()
		sqlQuery      string
		args          []any
		err           error
	)
	// Store SQL NULL for empty fields instead of jsonb 'null'.
	if len(record.Fields) == 0 {
		fieldsArg = nil
	}
	insertBuilder.InsertInto(tableAuditLogs)
	insertBuilder.Cols(
		"created_at", "action", "actor_id", "actor_name", "actor_email",
		"request_id", "source_ip_address", "status", "commit_id", "fields", "source",
	)
	insertBuilder.Values(
		time.Now().UTC(),
		record.Action,
		record.ActorID, record.ActorName, record.ActorEmail,
		record.RequestID, record.SourceIPAddress,
		string(record.Status),
		record.CommitID.String(), // commit_id is TEXT in the schema
		fieldsArg,
		string(record.Source),
	)
	sqlQuery, args = insertBuilder.Build()
	if _, err = s.db.Exec(ctx, sqlQuery, args...); err != nil {
		return mapError(err)
	}
	return nil
}

// mapError translates PostgreSQL driver errors into services sentinels.
func mapError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case errorCodeNotNullConstraint, errorCodeCheckConstraint:
			return fmt.Errorf("%w: %s", services.ErrInvalidAuditRecord, pgErr.Message)
		}
	}
	return fmt.Errorf("inserting audit log: %w", err)
}

// TODO(audit reads): This Store is currently write-only (InsertAuditLog), so it
// has no row-scanning struct. When a "query audit logs" read path is added,
// introduce an auditLogRow scan struct + toAuditRecord mapper here, mirroring
// the clientRow/toClient pattern in server/clients/internal/appdb/appdb.go.
//
//   1. Add scan struct with db: tags matching audit_logs columns:
//        type auditLogRow struct {
//            ID              int64          `db:"id"`               // BIGINT, NOT NULL
//            CreatedAt       time.Time      `db:"created_at"`       // TIMESTAMPTZ, NOT NULL
//            Action          string         `db:"action"`          // TEXT, NOT NULL
//            ActorID         null.String    `db:"actor_id"`
//            ActorName       null.String    `db:"actor_name"`
//            ActorEmail      null.String    `db:"actor_email"`
//            RequestID       null.String    `db:"request_id"`
//            SourceIPAddress null.String    `db:"source_ip_address"`
//            Status          null.String    `db:"status"`          // VARCHAR(15)
//            CommitID        null.String    `db:"commit_id"`       // TEXT (uuid as string)
//            Fields          map[string]any `db:"fields"`          // JSONB (pgx unmarshals)
//            Source          null.String    `db:"source"`          // VARCHAR(20)
//        }
//   2. Add toAuditRecord(row) that flattens null.* via ValueOrZero() and rebuilds
//      CommitID with uuid.FromString(row.CommitID.ValueOrZero()).
//   3. Scan with pgx.CollectRows / pgx.CollectOneRow + pgx.RowToStructByName[auditLogRow],
//      mapping pgx.ErrNoRows -> services.ErrNotFound (see fetchClient for reference).
//
// JSONB note: pgx v5's JSONBCodec unmarshals a jsonb column straight into
// map[string]any (json.Unmarshal), so Fields needs no wrapper type — the read
// mirror of the write path already in InsertAuditLog.
