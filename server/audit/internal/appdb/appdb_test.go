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

type recordingExecer struct {
	queries [][]any
	err     error
}

func (s *recordingExecer) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	s.queries = append(s.queries, append([]any{query}, args...))
	return pgconn.CommandTag{}, s.err
}

func TestStore_InsertAuditLog(t *testing.T) {
	var commitID = uuid.Must(uuid.NewV4())
	testCases := []struct {
		name           string
		fields         map[string]any
		expectedFields any
	}{
		{name: "maps populated record", fields: map[string]any{"key": "value"}, expectedFields: map[string]any{"key": "value"}},
		{name: "stores empty fields as null", fields: map[string]any{}, expectedFields: nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				database = &recordingExecer{}
				store    = NewStore(database)
				record   = services.AuditRecord{
					Action:          "POST/api/v2/roles/{role_id}",
					ActorID:         "actor-id",
					ActorName:       "actor-name",
					ActorEmail:      "actor@example.com",
					RequestID:       "request-id",
					SourceIPAddress: "10.0.0.1",
					Status:          services.StatusIntent,
					CommitID:        commitID,
					Fields:          testCase.fields,
					Source:          services.SourceMiddleware,
				}
			)

			require.NoError(t, store.InsertAuditLog(context.Background(), record))
			require.Len(t, database.queries, 1)
			assert.Contains(t, database.queries[0][0], "INSERT INTO audit_logs")
			assert.Equal(t, record.Action, database.queries[0][2])
			assert.Equal(t, record.Status, services.Status(database.queries[0][8].(string)))
			assert.Equal(t, record.CommitID.String(), database.queries[0][9])
			assert.Equal(t, testCase.expectedFields, database.queries[0][10])
			assert.Equal(t, record.Source, database.queries[0][11])
		})
	}
}

func TestStore_InsertAuditLogMapsErrors(t *testing.T) {
	var connectionError = errors.New("connection failed")
	testCases := []struct {
		name             string
		databaseError    error
		expectedSentinel error
	}{
		{name: "maps not-null violation", databaseError: &pgconn.PgError{Code: errorCodeNotNullConstraint}, expectedSentinel: services.ErrInvalidAuditRecord},
		{name: "maps invalid enum", databaseError: &pgconn.PgError{Code: errorCodeInvalidTextRepresentation}, expectedSentinel: services.ErrInvalidAuditRecord},
		{name: "wraps connection error", databaseError: connectionError, expectedSentinel: connectionError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var store = NewStore(&recordingExecer{err: testCase.databaseError})

			err := store.InsertAuditLog(context.Background(), services.AuditRecord{Action: "GET/api/v2/test"})
			require.Error(t, err)
			assert.ErrorIs(t, err, testCase.expectedSentinel)
		})
	}
}
