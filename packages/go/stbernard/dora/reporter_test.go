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

package dora

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTerminalReporter(t *testing.T) {
	now := time.Now()
	snapshot := MetricsSnapshot{
		PeriodStart:               now.AddDate(0, 0, -30),
		PeriodEnd:                 now,
		CalculatedAt:              now,
		DeploymentCount:           15,
		DeploymentFrequencyPerDay: 0.5,
		DeploymentTier:            string(TierHigh),
		LeadTimeP50Hours:          6.0,
		LeadTimeP90Hours:          12.0,
		LeadTimeP95Hours:          18.0,
		LeadTimeTier:              string(TierElite),
		FailedDeploymentCount:     1,
		ChangeFailureRate:         6.67,
		FailureRateTier:           string(TierHigh),
		IncidentCount:             1,
		MTTRHours:                 4.0,
		MedianTTRHours:            4.0,
		P95TTRHours:               4.0,
		RestoreTimeTier:           string(TierHigh),
		OverallTier:               string(TierHigh),
		AverageRCsPerRelease:      2.0,
		MedianRCsPerRelease:       2.0,
		TotalCommitsInPeriod:      75,
		AverageCommitsPerRelease:  5.0,
	}

	t.Run("terminal output with colors", func(t *testing.T) {
		reporter := NewTerminalReporter(true)
		var buf bytes.Buffer

		err := reporter.Report(snapshot, &buf)
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		output := buf.String()

		// Check for key sections
		if !strings.Contains(output, "DORA Metrics Report") {
			t.Error("Missing report header")
		}
		if !strings.Contains(output, "Overall Performance") {
			t.Error("Missing overall performance")
		}
		if !strings.Contains(output, "Deployment Frequency") {
			t.Error("Missing deployment frequency")
		}
		if !strings.Contains(output, "Lead Time for Changes") {
			t.Error("Missing lead time")
		}
		if !strings.Contains(output, "Change Failure Rate") {
			t.Error("Missing failure rate")
		}
		if !strings.Contains(output, "Time to Restore Service") {
			t.Error("Missing MTTR")
		}
		if !strings.Contains(output, "Quality Indicators") {
			t.Error("Missing quality metrics")
		}

		// Check for values
		if !strings.Contains(output, "0.50 per day") {
			t.Error("Missing deployment frequency value")
		}
		if !strings.Contains(output, "P50: 6.0h") {
			t.Error("Missing lead time P50")
		}
		if !strings.Contains(output, "6.7%") {
			t.Error("Missing failure rate percentage")
		}

		t.Logf("Terminal output:\n%s", output)
	})

	t.Run("terminal output without colors", func(t *testing.T) {
		reporter := NewTerminalReporter(false)
		var buf bytes.Buffer

		err := reporter.Report(snapshot, &buf)
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		output := buf.String()

		// Should not contain ANSI escape codes
		if strings.Contains(output, "\033[") {
			t.Error("Output contains ANSI codes when colors disabled")
		}

		// Should contain tier names in uppercase
		if !strings.Contains(output, "HIGH") {
			t.Error("Missing uppercase tier name")
		}

		// Log clean output for visual inspection
		t.Logf("Clean table output:\n%s", output)
	})

	t.Run("no incidents scenario", func(t *testing.T) {
		noIncidentSnapshot := snapshot
		noIncidentSnapshot.IncidentCount = 0
		noIncidentSnapshot.MTTRHours = 0
		noIncidentSnapshot.MedianTTRHours = 0
		noIncidentSnapshot.P95TTRHours = 0

		reporter := NewTerminalReporter(false)
		var buf bytes.Buffer

		err := reporter.Report(noIncidentSnapshot, &buf)
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		output := buf.String()

		if !strings.Contains(output, "No incidents") {
			t.Error("Missing 'No incidents' message")
		}
	})
}

func TestJSONReporter(t *testing.T) {
	now := time.Now()
	snapshot := MetricsSnapshot{
		PeriodStart:               now.AddDate(0, 0, -30),
		PeriodEnd:                 now,
		CalculatedAt:              now,
		DeploymentCount:           15,
		DeploymentFrequencyPerDay: 0.5,
		DeploymentTier:            string(TierHigh),
		LeadTimeP50Hours:          6.0,
		LeadTimeP90Hours:          12.0,
		LeadTimeP95Hours:          18.0,
		LeadTimeTier:              string(TierElite),
		FailedDeploymentCount:     1,
		ChangeFailureRate:         6.67,
		FailureRateTier:           string(TierHigh),
		IncidentCount:             1,
		MTTRHours:                 4.0,
		MedianTTRHours:            4.0,
		P95TTRHours:               4.0,
		RestoreTimeTier:           string(TierHigh),
		OverallTier:               string(TierHigh),
		AverageRCsPerRelease:      2.0,
		MedianRCsPerRelease:       2.0,
		TotalCommitsInPeriod:      75,
		AverageCommitsPerRelease:  5.0,
	}

	t.Run("pretty JSON output", func(t *testing.T) {
		reporter := NewJSONReporter(true)
		var buf bytes.Buffer

		err := reporter.Report(snapshot, &buf)
		if err != nil {
			t.Fatalf("Failed to generate JSON: %v", err)
		}

		// Parse JSON to verify structure
		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Check top-level structure
		if _, ok := result["report_metadata"]; !ok {
			t.Error("Missing report_metadata")
		}
		if _, ok := result["dora_metrics"]; !ok {
			t.Error("Missing dora_metrics")
		}
		if _, ok := result["quality_metrics"]; !ok {
			t.Error("Missing quality_metrics")
		}
		if _, ok := result["overall_tier"]; !ok {
			t.Error("Missing overall_tier")
		}

		t.Logf("JSON output:\n%s", buf.String())
	})

	t.Run("compact JSON output", func(t *testing.T) {
		reporter := NewJSONReporter(false)
		var buf bytes.Buffer

		err := reporter.Report(snapshot, &buf)
		if err != nil {
			t.Fatalf("Failed to generate JSON: %v", err)
		}

		// Should be single line (no pretty printing)
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("Expected 1 line, got %d", len(lines))
		}
	})
}
