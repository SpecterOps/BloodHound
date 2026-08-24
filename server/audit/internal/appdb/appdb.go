// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const (
	tableAuditLogs                     = "audit_logs"
	errorCodeNotNullConstraint         = "23502"
	errorCodeInvalidTextRepresentation = "22P02"

	colCreatedAt       = "created_at"
	colAction          = "action"
	colActorID         = "actor_id"
	colActorName       = "actor_name"
	colActorEmail      = "actor_email"
	colRequestID       = "request_id"
	colSourceIPAddress = "source_ip_address"
	colStatus          = "status"
	colCommitID        = "commit_id"
	colFields          = "fields"
	colSource          = "source"
)

// auditLogInsertColumns is the ordered column list written by InsertAuditLog. The
// order must match the InsertAuditLog Values call.
func auditLogInsertColumns() []string {
	return []string{
		colCreatedAt, colAction, colActorID, colActorName, colActorEmail,
		colRequestID, colSourceIPAddress, colStatus, colCommitID, colFields, colSource,
	}
}

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

// InsertAuditLog writes a single audit row (intent, success, or failure). id is
// auto-assigned by the sequence; created_at is set explicitly here.
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
	insertBuilder.Cols(auditLogInsertColumns()...)
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
		case errorCodeNotNullConstraint, errorCodeInvalidTextRepresentation:
			return fmt.Errorf("%w: %s", services.ErrInvalidAuditRecord, pgErr.Message)
		}
	}
	return fmt.Errorf("inserting audit log: %w", err)
}

// TODO(audit reads): This Store is write-only. When a read path is added,
// introduce an auditLogRow scan struct + toAuditRecord mapper here, mirroring the
// clientRow/toClient pattern in server/clients/internal/appdb/appdb.go.
