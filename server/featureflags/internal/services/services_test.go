// Copyright 2026 Specter Ops, Inc.
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
package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/specterops/bloodhound/server/featureflags/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serviceMocks bundles the generated mocks the Service depends on so tests can
// construct a service and set expectations through a single value.
type serviceMocks struct {
	database *mocks.MockDatabase
	analysis *mocks.MockAnalysisRequestSubmitter
}

// newServiceUnderTest builds a Service backed by freshly-created mocks, returning
// both so callers can wire expectations on the mocks and exercise the service.
func newServiceUnderTest(t *testing.T) (*services.Service, serviceMocks) {
	t.Helper()
	var (
		m = serviceMocks{
			database: mocks.NewMockDatabase(t),
			analysis: mocks.NewMockAnalysisRequestSubmitter(t),
		}
		svc = services.NewService(m.database, m.analysis)
	)
	return svc, m
}

func TestNewService(t *testing.T) {
	svc, _ := newServiceUnderTest(t)
	assert.NotNil(t, svc)
}

func TestNewService_NilDependenciesPanic(t *testing.T) {
	tests := []struct {
		name      string
		construct func(m serviceMocks)
		wantPanic string
	}{
		{
			name: "nil Database",
			construct: func(m serviceMocks) {
				services.NewService(nil, m.analysis)
			},
			wantPanic: "feature-flag: service requires a non-nil Database",
		},
		{
			name: "nil AnalysisRequestSubmitter",
			construct: func(m serviceMocks) {
				services.NewService(m.database, nil)
			},
			wantPanic: "feature-flag: service requires a non-nil AnalysisRequestSubmitter",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := serviceMocks{
				database: mocks.NewMockDatabase(t),
				analysis: mocks.NewMockAnalysisRequestSubmitter(t),
			}
			assert.PanicsWithValue(t, test.wantPanic, func() {
				test.construct(m)
			})
		})
	}
}

func TestService_GetFlagByKey(t *testing.T) {
	var (
		ctx  = context.Background()
		want = services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Enabled: true}
	)

	tests := []struct {
		name       string
		setupMocks func(m serviceMocks)
		wantFlag   services.FeatureFlag
		wantErr    error
	}{
		{
			name: "returns the flag from the database",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(ctx, services.FeatureOpenHoundSupport).Return(want, nil)
			},
			wantFlag: want,
		},
		{
			name: "propagates the database error",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(ctx, services.FeatureOpenHoundSupport).Return(services.FeatureFlag{}, services.ErrNotFound)
			},
			wantErr: services.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, m := newServiceUnderTest(t)
			test.setupMocks(m)

			got, err := svc.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantFlag, got)
			}
		})
	}
}

func TestService_IsEnabled(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name       string
		setupMocks func(m serviceMocks)
		want       bool
		wantErr    error
	}{
		{
			name: "true when the flag is enabled",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(ctx, services.FeatureOpenHoundSupport).Return(services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: true}, nil)
			},
			want: true,
		},
		{
			name: "false when the flag is disabled",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(ctx, services.FeatureOpenHoundSupport).Return(services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: false}, nil)
			},
			want: false,
		},
		{
			name: "propagates database errors",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(ctx, services.FeatureOpenHoundSupport).Return(services.FeatureFlag{}, dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, m := newServiceUnderTest(t)
			test.setupMocks(m)

			got, err := svc.IsEnabled(ctx, services.FeatureOpenHoundSupport)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestService_GetAllFlags(t *testing.T) {
	var (
		ctx           = context.Background()
		unexpectedErr = errors.New("connection refused")
		expected      = []services.FeatureFlag{
			{ID: 1, Key: services.FeatureOpenHoundSupport, Enabled: true, UserUpdatable: true},
			{ID: 2, Key: services.FeatureAlerts, Enabled: false, UserUpdatable: false},
		}
	)

	tests := []struct {
		name       string
		setupMocks func(m serviceMocks)
		wantResult []services.FeatureFlag
		wantErr    error
	}{
		{
			name: "returns all flags on success",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetAllFlags(ctx).Return(expected, nil)
			},
			wantResult: expected,
		},
		{
			name: "propagates database errors",
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetAllFlags(ctx).Return(nil, unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, m := newServiceUnderTest(t)
			test.setupMocks(m)

			got, err := svc.GetAllFlags(ctx)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, got)
			}
		})
	}
}

