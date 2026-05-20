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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisRequest(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	err := dbInst.RequestAnalysis(testCtx, "test", model.AnalysisStepAll)
	require.Nil(t, err)

	analysisRequest, err := dbInst.GetAnalysisRequest(testCtx)
	require.Nil(t, err)
	assert.Equal(t, analysisRequest.RequestType, model.AnalysisRequestAnalysis)
	assert.Equal(t, analysisRequest.RequestedBy, "test")
	assert.False(t, analysisRequest.RequestedAt.IsZero())

	err = dbInst.DeleteAnalysisRequest(testCtx)
	require.Nil(t, err)

	_, err = dbInst.GetAnalysisRequest(testCtx)
	assert.ErrorIs(t, err, database.ErrNotFound)
}

func TestAnalysisRequest_MergeAnalysisSteps(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	resetState := func(t *testing.T) {
		t.Helper()
		require.NoError(t, dbInst.DeleteAnalysisRequest(testCtx))
	}

	t.Run("subsequent request widens queued analysis_step bits", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisStepTaggingToCompletion))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisStepAll))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		assert.Equal(t, model.AnalysisStepAll, queued.AnalysisStep)
		assert.Equal(t, "tag-editor", queued.RequestedBy, "audit fields preserve the original requester")
	})

	t.Run("subsequent narrower request does not downgrade queued bits", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "admin", model.AnalysisStepAll))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisStepTaggingToCompletion))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		assert.Equal(t, model.AnalysisStepAll, queued.AnalysisStep, "narrower follow-up request must not downgrade queued steps")
		assert.Equal(t, "admin", queued.RequestedBy)
	})

	t.Run("identical bits are a no-op and leave the row untouched", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor-1", model.AnalysisStepTaggingToCompletion))
		original, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor-2", model.AnalysisStepTaggingToCompletion))
		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)

		assert.Equal(t, model.AnalysisStepTaggingToCompletion, queued.AnalysisStep)
		assert.Equal(t, "tag-editor-1", queued.RequestedBy)
		assert.Equal(t, original.RequestedAt, queued.RequestedAt, "row must not be updated when bits don't change")
	})

	t.Run("disjoint bits union into the queued row", func(t *testing.T) {
		resetState(t)

		require.NoError(t, dbInst.RequestAnalysis(testCtx, "post-proc", model.AnalysisStepADPostProcessing))
		require.NoError(t, dbInst.RequestAnalysis(testCtx, "tag-editor", model.AnalysisStepTaggingToCompletion))

		queued, err := dbInst.GetAnalysisRequest(testCtx)
		require.NoError(t, err)
		assert.Equal(t, model.AnalysisStepADPostProcessing|model.AnalysisStepTaggingToCompletion, queued.AnalysisStep)
		assert.Equal(t, "post-proc", queued.RequestedBy)
	})
}

// TestAnalysisRequestPrecedence exercises the precedence rules that govern
// what happens when a request lands on analysis_request_switch while another
// request is already pending.
func TestAnalysisRequestPrecedence(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)

		partialStep     = model.AnalysisStepTaggingToCompletion
		fullStep        = model.AnalysisStepAll
		deletionRequest = model.AnalysisRequest{
			RequestType:    model.AnalysisRequestDeletion,
			RequestedBy:    "deleter",
			DeleteAllGraph: true,
		}
	)

	type operation struct {
		isDeletion bool
		step       model.AnalysisStep
	}

	type want struct {
		requestType model.AnalysisRequestType
		step        model.AnalysisStep // only checked when requestType == AnalysisRequestAnalysis
	}

	testCases := []struct {
		name     string
		seed     *operation // nil means start with an empty table
		incoming operation
		want     want
	}{
		{
			name:     "empty table, incoming partial inserts partial",
			seed:     nil,
			incoming: operation{step: partialStep},
			want:     want{requestType: model.AnalysisRequestAnalysis, step: partialStep},
		},
		{
			name:     "empty table, incoming full inserts full",
			seed:     nil,
			incoming: operation{step: fullStep},
			want:     want{requestType: model.AnalysisRequestAnalysis, step: fullStep},
		},
		{
			name:     "existing partial, incoming full upgrades to full",
			seed:     &operation{step: partialStep},
			incoming: operation{step: fullStep},
			want:     want{requestType: model.AnalysisRequestAnalysis, step: fullStep},
		},
		{
			name:     "existing full, incoming partial stays full",
			seed:     &operation{step: fullStep},
			incoming: operation{step: partialStep},
			want:     want{requestType: model.AnalysisRequestAnalysis, step: fullStep},
		},
		{
			name:     "existing partial, incoming partial stays partial",
			seed:     &operation{step: partialStep},
			incoming: operation{step: partialStep},
			want:     want{requestType: model.AnalysisRequestAnalysis, step: partialStep},
		},
		{
			name:     "existing analysis, incoming deletion overwrites to deletion",
			seed:     &operation{step: fullStep},
			incoming: operation{isDeletion: true},
			want:     want{requestType: model.AnalysisRequestDeletion},
		},
		{
			name:     "existing deletion, incoming analysis stays deletion",
			seed:     &operation{isDeletion: true},
			incoming: operation{step: fullStep},
			want:     want{requestType: model.AnalysisRequestDeletion},
		},
	}

	apply := func(t *testing.T, action operation) {
		t.Helper()
		if action.isDeletion {
			require.Nil(t, dbInst.RequestCollectedGraphDataDeletion(testCtx, deletionRequest))
		} else {
			require.Nil(t, dbInst.RequestAnalysis(testCtx, "requester", action.step))
		}
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Nil(t, dbInst.DeleteAnalysisRequest(testCtx))

			if testCase.seed != nil {
				apply(t, *testCase.seed)
			}
			apply(t, testCase.incoming)

			got, err := dbInst.GetAnalysisRequest(testCtx)
			require.Nil(t, err)
			assert.Equal(t, testCase.want.requestType, got.RequestType)
			if testCase.want.requestType == model.AnalysisRequestAnalysis {
				assert.Equal(t, testCase.want.step, got.AnalysisStep,
					"expected analysis_step=%d, got %d", testCase.want.step, got.AnalysisStep)
			}
		})
	}
}
