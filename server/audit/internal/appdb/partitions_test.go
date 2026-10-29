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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_CreateNextPartition(t *testing.T) {
	testCases := []struct {
		name            string
		asOf            time.Time
		expectedQueries []string
		databaseError   error
		expectError     bool
	}{
		{
			name: "creates current and next month",
			asOf: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedQueries: []string{
				"CREATE TABLE IF NOT EXISTS audit_logs_2026_01 PARTITION OF audit_logs FOR VALUES FROM ('2026-01-01') TO ('2026-02-01')",
				"CREATE TABLE IF NOT EXISTS audit_logs_2026_02 PARTITION OF audit_logs FOR VALUES FROM ('2026-02-01') TO ('2026-03-01')",
			},
		},
		{
			name:          "returns database error",
			asOf:          time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			databaseError: errors.New("create failed"),
			expectError:   true,
			expectedQueries: []string{
				"CREATE TABLE IF NOT EXISTS audit_logs_2026_01 PARTITION OF audit_logs FOR VALUES FROM ('2026-01-01') TO ('2026-02-01')",
			},
		},
		{
			name: "crosses year boundary",
			asOf: time.Date(2026, time.December, 15, 0, 0, 0, 0, time.UTC),
			expectedQueries: []string{
				"CREATE TABLE IF NOT EXISTS audit_logs_2026_12 PARTITION OF audit_logs FOR VALUES FROM ('2026-12-01') TO ('2027-01-01')",
				"CREATE TABLE IF NOT EXISTS audit_logs_2027_01 PARTITION OF audit_logs FOR VALUES FROM ('2027-01-01') TO ('2027-02-01')",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				database = &recordingExecer{err: testCase.databaseError}
				store    = NewStore(database)
			)

			err := store.CreateNextPartition(context.Background(), testCase.asOf)
			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Len(t, database.queries, len(testCase.expectedQueries))
			for index, expectedQuery := range testCase.expectedQueries {
				assert.Equal(t, expectedQuery, database.queries[index][0])
			}
		})
	}
}

func TestStore_DropExpiredPartitions(t *testing.T) {
	testCases := []struct {
		name            string
		asOf            time.Time
		retentionMonths int
		expectedQueries []string
		databaseError   error
		expectError     bool
	}{
		{
			name:            "drops months before cutoff",
			asOf:            time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC),
			retentionMonths: 3,
			expectedQueries: []string{"DROP TABLE IF EXISTS audit_logs_2024_01", "DROP TABLE IF EXISTS audit_logs_2024_02"},
		},
		{name: "keeps partitions when cutoff precedes earliest month", asOf: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), retentionMonths: 3},
		{name: "rejects future time", asOf: time.Now().UTC().Add(time.Hour), retentionMonths: 3, expectError: true},
		{name: "rejects invalid retention", asOf: time.Now().UTC(), retentionMonths: 0, expectError: true},
		{
			name:            "returns database error",
			asOf:            time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC),
			retentionMonths: 3,
			databaseError:   errors.New("drop failed"),
			expectError:     true,
			expectedQueries: []string{"DROP TABLE IF EXISTS audit_logs_2024_01"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				database = &recordingExecer{err: testCase.databaseError}
				store    = NewStore(database)
			)

			err := store.DropExpiredPartitions(context.Background(), testCase.asOf, testCase.retentionMonths)
			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Len(t, database.queries, len(testCase.expectedQueries))
			for index, expectedQuery := range testCase.expectedQueries {
				assert.Equal(t, expectedQuery, database.queries[index][0])
			}
		})
	}
}

func TestPartition_PartitionName(t *testing.T) {
	assert.Equal(t, "audit_logs_2026_03", partitionName(time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)))
}

func TestPartition_FirstOfMonth(t *testing.T) {
	var input = time.Date(2026, time.March, 31, 23, 0, 0, 0, time.FixedZone("east", 2*60*60))
	assert.Equal(t, time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC), firstOfMonth(input))
}
