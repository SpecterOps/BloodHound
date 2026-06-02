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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	_ driver.Valuer    = AnalysisSteps{}
	_ sql.Scanner      = (*AnalysisSteps)(nil)
	_ json.Marshaler   = AnalysisSteps{}
	_ json.Unmarshaler = (*AnalysisSteps)(nil)
	_ fmt.Stringer     = AnalysisSteps{}
)

func TestAnalysisStepsFromEntrypoint(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		entrypoint    AnalysisEntrypoint
		expectedSteps AnalysisSteps
	}{
		{
			name:          "full entrypoint maps to all steps",
			entrypoint:    AnalysisEntrypointFull,
			expectedSteps: AnalysisStepAll,
		},
		{
			name:          "tagging entrypoint maps to tagging through completion",
			entrypoint:    AnalysisEntrypointTagging,
			expectedSteps: AnalysisStepTaggingToCompletion,
		},
		{
			name:          "unknown entrypoint defaults to full analysis",
			entrypoint:    AnalysisEntrypoint("unknown"),
			expectedSteps: AnalysisStepAll,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := AnalysisStepsFromEntrypoint(testCase.entrypoint)

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}

func TestAnalysisStepsString(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		analysisSteps AnalysisSteps
		expected      string
	}{
		{
			name:          "full analysis",
			analysisSteps: AnalysisStepAll,
			expected:      "ad_post_processing,azure_post_processing,tagging,analysis",
		},
		{
			name:          "tagging to completion",
			analysisSteps: AnalysisStepTaggingToCompletion,
			expected:      "tagging,analysis",
		},
		{
			name:          "empty",
			analysisSteps: AnalysisSteps{},
			expected:      "none",
		},
		{
			name:          "unknown bits",
			analysisSteps: AnalysisStepsFromBits(32),
			expected:      "unknown:32",
		},
		{
			name:          "known and unknown bits",
			analysisSteps: AnalysisStepsFromBits(AnalysisStepTagging.Bits() | 32),
			expected:      "tagging,unknown:32",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, testCase.analysisSteps.String())
		})
	}
}

func TestAnalysisStepsValue(t *testing.T) {
	t.Parallel()

	value, err := AnalysisStepTaggingToCompletion.Value()

	require.NoError(t, err)
	require.Equal(t, int64(AnalysisStepTaggingToCompletion.Bits()), value)
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
			value:    int64(AnalysisStepTaggingToCompletion.Bits()),
			expected: AnalysisStepTaggingToCompletion,
		},
		{
			name:     "int32",
			value:    int32(AnalysisStepTaggingToCompletion.Bits()),
			expected: AnalysisStepTaggingToCompletion,
		},
		{
			name:     "int",
			value:    AnalysisStepTaggingToCompletion.Bits(),
			expected: AnalysisStepTaggingToCompletion,
		},
		{
			name:     "bytes",
			value:    []byte("12"),
			expected: AnalysisStepTaggingToCompletion,
		},
		{
			name:     "string",
			value:    "12",
			expected: AnalysisStepTaggingToCompletion,
		},
		{
			name:                 "nil scans to empty steps",
			initialAnalysisSteps: AnalysisStepAll,
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

	payload, err := json.Marshal(AnalysisStepTaggingToCompletion)
	require.NoError(t, err)
	require.JSONEq(t, "12", string(payload))

	var analysisSteps AnalysisSteps
	err = json.Unmarshal(payload, &analysisSteps)
	require.NoError(t, err)
	require.Equal(t, AnalysisStepTaggingToCompletion, analysisSteps)

	err = json.Unmarshal([]byte(`"tagging"`), &analysisSteps)
	require.Error(t, err)
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
			firstSteps:    AnalysisStepsFromEntrypoint(AnalysisEntrypointFull),
			secondSteps:   AnalysisStepsFromEntrypoint(AnalysisEntrypointTagging),
			expectedSteps: AnalysisStepAll,
		},
		{
			name:          "full wins when requested after tagging",
			firstSteps:    AnalysisStepsFromEntrypoint(AnalysisEntrypointTagging),
			secondSteps:   AnalysisStepsFromEntrypoint(AnalysisEntrypointFull),
			expectedSteps: AnalysisStepAll,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := testCase.firstSteps.Merge(testCase.secondSteps)

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}
