// Copyright 2024 Specter Ops, Inc.
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

//go:build integration

package database_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/lib/pq"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/stretchr/testify/require"
)

func setVariableAnalysisModeFlag(t *testing.T, ctx context.Context, dbInst database.Database, enabled bool) {
	t.Helper()

	var variableAnalysisModeFlag appcfg.FeatureFlag
	if existingFlag, err := dbInst.GetFlagByKey(ctx, appcfg.FeatureVariableAnalysisMode); errors.Is(err, database.ErrNotFound) {
		variableAnalysisModeFlag = appcfg.FeatureFlag{
			Key:           appcfg.FeatureVariableAnalysisMode,
			Name:          "Variable Analysis Mode",
			Description:   "Enables analysis requests to run a subset of the analysis pipeline instead of always running the full pipeline.",
			UserUpdatable: false,
		}
	} else {
		require.NoError(t, err)
		variableAnalysisModeFlag = existingFlag
	}

	variableAnalysisModeFlag.Enabled = enabled
	require.NoError(t, dbInst.SetFlag(ctx, variableAnalysisModeFlag))
}

func TestAnalysisRequest(t *testing.T) {
	t.Parallel()

	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	err := dbInst.RequestAnalysis(testCtx, "test", model.AnalysisModeFull)
	require.Nil(t, err)

	analysisRequest, err := dbInst.GetAnalysisRequest(testCtx)
	require.Nil(t, err)
	require.Equal(t, analysisRequest.RequestType, model.AnalysisRequestAnalysis)
	require.Equal(t, analysisRequest.RequestedBy, "test")
	require.False(t, analysisRequest.RequestedAt.IsZero())

	err = dbInst.DeleteAnalysisRequest(testCtx)
	require.Nil(t, err)

	_, err = dbInst.GetAnalysisRequest(testCtx)
	require.ErrorIs(t, err, database.ErrNotFound)
}

func TestAnalysisRequest_MergeAnalysisSteps(t *testing.T) {
	t.Parallel()

	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	setVariableAnalysisModeFlag(t, testCtx, dbInst, true)

	resetState := func(t *testing.T) {
		t.Helper()
		require.NoError(t, dbInst.DeleteAnalysisRequest(testCtx))
	}

	t.Run("subsequent request widens queued analysis_step bits", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisModeNoPostProcessing))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisModeFull))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		require.Equal(t, model.AnalysisStepsFull(), queued.AnalysisSteps)
		require.Equal(t, "tag-editor", queued.RequestedBy, "audit fields preserve the original requester")
	})

	t.Run("subsequent narrower request does not downgrade queued bits", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisModeFull))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisModeNoPostProcessing))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		require.Equal(t, model.AnalysisStepsFull(), queued.AnalysisSteps, "narrower follow-up request must not downgrade queued steps")
		require.Equal(t, "admin", queued.RequestedBy)
	})

	t.Run("identical bits are a no-op and leave the row untouched", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor-1", model.AnalysisModeNoPostProcessing))
		original, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor-2", model.AnalysisModeNoPostProcessing))
		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)

		require.Equal(t, model.AnalysisStepsNoPostProcessing(), queued.AnalysisSteps)
		require.Equal(t, "tag-editor-1", queued.RequestedBy)
		require.Equal(t, original.RequestedAt, queued.RequestedAt, "row must not be updated when bits don't change")
	})

}

