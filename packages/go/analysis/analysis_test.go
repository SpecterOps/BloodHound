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
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
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
					operation: func(analysisPipelineRun) (pipelineStepStatus, []error) {
						calls = append(calls, "ad_post_processing")
						return pipelineStepStatusSuccess, nil
					},
				},
				{
					analysisStep: model.AnalysisStepAzurePostProcessing(),
					operation: func(analysisPipelineRun) (pipelineStepStatus, []error) {
						calls = append(calls, "azure_post_processing")
						return pipelineStepStatusSuccess, nil
					},
				},
				{
					analysisStep: model.AnalysisStepTagging(),
					operation: func(analysisPipelineRun) (pipelineStepStatus, []error) {
						calls = append(calls, "tagging")
						return pipelineStepStatusSuccess, nil
					},
				},
				{
					name: DataQuality,
					operation: func(analysisPipelineRun) (pipelineStepStatus, []error) {
						calls = append(calls, "data_quality")
						return pipelineStepStatusSuccess, nil
					},
				},
			}.dispatchAnalysisSteps(analysisPipelineRun{
				ctx:           context.Background(),
				analysisSteps: testCase.analysisSteps,
				analysisErrs:  &analysisErrors{},
			})

			require.Equal(t, testCase.expectedCalls, calls)
			require.Empty(t, pipelineResult.Errors())
		})
	}
}

func TestAnalysisPipelineStepShouldRun(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		pipelineStep  analysisPipelineStep
		analysisSteps model.AnalysisSteps
		expected      bool
	}{
		{
			name:          "non selectable step always runs",
			pipelineStep:  analysisPipelineStep{name: DataQuality},
			analysisSteps: model.AnalysisSteps{},
			expected:      true,
		},
		{
			name:          "selected step runs",
			pipelineStep:  analysisPipelineStep{analysisStep: model.AnalysisStepTagging()},
			analysisSteps: model.AnalysisStepsNoPostProcessing(),
			expected:      true,
		},
		{
			name:          "unselected step does not run",
			pipelineStep:  analysisPipelineStep{analysisStep: model.AnalysisStepADPostProcessing()},
			analysisSteps: model.AnalysisStepsNoPostProcessing(),
			expected:      false,
		},
		{
			name:          "selected full-analysis step runs",
			pipelineStep:  analysisPipelineStep{analysisStep: model.AnalysisStepAzurePostProcessing()},
			analysisSteps: model.AnalysisStepsFull(),
			expected:      true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, testCase.pipelineStep.shouldRun(testCase.analysisSteps))
		})
	}
}

func TestAnalysisPipelineStepString(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name         string
		pipelineStep analysisPipelineStep
		expected     string
	}{
		{
			name:         "known analysis step uses model step name",
			pipelineStep: analysisPipelineStep{analysisStep: model.AnalysisStepTagging(), name: "ignored"},
			expected:     "tagging",
		},
		{
			name:         "non selectable step uses explicit name",
			pipelineStep: analysisPipelineStep{name: DataQuality},
			expected:     DataQuality,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, testCase.pipelineStep.String())
		})
	}
}

func TestAnalysisPipelineString(t *testing.T) {
	t.Parallel()

	pipeline := analysisPipeline{
		{analysisStep: model.AnalysisStepADPostProcessing()},
		{analysisStep: model.AnalysisStepTagging()},
		{name: DataQuality},
	}

	require.Equal(t, "ad_post_processing,tagging,data_quality", pipeline.String())
}

func TestAnalysisPipelineStepResultString(t *testing.T) {
	t.Parallel()

	result := analysisPipelineStepResult{
		name:   "tagging",
		status: pipelineStepStatusSuccess,
	}

	require.Equal(t, "tagging:success", result.String())
}

func TestAnalysisPipelineResultString(t *testing.T) {
	t.Parallel()

	result := analysisPipelineResult{
		{name: "ad_post_processing", status: pipelineStepStatusSkipped},
		{name: "tagging", status: pipelineStepStatusSuccess},
		{name: DataQuality, status: pipelineStepStatusFailed},
	}

	require.Equal(t, "ad_post_processing:skipped,tagging:success,data_quality:failed", result.String())
}

func TestAnalysisPipelineResultErrors(t *testing.T) {
	t.Parallel()

	var (
		firstError  = errors.New("first")
		secondError = errors.New("second")
		result      = analysisPipelineResult{
			{name: "ad_post_processing", errors: []error{firstError}},
			{name: "tagging"},
			{name: DataQuality, errors: []error{secondError}},
		}
	)

	collectedErrors := result.Errors()

	require.Len(t, collectedErrors, 2)
	require.ErrorIs(t, collectedErrors[0], firstError)
	require.ErrorIs(t, collectedErrors[1], secondError)
}

func TestAnalysisErrorsEvaluateErrors(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		errs        analysisErrors
		expectedErr error
	}{
		{
			name: "no errors succeeds",
		},
		{
			name: "all full-failure fields failed",
			errs: analysisErrors{
				adPost:      true,
				azurePost:   true,
				agi:         true,
				dataQuality: true,
			},
			expectedErr: ErrAnalysisFailed,
		},
		{
			name: "ad post failure partially completes",
			errs: analysisErrors{
				adPost: true,
			},
			expectedErr: ErrAnalysisPartiallyCompleted,
		},
		{
			name: "azure post failure partially completes",
			errs: analysisErrors{
				azurePost: true,
			},
			expectedErr: ErrAnalysisPartiallyCompleted,
		},
		{
			name: "agi failure partially completes",
			errs: analysisErrors{
				agi: true,
			},
			expectedErr: ErrAnalysisPartiallyCompleted,
		},
		{
			name: "agt partial failure partially completes",
			errs: analysisErrors{
				agtPartial: true,
			},
			expectedErr: ErrAnalysisPartiallyCompleted,
		},
		{
			name: "data quality failure partially completes",
			errs: analysisErrors{
				dataQuality: true,
			},
			expectedErr: ErrAnalysisPartiallyCompleted,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.errs.evaluateErrors()

			if testCase.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, testCase.expectedErr)
			}
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
		require.NotNil(t, pipelineStep.operation, "BHCE pipeline step %q must have an operation", stepName)

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
