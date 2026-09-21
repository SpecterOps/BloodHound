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
	t.Parallel()

	svc, _ := newServiceUnderTest(t)
	assert.NotNil(t, svc)
}

func TestNewService_NilDependenciesPanic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		construct func(m serviceMocks)
		wantPanic string
	}{
		{
			name: "Error: nil Database",
			construct: func(m serviceMocks) {
				services.NewService(nil, m.analysis)
			},
			wantPanic: "feature-flag: service requires a non-nil Database",
		},
		{
			name: "Error: nil AnalysisRequestSubmitter",
			construct: func(m serviceMocks) {
				services.NewService(m.database, nil)
			},
			wantPanic: "feature-flag: service requires a non-nil AnalysisRequestSubmitter",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()

	type args struct {
		ctx context.Context
		key string
	}
	type want struct {
		flag services.FeatureFlag
		err  error
	}

	var expectedFlag = services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Enabled: true}

	tests := []struct {
		name       string
		args       args
		setupMocks func(m serviceMocks)
		want       want
	}{
		{
			name: "Success: returns the flag from the database",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(context.Background(), services.FeatureOpenHoundSupport).Return(expectedFlag, nil)
			},
			want: want{flag: expectedFlag},
		},
		{
			name: "Error: propagates the database error",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(context.Background(), services.FeatureOpenHoundSupport).Return(services.FeatureFlag{}, services.ErrNotFound)
			},
			want: want{err: services.ErrNotFound},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			svc, m := newServiceUnderTest(t)
			test.setupMocks(m)

			got, err := svc.GetFlagByKey(test.args.ctx, test.args.key)
			if test.want.err != nil {
				assert.ErrorIs(t, err, test.want.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.flag, got)
			}
		})
	}
}

func TestService_IsEnabled(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		key string
	}
	type want struct {
		enabled bool
		err     error
	}

	var dbErr = errors.New("connection refused")

	tests := []struct {
		name       string
		args       args
		setupMocks func(m serviceMocks)
		want       want
	}{
		{
			name: "Success: returns true when the flag is enabled",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(context.Background(), services.FeatureOpenHoundSupport).Return(services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: true}, nil)
			},
			want: want{enabled: true},
		},
		{
			name: "Success: returns false when the flag is disabled",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(context.Background(), services.FeatureOpenHoundSupport).Return(services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: false}, nil)
			},
			want: want{enabled: false},
		},
		{
			name: "Error: propagates database errors",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByKey(context.Background(), services.FeatureOpenHoundSupport).Return(services.FeatureFlag{}, dbErr)
			},
			want: want{err: dbErr},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			svc, m := newServiceUnderTest(t)
			test.setupMocks(m)

			got, err := svc.IsEnabled(test.args.ctx, test.args.key)
			if test.want.err != nil {
				assert.ErrorIs(t, err, test.want.err)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.enabled, got)
			}
		})
	}
}

