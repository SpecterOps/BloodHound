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

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/specterops/bloodhound/server/appcfg/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dbErr = errors.New("database error")
)

func TestService_GetDatapipeStatus(t *testing.T) {
	var (
		ctx         = context.Background()
		updatedAt   = time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
		completedAt = time.Date(2026, 6, 18, 11, 0, 0, 0, time.UTC)
		startedAt   = time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		optimizedAt = time.Date(2026, 6, 18, 9, 30, 0, 0, time.UTC)
		nextRun     = null.TimeFrom(time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC))
		expected    = services.DatapipeStatus{
			Status:                  services.DatapipeStatusIdle,
			UpdatedAt:               updatedAt,
			LastCompleteAnalysisAt:  completedAt,
			LastAnalysisRunAt:       startedAt,
			LastCompleteOptimizeAt:  optimizedAt,
			NextScheduledAnalysisAt: nextRun,
		}
	)

	tests := []struct {
		name       string
		setupMock  func(*mocks.MockDatabase)
		wantResult services.DatapipeStatus
		wantErr    error
	}{
		{
			name: "returns datapipe status on success",
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetDatapipeStatus", ctx).Return(expected, nil)
			},
			wantResult: expected,
		},
		{
			name: "returns ErrNotFound when database returns ErrNotFound",
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetDatapipeStatus", ctx).Return(services.DatapipeStatus{}, services.ErrNotFound)
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "propagates database errors",
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetDatapipeStatus", ctx).Return(services.DatapipeStatus{}, dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDatabase(t)
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

func setUpMockWithParam(t *testing.T, ctx context.Context, paramKey services.ParameterKey, value any) *services.Service {
	t.Helper()
	object, err := types.NewJSONBObject(value)
	require.Nil(t, err)
	parameter := services.Parameter{
		Key:         paramKey,
		Name:        "",
		Description: "",
		Value:       object,
	}
	mockDB := mocks.NewMockDatabase(t)
	mockDB.EXPECT().GetConfigurationParameter(ctx, paramKey).Return(parameter, error(nil))
	return services.NewService(mockDB)
}

func TestService_GetConfig(t *testing.T) {
	t.Run("test GraphStorageOptimization - basic", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		service := setUpMockWithParam(t, ctx, services.GraphStorageOptimizationKey, map[string]any{
			"after_boot":           true,
			"after_analysis":       true,
			"min_interval_seconds": 8000,
		})
		parameterVal, err := services.GetConfig[services.POCGraphStorageOptimizationParam](ctx, service, services.GraphStorageOptimizationKey)

		require.NoError(t, err)
		assert.True(t, parameterVal.AfterBoot)
		assert.True(t, parameterVal.AfterAnalysis)
		assert.Equal(t, 8000, parameterVal.MinIntervalSeconds)
	})

	t.Run("test GraphStorageOptimization - default for invalid value", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		service := setUpMockWithParam(t, ctx, services.GraphStorageOptimizationKey, map[string]any{
			"after_boot":           true,
			"after_analysis":       true,
			"min_interval_seconds": -10,
		})
		parameterVal, err := services.GetConfig[services.POCGraphStorageOptimizationParam](ctx, service, services.GraphStorageOptimizationKey)

		require.NoError(t, err)
		assert.True(t, parameterVal.AfterBoot)
		assert.True(t, parameterVal.AfterAnalysis)
		assert.Equal(t, 86400, parameterVal.MinIntervalSeconds)
	})

	t.Run("test GraphStorageOptimization - default for error", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		mockDB := mocks.NewMockDatabase(t)
		mockDB.EXPECT().GetConfigurationParameter(ctx, services.GraphStorageOptimizationKey).Return(services.Parameter{}, dbErr)
		service := services.NewService(mockDB)

		parameterVal, err := services.GetConfig[services.POCGraphStorageOptimizationParam](ctx, service, services.GraphStorageOptimizationKey)

		var getConfigError = services.GetConfigError{}
		require.Error(t, err)
		require.ErrorAs(t, err, &getConfigError)
		assert.True(t, getConfigError.AppliedDefault)
		assert.Equal(t, services.GraphStorageOptimizationKey, getConfigError.ParameterKey)
		assert.False(t, parameterVal.AfterBoot)
		assert.False(t, parameterVal.AfterAnalysis)
		assert.Equal(t, 86400, parameterVal.MinIntervalSeconds)
	})

	t.Run("test SupportAccountProvisioning - converts ISO duration into duration", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		service := setUpMockWithParam(t, ctx, services.SupportAccountProvisioningKey, map[string]any{
			"enabled":     true,
			"session_ttl": "P10D",
		})
		parameterVal, err := services.GetConfig[services.POCSupportAccountProvisioningParam](ctx, service, services.SupportAccountProvisioningKey)

		require.NoError(t, err)
		assert.True(t, parameterVal.Enabled)
		assert.Equal(t, time.Hour*24*10, time.Duration(parameterVal.SessionTTL))
	})

	t.Run("test type mismatch failure", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		mockDB := mocks.NewMockDatabase(t)
		service := services.NewService(mockDB)
		_, err := services.GetConfig[services.POCSupportAccountProvisioningParam](ctx, service, services.GraphStorageOptimizationKey)

		require.Error(t, err)
		var getConfigError = services.GetConfigError{}
		require.ErrorAs(t, err, &getConfigError)
		assert.False(t, getConfigError.AppliedDefault)
		assert.Equal(t, services.GraphStorageOptimizationKey, getConfigError.ParameterKey)
	})
}
