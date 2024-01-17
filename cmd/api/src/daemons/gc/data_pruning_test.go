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
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGC_NewDataPruningDaemon(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	daemon := NewDataPruningDaemon(mocks.NewMockDatabase(mockCtrl))
	require.NotNil(t, daemon)
}

func TestGC_Name(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	daemon := NewDataPruningDaemon(mocks.NewMockDatabase(mockCtrl))
	require.NotNil(t, daemon)

	result := daemon.Name()
	require.Equal(t, "Data Pruning Daemon", result)
}

func TestGC_Start(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocks.NewMockDatabase(mockCtrl)

	mockDB.EXPECT().SweepSessions().Do(func() {
		// simulate some work being done
		time.Sleep(1 * time.Millisecond)
	})
	mockDB.EXPECT().SweepAssetGroupCollections().Do(func() {
		time.Sleep(1 * time.Millisecond)
	})

	daemon := NewDataPruningDaemon(mockDB)
	require.NotNil(t, daemon)

	go func() {
		// simulate the daemon running for 1 second and then quitting
		time.Sleep(1 * time.Second)
		daemon.exitC <- struct{}{}
	}()

	daemon.Start()
}
