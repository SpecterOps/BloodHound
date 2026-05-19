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

package analysis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL strings expected by the Store. These are compared via
// pgxmock.QueryMatcherEqual, which whitespace-normalises both sides, so
// formatting (tabs, newlines) is not load-bearing — token order, table
// name, column names, parameter shape and the ON CONFLICT clause are.
const (
	expectedSelectSQL = `SELECT requested_by, request_type, requested_at, delete_all_graph, delete_sourceless_graph, delete_source_kinds, delete_relationships FROM analysis_request_switch LIMIT 1`

	expectedInsertSQL = `INSERT INTO analysis_request_switch("requested_by", "request_type", "requested_at", "delete_all_graph", "delete_sourceless_graph", "delete_source_kinds", "delete_relationships") VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (singleton) DO NOTHING`
)

func newTestStore(t *testing.T) (*Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return &Store{db: pool}, pool
}

func analysisRequestRowColumns() []string {
	return []string{
		"requested_by", "request_type", "requested_at",
		"delete_all_graph", "delete_sourceless_graph",
		"delete_source_kinds", "delete_relationships",
	}
}

func TestStore_GetAnalysisRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully returns an analysis request", func(t *testing.T) {
		var (
			expected = models.RequestedAnalysis{
				RequestedBy:           "test-user",
				RequestType:           models.RequestedAnalysisTypeAnalysis,
				RequestedAt:           time.Now(),
				DeleteAllGraph:        true,
				DeleteSourcelessGraph: false,
				DeleteSourceKinds:     []string{"AZBase"},
				DeleteRelationships:   []string{"HasSession"},
			}
			store, pool = newTestStore(t)
		)

		pool.ExpectQuery(expectedSelectSQL).WillReturnRows(
			pool.NewRows(analysisRequestRowColumns()).AddRow(
				expected.RequestedBy,
				string(expected.RequestType),
				expected.RequestedAt,
				expected.DeleteAllGraph,
				expected.DeleteSourcelessGraph,
				expected.DeleteSourceKinds,
				expected.DeleteRelationships,
			),
		)

		result, err := store.GetAnalysisRequest(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when no rows are found", func(t *testing.T) {
		store, pool := newTestStore(t)

		pool.ExpectQuery(expectedSelectSQL).WillReturnError(pgx.ErrNoRows)

		_, err := store.GetAnalysisRequest(ctx)
		assert.ErrorIs(t, err, ErrNotFound)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates other database errors", func(t *testing.T) {
		var (
			expectedErr = errors.New("connection refused")
			store, pool = newTestStore(t)
		)

		pool.ExpectQuery(expectedSelectSQL).WillReturnError(expectedErr)

		_, err := store.GetAnalysisRequest(ctx)
		assert.ErrorIs(t, err, expectedErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestStore_CreateAnalysisRequest(t *testing.T) {
	var (
		ctx       = context.Background()
		requester = "test-user"
	)

	t.Run("returns created=true and the new request when no row exists", func(t *testing.T) {
		store, pool := newTestStore(t)

		pool.ExpectExec(expectedInsertSQL).
			WithArgs(
				requester,
				string(models.RequestedAnalysisTypeAnalysis),
				pgxmock.AnyArg(), // now
				false,
				false,
				[]string{},
				[]string{},
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		pool.ExpectQuery(expectedSelectSQL).WillReturnRows(
			pool.NewRows(analysisRequestRowColumns()).AddRow(
				requester,
				string(models.RequestedAnalysisTypeAnalysis),
				time.Now().UTC(),
				false,
				false,
				[]string{},
				[]string{},
			),
		)

		current, created, err := store.CreateAnalysisRequest(ctx, requester)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, requester, current.RequestedBy)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("returns created=false and the existing request when a row already exists", func(t *testing.T) {
		var (
			store, pool = newTestStore(t)
			existing    = models.RequestedAnalysis{
				RequestedBy: "other-user",
				RequestType: models.RequestedAnalysisTypeAnalysis,
				RequestedAt: time.Now().UTC(),
			}
		)

		// ON CONFLICT DO NOTHING reports 0 rows affected when the singleton row
		// already exists — the INSERT is still issued, the conflict is silently
		// resolved at the DB level.
		pool.ExpectExec(expectedInsertSQL).
			WithArgs(
				requester,
				string(models.RequestedAnalysisTypeAnalysis),
				pgxmock.AnyArg(),
				false,
				false,
				[]string{},
				[]string{},
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 0))

		pool.ExpectQuery(expectedSelectSQL).WillReturnRows(
			pool.NewRows(analysisRequestRowColumns()).AddRow(
				existing.RequestedBy,
				string(existing.RequestType),
				existing.RequestedAt,
				false,
				false,
				[]string{},
				[]string{},
			),
		)

		current, created, err := store.CreateAnalysisRequest(ctx, requester)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, existing.RequestedBy, current.RequestedBy)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates insert errors", func(t *testing.T) {
		var (
			expectedErr = errors.New("insert failed")
			store, pool = newTestStore(t)
		)

		pool.ExpectExec(expectedInsertSQL).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
				pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(expectedErr)

		_, _, err := store.CreateAnalysisRequest(ctx, requester)
		assert.ErrorIs(t, err, expectedErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates select errors after a successful insert", func(t *testing.T) {
		var (
			expectedErr = errors.New("connection lost")
			store, pool = newTestStore(t)
		)

		pool.ExpectExec(expectedInsertSQL).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
				pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		pool.ExpectQuery(expectedSelectSQL).WillReturnError(expectedErr)

		_, _, err := store.CreateAnalysisRequest(ctx, requester)
		assert.ErrorIs(t, err, expectedErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})
}
