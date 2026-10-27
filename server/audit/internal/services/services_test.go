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

package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/specterops/bloodhound/server/audit/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ListAuditLogs(t *testing.T) {
	var (
		ctx           = context.Background()
		query         = services.ListQuery{Before: time.Now(), Limit: 1000}
		expectedLogs  = []services.Log{{ID: 1, Action: "CreateUser"}}
		expectedCount = 1
		databaseError = errors.New("database unavailable")
	)

	tests := []struct {
		name          string
		setupMocks    func(database *mocks.MockDatabase)
		expectedLogs  []services.Log
		expectedCount int
		expectedError error
	}{
		{
			name: "Success: returns database results",
			setupMocks: func(database *mocks.MockDatabase) {
				database.EXPECT().ListAuditLogs(ctx, query).Return(expectedLogs, expectedCount, nil)
			},
			expectedLogs: expectedLogs, expectedCount: expectedCount,
		},
		{
			name: "Error: propagates database error",
			setupMocks: func(database *mocks.MockDatabase) {
				database.EXPECT().ListAuditLogs(ctx, query).Return(nil, 0, databaseError)
			},
			expectedError: databaseError,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			database := mocks.NewMockDatabase(t)
			testCase.setupMocks(database)

			logs, count, err := services.NewService(database).ListAuditLogs(ctx, query)

			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedLogs, logs)
			assert.Equal(t, testCase.expectedCount, count)
		})
	}
}