func TestService_GetAllFlags(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}
	type want struct {
		flags []services.FeatureFlag
		err   error
	}

	var (
		unexpectedErr = errors.New("connection refused")
		expected      = []services.FeatureFlag{
			{ID: 1, Key: services.FeatureOpenHoundSupport, Enabled: true, UserUpdatable: true},
			{ID: 2, Key: services.FeatureAlerts, Enabled: false, UserUpdatable: false},
		}
	)

	tests := []struct {
		name       string
		args       args
		setupMocks func(m serviceMocks)
		want       want
	}{
		{
			name: "Success: returns all flags",
			args: args{ctx: context.Background()},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetAllFlags(context.Background()).Return(expected, nil)
			},
			want: want{flags: expected},
		},
		{
			name: "Error: propagates database errors",
			args: args{ctx: context.Background()},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetAllFlags(context.Background()).Return(nil, unexpectedErr)
			},
			want: want{err: unexpectedErr},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			svc, m := newServiceUnderTest(t)
			test.setupMocks(m)

			got, err := svc.GetAllFlags(test.args.ctx)
			if test.want.err != nil {
				assert.ErrorIs(t, err, test.want.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.flags, got)
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

	type args struct {
		ctx       context.Context
		featureID int32
	}
	type want struct {
		flag        services.FeatureFlag
		err         error
		rollbackErr error
	}
	type testCase struct {
		name       string
		args       args
		setupMocks func(m serviceMocks)
		want       want
	}

	toggledUpdatableFlag := updatableFlag
	toggledUpdatableFlag.Enabled = true

	enabledFindingsPrioritizationFlag := findingsPrioritizationFlag
	enabledFindingsPrioritizationFlag.Enabled = true

	testCases := []testCase{
		{
			name: "Success: toggles the flag and returns the updated value",
			args: args{ctx: ctx, featureID: updatableFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, updatableFlag.ID).Return(updatableFlag, nil)
				m.database.EXPECT().SetFlag(ctx, toggledUpdatableFlag).Return(nil)
			},
			want: want{flag: toggledUpdatableFlag},
		},
		{
			name: "Success: requests no-post-processing analysis when findings prioritization is enabled",
			args: args{ctx: ctx, featureID: findingsPrioritizationFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				m.analysis.EXPECT().SubmitAnalysisRequest(ctx, services.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(nil)
			},
			want: want{flag: enabledFindingsPrioritizationFlag},
		},
		{
			name: "Success: does not request analysis when findings prioritization is disabled",
			args: args{ctx: ctx, featureID: enabledFindingsPrioritizationFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, enabledFindingsPrioritizationFlag.ID).Return(enabledFindingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(nil)
			},
			want: want{flag: findingsPrioritizationFlag},
		},
		{
			name: "Error: returns ErrNotUserUpdatable when the flag is not user updatable",
			args: args{ctx: ctx, featureID: nonUpdatableFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, nonUpdatableFlag.ID).Return(nonUpdatableFlag, nil)
			},
			want: want{flag: nonUpdatableFlag, err: services.ErrNotUserUpdatable},
		},
		{
			name: "Error: propagates errors from GetFlagByID",
			args: args{ctx: ctx, featureID: 99},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, int32(99)).Return(services.FeatureFlag{}, unexpectedErr)
			},
			want: want{err: unexpectedErr},
		},
		{
			name: "Error: propagates errors from SetFlag",
			args: args{ctx: ctx, featureID: updatableFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, updatableFlag.ID).Return(updatableFlag, nil)
				m.database.EXPECT().SetFlag(ctx, toggledUpdatableFlag).Return(setFlagErr)
			},
			want: want{flag: toggledUpdatableFlag, err: setFlagErr},
		},
		{
			name: "Error: propagates errors from SubmitAnalysisRequest",
			args: args{ctx: ctx, featureID: findingsPrioritizationFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				m.analysis.EXPECT().SubmitAnalysisRequest(ctx, services.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(requestAnalysisErr)
				m.database.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(nil)
			},
			want: want{flag: findingsPrioritizationFlag, err: requestAnalysisErr},
		},
		{
			name: "Error: propagates rollback errors from SubmitAnalysisRequest failure",
			args: args{ctx: ctx, featureID: findingsPrioritizationFlag.ID},
			setupMocks: func(m serviceMocks) {
				m.database.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				m.database.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				m.analysis.EXPECT().SubmitAnalysisRequest(ctx, services.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(requestAnalysisErr)
				m.database.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(rollbackErr)
			},
			want: want{flag: findingsPrioritizationFlag, err: requestAnalysisErr, rollbackErr: rollbackErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc, m := newServiceUnderTest(t)
			testCase.setupMocks(m)

			got, err := svc.ToggleFlag(testCase.args.ctx, testCase.args.featureID)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				require.NoError(t, err)
			}
			if testCase.want.rollbackErr != nil {
				assert.ErrorIs(t, err, testCase.want.rollbackErr)
			}
			assert.Equal(t, testCase.want.flag, got)
		})
	}
}
