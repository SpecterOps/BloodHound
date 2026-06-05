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

package analysis

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestDispatchAnalysisSteps(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		analysisSteps model.AnalysisSteps
		expectedCalls []string
	}{
		{
			name:          "full analysis dispatches every CE analysis stage",
			analysisSteps: model.AnalysisStepsFull(),
			expectedCalls: []string{"ad_post_processing", "azure_post_processing", "tagging", "data_quality"},
		},
		{
			name:          "no post processing skips post-processing",
			analysisSteps: model.AnalysisStepsNoPostProcessing(),
			expectedCalls: []string{"tagging", "data_quality"},
		},
		{
			name:          "empty steps still perform post-run data quality bookkeeping",
			analysisSteps: model.AnalysisSteps{},
			expectedCalls: []string{"data_quality"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var calls []string

			pipelineResult := analysisPipeline{
				{
					analysisStep: model.AnalysisStepADPostProcessing(),
					operation: func(context.Context, database.Database, graph.Database) (operationStatus, []error) {
						calls = append(calls, "ad_post_processing")
						return operationStatusSuccess, nil
					},
				},
				{
					analysisStep: model.AnalysisStepAzurePostProcessing(),
					operation: func(context.Context, database.Database, graph.Database) (operationStatus, []error) {
						calls = append(calls, "azure_post_processing")
						return operationStatusSuccess, nil
					},
				},
				{
					analysisStep: model.AnalysisStepTagging(),
					operation: func(context.Context, database.Database, graph.Database) (operationStatus, []error) {
						calls = append(calls, "tagging")
						return operationStatusSuccess, nil
					},
				},
				{
					name: DataQuality,
					operation: func(context.Context, database.Database, graph.Database) (operationStatus, []error) {
						calls = append(calls, "data_quality")
						return operationStatusSuccess, nil
					},
				},
			}.dispatchAnalysisSteps(analysisPipelineRun{
				ctx:           context.Background(),
				analysisSteps: testCase.analysisSteps,
			})

			require.Equal(t, testCase.expectedCalls, calls)
			require.NoError(t, pipelineResult.Err())
		})
	}
}

func TestNewPipelineHandlesAllBHCEAnalysisSteps(t *testing.T) {
	t.Parallel()

	var (
		handledSteps = map[model.AnalysisStep]string{}
		// Generate findings is implemented by the BHE Butterfly pipeline, not the BHCE pipeline.
		unsupportedSteps = map[model.AnalysisStep]string{
			model.AnalysisStepGenerateFindings(): "BHE only",
		}
	)

	for _, pipelineStep := range newPipeline() {
		if pipelineStep.analysisStep == 0 {
			continue
		}

		stepName, present := model.AnalysisStepName(pipelineStep.analysisStep)
		require.True(t, present, "BHCE pipeline step %d must have an analysis step name", pipelineStep.analysisStep)
		require.NotContains(t, handledSteps, pipelineStep.analysisStep, "BHCE pipeline handles analysis step %q more than once", stepName)

		handledSteps[pipelineStep.analysisStep] = stepName
	}

	for stepBits := 1; stepBits <= model.AnalysisStepsFull().Bits(); stepBits = stepBits << 1 {
		step := model.AnalysisStep(stepBits)
		stepName, present := model.AnalysisStepName(step)

		require.True(t, present, "analysis step %d must have a name", step)

		if _, handled := handledSteps[step]; handled {
			continue
		}

		if _, unsupported := unsupportedSteps[step]; unsupported {
			continue
		}

		require.Failf(t, "missing BHCE pipeline step", "analysis step %q must be handled by newPipeline or listed as unsupported", stepName)
	}

	for unsupportedStep, reason := range unsupportedSteps {
		stepName, present := model.AnalysisStepName(unsupportedStep)

		require.True(t, present, "unsupported BHCE analysis step %d must have a name", unsupportedStep)
		require.True(t, model.AnalysisStepsFull().Has(unsupportedStep), "unsupported BHCE analysis step %q must be part of full analysis", stepName)
		require.NotContains(t, handledSteps, unsupportedStep, "unsupported BHCE analysis step %q is also handled by newPipeline; remove the unsupported entry: %s", stepName, reason)
	}
}
