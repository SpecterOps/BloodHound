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
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_GetAnalysisRequest(t *testing.T) {
	var (
		ctx               = context.Background()
		expectedSQL, _, _ = psql.Select(
			sm.Columns(
				"requested_by",
				"request_type",
				"requested_at",
				"delete_all_graph",
				"delete_sourceless_graph",
				"delete_source_kinds",
				"delete_relationships",
			),
			sm.From("analysis_request_switch"),
			sm.Limit(1),
		).Build(ctx)
	)

	newStore := func(t *testing.T) (*Store, pgxmock.PgxPoolIface) {
		pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		t.Cleanup(pool.Close)
		return &Store{db: pool}, pool
	}

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
			store, pool = newStore(t)
		)

		pool.ExpectQuery(expectedSQL).WillReturnRows(
			pool.NewRows([]string{
				"requested_by",
				"request_type",
				"requested_at",
				"delete_all_graph",
				"delete_sourceless_graph",
				"delete_source_kinds",
				"delete_relationships",
			}).AddRow(
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
		store, pool := newStore(t)

		pool.ExpectQuery(expectedSQL).WillReturnError(pgx.ErrNoRows)

		_, err := store.GetAnalysisRequest(ctx)
		assert.ErrorIs(t, err, ErrNotFound)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates other database errors", func(t *testing.T) {
		var (
			expectedErr = errors.New("connection refused")
			store, pool = newStore(t)
		)

		pool.ExpectQuery(expectedSQL).WillReturnError(expectedErr)

		_, err := store.GetAnalysisRequest(ctx)
		assert.ErrorIs(t, err, expectedErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestStore_CreateAnalysisRequest(t *testing.T) {
	var (
		ctx               = context.Background()
		requester         = "test-user"
		selectSQL, _, _   = psql.Select(
			sm.Columns(
				"requested_by",
				"request_type",
				"requested_at",
				"delete_all_graph",
				"delete_sourceless_graph",
				"delete_source_kinds",
				"delete_relationships",
			),
			sm.From("analysis_request_switch"),
			sm.Limit(1),
		).Build(ctx)
		insertSQL, _, _ = psql.Insert(
			im.Into("analysis_request_switch",
				"requested_by",
				"request_type",
				"requested_at",
				"delete_all_graph",
				"delete_sourceless_graph",
				"delete_source_kinds",
				"delete_relationships",
			),
			im.Values(psql.Arg(
				"", "", time.Time{}, false, false, []string{}, []string{},
			)),
		).Build(ctx)
	)

	newStore := func(t *testing.T) (*Store, pgxmock.PgxPoolIface) {
		pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		t.Cleanup(pool.Close)
		return &Store{db: pool}, pool
	}

	t.Run("inserts row when no request exists", func(t *testing.T) {
		store, pool := newStore(t)

		pool.ExpectQuery(selectSQL).WillReturnError(pgx.ErrNoRows)
		pool.ExpectExec(insertSQL).
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

		require.NoError(t, store.CreateAnalysisRequest(ctx, requester))
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("no-ops when a request already exists", func(t *testing.T) {
		var (
			store, pool = newStore(t)
			existing    = models.RequestedAnalysis{
				RequestedBy: "other-user",
				RequestType: models.RequestedAnalysisTypeAnalysis,
				RequestedAt: time.Now(),
			}
		)

		pool.ExpectQuery(selectSQL).WillReturnRows(
			pool.NewRows([]string{
				"requested_by", "request_type", "requested_at",
				"delete_all_graph", "delete_sourceless_graph",
				"delete_source_kinds", "delete_relationships",
			}).AddRow(
				existing.RequestedBy, string(existing.RequestType), existing.RequestedAt,
				false, false, []string{}, []string{},
			),
		)
		// No Exec expectation — the insert must NOT be called.

		require.NoError(t, store.CreateAnalysisRequest(ctx, requester))
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates select error", func(t *testing.T) {
		var (
			expectedErr = errors.New("db unavailable")
			store, pool = newStore(t)
		)

		pool.ExpectQuery(selectSQL).WillReturnError(expectedErr)

		err := store.CreateAnalysisRequest(ctx, requester)
		assert.ErrorIs(t, err, expectedErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("propagates insert error", func(t *testing.T) {
		var (
			expectedErr = errors.New("insert failed")
			store, pool = newStore(t)
		)

		pool.ExpectQuery(selectSQL).WillReturnError(pgx.ErrNoRows)
		pool.ExpectExec(insertSQL).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
				pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(expectedErr)

		err := store.CreateAnalysisRequest(ctx, requester)
		assert.ErrorIs(t, err, expectedErr)
		require.NoError(t, pool.ExpectationsWereMet())
	})
}
