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
	"encoding/json"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const (
	tableAuditLogs                = "audit_logs"
	auditLogColumnCreatedAt       = "created_at"
	auditLogColumnActorID         = "actor_id"
	auditLogColumnActorName       = "actor_name"
	auditLogColumnActorEmail      = "actor_email"
	auditLogColumnAction          = "action"
	auditLogColumnFields          = "fields"
	auditLogColumnRequestID       = "request_id"
	auditLogColumnSourceIPAddress = "source_ip_address"
	auditLogColumnStatus          = "status"
	auditLogColumnCommitID        = "commit_id"
	sqlCurrentTimestamp           = "current_timestamp"
)

// pgxQuerier lists the pgx methods used by the audit persistence slice and
// supports future audit-log queries and transactional operations.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Store struct {
	db pgxQuerier
}

func NewStore(databaseInterface pgxQuerier) *Store {
	return &Store{db: databaseInterface}
}

func (s *Store) CreateAuditLog(ctx context.Context, auditLog services.AuditLog) error {
	var (
		insertBuilder = sqlbuilder.PostgreSQL.NewInsertBuilder()
		fields        []byte
		sqlQuery      string
		arguments     []any
		err           error
	)

	if fields, err = json.Marshal(auditLog.Fields); err != nil {
		return fmt.Errorf("marshalling audit fields: %w", err)
	}

	insertBuilder.InsertInto(tableAuditLogs)
	insertBuilder.Cols(
		auditLogColumnCreatedAt,
		auditLogColumnActorID,
		auditLogColumnActorName,
		auditLogColumnActorEmail,
		auditLogColumnAction,
		auditLogColumnFields,
		auditLogColumnRequestID,
		auditLogColumnSourceIPAddress,
		auditLogColumnStatus,
		auditLogColumnCommitID,
	)
	insertBuilder.Values(
		sqlbuilder.Raw(sqlCurrentTimestamp),
		auditLog.ActorID,
		auditLog.ActorName,
		auditLog.ActorEmail,
		string(auditLog.Action),
		string(fields),
		auditLog.RequestID,
		auditLog.SourceIPAddress,
		string(auditLog.Status),
		auditLog.CommitID.String(),
	)

	sqlQuery, arguments = insertBuilder.Build()
	if _, err = s.db.Exec(ctx, sqlQuery, arguments...); err != nil {
		return fmt.Errorf("inserting audit log: %w", err)
	}

	return nil
}
