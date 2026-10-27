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
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const auditLogsUnion = `(
SELECT id, created_at, COALESCE(actor_id, '') AS actor_id, COALESCE(actor_name, '') AS actor_name,
       COALESCE(actor_email, '') AS actor_email, action, COALESCE(fields, '{}'::jsonb) AS fields,
       COALESCE(request_id, '') AS request_id, COALESCE(source_ip_address, '') AS source_ip_address,
       COALESCE(status::text, '') AS status, COALESCE(commit_id, '00000000-0000-0000-0000-000000000000') AS commit_id
FROM audit_logs_old
UNION ALL
SELECT id, created_at, COALESCE(actor_id, '') AS actor_id, COALESCE(actor_name, '') AS actor_name,
       COALESCE(actor_email, '') AS actor_email, action, COALESCE(fields, '{}'::jsonb) AS fields,
       COALESCE(request_id, '') AS request_id, COALESCE(source_ip_address, '') AS source_ip_address,
       COALESCE(status::text, '') AS status, COALESCE(commit_id, '00000000-0000-0000-0000-000000000000') AS commit_id
FROM audit_logs
) AS combined_audit_logs`

var auditLogColumns = map[string]string{
	"id":                "id",
	"actor_id":          "actor_id",
	"actor_name":        "actor_name",
	"actor_email":       "actor_email",
	"action":            "action",
	"request_id":        "request_id",
	"created_at":        "created_at",
	"source_ip_address": "source_ip_address",
	"status":            "status",
}

var selectedAuditLogColumns = []string{
	"id", "created_at", "actor_id", "actor_name", "actor_email", "action", "fields",
	"request_id", "source_ip_address", "status", "commit_id",
}

type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type auditLogRow struct {
	ID              int64          `db:"id"`
	CreatedAt       time.Time      `db:"created_at"`
	ActorID         string         `db:"actor_id"`
	ActorName       string         `db:"actor_name"`
	ActorEmail      string         `db:"actor_email"`
	Action          string         `db:"action"`
	Fields          map[string]any `db:"fields"`
	RequestID       string         `db:"request_id"`
	SourceIPAddress string         `db:"source_ip_address"`
	Status          string         `db:"status"`
	CommitID        string         `db:"commit_id"`
}

type auditLogCount struct {
	Count int64 `db:"count"`
}

// Store implements audit persistence.
type Store struct {
	database pgxQuerier
}

// NewStore returns a Store backed by databaseInterface.
func NewStore(databaseInterface pgxQuerier) *Store {
	return &Store{database: databaseInterface}
}

// ListAuditLogs queries the legacy and partitioned tables as one logical dataset.
func (s *Store) ListAuditLogs(ctx context.Context, query services.ListQuery) ([]services.Log, int, error) {
	var (
		countBuilder *sqlbuilder.SelectBuilder
		listBuilder  *sqlbuilder.SelectBuilder
		countRows    pgx.Rows
		listRows     pgx.Rows
		countRow     auditLogCount
		rows         []auditLogRow
		result       []services.Log
		orderBy      []string
		countSQL     string
		countArgs    []any
		listSQL      string
		listArgs     []any
		err          error
	)

	if countBuilder, err = newAuditLogSelectBuilder(query); err != nil {
		return nil, 0, err
	}
	if listBuilder, err = newAuditLogSelectBuilder(query); err != nil {
		return nil, 0, err
	}

	countBuilder.Select("COUNT(*) AS count")
	countSQL, countArgs = countBuilder.Build()
	if countRows, err = s.database.Query(ctx, countSQL, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("counting audit logs: %w", err)
	}
	if countRow, err = pgx.CollectOneRow(countRows, pgx.RowToStructByName[auditLogCount]); err != nil {
		return nil, 0, fmt.Errorf("collecting audit log count: %w", err)
	}

	listBuilder.Select(selectedAuditLogColumns...)
	if orderBy, err = buildOrderBy(query.Sort); err != nil {
		return nil, 0, err
	}
	listBuilder.OrderBy(orderBy...)
	if query.Skip > 0 {
		listBuilder.Offset(query.Skip)
	}
	if query.Limit > 0 {
		listBuilder.Limit(query.Limit)
	}

	listSQL, listArgs = listBuilder.Build()
	if listRows, err = s.database.Query(ctx, listSQL, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("listing audit logs: %w", err)
	}
	if rows, err = pgx.CollectRows(listRows, pgx.RowToStructByName[auditLogRow]); err != nil {
		return nil, 0, fmt.Errorf("collecting audit logs: %w", err)
	}

	result = make([]services.Log, 0, len(rows))
	for _, row := range rows {
		log, conversionErr := toAuditLog(row)
		if conversionErr != nil {
			return nil, 0, conversionErr
		}
		result = append(result, log)
	}

	return result, int(countRow.Count), nil
}

