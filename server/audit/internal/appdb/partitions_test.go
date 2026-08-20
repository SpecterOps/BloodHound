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
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingQuerier is a pgxQuerier that captures the SQL and args passed to Exec
// so the emitted DDL/DML can be asserted without a live database. execErr, when
// set, is returned from Exec to exercise error-mapping paths. Query/QueryRow are
// unused by the audit Store and return zero values.
type recordingQuerier struct {
	execSQL  []string
	execArgs [][]any
	execErr  error
}

func (s *recordingQuerier) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, nil
}

func (s *recordingQuerier) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return nil
}

func (s *recordingQuerier) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	s.execSQL = append(s.execSQL, sql)
	s.execArgs = append(s.execArgs, args)
	return pgconn.CommandTag{}, s.execErr
}

func TestPartitionName(t *testing.T) {
	var cases = []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"january", time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC), "audit_logs_2024_01"},
		{"december", time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC), "audit_logs_2026_12"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, partitionName(tc.input))
		})
	}
}

func TestFirstOfMonth(t *testing.T) {
	var cases = []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "mid month normalizes to first at midnight utc",
			input:    time.Date(2026, time.March, 15, 13, 45, 12, 0, time.UTC),
			expected: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "non-utc zone is converted to utc first",
			input:    time.Date(2026, time.March, 31, 23, 0, 0, 0, time.FixedZone("east", 2*60*60)),
			expected: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, firstOfMonth(tc.input))
		})
	}
}

func TestStore_PreCreateNextPartition(t *testing.T) {
	var cases = []struct {
		name        string
		asOf        time.Time
		expectedSQL string
	}{
		{
			name:        "mid month creates next month partition",
			asOf:        time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedSQL: `CREATE TABLE IF NOT EXISTS audit_logs_2026_02 PARTITION OF audit_logs FOR VALUES FROM ('2026-02-01') TO ('2026-03-01')`,
		},
		{
			name:        "december rolls over into next year",
			asOf:        time.Date(2026, time.December, 20, 0, 0, 0, 0, time.UTC),
			expectedSQL: `CREATE TABLE IF NOT EXISTS audit_logs_2027_01 PARTITION OF audit_logs FOR VALUES FROM ('2027-01-01') TO ('2027-02-01')`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				querier = &recordingQuerier{}
				store   = NewStore(querier)
			)

			require.NoError(t, store.PreCreateNextPartition(context.Background(), tc.asOf))
			require.Len(t, querier.execSQL, 1)
			assert.Equal(t, tc.expectedSQL, querier.execSQL[0])
		})
	}
}

func TestStore_DropExpiredPartitions(t *testing.T) {
	var cases = []struct {
		name            string
		asOf            time.Time
		retentionMonths int
		// expectedDrops are the partition names expected to be dropped, in order.
		expectedDrops []string
	}{
		{
			// cutoff = firstOfMonth(2024-06) - 3mo = 2024-03-01, so only months
			// strictly before March 2024 are dropped, starting from
			// earliestPartitionMonth (2024-01). March 2024 is retained.
			name:            "drops only months before the retention cutoff",
			asOf:            time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC),
			retentionMonths: 3,
			expectedDrops: []string{
				"DROP TABLE IF EXISTS audit_logs_2024_01",
				"DROP TABLE IF EXISTS audit_logs_2024_02",
			},
		},
		{
			// cutoff = earliestPartitionMonth itself, so nothing is expired.
			name:            "nothing dropped when cutoff is at earliest partition",
			asOf:            time.Date(2024, time.April, 10, 0, 0, 0, 0, time.UTC),
			retentionMonths: 3,
			expectedDrops:   nil,
		},
		{
			// cutoff before earliestPartitionMonth: loop never runs.
			name:            "nothing dropped when cutoff precedes earliest partition",
			asOf:            time.Date(2024, time.January, 5, 0, 0, 0, 0, time.UTC),
			retentionMonths: 3,
			expectedDrops:   nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				querier = &recordingQuerier{}
				store   = NewStore(querier)
			)

			require.NoError(t, store.DropExpiredPartitions(context.Background(), tc.asOf, tc.retentionMonths))
			assert.Equal(t, tc.expectedDrops, querier.execSQL)
		})
	}
}

// TestStore_DropExpiredPartitions_Idempotent verifies that a second sweep with
// the same inputs issues the same idempotent DROP TABLE IF EXISTS statements,
// mirroring the daemon calling the sweep on every tick.
func TestStore_DropExpiredPartitions_Idempotent(t *testing.T) {
	var (
		querier = &recordingQuerier{}
		store   = NewStore(querier)
		asOf    = time.Date(2024, time.April, 15, 0, 0, 0, 0, time.UTC)
	)

	require.NoError(t, store.DropExpiredPartitions(context.Background(), asOf, 3))
	first := append([]string(nil), querier.execSQL...)

	querier.execSQL = nil
	require.NoError(t, store.DropExpiredPartitions(context.Background(), asOf, 3))

	assert.Equal(t, first, querier.execSQL)
}
