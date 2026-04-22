// Copyright 2025 Specter Ops, Inc.
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
package changelog

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestIngestionCoordinator_FlushBuffer(t *testing.T) {
	t.Parallel()

	var (
		deadlockErr     = &pgconn.PgError{Code: pgDeadlockErrorCode}
		nonRetryableErr = errors.New("connection refused")
		retryConfig     = RetryConfig{MaxRetries: 3, BaseDelay: 0, MaxJitter: 0}
	)

	tests := []struct {
		name   string
		setup  func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator
		assert func(t *testing.T, coordinator *ingestionCoordinator, flushErr error)
	}{
		{
			name: "empty buffer is a no-op",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.NoError(t, flushErr)
				assert.Empty(t, coordinator.buffer)
			},
		},
		{
			name: "succeeds on first attempt writing lastseen property to unchanged nodes/edges",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nil)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				coordinator.buffer = make([]Change, 5)
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.NoError(t, flushErr)
				assert.Empty(t, coordinator.buffer)
			},
		},
		{
			name: "retries on deadlock and succeeds",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				gomock.InOrder(
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(deadlockErr),
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nil),
				)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				coordinator.buffer = make([]Change, 5)
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.NoError(t, flushErr)
				assert.Empty(t, coordinator.buffer)
			},
		},
		{
			name: "retries multiple deadlocks and succeeds",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				gomock.InOrder(
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(deadlockErr),
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(deadlockErr),
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nil),
				)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				coordinator.buffer = make([]Change, 5)
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.NoError(t, flushErr)
				assert.Empty(t, coordinator.buffer)
			},
		},
		{
			name: "all retries exhausted on deadlock",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(deadlockErr).Times(3)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				coordinator.buffer = make([]Change, 5)
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.Error(t, flushErr)
				assert.Contains(t, flushErr.Error(), "40P01")
				assert.Empty(t, coordinator.buffer)
			},
		},
		{
			name: "non-retryable error stops immediately",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nonRetryableErr).Times(1)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				coordinator.buffer = make([]Change, 5)
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.Error(t, flushErr)
				assert.Contains(t, flushErr.Error(), "connection refused")
				assert.Empty(t, coordinator.buffer)
			},
		},
		{
			name: "deadlock then non-retryable error stops retrying",
			setup: func(t *testing.T, mockCtrl *gomock.Controller) *ingestionCoordinator {
				t.Helper()
				mockDB := graphmocks.NewMockDatabase(mockCtrl)
				gomock.InOrder(
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(deadlockErr),
					mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nonRetryableErr),
				)
				coordinator := newIngestionCoordinator(mockDB, retryConfig)
				coordinator.batchSize = 10
				coordinator.buffer = make([]Change, 5)
				return coordinator
			},
			assert: func(t *testing.T, coordinator *ingestionCoordinator, flushErr error) {
				t.Helper()
				assert.Error(t, flushErr)
				assert.Contains(t, flushErr.Error(), "connection refused")
				assert.Empty(t, coordinator.buffer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			coordinator := tt.setup(t, mockCtrl)
			flushErr := coordinator.flushBuffer(context.Background(), true)
			tt.assert(t, coordinator, flushErr)
		})
	}
}

func TestIsDeadlockError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			expected: false,
		},
		{
			name:     "pg deadlock error",
			err:      &pgconn.PgError{Code: "40P01"},
			expected: true,
		},
		{
			name:     "wrapped pg deadlock error",
			err:      fmt.Errorf("batch failed: %w", &pgconn.PgError{Code: "40P01"}),
			expected: true,
		},
		{
			name:     "pg error with different code",
			err:      &pgconn.PgError{Code: "23505"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, isDeadlockError(tt.err))
		})
	}
}
