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
			name:          "tagging to completion skips post-processing",
			analysisSteps: model.AnalysisStepsNoPostProcessing(),
			expectedCalls: []string{"tagging", "data_quality"},
		},
		{
			name:          "single selected stage only dispatches that stage",
			analysisSteps: model.AnalysisStepADPostProcessing(),
			expectedCalls: []string{"ad_post_processing", "data_quality"},
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

			dispatchAnalysisSteps(testCase.analysisSteps, analysisStepDispatch{
				adPostProcessing: func() {
					calls = append(calls, "ad_post_processing")
				},
				azurePostProcessing: func() {
					calls = append(calls, "azure_post_processing")
				},
				tagging: func() {
					calls = append(calls, "tagging")
				},
				saveDataQuality: func() {
					calls = append(calls, "data_quality")
				},
			})

			require.Equal(t, testCase.expectedCalls, calls)
		})
	}
}
