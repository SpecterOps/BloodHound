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
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_InsertAuditLog_ColumnsAndArgs(t *testing.T) {
	var commitID = uuid.Must(uuid.NewV4())

	tests := []struct {
		name   string
		record services.AuditRecord
	}{
		{
			name: "maps every column arg by position",
			record: services.AuditRecord{
				Action:          "POST /api/v2/roles/{role_id}",
				ActorID:         "actor-id",
				ActorName:       "actor-name",
				ActorEmail:      "actor@example.com",
				RequestID:       "req-1",
				SourceIPAddress: "10.0.0.1",
				Status:          services.StatusIntent,
				CommitID:        commitID,
				Fields:          map[string]any{"key": "value"},
				Source:          services.SourceMiddleware,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				querier = &recordingQuerier{}
				store   = NewStore(querier)
			)

			require.NoError(t, store.InsertAuditLog(context.Background(), tt.record))
			require.Len(t, querier.execSQL, 1)
			assert.Contains(t, querier.execSQL[0], "INSERT INTO audit_logs")

			// created_at is generated inside the method (time.Now), so assert on
			// the remaining args by position: created_at is args[0].
			args := querier.execArgs[0]
			require.Len(t, args, 11)
			assert.Equal(t, tt.record.Action, args[1])
			assert.Equal(t, tt.record.ActorID, args[2])
			assert.Equal(t, tt.record.ActorName, args[3])
			assert.Equal(t, tt.record.ActorEmail, args[4])
			assert.Equal(t, tt.record.RequestID, args[5])
			assert.Equal(t, tt.record.SourceIPAddress, args[6])
			assert.Equal(t, string(tt.record.Status), args[7])
			assert.Equal(t, tt.record.CommitID.String(), args[8])
			assert.Equal(t, tt.record.Fields, args[9])
			assert.Equal(t, string(tt.record.Source), args[10])
		})
	}
}

// TestStore_InsertAuditLog_EmptyFieldsBecomeNull verifies the branch that stores
// SQL NULL (nil arg) rather than a jsonb 'null' when Fields is empty.
func TestStore_InsertAuditLog_EmptyFieldsBecomeNull(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]any
	}{
		{name: "nil fields", fields: nil},
		{name: "empty fields", fields: map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				querier = &recordingQuerier{}
				store   = NewStore(querier)
			)

			require.NoError(t, store.InsertAuditLog(context.Background(), services.AuditRecord{
				Action: "GET /x",
				Status: services.StatusSuccess,
				Fields: tt.fields,
			}))

			require.Len(t, querier.execArgs, 1)
			assert.Nil(t, querier.execArgs[0][9], "empty fields should be stored as SQL NULL")
		})
	}
}

func TestStore_InsertAuditLog_MapsErrors(t *testing.T) {
	var sentinel = errors.New("connection refused")

	tests := []struct {
		name          string
		execErr       error
		wantSentinel  bool
		wantWrapped   error
		wantErrString string
	}{
		{
			name:         "not null constraint maps to the invalid record sentinel",
			execErr:      &pgconn.PgError{Code: errorCodeNotNullConstraint, Message: "boom"},
			wantSentinel: true,
		},
		{
			name:         "invalid enum value maps to the invalid record sentinel",
			execErr:      &pgconn.PgError{Code: errorCodeInvalidTextRepresentation, Message: "boom"},
			wantSentinel: true,
		},
		{
			name:          "unmapped error is wrapped and not the invalid record sentinel",
			execErr:       sentinel,
			wantWrapped:   sentinel,
			wantErrString: "inserting audit log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				querier = &recordingQuerier{execErr: tt.execErr}
				store   = NewStore(querier)
			)

			err := store.InsertAuditLog(context.Background(), services.AuditRecord{Action: "GET /x"})
			require.Error(t, err)

			if tt.wantSentinel {
				assert.ErrorIs(t, err, services.ErrInvalidAuditRecord)
				return
			}

			assert.NotErrorIs(t, err, services.ErrInvalidAuditRecord)
			assert.ErrorIs(t, err, tt.wantWrapped)
			assert.Contains(t, err.Error(), tt.wantErrString)
		})
	}
}
