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

// fakeFlagDatabase is a minimal in-memory implementation of the Database port
// used to drive the Service use cases without a real connection pool.
type fakeFlagDatabase struct {
	flag services.FeatureFlag
	err  error
}

type stubAnalysisRequestSubmitter struct {
	called       bool
	requestedBy  string
	analysisMode model.AnalysisMode
	err          error
}

func (f fakeFlagDatabase) GetFlagByKey(_ context.Context, _ string) (services.FeatureFlag, error) {
	return f.flag, f.err
}

func (f fakeFlagDatabase) GetFlagByID(_ context.Context, _ int32) (services.FeatureFlag, error) {
	return f.flag, f.err
}

func (f fakeFlagDatabase) GetAllFlags(_ context.Context) ([]services.FeatureFlag, error) {
	return nil, f.err
}

func (f fakeFlagDatabase) SetFlag(_ context.Context, _ services.FeatureFlag) error {
	return f.err
}

func (s *stubAnalysisRequestSubmitter) SubmitAnalysisRequest(_ context.Context, requestedBy string, analysisMode model.AnalysisMode) error {
	s.called = true
	s.requestedBy = requestedBy
	s.analysisMode = analysisMode
	return s.err
}

func TestNewService(t *testing.T) {
	mockDb := mocks.NewMockDatabase(t)
	assert.NotNil(t, services.NewService(mockDb, &stubAnalysisRequestSubmitter{}))
}

func TestService_GetFlagByKey(t *testing.T) {
	var (
		ctx     = context.Background()
		want    = services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Enabled: true}
		notFErr = services.ErrNotFound
	)

	t.Run("returns the flag from the database", func(t *testing.T) {
		svc := services.NewService(fakeFlagDatabase{flag: want}, &stubAnalysisRequestSubmitter{})

		got, err := svc.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates the database error", func(t *testing.T) {
		svc := services.NewService(fakeFlagDatabase{err: notFErr}, &stubAnalysisRequestSubmitter{})

		_, err := svc.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)

		assert.ErrorIs(t, err, notFErr)
	})
}

