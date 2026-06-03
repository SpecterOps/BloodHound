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

package model_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/bits"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

var (
	_ driver.Valuer    = model.AnalysisSteps{}
	_ sql.Scanner      = (*model.AnalysisSteps)(nil)
	_ json.Marshaler   = model.AnalysisSteps{}
	_ json.Unmarshaler = (*model.AnalysisSteps)(nil)
	_ fmt.Stringer     = model.AnalysisSteps{}
)

func TestAnalysisStepsFromMode(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		mode          model.AnalysisMode
		expectedSteps model.AnalysisSteps
	}{
		{
			name:          "full mode maps to all steps",
			mode:          model.AnalysisModeFull,
			expectedSteps: model.AnalysisStepAll,
		},
		{
			name:          "tagging mode maps to tagging through completion",
			mode:          model.AnalysisModeTaggingOnwards,
			expectedSteps: model.AnalysisStepTaggingToCompletion,
		},
		{
			name:          "unknown mode defaults to full analysis",
			mode:          model.AnalysisMode("unknown"),
			expectedSteps: model.AnalysisStepAll,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := model.AnalysisStepsFromMode(testCase.mode)

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}

func TestAnalysisStepsString(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		analysisSteps model.AnalysisSteps
		expected      string
	}{
		{
			name:          "full analysis",
			analysisSteps: model.AnalysisStepAll,
			expected:      "ad_post_processing,azure_post_processing,tagging,analysis",
		},
		{
			name:          "tagging to completion",
			analysisSteps: model.AnalysisStepTaggingToCompletion,
			expected:      "tagging,analysis",
		},
		{
			name:          "empty",
			analysisSteps: model.AnalysisSteps{},
			expected:      "none",
		},
		{
			name:          "unknown bits",
			analysisSteps: model.AnalysisStepsFromBits(32),
			expected:      "unknown:32",
		},
		{
			name:          "known and unknown bits",
			analysisSteps: model.AnalysisStepsFromBits(model.AnalysisStepTagging.Bits() | 32),
			expected:      "tagging,unknown:32",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, testCase.analysisSteps.String())
		})
	}
}

func TestAnalysisStepsStringCoversAllSteps(t *testing.T) {
	stepNames := strings.Split(model.AnalysisStepAll.String(), ",")

	require.Len(t, stepNames, bits.OnesCount(uint(model.AnalysisStepAll.Bits())))
	require.NotContains(t, model.AnalysisStepAll.String(), "unknown:")
	require.NotContains(t, model.AnalysisStepAll.String(), "none")
}

func TestAnalysisStepsValue(t *testing.T) {
	t.Parallel()

	value, err := model.AnalysisStepTaggingToCompletion.Value()

	require.NoError(t, err)
	require.Equal(t, int64(model.AnalysisStepTaggingToCompletion.Bits()), value)
}

func TestAnalysisStepsScan(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name                 string
		initialAnalysisSteps model.AnalysisSteps
		value                any
		expected             model.AnalysisSteps
		expectError          bool
	}{
		{
			name:     "int64",
			value:    int64(model.AnalysisStepTaggingToCompletion.Bits()),
			expected: model.AnalysisStepTaggingToCompletion,
		},
		{
			name:     "int32",
			value:    int32(model.AnalysisStepTaggingToCompletion.Bits()),
			expected: model.AnalysisStepTaggingToCompletion,
		},
		{
			name:     "int",
			value:    model.AnalysisStepTaggingToCompletion.Bits(),
			expected: model.AnalysisStepTaggingToCompletion,
		},
		{
			name:     "bytes",
			value:    []byte("12"),
			expected: model.AnalysisStepTaggingToCompletion,
		},
		{
			name:     "string",
			value:    "12",
			expected: model.AnalysisStepTaggingToCompletion,
		},
		{
			name:                 "nil scans to empty steps",
			initialAnalysisSteps: model.AnalysisStepAll,
			value:                nil,
			expected:             model.AnalysisSteps{},
		},
		{
			name:        "invalid type returns an error",
			value:       true,
			expectError: true,
		},
		{
			name:        "invalid numeric text returns an error",
			value:       "not-a-number",
			expectError: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var analysisSteps = testCase.initialAnalysisSteps

			err := analysisSteps.Scan(testCase.value)

			if testCase.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected, analysisSteps)
		})
	}
}

func TestAnalysisStepsJSON(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(model.AnalysisStepTaggingToCompletion)
	require.NoError(t, err)
	require.JSONEq(t, "12", string(payload))

	var analysisSteps model.AnalysisSteps
	err = json.Unmarshal(payload, &analysisSteps)
	require.NoError(t, err)
	require.Equal(t, model.AnalysisStepTaggingToCompletion, analysisSteps)

	err = json.Unmarshal([]byte(`"tagging"`), &analysisSteps)
	require.Error(t, err)
}

func TestAnalysisStepsMerge(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		firstSteps    model.AnalysisSteps
		secondSteps   model.AnalysisSteps
		expectedSteps model.AnalysisSteps
	}{
		{
			name:          "full wins when requested before tagging",
			firstSteps:    model.AnalysisStepsFromMode(model.AnalysisModeFull),
			secondSteps:   model.AnalysisStepsFromMode(model.AnalysisModeTaggingOnwards),
			expectedSteps: model.AnalysisStepAll,
		},
		{
			name:          "full wins when requested after tagging",
			firstSteps:    model.AnalysisStepsFromMode(model.AnalysisModeTaggingOnwards),
			secondSteps:   model.AnalysisStepsFromMode(model.AnalysisModeFull),
			expectedSteps: model.AnalysisStepAll,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := testCase.firstSteps.Merge(testCase.secondSteps)

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}
