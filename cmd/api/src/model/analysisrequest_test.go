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

package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ driver.Valuer    = AnalysisSteps{}
	_ sql.Scanner      = (*AnalysisSteps)(nil)
	_ json.Marshaler   = AnalysisSteps{}
	_ json.Unmarshaler = (*AnalysisSteps)(nil)
)

func TestAnalysisStepsFromMode(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		mode          AnalysisMode
		expectedSteps AnalysisSteps
	}{
		{
			name:          "full mode maps to all steps",
			mode:          AnalysisModeFull,
			expectedSteps: AnalysisStepsFull(),
		},
		{
			name:          "tagging mode maps to tagging through completion",
			mode:          AnalysisModeTaggingOnwards,
			expectedSteps: AnalysisStepsTaggingOnwards(),
		},
		{
			name:          "unknown mode defaults to full analysis",
			mode:          AnalysisMode("unknown"),
			expectedSteps: AnalysisStepsFull(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := testCase.mode.AnalysisStepsFromMode()

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}

func TestAnalysisStepsValue(t *testing.T) {
	t.Parallel()

	value, err := AnalysisStepsTaggingOnwards().Value()

	require.NoError(t, err)
	require.Equal(t, int32(AnalysisStepsTaggingOnwards().Bits()), value)
}

func TestAnalysisStepsScan(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name                 string
		initialAnalysisSteps AnalysisSteps
		value                any
		expected             AnalysisSteps
		expectError          bool
	}{
		{
			name:     "int64",
			value:    int64(AnalysisStepsTaggingOnwards().Bits()),
			expected: AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "int32",
			value:    int32(AnalysisStepsTaggingOnwards().Bits()),
			expected: AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "int",
			value:    AnalysisStepsTaggingOnwards().Bits(),
			expected: AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "bytes",
			value:    []byte("12"),
			expected: AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "string",
			value:    "12",
			expected: AnalysisStepsTaggingOnwards(),
		},
		{
			name:                 "nil scans to empty steps",
			initialAnalysisSteps: AnalysisStepsFull(),
			value:                nil,
			expected:             AnalysisSteps{},
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

	payload, err := json.Marshal(AnalysisStepsTaggingOnwards())
	require.NoError(t, err)
	require.JSONEq(t, "12", string(payload))

	var analysisSteps AnalysisSteps
	err = json.Unmarshal(payload, &analysisSteps)
	require.NoError(t, err)
	require.Equal(t, AnalysisStepsTaggingOnwards(), analysisSteps)

	err = json.Unmarshal([]byte(`"tagging"`), &analysisSteps)
	require.Error(t, err)
}

func TestAnalysisStepNames_ContainsNameForEachDefinedBit(t *testing.T) {
	t.Parallel()

	for i := 1; i < int(analysisSentinel); i = i << 1 {
		_, present := analysisStepsFromBits(i).GetNameOfAnalysisStep()

		assert.True(t, present, "analysisStepNames is missing a name for step with bits %d", i)

	}
}

func TestAnalysisStepsMerge(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		firstSteps    AnalysisSteps
		secondSteps   AnalysisSteps
		expectedSteps AnalysisSteps
	}{
		{
			name:          "full wins when requested before tagging",
			firstSteps:    AnalysisModeFull.AnalysisStepsFromMode(),
			secondSteps:   AnalysisModeTaggingOnwards.AnalysisStepsFromMode(),
			expectedSteps: AnalysisStepsFull(),
		},
		{
			name:          "full wins when requested after tagging",
			firstSteps:    AnalysisModeTaggingOnwards.AnalysisStepsFromMode(),
			secondSteps:   AnalysisModeFull.AnalysisStepsFromMode(),
			expectedSteps: AnalysisStepsFull(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := testCase.firstSteps.Merge(testCase.secondSteps)

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}
