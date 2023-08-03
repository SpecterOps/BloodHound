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

package database_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/database"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/log/mocks"

	"github.com/specterops/bloodhound/log"
)

func TestGormLogAdapter_Info(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockEvent      = mocks.NewMockEvent(mockCtrl)
		gormLogAdapter = database.GormLogAdapter{
			SlowQueryWarnThreshold:  time.Minute,
			SlowQueryErrorThreshold: time.Minute,
		}
	)

	log.ConfigureDefaults()

	mockEvent.EXPECT().Msgf("message %d %s %f", 1, "arg", 2.0).Times(1)
	gormLogAdapter.Log(mockEvent, "message %d %s %f", 1, "arg", 2.0)
}
