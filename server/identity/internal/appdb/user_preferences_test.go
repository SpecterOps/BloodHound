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

package appdb_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/require"
)

const preferencesSelectSQL = `SELECT COALESCE(p.ui_storage, ''::bytea)
    FROM users u LEFT JOIN user_preferences p ON p.user_id = u.id WHERE u.id = $1`
const preferencesUpsertSQL = `INSERT INTO user_preferences (user_id, ui_storage)
    VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE
    SET ui_storage = EXCLUDED.ui_storage, updated_at = clock_timestamp()`
const preferencesAuditSQL = `INSERT INTO audit_logs
    (created_at, actor_id, actor_name, actor_email, action, fields,
    request_id, source_ip_address, status, commit_id)
    VALUES (clock_timestamp(), $1, $2, $3, $4, $5, $6, $7, $8, $9)`

func TestStore_GetUserPreferences(t *testing.T) {
	var (
		userID        = uuid.Must(uuid.NewV4())
		databaseError = errors.New("database unavailable")
	)
	for _, testCase := range []struct {
		name          string
		storage       []byte
		missing       bool
		queryError    error
		rowError      error
		expectedError error
	}{
		{name: "opaque bytes", storage: []byte{0, 255, 128, 10}},
		{name: "no saved preferences", storage: []byte{}},
		{name: "missing user", missing: true, expectedError: services.ErrUserNotFound},
		{name: "query failure", queryError: databaseError, expectedError: databaseError},
		{name: "read failure", rowError: databaseError, expectedError: databaseError},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			if testCase.queryError != nil {
				pool.ExpectQuery(preferencesSelectSQL).WithArgs(userID.String()).WillReturnError(testCase.queryError)
			} else {
				rows := pool.NewRows([]string{"ui_storage"})
				if !testCase.missing {
					rows.AddRow(testCase.storage)
				}
				if testCase.rowError != nil {
					rows.RowError(0, testCase.rowError)
				}
				pool.ExpectQuery(preferencesSelectSQL).WithArgs(userID.String()).WillReturnRows(rows)
			}
			storage, err := store.GetUserPreferences(context.Background(), userID)
			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.storage, storage)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

type preferencesAuditFields struct {
	userID uuid.UUID
	size   int
}

// Match ensures only metadata is logged, without exposing opaque payload bytes.
func (s preferencesAuditFields) Match(value any) bool {
	var fields map[string]any
	text, ok := value.(string)
	return ok && json.Unmarshal([]byte(text), &fields) == nil && len(fields) == 2 &&
		fields["user_id"] == s.userID.String() && fields["ui_storage_bytes"] == float64(s.size)
}

func TestStore_UpsertUserPreferences(t *testing.T) {
	var (
		userID        = uuid.Must(uuid.NewV4())
		storage       = []byte{0, 255, 128, 10}
		databaseError = errors.New("database unavailable")
		ctx           = bhctx.Set(context.Background(), &bhctx.Context{
			RequestID: "preferences-request",
			RequestIP: "127.0.0.1",
			AuthCtx: auth.Context{Owner: model.User{
				Unique: model.Unique{ID: userID}, PrincipalName: "preferences-user",
			}},
		})
	)
	for _, testCase := range []struct {
		name          string
		empty         bool
		beginError    error
		writeError    error
		auditError    error
		commitError   error
		expectedError error
	}{
		{name: "commit bytes and audit metadata"},
		{name: "nil becomes empty bytes", empty: true},
		{name: "begin failure", beginError: databaseError, expectedError: databaseError},
		{name: "write failure rolls back", writeError: databaseError, expectedError: databaseError},
		{name: "missing user rolls back", writeError: &pgconn.PgError{Code: "23503", ConstraintName: "user_preferences_user_id_fkey"}, expectedError: services.ErrUserNotFound},
		{name: "size constraint rolls back", writeError: &pgconn.PgError{Code: "23514", ConstraintName: "user_preferences_ui_storage_size"}, expectedError: services.ErrUserPreferencesTooLarge},
		{name: "audit failure rolls back preferences", auditError: databaseError, expectedError: databaseError},
		{name: "commit failure", commitError: databaseError, expectedError: databaseError},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				store, pool     = newTestStore(t)
				payload         = storage
				expectedPayload = storage
			)
			if testCase.empty {
				payload = nil
				expectedPayload = []byte{}
			}
			if testCase.beginError != nil {
				pool.ExpectBegin().WillReturnError(testCase.beginError)
			} else {
				pool.ExpectBegin()
				if testCase.writeError != nil {
					pool.ExpectExec(preferencesUpsertSQL).WithArgs(userID.String(), expectedPayload).WillReturnError(testCase.writeError)
				} else {
					pool.ExpectExec(preferencesUpsertSQL).WithArgs(userID.String(), expectedPayload).WillReturnResult(pgxmock.NewResult("INSERT", 1))
					auditExpectation := pool.ExpectExec(preferencesAuditSQL).WithArgs(
						userID.String(), "preferences-user", "", string(model.AuditLogActionUpdateUserPreferences),
						preferencesAuditFields{userID: userID, size: len(payload)}, "preferences-request", "127.0.0.1",
						string(model.AuditLogStatusSuccess), pgxmock.AnyArg(),
					)
					if testCase.auditError != nil {
						auditExpectation.WillReturnError(testCase.auditError)
					} else {
						auditExpectation.WillReturnResult(pgxmock.NewResult("INSERT", 1))
						if testCase.commitError != nil {
							pool.ExpectCommit().WillReturnError(testCase.commitError)
						} else {
							pool.ExpectCommit()
						}
					}
				}
				pool.ExpectRollback()
			}
			err := store.UpsertUserPreferences(ctx, userID, payload)
			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_UpsertUserPreferencesRequiresAuditActor(t *testing.T) {
	store, pool := newTestStore(t)
	require.ErrorContains(t, store.UpsertUserPreferences(context.Background(), uuid.Must(uuid.NewV4()), []byte{1}), "authenticated user")
	require.NoError(t, pool.ExpectationsWereMet())
}
