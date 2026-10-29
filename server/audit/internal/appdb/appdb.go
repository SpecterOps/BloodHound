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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const (
	tableAuditLogs                     = "audit_logs"
	errorCodeNotNullConstraint         = "23502"
	errorCodeInvalidTextRepresentation = "22P02"
)

type pgxExecer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Store struct {
	database pgxExecer
}

func NewStore(database pgxExecer) *Store {
	return &Store{database: database}
}

func (s *Store) InsertAuditLog(ctx context.Context, record services.AuditRecord) error {
	var (
		fields  any = record.Fields
		builder     = sqlbuilder.PostgreSQL.NewInsertBuilder()
	)

	if len(record.Fields) == 0 {
		fields = nil
	}

	builder.InsertInto(tableAuditLogs)
	builder.Cols(
		"created_at", "action", "actor_id", "actor_name", "actor_email",
		"request_id", "source_ip_address", "status", "commit_id", "fields", "source",
	)
	builder.Values(
		time.Now().UTC(), record.Action, record.ActorID, record.ActorName, record.ActorEmail,
		record.RequestID, record.SourceIPAddress, string(record.Status), record.CommitID.String(), fields, record.Source,
	)

	query, args := builder.Build()
	if _, err := s.database.Exec(ctx, query, args...); err != nil {
		return mapError(err)
	}

	return nil
}

func mapError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case errorCodeNotNullConstraint, errorCodeInvalidTextRepresentation:
			return fmt.Errorf("%w: %s", services.ErrInvalidAuditRecord, postgresError.Message)
		}
	}

	return fmt.Errorf("inserting audit log: %w", err)
}
