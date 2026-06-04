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
			expectedSteps: model.AnalysisStepsFull(),
		},
		{
			name:          "tagging mode maps to tagging through completion",
			mode:          model.AnalysisModeTaggingOnwards,
			expectedSteps: model.AnalysisStepsTaggingOnwards(),
		},
		{
			name:          "unknown mode defaults to full analysis",
			mode:          model.AnalysisMode("unknown"),
			expectedSteps: model.AnalysisStepsFull(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := testCase.mode.AnalysisStepsFromMode()

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
			analysisSteps: model.AnalysisStepsFull(),
			expected:      "ad_post_processing,azure_post_processing,tagging,analysis",
		},
		{
			name:          "tagging to completion",
			analysisSteps: model.AnalysisStepsTaggingOnwards(),
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
			analysisSteps: model.AnalysisStepTagging().Merge(model.AnalysisStepsFromBits(32)),
			expected:      "tagging,unknown:32",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, testCase.analysisSteps.String())
		})
	}
}

func TestOrderedAnalysisStepDefinitions(t *testing.T) {
	t.Parallel()

	var (
		definedSteps model.AnalysisSteps
		seenBits     = map[int]string{}
		seenNames    = map[string]struct{}{}
	)

	stepDefinitions := model.AnalysisDefinition()
	require.NotEmpty(t, stepDefinitions)

	for _, stepDefinition := range stepDefinitions {
		require.False(t, stepDefinition.Step.IsEmpty(), "analysis step definition %q has no bits", stepDefinition.Name)
		require.Equal(t, 1, bits.OnesCount(uint(stepDefinition.Step.Bits())), "analysis step definition %q must contain exactly one bit", stepDefinition.Name)
		require.NotEmpty(t, stepDefinition.Name)
		require.NotContains(t, seenBits, stepDefinition.Step.Bits(), "analysis step bit is defined more than once")
		require.NotContains(t, seenNames, stepDefinition.Name, "analysis step name is defined more than once")

		definedSteps = definedSteps.Merge(stepDefinition.Step)
		seenBits[stepDefinition.Step.Bits()] = stepDefinition.Name
		seenNames[stepDefinition.Name] = struct{}{}
	}

	require.Equal(t, model.AnalysisStepsFull(), definedSteps)
}

func TestOrderedAnalysisStepDefinitionsReturnsCopy(t *testing.T) {
	t.Parallel()

	stepDefinitions := model.AnalysisDefinition()
	require.NotEmpty(t, stepDefinitions)

	stepDefinitions[0].Name = "changed"

	require.NotEqual(t, "changed", model.AnalysisDefinition()[0].Name)
}

func TestAnalysisStepHelpersUseDefinitions(t *testing.T) {
	t.Parallel()

	var allDefinedSteps model.AnalysisSteps
	for _, stepDefinition := range model.AnalysisDefinition() {
		allDefinedSteps = allDefinedSteps.Merge(stepDefinition.Step)
	}

	taggingToCompletion := analysisStepsFromDefinition(t, model.AnalysisStepTagging())

	require.Equal(t, allDefinedSteps, model.AnalysisStepsFull())
	require.Equal(t, allDefinedSteps, model.AnalysisStepsFull())
	require.Equal(t, allDefinedSteps, model.AnalysisModeFull.AnalysisStepsFromMode())
	require.Equal(t, taggingToCompletion, model.AnalysisStepsTaggingOnwards())
	require.Equal(t, taggingToCompletion, model.AnalysisModeTaggingOnwards.AnalysisStepsFromMode())
}

func TestAnalysisStepsStringUsesDefinitions(t *testing.T) {
	t.Parallel()

	var (
		allDefinedSteps model.AnalysisSteps
		stepNames       []string
	)

	for _, stepDefinition := range model.AnalysisDefinition() {
		allDefinedSteps = allDefinedSteps.Merge(stepDefinition.Step)
		stepNames = append(stepNames, stepDefinition.Name)
	}

	require.Equal(t, strings.Join(stepNames, ","), allDefinedSteps.String())
	require.Len(t, stepNames, bits.OnesCount(uint(model.AnalysisStepsFull().Bits())))
	require.NotContains(t, model.AnalysisStepsFull().String(), "unknown:")
	require.NotContains(t, model.AnalysisStepsFull().String(), "none")
}

func analysisStepsFromDefinition(t *testing.T, startStep model.AnalysisSteps) model.AnalysisSteps {
	t.Helper()

	var (
		foundStartStep bool
		analysisSteps  model.AnalysisSteps
	)

	for _, stepDefinition := range model.AnalysisDefinition() {
		if stepDefinition.Step == startStep {
			foundStartStep = true
		}

		if foundStartStep {
			analysisSteps = analysisSteps.Merge(stepDefinition.Step)
		}
	}

	require.True(t, foundStartStep, "analysis step definition is missing start step %s", startStep)
	return analysisSteps
}

func TestAnalysisStepsValue(t *testing.T) {
	t.Parallel()

	value, err := model.AnalysisStepsTaggingOnwards().Value()

	require.NoError(t, err)
	require.Equal(t, int32(model.AnalysisStepsTaggingOnwards().Bits()), value)
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
			value:    int64(model.AnalysisStepsTaggingOnwards().Bits()),
			expected: model.AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "int32",
			value:    int32(model.AnalysisStepsTaggingOnwards().Bits()),
			expected: model.AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "int",
			value:    model.AnalysisStepsTaggingOnwards().Bits(),
			expected: model.AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "bytes",
			value:    []byte("12"),
			expected: model.AnalysisStepsTaggingOnwards(),
		},
		{
			name:     "string",
			value:    "12",
			expected: model.AnalysisStepsTaggingOnwards(),
		},
		{
			name:                 "nil scans to empty steps",
			initialAnalysisSteps: model.AnalysisStepsFull(),
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

	payload, err := json.Marshal(model.AnalysisStepsTaggingOnwards())
	require.NoError(t, err)
	require.JSONEq(t, "12", string(payload))

	var analysisSteps model.AnalysisSteps
	err = json.Unmarshal(payload, &analysisSteps)
	require.NoError(t, err)
	require.Equal(t, model.AnalysisStepsTaggingOnwards(), analysisSteps)

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
			firstSteps:    model.AnalysisModeFull.AnalysisStepsFromMode(),
			secondSteps:   model.AnalysisModeTaggingOnwards.AnalysisStepsFromMode(),
			expectedSteps: model.AnalysisStepsFull(),
		},
		{
			name:          "full wins when requested after tagging",
			firstSteps:    model.AnalysisModeTaggingOnwards.AnalysisStepsFromMode(),
			secondSteps:   model.AnalysisModeFull.AnalysisStepsFromMode(),
			expectedSteps: model.AnalysisStepsFull(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysisSteps := testCase.firstSteps.Merge(testCase.secondSteps)

			require.Equal(t, testCase.expectedSteps, analysisSteps)
		})
	}
}
