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

package appdb_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pashagolub/pgxmock/v6"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/audit/internal/appdb"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedInsertSQL = `INSERT INTO audit_logs (created_at, actor_id, actor_name, actor_email, action, fields, request_id, source_ip_address, status, commit_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func TestStore_CreateAuditLog(t *testing.T) {
	t.Parallel()

	var (
		ctx         = context.Background()
		createdAt   = time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
		commitID    = uuid.Must(uuid.NewV4())
		databaseErr = errors.New("database unavailable")
		auditLog    = services.Log{
			CreatedAt:       createdAt,
			ActorID:         "actor-id",
			ActorName:       "actor",
			ActorEmail:      "actor@example.com",
			Action:          model.AuditLogActionUpdateParameter,
			Fields:          map[string]any{"key": "value"},
			RequestID:       "request-id",
			SourceIPAddress: "127.0.0.1",
			Status:          model.AuditLogStatusIntent,
			CommitID:        commitID,
		}
	)

	tests := []struct {
		name       string
		setupMocks func(pgxmock.PgxPoolIface)
		fields     map[string]any
		expected   struct {
			error         error
			errorContains string
		}
	}{
		{
			name: "Success: inserts audit log",
			setupMocks: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedInsertSQL).
					WithArgs(
						createdAt, "actor-id", "actor", "actor@example.com",
						string(model.AuditLogActionUpdateParameter), `{"key":"value"}`,
						"request-id", "127.0.0.1", string(model.AuditLogStatusIntent), commitID.String(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Error: database insert failure",
			setupMocks: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedInsertSQL).
					WithArgs(
						createdAt, "actor-id", "actor", "actor@example.com",
						string(model.AuditLogActionUpdateParameter), `{"key":"value"}`,
						"request-id", "127.0.0.1", string(model.AuditLogStatusIntent), commitID.String(),
					).
					WillReturnError(databaseErr)
			},
			expected: struct {
				error         error
				errorContains string
			}{error: databaseErr},
		},
		{
			name:   "Error: audit fields cannot be marshaled",
			fields: map[string]any{"invalid": func() {}},
			expected: struct {
				error         error
				errorContains string
			}{errorContains: "marshalling audit fields"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			t.Cleanup(pool.Close)
			if testCase.setupMocks != nil {
				testCase.setupMocks(pool)
			}
			store := appdb.NewStore(pool)
			input := auditLog
			if testCase.fields != nil {
				input.Fields = testCase.fields
			}

			err = store.CreateAuditLog(ctx, input)

			switch {
			case testCase.expected.error != nil:
				assert.ErrorIs(t, err, testCase.expected.error)
			case testCase.expected.errorContains != "":
				assert.ErrorContains(t, err, testCase.expected.errorContains)
			default:
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
