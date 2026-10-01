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

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

// GetUserPreferences distinguishes an absent user from an existing user with
// no saved preferences. The latter returns an empty byte slice without writing.
func (s *Store) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	rows, err := s.db.Query(ctx, `
        SELECT COALESCE(p.ui_storage, ''::bytea)
        FROM users u LEFT JOIN user_preferences p ON p.user_id = u.id
        WHERE u.id = $1`, userID.String())
	if err != nil {
		return nil, fmt.Errorf("querying user preferences: %w", err)
	}
	storage, err := pgx.CollectOneRow(rows, pgx.RowTo[[]byte])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, services.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("reading user preferences: %w", err)
	}
	return storage, nil
}

// UpsertUserPreferences atomically replaces the opaque UI storage and records
// the target and byte count in the audit log. Blob contents are never logged.
func (s *Store) UpsertUserPreferences(ctx context.Context, userID uuid.UUID, storage []byte) error {
	var (
		requestContext = bhctx.Get(ctx)
		actor, isUser  = auth.GetUserFromAuthCtx(requestContext.AuthCtx)
		postgresError  *pgconn.PgError
	)

	if !isUser {
		return errors.New("user preferences audit requires an authenticated user")
	}
	if storage == nil {
		storage = []byte{}
	}
	auditFields, err := json.Marshal(model.AuditData{
		"user_id":          userID.String(),
		"ui_storage_bytes": len(storage),
	})
	if err != nil {
		return fmt.Errorf("encoding user preferences audit metadata: %w", err)
	}
	commitID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("generating user preferences audit commit ID: %w", err)
	}
	transaction, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning user preferences transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err = transaction.Exec(ctx, `
        INSERT INTO user_preferences (user_id, ui_storage)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO UPDATE
        SET ui_storage = EXCLUDED.ui_storage, updated_at = clock_timestamp()`, userID.String(), storage); err != nil {
		if errors.As(err, &postgresError) {
			switch {
			case postgresError.Code == "23503" && postgresError.ConstraintName == "user_preferences_user_id_fkey":
				return services.ErrUserNotFound
			case postgresError.Code == "23514" && postgresError.ConstraintName == "user_preferences_ui_storage_size":
				return services.ErrUserPreferencesTooLarge
			}
		}
		return fmt.Errorf("upserting user preferences: %w", err)
	}

	if _, err = transaction.Exec(ctx, `
        INSERT INTO audit_logs
            (created_at, actor_id, actor_name, actor_email, action, fields,
             request_id, source_ip_address, status, commit_id)
        VALUES (clock_timestamp(), $1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		actor.ID.String(), actor.PrincipalName, actor.EmailAddress.ValueOrZero(),
		string(model.AuditLogActionUpdateUserPreferences), string(auditFields),
		requestContext.RequestID, requestContext.RequestIP,
		string(model.AuditLogStatusSuccess), commitID.String()); err != nil {
		return fmt.Errorf("auditing user preferences: %w", err)
	}
	if err = transaction.Commit(ctx); err != nil {
		return fmt.Errorf("committing user preferences: %w", err)
	}
	return nil
}