func toAuditLog(row auditLogRow) (services.Log, error) {
	commitID, err := uuid.FromString(row.CommitID)
	if err != nil {
		return services.Log{}, fmt.Errorf("parsing audit log commit id: %w", err)
	}

	return services.Log{
		ID:              row.ID,
		CreatedAt:       row.CreatedAt,
		ActorID:         row.ActorID,
		ActorName:       row.ActorName,
		ActorEmail:      row.ActorEmail,
		Action:          row.Action,
		Fields:          row.Fields,
		RequestID:       row.RequestID,
		SourceIPAddress: row.SourceIPAddress,
		Status:          row.Status,
		CommitID:        commitID,
	}, nil
}

func newAuditLogSelectBuilder(query services.ListQuery) (*sqlbuilder.SelectBuilder, error) {
	var builder = sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.From(auditLogsUnion)
	builder.Where(builder.Between("created_at", query.After, query.Before))
	if err := applyFilters(builder, query.Filters); err != nil {
		return nil, err
	}
	return builder, nil
}

func applyFilters(builder *sqlbuilder.SelectBuilder, filters params.Filters) error {
	for field, fieldFilters := range filters {
		column, valid := auditLogColumns[field]
		if !valid {
			return fmt.Errorf("audit filter references unknown field %q", field)
		}
		expressions := make([]string, 0, len(fieldFilters))
		for _, filter := range fieldFilters {
			var expression string
			switch filter.Operator {
			case params.Equals:
				expression = builder.Equal(column, filter.Value)
			case params.NotEquals:
				expression = builder.NotEqual(column, filter.Value)
			case params.GreaterThan:
				expression = builder.GreaterThan(column, filter.Value)
			case params.GreaterThanOrEquals:
				expression = builder.GreaterEqualThan(column, filter.Value)
			case params.LessThan:
				expression = builder.LessThan(column, filter.Value)
			case params.LessThanOrEquals:
				expression = builder.LessEqualThan(column, filter.Value)
			default:
				return fmt.Errorf("audit filter uses unsupported operator %q", filter.Operator)
			}
			expressions = append(expressions, expression)
		}
		if len(fieldFilters) > 0 && fieldFilters[0].SetOperator == params.FilterOr {
			builder.Where(builder.Or(expressions...))
		} else {
			builder.Where(builder.And(expressions...))
		}
	}
	return nil
}

func buildOrderBy(sortItems params.SortItems) ([]string, error) {
	var orderBy = make([]string, 0, len(sortItems))
	for _, sortItem := range sortItems {
		column, valid := auditLogColumns[sortItem.Field]
		if !valid {
			return nil, fmt.Errorf("audit sort references unknown field %q", sortItem.Field)
		}
		if sortItem.Direction == params.Descending {
			orderBy = append(orderBy, column+" DESC")
		} else {
			orderBy = append(orderBy, column+" ASC")
		}
	}
	return orderBy, nil
}
