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

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/mocks"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_GetDatapipeStatus(t *testing.T) {
	var (
		ctx         = context.Background()
		dbErr       = errors.New("database error")
		updatedAt   = time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
		completedAt = time.Date(2026, 6, 18, 11, 0, 0, 0, time.UTC)
		startedAt   = time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
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
		name        string
		setupMock   func(*mocks.MockDatabase)
		wantResult  services.DatapipeStatus
		wantErr     error
	}{
		{
			name: "returns datapipe status on success",
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetDatapipeStatus(ctx).Return(expected, nil)
			},
			wantResult: expected,
		},
		{
			name: "returns ErrNotFound when database returns ErrNotFound",
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetDatapipeStatus(ctx).Return(services.DatapipeStatus{}, services.ErrNotFound)
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "propagates database errors",
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetDatapipeStatus(ctx).Return(services.DatapipeStatus{}, dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDatabase(ctrl)
			tt.setupMock(mockDB)

			svc := services.NewService(mockDB)
			result, err := svc.GetDatapipeStatus(ctx)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
