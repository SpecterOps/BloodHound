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
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/packages/go/duration"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/specterops/bloodhound/server/appcfg/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ErrDB = errors.New("database error")
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
				mockDB.On("GetDatapipeStatus", ctx).Return(services.DatapipeStatus{}, ErrDB)
			},
			wantErr: ErrDB,
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

func TestService_ValidateValue(t *testing.T) {
	mockDB := mocks.NewMockDatabase(t)
	svc := services.NewService(mockDB)

	// specific validator tag rules are already tested elsewhere
	var cases = []struct {
		name            string
		parameterKey    services.ParameterKey
		input           string
		expectError     bool
		expectErrorLike string
	}{
		{
			name:         "happy: input matches type",
			parameterKey: services.APITokenExpiration,
			input:        `{"enabled":true, "expiration_period":200}`,
			expectError:  false,
		},
		{
			name:         "happy: input matches type with a different parameter",
			parameterKey: services.ClientMetricsKey,
			input:        `{"enabled":true}`,
			expectError:  false,
		},
		{
			name:         "happy: duration validation passes",
			parameterKey: services.PruneTTL,
			input:        `{"base_ttl": "P14D","has_session_edge_ttl": "P3D"}`,
			expectError:  false,
		},
		{
			name:            "sad: duration validation fails",
			parameterKey:    services.PruneTTL,
			input:           `{"base_ttl": "P14D","has_session_edge_ttl": "P100D"}`,
			expectError:     true,
			expectErrorLike: "HasSessionEdgeTTL: must be <= P7D",
		},
		{
			name:            "sad: too few fields",
			parameterKey:    services.APITokenExpiration,
			input:           `{"expiration_period":200}`,
			expectError:     true,
			expectErrorLike: "value property contains an invalid field",
		},
		{
			name:            "sad: too many fields",
			parameterKey:    services.APITokenExpiration,
			input:           `{"enabled":true, "expiration_period":200, "blah":"blah"}`,
			expectError:     true,
			expectErrorLike: "value property contains an invalid field",
		},
		{
			name:            "sad: value out of range",
			parameterKey:    services.APITokenExpiration,
			input:           `{"enabled":true, "expiration_period":500}`,
			expectError:     true,
			expectErrorLike: "ExpirationPeriod: must be <= 365",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			var object types.JSONBObject
			err := json.Unmarshal([]byte(cc.input), &object)
			require.NoError(t, err)

			errs := svc.ValidateValue(cc.parameterKey, object)
			if !cc.expectError {
				assert.Nil(t, errs)
			} else {
				assert.Contains(t, errs.Error(), cc.expectErrorLike)
			}
		})
	}
}

func TestService_GetConfig(t *testing.T) {
	// Consider testing actual param configurations in future rather than fixtures
	// (once those exist), and making param definition types and vars private
	// I tend to think exhaustively testing each param will not be worthwhile; but rather
	// testing one use-case of all supported GetParam functionality (defaults,
	// custom normalization, etc.)

	var (
		fixtureKey1 = services.ParameterKey("testKey1")
		fixtureKey2 = services.ParameterKey("testKey2")
	)

	type fixtureParam1 struct {
		AfterBoot          bool `json:"after_boot"`
		MinIntervalSeconds int  `json:"min_interval_seconds"`
	}

	type fixtureParam2 struct {
		Enabled    bool                 `json:"enabled,omitempty"`
		SessionTTL duration.ISODuration `json:"session_ttl,omitempty"`
	}

	originalDefinitions := services.ParamTypeDefinitions
	defer func() { services.ParamTypeDefinitions = originalDefinitions }()

	services.ParamTypeDefinitions = map[services.ParameterKey]services.ParamTypeDefinition{
		services.ParameterKey(fixtureKey1): {
			AllowAPIAccess: false,
			HydrationRules: services.ParamTypeHydrationRules[fixtureParam1]{
				Default: fixtureParam1{
					AfterBoot:          false,
					MinIntervalSeconds: 86400,
				},
				Normalize: func(g *fixtureParam1) {
					if g.MinIntervalSeconds < 0 {
						g.MinIntervalSeconds = 86400
					}
				},
			},
		},
		services.ParameterKey(fixtureKey2): {
			AllowAPIAccess: true,
			HydrationRules: services.ParamTypeHydrationRules[fixtureParam2]{
				Default: fixtureParam2{
					Enabled:    true,
					SessionTTL: duration.ISODuration(time.Hour * 2),
				},
			},
		},
	}

	t.Run("test happy path", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		service := setUpMockWithParam(t, ctx, fixtureKey1, map[string]any{
			"after_boot":           true,
			"min_interval_seconds": 8000,
		})
		parameterVal, err := service.GetConfig[fixtureParam1](ctx, fixtureKey1)

		require.NoError(t, err)
		assert.True(t, parameterVal.AfterBoot)
		assert.Equal(t, 8000, parameterVal.MinIntervalSeconds)
	})

	t.Run("default for invalid value", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		service := setUpMockWithParam(t, ctx, fixtureKey1, map[string]any{
			"after_boot":           true,
			"after_analysis":       true,
			"min_interval_seconds": -10,
		})
		parameterVal, err := service.GetConfig[fixtureParam1](ctx, fixtureKey1)

		require.NoError(t, err)
		assert.True(t, parameterVal.AfterBoot)
		assert.Equal(t, 86400, parameterVal.MinIntervalSeconds)
	})

	t.Run("default for error", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		mockDB := mocks.NewMockDatabase(t)
		mockDB.EXPECT().GetConfigurationParameter(ctx, fixtureKey1).Return(services.Parameter{}, ErrDB)
		service := services.NewService(mockDB)

		parameterVal, err := service.GetConfig[fixtureParam1](ctx, fixtureKey1)

		var getConfigError = services.GetConfigError{}
		require.Error(t, err)
		require.ErrorAs(t, err, &getConfigError)
		assert.True(t, getConfigError.AppliedDefault)
		assert.Equal(t, fixtureKey1, getConfigError.ParameterKey)
		assert.False(t, parameterVal.AfterBoot)
		assert.Equal(t, 86400, parameterVal.MinIntervalSeconds)
	})

	t.Run("converts ISO duration into duration", func(t *testing.T) {
		var (
			ctx = context.Background()
		)
		service := setUpMockWithParam(t, ctx, fixtureKey2, map[string]any{
			"enabled":     true,
			"session_ttl": "P10D",
		})
		parameterVal, err := service.GetConfig[fixtureParam2](ctx, fixtureKey2)

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
		_, err := service.GetConfig[fixtureParam2](ctx, fixtureKey1)

		require.Error(t, err)
		var getConfigError = services.GetConfigError{}
		require.ErrorAs(t, err, &getConfigError)
		assert.False(t, getConfigError.AppliedDefault)
		assert.Equal(t, fixtureKey1, getConfigError.ParameterKey)
	})
}

func TestService_IsAPIAllowedKey(t *testing.T) {
	// This function is quite trivial. Just sanity testing one true and false case.

	var (
		mockDB  = mocks.NewMockDatabase(t)
		service = services.NewService(mockDB)
	)

	t.Run("true case", func(t *testing.T) {
		assert.True(t, service.IsAPIAllowedKey(services.PasswordExpirationWindow))
	})

	t.Run("false case", func(t *testing.T) {
		assert.False(t, service.IsAPIAllowedKey(services.TrustedProxiesConfig))
	})
}
