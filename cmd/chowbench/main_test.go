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

package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errWriteFailed = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

func TestSummarizeDurations(t *testing.T) {
	summary := summarizeDurations([]time.Duration{
		3 * time.Millisecond,
		1 * time.Millisecond,
		2 * time.Millisecond,
	})

	assert.Equal(t, 2*time.Millisecond, summary.Avg)
	assert.Equal(t, time.Millisecond, summary.Min)
	assert.Equal(t, 3*time.Millisecond, summary.Max)
}

func TestRunReturnsWriteError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "payload.json")
	require.NoError(t, os.WriteFile(file, []byte(`{"graph":{"nodes":[]}}`), 0o600))

	err := run(failingWriter{}, []string{file}, 1, 0, false)

	assert.ErrorIs(t, err, errWriteFailed)
}

func TestSummarizeDurationsEmpty(t *testing.T) {
	assert.Equal(t, durationSummary{}, summarizeDurations(nil))
}

func TestStatusForValidationResult(t *testing.T) {
	assertions := []struct {
		name           string
		report         payload.ValidationReport
		err            error
		expectedStatus string
		expectedErr    string
	}{
		{
			name:           "valid file",
			expectedStatus: "ok",
		},
		{
			name: "validation errors",
			report: payload.ValidationReport{
				ValidationErrors: []payload.ValidationError{{Location: "/graph/nodes[0]"}},
			},
			err:            payload.ErrValidationErrors,
			expectedStatus: "validation_error",
			expectedErr:    payload.ErrValidationErrors.Error(),
		},
		{
			name: "critical errors",
			report: payload.ValidationReport{
				CriticalErrors: []payload.CriticalError{{Message: "bad file"}},
			},
			err:            payload.ErrInvalidFileConfiguration,
			expectedStatus: "critical_error",
			expectedErr:    payload.ErrInvalidFileConfiguration.Error(),
		},
		{
			name:           "harness error",
			err:            errors.New("open file: permission denied"),
			expectedStatus: "error",
			expectedErr:    "open file: permission denied",
		},
	}

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			status, errText := statusForValidationResult(assertion.report, assertion.err)

			assert.Equal(t, assertion.expectedStatus, status)
			assert.Equal(t, assertion.expectedErr, errText)
		})
	}
}

func TestExitErrorForResults(t *testing.T) {
	assertions := []struct {
		name        string
		results     []benchmarkResult
		strict      bool
		expectError bool
	}{
		{
			name: "valid files",
			results: []benchmarkResult{
				{Status: "ok"},
			},
		},
		{
			name: "validation errors are allowed by default",
			results: []benchmarkResult{
				{Status: "validation_error"},
			},
		},
		{
			name: "validation errors fail in strict mode",
			results: []benchmarkResult{
				{Status: "validation_error"},
			},
			strict:      true,
			expectError: true,
		},
		{
			name: "harness errors always fail",
			results: []benchmarkResult{
				{Status: "error"},
			},
			expectError: true,
		},
	}

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			err := exitErrorForResults(assertion.results, assertion.strict)
			if assertion.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteResultsReturnsWriteError(t *testing.T) {
	err := writeResults(failingWriter{}, []benchmarkResult{{File: "payload.json", Status: "ok"}})

	assert.ErrorIs(t, err, errWriteFailed)
}
