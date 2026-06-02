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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/stretchr/testify/require"
)

func setVariableAnalysisEntrypointFlag(t *testing.T, ctx context.Context, dbInst database.Database, enabled bool) {
	t.Helper()

	var variableAnalysisEntrypointFlag appcfg.FeatureFlag
	if existingFlag, err := dbInst.GetFlagByKey(ctx, appcfg.FeatureVariableAnalysisEntrypoint); errors.Is(err, database.ErrNotFound) {
		variableAnalysisEntrypointFlag = appcfg.FeatureFlag{
			Key:           appcfg.FeatureVariableAnalysisEntrypoint,
			Name:          "Variable Analysis Entrypoint",
			Description:   "Enables analysis requests to run a subset of the analysis pipeline instead of always running the full pipeline.",
			UserUpdatable: false,
		}
	} else {
		require.NoError(t, err)
		variableAnalysisEntrypointFlag = existingFlag
	}

	variableAnalysisEntrypointFlag.Enabled = enabled
	require.NoError(t, dbInst.SetFlag(ctx, variableAnalysisEntrypointFlag))
}

func TestAnalysisRequest(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	err := dbInst.RequestAnalysis(testCtx, "test", model.AnalysisEntrypointFull)
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
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	setVariableAnalysisEntrypointFlag(t, testCtx, dbInst, true)

	resetState := func(t *testing.T) {
		t.Helper()
		require.NoError(t, dbInst.DeleteAnalysisRequest(testCtx))
	}

	t.Run("subsequent request widens queued analysis_step bits", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisEntrypointTagging))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisEntrypointFull))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		require.Equal(t, model.AnalysisStepAll, queued.AnalysisSteps)
		require.Equal(t, "tag-editor", queued.RequestedBy, "audit fields preserve the original requester")
	})

	t.Run("subsequent narrower request does not downgrade queued bits", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisEntrypointFull))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisEntrypointTagging))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		require.Equal(t, model.AnalysisStepAll, queued.AnalysisSteps, "narrower follow-up request must not downgrade queued steps")
		require.Equal(t, "admin", queued.RequestedBy)
	})

	t.Run("identical bits are a no-op and leave the row untouched", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor-1", model.AnalysisEntrypointTagging))
		original, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor-2", model.AnalysisEntrypointTagging))
		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)

		require.Equal(t, model.AnalysisStepTaggingToCompletion, queued.AnalysisSteps)
		require.Equal(t, "tag-editor-1", queued.RequestedBy)
		require.Equal(t, original.RequestedAt, queued.RequestedAt, "row must not be updated when bits don't change")
	})

}

func TestAnalysisRequest_DisabledVariableAnalysisEntrypointQueuesFullAnalysis(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	setVariableAnalysisEntrypointFlag(t, testCtx, dbInst, false)

	require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisEntrypointTagging))

	queued, err := dbInst.GetAnalysisRequest(testCtx)
	require.NoError(t, err)
	require.Equal(t, model.AnalysisStepAll, queued.AnalysisSteps)
}
