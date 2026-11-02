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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

const (
	tableParameters = "parameters"
	tableAuditLogs  = "audit_logs"
)

type ParameterRow struct {
	ID          int32             `db:"id"`
	Key         string            `db:"key"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
	Value       types.JSONBObject `db:"value"`

	CreatedAt null.Time `db:"created_at"`
	UpdatedAt null.Time `db:"updated_at"`
	// DeletedAt is not actually stored in the DB
}

func toParameter(row ParameterRow) services.Parameter {
	return services.Parameter{
		ID:          row.ID,
		Key:         services.ParameterKey(row.Key),
		Name:        row.Name,
		Description: row.Description,
		Value:       row.Value,
		CreatedAt:   row.CreatedAt.ValueOrZero(),
		UpdatedAt:   row.UpdatedAt.ValueOrZero(),
	}
}

func (s *Store) GetConfigurationParameter(ctx context.Context, parameterKey services.ParameterKey) (services.Parameter, error) {
	var (
		row  ParameterRow
		rows pgx.Rows
		err  error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"id",
		"key",
		"name",
		"description",
		"value",
		"created_at",
		"updated_at",
	)
	selectBuilder.From(tableParameters)
	selectBuilder.Where(selectBuilder.Equal("key", string(parameterKey)))
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return services.Parameter{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[ParameterRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Parameter{}, services.ErrNotFound
	}
	if err != nil {
		return services.Parameter{}, fmt.Errorf("reading rows: %w", err)
	}

	return toParameter(row), nil
}

func (s *Store) GetAllConfigurationParameters(ctx context.Context) (services.Parameters, error) {
	var (
		parameterRows []ParameterRow
		rows          pgx.Rows
		err           error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"id",
		"key",
		"name",
		"description",
		"value",
		"created_at",
		"updated_at",
	)
	selectBuilder.From(tableParameters)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	} else if parameterRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[ParameterRow]); err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	}

	returnParameters := []services.Parameter{}

	for _, row := range parameterRows {
		returnParameters = append(returnParameters, toParameter(row))
	}

	return services.Parameters(returnParameters), nil
}

// TODO: This will be used in middleware and will need to be removed when implemented
func insertAuditLog(ctx context.Context, tx pgx.Tx, parameterKey services.ParameterKey, value types.JSONBObject) error {
	var (
		commitID, err = uuid.NewV4()
		bheCtx        = bhctx.Get(ctx)
		user, isUser  = auth.GetUserFromAuthCtx(bheCtx.AuthCtx)
	)
	if err != nil {
		return fmt.Errorf("generating commit id: %w", err)
	}
	if !isUser {
		return fmt.Errorf("no authenticated user on context")
	}

	auditData := model.AuditData{
		"key":   parameterKey,
		"value": value,
	}

	fields, err := json.Marshal(auditData)
	if err != nil {
		return fmt.Errorf("marshalling audit fields: %w", err)
	}

	insertBuilder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	insertBuilder.InsertInto(tableAuditLogs)
	insertBuilder.Cols(
		"created_at", "actor_id", "actor_name", "actor_email",
		"action", "fields", "request_id", "source_ip_address",
		"status", "commit_id",
	)
	insertBuilder.Values(
		time.Now().UTC(),
		user.ID.String(),
		user.PrincipalName,
		user.EmailAddress.ValueOrZero(),
		string(model.AuditLogActionUpdateParameter),
		string(fields),
		bheCtx.RequestID,
		bheCtx.RequestIP,
		string(model.AuditLogStatusSuccess),
		commitID.String(),
	)

	sqlQuery, args := insertBuilder.Build()
	_, err = tx.Exec(ctx, sqlQuery, args...)
	return err
}

func (s *Store) UpsertConfigurationParameter(ctx context.Context, parameterKey services.ParameterKey, value types.JSONBObject) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		// Rollback is a no-op once the transaction has been committed.
		_ = tx.Rollback(ctx)
	}()

	now := time.Now().UTC()

	insertBuilder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	insertBuilder.InsertInto(tableParameters)
	insertBuilder.Cols("key", "value", "name", "description", "created_at", "updated_at")
	insertBuilder.Values(parameterKey, value, "", "", now, now)
	insertBuilder.SQL(`
		ON CONFLICT (key) DO UPDATE SET
        value = EXCLUDED.value,
        updated_at = EXCLUDED.updated_at
	`)

	sqlQuery, args := insertBuilder.Build()
	if _, err = tx.Exec(ctx, sqlQuery, args...); err != nil {
		return err
	}

	if err = insertAuditLog(ctx, tx, parameterKey, value); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
