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

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/appdb"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL strings expected by the Store. These are compared via
// pgxmock.QueryMatcherEqual, which whitespace-normalises both sides.
const (
	expectedGetDatapipeStatusSQL = `SELECT status, updated_at, last_complete_analysis_at, last_analysis_run_at, next_scheduled_analysis_at FROM datapipe_status LIMIT $1`
)

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func datapipeStatusRowColumns() []string {
	return []string{
		"status", "updated_at", "last_complete_analysis_at",
		"last_analysis_run_at", "next_scheduled_analysis_at",
	}
}

func TestStore_GetDatapipeStatus(t *testing.T) {
	var (
		ctx         = context.Background()
		dbErr       = errors.New("connection refused")
		updatedAt   = time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
		completedAt = null.TimeFrom(time.Date(2026, 6, 18, 11, 0, 0, 0, time.UTC))
		startedAt   = null.TimeFrom(time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC))
		nextRun     = null.TimeFrom(time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC))
		expected    = services.DatapipeStatus{
			Status:                  services.DatapipeStatusIdle,
			UpdatedAt:               updatedAt,
			LastCompleteAnalysisAt:  completedAt,
			LastAnalysisRunAt:       startedAt,
			NextScheduledAnalysisAt: nextRun,
		}
	)

	tests := []struct {
		name            string
		expectations    func(pool pgxmock.PgxPoolIface)
		wantResult      services.DatapipeStatus
		wantErr         error
		wantErrContains string
	}{
		{
			name: "returns the datapipe status on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetDatapipeStatusSQL).WithArgs(1).WillReturnRows(
					pool.NewRows(datapipeStatusRowColumns()).AddRow(
						string(expected.Status),
						expected.UpdatedAt,
						expected.LastCompleteAnalysisAt,
						expected.LastAnalysisRunAt,
						expected.NextScheduledAnalysisAt,
					),
				)
			},
			wantResult: expected,
		},
		{
			name: "returns ErrNotFound when no row exists",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetDatapipeStatusSQL).WithArgs(1).WillReturnRows(
					pool.NewRows(datapipeStatusRowColumns()),
				)
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "wraps database errors when query fails",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetDatapipeStatusSQL).WithArgs(1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name: "wraps scanning errors when column type mismatch occurs",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetDatapipeStatusSQL).WithArgs(1).WillReturnRows(
					pool.NewRows([]string{"status"}).AddRow(123), // missing required columns
				)
			},
			wantErrContains: "reading rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			result, err := store.GetDatapipeStatus(ctx)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.wantErrContains != "" {
				assert.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
