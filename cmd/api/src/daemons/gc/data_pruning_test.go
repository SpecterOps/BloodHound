// Copyright 2023 Specter Ops, Inc.
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

package gc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// fakeAuditMaintainer is a hand-rolled audit.Maintainer used to assert the
// daemon drives partition maintenance. It records the calls and the
// retentionMonths it was invoked with.
type fakeAuditMaintainer struct {
	mu              sync.Mutex
	preCreateCalls  int
	dropCalls       int
	retentionMonths int
}

func (s *fakeAuditMaintainer) PreCreateNextPartition(_ context.Context, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preCreateCalls++
	return nil
}

func (s *fakeAuditMaintainer) DropExpiredPartitions(_ context.Context, _ time.Time, retentionMonths int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropCalls++
	s.retentionMonths = retentionMonths
	return nil
}

func TestGC_NewDataPruningDaemon(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	daemon := NewDataPruningDaemon(mocks.NewMockDatabase(mockCtrl), nil)
	require.NotNil(t, daemon)
}

func TestGC_Name(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	daemon := NewDataPruningDaemon(mocks.NewMockDatabase(mockCtrl), nil)
	require.NotNil(t, daemon)

	result := daemon.Name()
	require.Equal(t, "Data Pruning Daemon", result)
}

func TestGC_Start(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocks.NewMockDatabase(mockCtrl)

	mockDB.EXPECT().SweepSessions(gomock.Any()).Do(func(ctx context.Context) {
		// simulate some work being done
		time.Sleep(1 * time.Millisecond)
	})
	mockDB.EXPECT().SweepAssetGroupCollections(gomock.Any()).Do(func(ctx context.Context) {
		time.Sleep(1 * time.Millisecond)
	})

	daemon := NewDataPruningDaemon(mockDB, nil)
	require.NotNil(t, daemon)

	go func() {
		// simulate the daemon running for 1 second and then quitting
		time.Sleep(1 * time.Second)
		daemon.exitC <- struct{}{}
	}()

	daemon.Start(context.Background())
}

func TestGC_SweepAuditPartitions(t *testing.T) {
	t.Run("nil maintainer is a no-op", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		daemon := NewDataPruningDaemon(mocks.NewMockDatabase(mockCtrl), nil)
		require.NotPanics(t, func() {
			daemon.sweepAuditPartitions(context.Background())
		})
	})

	t.Run("drives partition maintenance with configured maintainer", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		maintainer := &fakeAuditMaintainer{}
		daemon := NewDataPruningDaemon(mocks.NewMockDatabase(mockCtrl), maintainer)

		daemon.sweepAuditPartitions(context.Background())

		require.Equal(t, 1, maintainer.preCreateCalls)
		require.Equal(t, 1, maintainer.dropCalls)
		require.Equal(t, defaultAuditRetentionMonths, maintainer.retentionMonths)
	})
}
