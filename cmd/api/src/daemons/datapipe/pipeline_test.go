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

package datapipe

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/job"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestBHCEPipeline_Analyze verifies that Analyze decides whether to run the analysis pipeline based on
// the queued AnalysisRequest's RequestType and whether ingest jobs are waiting. Deletion requests must
// not trigger analysis on their own; they are handled by DeleteData on a separate path.
//
// Each subtest sets DisableAnalysis=true so that when the inner analyze() runs it returns
// ErrAnalysisDisabled after calling DeleteAnalysisRequest, without touching the graph database.
// The presence (or absence) of a DeleteAnalysisRequest expectation is what proves whether the inner
// analyze() executed.
func TestBHCEPipeline_Analyze(t *testing.T) {
	t.Run("deletion request queued and no ingest jobs waiting returns no-op", func(t *testing.T) {
		var (
			mockCtrl       = gomock.NewController(t)
			ctx            = context.Background()
			mockDatabase   = dbmocks.NewMockDatabase(mockCtrl)
			pipelineUnderTest = &BHCEPipeline{
				db:         mockDatabase,
				jobService: job.NewJobService(ctx, mockDatabase),
				cfg:        config.Configuration{DisableAnalysis: true},
			}
		)
		defer mockCtrl.Finish()

		mockDatabase.EXPECT().GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing).Return(nil, nil)
		mockDatabase.EXPECT().GetAnalysisRequest(ctx).Return(model.AnalysisRequest{
			RequestType:  model.AnalysisRequestDeletion,
			AnalysisStep: model.AnalysisStepAll,
		}, nil)

		require.NoError(t, pipelineUnderTest.Analyze(ctx))
	})

	t.Run("deletion request queued with ingest jobs waiting runs full analysis", func(t *testing.T) {
		var (
			mockCtrl       = gomock.NewController(t)
			ctx            = context.Background()
			mockDatabase   = dbmocks.NewMockDatabase(mockCtrl)
			pipelineUnderTest = &BHCEPipeline{
				db:         mockDatabase,
				jobService: job.NewJobService(ctx, mockDatabase),
				cfg:        config.Configuration{DisableAnalysis: true},
			}
		)
		defer mockCtrl.Finish()

		mockDatabase.EXPECT().GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing).
			Return([]model.IngestJob{{}}, nil)
		mockDatabase.EXPECT().GetAnalysisRequest(ctx).Return(model.AnalysisRequest{
			RequestType:  model.AnalysisRequestDeletion,
			AnalysisStep: model.AnalysisStepAll,
		}, nil)
		mockDatabase.EXPECT().DeleteAnalysisRequest(ctx).Return(nil)

		require.ErrorIs(t, pipelineUnderTest.Analyze(ctx), ErrAnalysisDisabled)
	})

	t.Run("analysis request queued runs analysis with queued steps", func(t *testing.T) {
		var (
			mockCtrl       = gomock.NewController(t)
			ctx            = context.Background()
			mockDatabase   = dbmocks.NewMockDatabase(mockCtrl)
			pipelineUnderTest = &BHCEPipeline{
				db:         mockDatabase,
				jobService: job.NewJobService(ctx, mockDatabase),
				cfg:        config.Configuration{DisableAnalysis: true},
			}
		)
		defer mockCtrl.Finish()

		mockDatabase.EXPECT().GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing).Return(nil, nil)
		mockDatabase.EXPECT().GetAnalysisRequest(ctx).Return(model.AnalysisRequest{
			RequestType:  model.AnalysisRequestAnalysis,
			AnalysisStep: model.AnalysisStepTaggingToCompletion,
		}, nil)
		mockDatabase.EXPECT().DeleteAnalysisRequest(ctx).Return(nil)

		require.ErrorIs(t, pipelineUnderTest.Analyze(ctx), ErrAnalysisDisabled)
	})

	t.Run("no queued request and no ingest jobs waiting returns no-op", func(t *testing.T) {
		var (
			mockCtrl       = gomock.NewController(t)
			ctx            = context.Background()
			mockDatabase   = dbmocks.NewMockDatabase(mockCtrl)
			pipelineUnderTest = &BHCEPipeline{
				db:         mockDatabase,
				jobService: job.NewJobService(ctx, mockDatabase),
				cfg:        config.Configuration{DisableAnalysis: true},
			}
		)
		defer mockCtrl.Finish()

		mockDatabase.EXPECT().GetIngestJobsWithStatus(ctx, model.JobStatusAnalyzing).Return(nil, nil)
		mockDatabase.EXPECT().GetAnalysisRequest(ctx).Return(model.AnalysisRequest{}, database.ErrNotFound)

		require.NoError(t, pipelineUnderTest.Analyze(ctx))
	})
}
