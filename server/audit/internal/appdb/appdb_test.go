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
	"regexp"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingDatabase struct {
	queries [][]any
	err     error
}

type recordingExecer = recordingDatabase

func (s *recordingDatabase) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	s.queries = append(s.queries, append([]any{query}, args...))
	return pgconn.CommandTag{}, s.err
}

func (s *recordingDatabase) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func TestStore_ListAuditLogs(t *testing.T) {
	t.Parallel()
	var (
		ctx      = context.Background()
		before   = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
		after    = before.Add(-time.Hour)
		commitID = uuid.Must(uuid.NewV4())
		query    = services.ListQuery{
			Before: before, After: after, Skip: 3, Limit: 10,
			Filters: params.Filters{"actor_name": {{Field: "actor_name", Operator: params.Equals, Value: "Ada", SetOperator: params.FilterAnd}}},
			Sort:    params.SortItems{{Field: "created_at", Direction: params.Descending}},
		}
	)

	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	countBuilder, err := newAuditLogSelectBuilder(query)
	require.NoError(t, err)
	countBuilder.Select("COUNT(*) AS count")
	countSQL, countArgs := countBuilder.Build()
	pool.ExpectQuery(regexp.QuoteMeta(countSQL)).WithArgs(countArgs...).WillReturnRows(pool.NewRows([]string{"count"}).AddRow(int64(2)))

	listBuilder, err := newAuditLogSelectBuilder(query)
	require.NoError(t, err)
	listBuilder.Select(selectedAuditLogColumns...).OrderBy("created_at DESC").Offset(3).Limit(10)
	listSQL, listArgs := listBuilder.Build()
	pool.ExpectQuery(regexp.QuoteMeta(listSQL)).WithArgs(listArgs...).WillReturnRows(
		pool.NewRows(selectedAuditLogColumns).AddRow(
			int64(7), before.Add(-time.Minute), "actor-id", "Ada", "ada@example.com", "CreateUser",
			map[string]any{"name": "user"}, "request-id", "127.0.0.1", "success", commitID.String(),
		),
	)

	logs, count, err := NewStore(pool).ListAuditLogs(ctx, query)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	require.Len(t, logs, 1)
	assert.Equal(t, int64(7), logs[0].ID)
	assert.Equal(t, commitID, logs[0].CommitID)
	assert.Contains(t, countSQL, "FROM audit_logs_old")
	assert.Contains(t, countSQL, "FROM audit_logs")
	require.NoError(t, pool.ExpectationsWereMet())
}

func TestApplyFilters_RejectsUnknownFields(t *testing.T) {
	t.Parallel()
	_, err := newAuditLogSelectBuilder(services.ListQuery{
		Filters: params.Filters{"unknown": {{Field: "unknown", Operator: params.Equals, Value: "value"}}},
	})
	assert.ErrorContains(t, err, "unknown field")
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
				database = &recordingDatabase{}
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
				}
			)

			require.NoError(t, store.InsertAuditLog(context.Background(), record))
			require.Len(t, database.queries, 1)
			assert.Contains(t, database.queries[0][0], "INSERT INTO audit_logs")
			assert.Equal(t, record.Action, database.queries[0][2])
			assert.Equal(t, record.Status, services.Status(database.queries[0][8].(string)))
			assert.Equal(t, record.CommitID.String(), database.queries[0][9])
			assert.Equal(t, testCase.expectedFields, database.queries[0][10])
			require.Len(t, database.queries[0], 11)
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
			var store = NewStore(&recordingDatabase{err: testCase.databaseError})

			err := store.InsertAuditLog(context.Background(), services.AuditRecord{Action: "GET/api/v2/test"})
			require.Error(t, err)
			assert.ErrorIs(t, err, testCase.expectedSentinel)
		})
	}
}
