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
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/analysis/internal/appdb"
	"github.com/specterops/bloodhound/server/analysis/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL strings expected by the Store. These are compared via
// pgxmock.QueryMatcherEqual, which whitespace-normalises both sides, so
// formatting (tabs, newlines) is not load-bearing — token order, table
// name, column names, parameter shape and the ON CONFLICT clause are.
const (
	expectedSelectSQL = `SELECT requested_by, request_type, requested_at, delete_all_graph, delete_sourceless_graph, delete_source_kinds, delete_relationships FROM analysis_request_switch LIMIT $1`

	expectedInsertSQL = `INSERT INTO analysis_request_switch (requested_by, request_type, requested_at, delete_all_graph, delete_sourceless_graph, delete_source_kinds, delete_relationships) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (singleton) DO NOTHING`

	expectedDeleteSQL = `DELETE FROM analysis_request_switch`
)

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func analysisRequestRowColumns() []string {
	return []string{
		"requested_by", "request_type", "requested_at",
		"delete_all_graph", "delete_sourceless_graph",
		"delete_source_kinds", "delete_relationships",
	}
}

func TestStore_GetAnalysisRequest(t *testing.T) {
	var (
		ctx      = context.Background()
		dbErr    = errors.New("connection refused")
		expected = services.RequestedAnalysis{
			RequestedBy:           "test-user",
			RequestType:           services.RequestedAnalysisTypeAnalysis,
			RequestedAt:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DeleteAllGraph:        true,
			DeleteSourcelessGraph: false,
			DeleteSourceKinds:     []string{"AZBase"},
			DeleteRelationships:   []string{"HasSession"},
		}
	)

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      services.RequestedAnalysis
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns the analysis request on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(
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
			},
			wantResult: expected,
		},
		{
			name: "maps CollectOneRow pgx.ErrNoRows to services.ErrNoPendingRequest",
			expectations: func(pool pgxmock.PgxPoolIface) {
				// Query succeeds but returns zero rows; CollectOneRow sees no data and returns pgx.ErrNoRows
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(
					pool.NewRows(analysisRequestRowColumns()),
				)
			},
			wantErr: services.ErrNoPendingRequest,
		},
		{
			name: "wraps CollectOneRow iteration error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				// Query succeeds but the rows object carries a close error that
				// pgx.CollectOneRow surfaces via rows.Err() when Next() returns false.
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(
					pool.NewRows(analysisRequestRowColumns()).CloseError(errors.New("forced iteration error")),
				)
			},
			wantErrContains: "reading rows:",
		},
		{
			name: "propagates other database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.GetAnalysisRequest(ctx)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_CreateAnalysisRequest(t *testing.T) {
	var (
		ctx           = context.Background()
		requester     = "test-user"
		otherUser     = "other-user"
		beginTxErr    = errors.New("begin tx failed")
		insertErr     = errors.New("insert failed")
		postInsertErr = errors.New("connection lost")
		// time.Now().UTC() is called inside CreateAnalysisRequest so we can't
		// pin the exact value from outside — AnyArg is the right tool here.
		insertArgs = []any{
			requester,
			string(services.RequestedAnalysisTypeAnalysis),
			pgxmock.AnyArg(), // now
			false,
			false,
			[]string{},
			[]string{},
		}
		existingRow = func(pool pgxmock.PgxPoolIface) *pgxmock.Rows {
			return pool.NewRows(analysisRequestRowColumns()).AddRow(
				otherUser, string(services.RequestedAnalysisTypeAnalysis), time.Now().UTC(),
				false, false, []string{}, []string{},
			)
		}
		createdRow = func(pool pgxmock.PgxPoolIface) *pgxmock.Rows {
			return pool.NewRows(analysisRequestRowColumns()).AddRow(
				requester, string(services.RequestedAnalysisTypeAnalysis), time.Now().UTC(),
				false, false, []string{}, []string{},
			)
		}
	)

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantCreated     bool
		wantRequester   string
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns created=true and the new request when no row exists",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectExec(expectedInsertSQL).WithArgs(insertArgs...).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(createdRow(pool))
				pool.ExpectCommit()
			},
			wantCreated:   true,
			wantRequester: requester,
		},
		{
			name: "returns created=false and the existing request when a row already exists",
			// ON CONFLICT DO NOTHING reports 0 rows affected when the singleton
			// row already exists — the INSERT is still issued, the conflict is
			// silently resolved at the DB level.
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectExec(expectedInsertSQL).WithArgs(insertArgs...).
					WillReturnResult(pgxmock.NewResult("INSERT", 0))
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(existingRow(pool))
				pool.ExpectCommit()
			},
			wantCreated:   false,
			wantRequester: otherUser,
		},
		{
			name: "wraps BeginTx errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{}).WillReturnError(beginTxErr)
			},
			wantErrContains: "beginning transaction:",
		},
		{
			name: "propagates insert errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectExec(expectedInsertSQL).WithArgs(insertArgs...).WillReturnError(insertErr)
				pool.ExpectRollback()
			},
			wantErr: insertErr,
		},
		{
			name: "propagates select errors after a successful insert",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectExec(expectedInsertSQL).WithArgs(insertArgs...).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnError(postInsertErr)
				pool.ExpectRollback()
			},
			wantErr: postInsertErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			current, created, err := store.CreateAnalysisRequest(ctx, requester)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantCreated, created)
				assert.Equal(t, tt.wantRequester, current.RequestedBy)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_DeleteAnalysisRequest(t *testing.T) {
	var (
		ctx         = context.Background()
		beginTxErr  = errors.New("begin tx failed")
		selectErr   = errors.New("select failed")
		deleteErr   = errors.New("delete failed")
		analysisRow = func(pool pgxmock.PgxPoolIface) *pgxmock.Rows {
			return pool.NewRows(analysisRequestRowColumns()).AddRow(
				"test-user", string(services.RequestedAnalysisTypeAnalysis), time.Now().UTC(),
				false, false, []string{}, []string{},
			)
		}
		deletionRow = func(pool pgxmock.PgxPoolIface) *pgxmock.Rows {
			return pool.NewRows(analysisRequestRowColumns()).AddRow(
				"test-user", string(services.RequestedAnalysisTypeDeletion), time.Now().UTC(),
				false, false, []string{}, []string{},
			)
		}
	)

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantErr         error
		wantErrContains string
	}{
		{
			name: "deletes an analysis request successfully",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(analysisRow(pool))
				pool.ExpectExec(expectedDeleteSQL).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				pool.ExpectCommit()
			},
		},
		{
			name: "returns ErrDeletionRequestPending when a deletion request is pending",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(deletionRow(pool))
				pool.ExpectRollback()
			},
			wantErr: services.ErrDeletionRequestPending,
		},
		{
			name: "returns ErrNoPendingRequest when no request exists",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(
					pool.NewRows(analysisRequestRowColumns()),
				)
				pool.ExpectRollback()
			},
			wantErr: services.ErrNoPendingRequest,
		},
		{
			name: "wraps BeginTx errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{}).WillReturnError(beginTxErr)
			},
			wantErrContains: "beginning transaction:",
		},
		{
			name: "propagates SELECT errors within the transaction",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnError(selectErr)
				pool.ExpectRollback()
			},
			wantErr: selectErr,
		},
		{
			name: "wraps DELETE exec errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedSelectSQL).WithArgs(1).WillReturnRows(analysisRow(pool))
				pool.ExpectExec(expectedDeleteSQL).WillReturnError(deleteErr)
				pool.ExpectRollback()
			},
			wantErrContains: "deleting analysis request:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			err := store.DeleteAnalysisRequest(ctx)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
