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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const tableAuditLogs = "audit_logs"

type pgxExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Store struct {
	database pgxExecer
}

func NewStore(databaseInterface pgxExecer) *Store {
	return &Store{database: databaseInterface}
}

func (s *Store) CreateAuditLog(ctx context.Context, auditLog services.Log) error {
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
		"created_at", "actor_id", "actor_name", "actor_email",
		"action", "fields", "request_id", "source_ip_address",
		"status", "commit_id",
	)
	insertBuilder.Values(
		auditLog.CreatedAt,
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
	if _, err = s.database.Exec(ctx, sqlQuery, arguments...); err != nil {
		return fmt.Errorf("inserting audit log: %w", err)
	}

	return nil
}