func TestService_ToggleFlag(t *testing.T) {
	t.Parallel()

	var (
		ctx                = context.Background()
		unexpectedErr      = errors.New("connection refused")
		setFlagErr         = errors.New("set flag failed")
		rollbackErr        = errors.New("rollback set flag failed")
		requestAnalysisErr = errors.New("request analysis failed")
		updatableFlag      = services.FeatureFlag{
			ID:            7,
			Key:           services.FeatureOpenHoundSupport,
			Enabled:       false,
			UserUpdatable: true,
		}
		nonUpdatableFlag = services.FeatureFlag{
			ID:            8,
			Key:           services.FeatureAlerts,
			Enabled:       true,
			UserUpdatable: false,
		}
		findingsPrioritizationFlag = services.FeatureFlag{
			ID:            9,
			Key:           services.FeatureFindingsPrioritizationV0,
			Enabled:       false,
			UserUpdatable: true,
		}
	)

	type testCase struct {
		name       string
		featureID  int32
		setupMocks func(m serviceMocks)
		assert     func(t *testing.T, got services.FeatureFlag, err error)
	}

	toggledUpdatableFlag := updatableFlag
	toggledUpdatableFlag.Enabled = true

	enabledFindingsPrioritizationFlag := findingsPrioritizationFlag
	enabledFindingsPrioritizationFlag.Enabled = true

	testCases := []testCase{
		{
			name:      "toggles the flag and returns the updated value",
			featureID: updatableFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, updatableFlag.ID).Return(updatableFlag, nil)
				m.database.EXPECT().SetFlag(ctx, toggledUpdatableFlag).Return(nil)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				require.NoError(t, err)
				assert.Equal(t, toggledUpdatableFlag, got)
			},
		},
		{
			name:      "requests no-post-processing analysis when findings prioritization is enabled",
			featureID: findingsPrioritizationFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				m.analysis.EXPECT().SubmitAnalysisRequest(ctx, services.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(nil)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				require.NoError(t, err)
				assert.Equal(t, enabledFindingsPrioritizationFlag, got)
			},
		},
		{
			name:      "does not request analysis when findings prioritization is disabled",
			featureID: enabledFindingsPrioritizationFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, enabledFindingsPrioritizationFlag.ID).Return(enabledFindingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(nil)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				require.NoError(t, err)
				assert.Equal(t, findingsPrioritizationFlag, got)
			},
		},
		{
			name:      "returns ErrNotUserUpdatable when the flag is not user updatable",
			featureID: nonUpdatableFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, nonUpdatableFlag.ID).Return(nonUpdatableFlag, nil)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, services.ErrNotUserUpdatable)
				assert.Equal(t, nonUpdatableFlag, got)
			},
		},
		{
			name:      "propagates errors from GetFlagByID",
			featureID: 99,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, int32(99)).Return(services.FeatureFlag{}, unexpectedErr)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, unexpectedErr)
				assert.Equal(t, services.FeatureFlag{}, got)
			},
		},
		{
			name:      "propagates errors from SetFlag",
			featureID: updatableFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, updatableFlag.ID).Return(updatableFlag, nil)
				m.database.EXPECT().SetFlag(ctx, toggledUpdatableFlag).Return(setFlagErr)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, setFlagErr)
				assert.Equal(t, toggledUpdatableFlag, got)
			},
		},
		{
			name:      "propagates errors from SubmitAnalysisRequest",
			featureID: findingsPrioritizationFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				m.analysis.EXPECT().SubmitAnalysisRequest(ctx, services.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(requestAnalysisErr)
				m.database.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(nil)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, requestAnalysisErr)
				assert.Equal(t, findingsPrioritizationFlag, got)
			},
		},
		{
			name:      "propagates rollback errors from SubmitAnalysisRequest failure",
			featureID: findingsPrioritizationFlag.ID,
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				m.analysis.EXPECT().SubmitAnalysisRequest(ctx, services.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(requestAnalysisErr)
				m.database.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(rollbackErr)
			},
			assert: func(t *testing.T, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, requestAnalysisErr)
				assert.ErrorIs(t, err, rollbackErr)
				assert.Equal(t, findingsPrioritizationFlag, got)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc, m := newServiceUnderTest(t)
			testCase.setupMocks(m)

			got, err := svc.ToggleFlag(ctx, testCase.featureID)
			testCase.assert(t, got, err)
		})
	}
}