func TestAnalysisRequest_RequestTypePrecedence(t *testing.T) {
	t.Parallel()

	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	setVariableAnalysisModeFlag(t, testCtx, dbInst, true)

	resetState := func(t *testing.T) {
		t.Helper()
		require.NoError(t, dbInst.DeleteAnalysisRequest(testCtx))
	}

	deletionRequest := func(requestedBy string) model.AnalysisRequest {
		return model.AnalysisRequest{
			RequestType:           model.AnalysisRequestDeletion,
			RequestedBy:           requestedBy,
			DeleteAllGraph:        true,
			DeleteSourcelessGraph: true,
			DeleteSourceKinds:     pq.StringArray{"source-kind-" + requestedBy},
			DeleteRelationships:   pq.StringArray{"relationship-" + requestedBy},
		}
	}

	t.Run("deletion request is queued when no request exists", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestCollectedGraphDataDeletion(testCtx, deletionRequest("deleter")))

		queued, found := dbInst.HasCollectedGraphDataDeletionRequest(testCtx)
		require.True(t, found)
		require.Equal(t, model.AnalysisRequestDeletion, queued.RequestType)
		require.Equal(t, "deleter", queued.RequestedBy)
		require.True(t, queued.DeleteAllGraph)
		require.True(t, queued.DeleteSourcelessGraph)
		require.Equal(t, pq.StringArray{"source-kind-deleter"}, queued.DeleteSourceKinds)
		require.Equal(t, pq.StringArray{"relationship-deleter"}, queued.DeleteRelationships)
		require.False(t, queued.RequestedAt.IsZero())
	})

	t.Run("deletion request overwrites queued analysis request", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisModeNoPostProcessing))
		require.NoError(t, dbInst.RequestCollectedGraphDataDeletion(testCtx, deletionRequest("deleter")))

		queued, found := dbInst.HasCollectedGraphDataDeletionRequest(testCtx)
		require.True(t, found)
		require.Equal(t, model.AnalysisRequestDeletion, queued.RequestType)
		require.Equal(t, "deleter", queued.RequestedBy)
		require.Equal(t, pq.StringArray{"source-kind-deleter"}, queued.DeleteSourceKinds)
		require.Equal(t, pq.StringArray{"relationship-deleter"}, queued.DeleteRelationships)
	})

	t.Run("analysis request does not overwrite queued deletion request", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestCollectedGraphDataDeletion(testCtx, deletionRequest("deleter")))
		original, found := dbInst.HasCollectedGraphDataDeletionRequest(testCtx)
		require.True(t, found)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisModeFull))

		queued, found := dbInst.HasCollectedGraphDataDeletionRequest(testCtx)
		require.True(t, found)
		require.Equal(t, original.RequestedBy, queued.RequestedBy)
		require.Equal(t, original.RequestedAt, queued.RequestedAt)
		require.Equal(t, original.DeleteSourceKinds, queued.DeleteSourceKinds)
		require.Equal(t, original.DeleteRelationships, queued.DeleteRelationships)
	})

	t.Run("deletion request does not overwrite queued deletion request", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestCollectedGraphDataDeletion(testCtx, deletionRequest("first-deleter")))
		original, found := dbInst.HasCollectedGraphDataDeletionRequest(testCtx)
		require.True(t, found)

		require.NoError(t, dbInst.RequestCollectedGraphDataDeletion(testCtx, deletionRequest("second-deleter")))

		queued, found := dbInst.HasCollectedGraphDataDeletionRequest(testCtx)
		require.True(t, found)
		require.Equal(t, original.RequestedBy, queued.RequestedBy)
		require.Equal(t, original.RequestedAt, queued.RequestedAt)
		require.Equal(t, original.DeleteSourceKinds, queued.DeleteSourceKinds)
		require.Equal(t, original.DeleteRelationships, queued.DeleteRelationships)
	})
}

func TestAnalysisRequest_ConcurrentAnalysisRequestsMerge(t *testing.T) {
	t.Parallel()

	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		errs    = make(chan error, 20)
		wg      sync.WaitGroup
	)

	setVariableAnalysisModeFlag(t, testCtx, dbInst, true)
	require.NoError(t, dbInst.DeleteAnalysisRequest(testCtx))

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			analysisMode := model.AnalysisModeNoPostProcessing
			if index%2 == 0 {
				analysisMode = model.AnalysisModeFull
			}

			errs <- dbInst.RequestAnalysis(testCtx, "requester", analysisMode)
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	queued, err := dbInst.GetAnalysisRequest(testCtx)
	require.NoError(t, err)
	require.Equal(t, model.AnalysisStepsFull(), queued.AnalysisSteps)
}

func TestAnalysisRequest_DisabledVariableAnalysisModeQueuesFullAnalysis(t *testing.T) {
	t.Parallel()

	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	setVariableAnalysisModeFlag(t, testCtx, dbInst, false)

	require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisModeNoPostProcessing))

	queued, err := dbInst.GetAnalysisRequest(testCtx)
	require.NoError(t, err)
	require.Equal(t, model.AnalysisStepsFull(), queued.AnalysisSteps)
}