func TestService_IsEnabled(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name    string
		db      fakeFlagDatabase
		want    bool
		wantErr error
	}{
		{
			name: "true when the flag is enabled",
			db:   fakeFlagDatabase{flag: services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: true}},
			want: true,
		},
		{
			name: "false when the flag is disabled",
			db:   fakeFlagDatabase{flag: services.FeatureFlag{Key: services.FeatureOpenHoundSupport, Enabled: false}},
			want: false,
		},
		{
			name:    "propagates database errors",
			db:      fakeFlagDatabase{err: dbErr},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := services.NewService(tt.db, &stubAnalysisRequestSubmitter{})
			got, err := svc.IsEnabled(ctx, services.FeatureOpenHoundSupport)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
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
		dbResult   []services.FeatureFlag
		dbErr      error
		wantResult []services.FeatureFlag
		wantErr    error
	}{
		{
			name:       "returns all flags on success",
			dbResult:   expected,
			wantResult: expected,
		},
		{
			name:    "propagates database errors",
			dbErr:   unexpectedErr,
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock, &stubAnalysisRequestSubmitter{})
			)

			databaseMock.EXPECT().GetAllFlags(ctx).Return(tt.dbResult, tt.dbErr)

			got, err := svc.GetAllFlags(ctx)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, got)
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
		name                 string
		featureID            int32
		analysisRequesterErr error
		setupMocks           func(databaseMock *mocks.MockDatabase)
		assert               func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error)
	}

	toggledUpdatableFlag := updatableFlag
	toggledUpdatableFlag.Enabled = true

	enabledFindingsPrioritizationFlag := findingsPrioritizationFlag
	enabledFindingsPrioritizationFlag.Enabled = true

	testCases := []testCase{
		{
			name:      "toggles the flag and returns the updated value",
			featureID: updatableFlag.ID,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, updatableFlag.ID).Return(updatableFlag, nil)
				databaseMock.EXPECT().SetFlag(ctx, toggledUpdatableFlag).Return(nil)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				require.NoError(t, err)
				assert.False(t, analysisRequester.called)
				assert.Equal(t, toggledUpdatableFlag, got)
			},
		},
		{
			name:      "requests no-post-processing analysis when findings prioritization is enabled",
			featureID: findingsPrioritizationFlag.ID,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				databaseMock.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				require.NoError(t, err)
				assert.True(t, analysisRequester.called)
				assert.Equal(t, services.PrioritizationFlagRequestSource, analysisRequester.requestedBy)
				assert.Equal(t, model.AnalysisModeNoPostProcessing, analysisRequester.analysisMode)
				assert.Equal(t, enabledFindingsPrioritizationFlag, got)
			},
		},
		{
			name:      "does not request analysis when findings prioritization is disabled",
			featureID: enabledFindingsPrioritizationFlag.ID,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, enabledFindingsPrioritizationFlag.ID).Return(enabledFindingsPrioritizationFlag, nil)
				databaseMock.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(nil)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				require.NoError(t, err)
				assert.False(t, analysisRequester.called)
				assert.Equal(t, findingsPrioritizationFlag, got)
			},
		},
		{
			name:      "returns ErrNotUserUpdatable when the flag is not user updatable",
			featureID: nonUpdatableFlag.ID,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, nonUpdatableFlag.ID).Return(nonUpdatableFlag, nil)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, services.ErrNotUserUpdatable)
				assert.False(t, analysisRequester.called)
				assert.Equal(t, nonUpdatableFlag, got)
			},
		},
		{
			name:      "propagates errors from GetFlagByID",
			featureID: 99,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, int32(99)).Return(services.FeatureFlag{}, unexpectedErr)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, unexpectedErr)
				assert.False(t, analysisRequester.called)
				assert.Equal(t, services.FeatureFlag{}, got)
			},
		},
		{
			name:      "propagates errors from SetFlag",
			featureID: updatableFlag.ID,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, updatableFlag.ID).Return(updatableFlag, nil)
				databaseMock.EXPECT().SetFlag(ctx, toggledUpdatableFlag).Return(setFlagErr)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, setFlagErr)
				assert.False(t, analysisRequester.called)
				assert.Equal(t, toggledUpdatableFlag, got)
			},
		},
		{
			name:                 "propagates errors from SubmitAnalysisRequest",
			featureID:            findingsPrioritizationFlag.ID,
			analysisRequesterErr: requestAnalysisErr,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				databaseMock.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				databaseMock.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(nil)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, requestAnalysisErr)
				assert.True(t, analysisRequester.called)
				assert.Equal(t, findingsPrioritizationFlag, got)
			},
		},
		{
			name:                 "propagates rollback errors from SubmitAnalysisRequest failure",
			featureID:            findingsPrioritizationFlag.ID,
			analysisRequesterErr: requestAnalysisErr,
			setupMocks: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetFlagByID(ctx, findingsPrioritizationFlag.ID).Return(findingsPrioritizationFlag, nil)
				databaseMock.EXPECT().SetFlag(ctx, enabledFindingsPrioritizationFlag).Return(nil)
				databaseMock.EXPECT().SetFlag(ctx, findingsPrioritizationFlag).Return(rollbackErr)
			},
			assert: func(t *testing.T, analysisRequester *stubAnalysisRequestSubmitter, got services.FeatureFlag, err error) {
				assert.ErrorIs(t, err, requestAnalysisErr)
				assert.ErrorIs(t, err, rollbackErr)
				assert.True(t, analysisRequester.called)
				assert.Equal(t, findingsPrioritizationFlag, got)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock      = mocks.NewMockDatabase(t)
				analysisRequester = &stubAnalysisRequestSubmitter{err: testCase.analysisRequesterErr}
				svc               = services.NewService(databaseMock, analysisRequester)
			)

			testCase.setupMocks(databaseMock)

			got, err := svc.ToggleFlag(ctx, testCase.featureID)
			testCase.assert(t, analysisRequester, got, err)
		})
	}
}
