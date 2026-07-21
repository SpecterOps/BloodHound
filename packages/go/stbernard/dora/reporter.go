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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Reporter defines the interface for outputting metrics
type Reporter interface {
	Report(snapshot MetricsSnapshot, writer io.Writer) error
}

// TerminalReporter outputs metrics to terminal with formatting
type TerminalReporter struct {
	UseColor bool
}

// NewTerminalReporter creates a new terminal reporter
func NewTerminalReporter(useColor bool) *TerminalReporter {
	return &TerminalReporter{UseColor: useColor}
}

// Report outputs metrics in a formatted terminal display
func (s *TerminalReporter) Report(snapshot MetricsSnapshot, writer io.Writer) error {
	var sb strings.Builder

	// Header
	sb.WriteString(s.bold("\n╔════════════════════════════════════════════════════════════╗\n"))
	sb.WriteString(s.bold("║            DORA Metrics Report                             ║\n"))
	sb.WriteString(s.bold("╚════════════════════════════════════════════════════════════╝\n\n"))

	// Period
	sb.WriteString(s.dim(fmt.Sprintf("Period: %s to %s\n",
		snapshot.PeriodStart.Format("2006-01-02"),
		snapshot.PeriodEnd.Format("2006-01-02"))))
	sb.WriteString(s.dim(fmt.Sprintf("Generated: %s\n\n",
		snapshot.CalculatedAt.Format("2006-01-02 15:04:05"))))

	// Overall Performance
	sb.WriteString(s.bold("Overall Performance: "))
	sb.WriteString(s.colorTier(snapshot.OverallTier))
	sb.WriteString("\n\n")

	// DORA Metrics
	sb.WriteString(s.bold("═══ DORA Metrics ═══\n\n"))

	// 1. Deployment Frequency
	sb.WriteString(s.formatMetric(
		"1. Deployment Frequency",
		fmt.Sprintf("%.2f per day (%d total)", snapshot.DeploymentFrequencyPerDay, snapshot.DeploymentCount),
		snapshot.DeploymentTier,
	))

	// 2. Lead Time for Changes
	leadTimeDesc := fmt.Sprintf("P50: %.1fh | P90: %.1fh | P95: %.1fh",
		snapshot.LeadTimeP50Hours,
		snapshot.LeadTimeP90Hours,
		snapshot.LeadTimeP95Hours)
	sb.WriteString(s.formatMetric(
		"2. Lead Time for Changes",
		leadTimeDesc,
		snapshot.LeadTimeTier,
	))

	// 3. Change Failure Rate
	sb.WriteString(s.formatMetric(
		"3. Change Failure Rate",
		fmt.Sprintf("%.1f%% (%d failures / %d deployments)",
			snapshot.ChangeFailureRate,
			snapshot.FailedDeploymentCount,
			snapshot.DeploymentCount),
		snapshot.FailureRateTier,
	))

	// 4. Time to Restore Service
	if snapshot.IncidentCount > 0 {
		mttrDesc := fmt.Sprintf("Mean: %.1fh | Median: %.1fh | P95: %.1fh (%d incidents)",
			snapshot.MTTRHours,
			snapshot.MedianTTRHours,
			snapshot.P95TTRHours,
			snapshot.IncidentCount)
		sb.WriteString(s.formatMetric(
			"4. Time to Restore Service",
			mttrDesc,
			snapshot.RestoreTimeTier,
		))
	} else {
		sb.WriteString(s.formatMetric(
			"4. Time to Restore Service",
			"No incidents in period 🎉",
			"",
		))
	}

	// Quality Metrics
	if snapshot.TotalCommitsInPeriod > 0 {
		sb.WriteString("\n")
		sb.WriteString(s.bold("═══ Quality Indicators ═══\n\n"))

		sb.WriteString(s.formatQualityMetric(
			"Release Iterations",
			fmt.Sprintf("Avg: %.1f RCs | Median: %.1f RCs",
				snapshot.AverageRCsPerRelease,
				snapshot.MedianRCsPerRelease),
			s.interpretRCs(snapshot.MedianRCsPerRelease),
		))

		sb.WriteString(s.formatQualityMetric(
			"Batch Size",
			fmt.Sprintf("%.1f commits/release (%d total commits)",
				snapshot.AverageCommitsPerRelease,
				snapshot.TotalCommitsInPeriod),
			s.interpretBatchSize(snapshot.AverageCommitsPerRelease),
		))
	}

	sb.WriteString("\n")

	// Write to output
	_, err := writer.Write([]byte(sb.String()))
	return err
}

// formatMetric formats a DORA metric line
func (s *TerminalReporter) formatMetric(name, value, tier string) string {
	var sb strings.Builder
	sb.WriteString(s.bold(name))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("   %s\n", value))
	if tier != "" {
		sb.WriteString(fmt.Sprintf("   Tier: %s\n", s.colorTier(tier)))
	}
	sb.WriteString("\n")
	return sb.String()
}

// formatQualityMetric formats a quality indicator line
func (s *TerminalReporter) formatQualityMetric(name, value, interpretation string) string {
	var sb strings.Builder
	sb.WriteString(s.bold(name))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("   %s\n", value))
	if interpretation != "" {
		sb.WriteString(fmt.Sprintf("   %s\n", s.dim(interpretation)))
	}
	sb.WriteString("\n")
	return sb.String()
}

// colorTier returns tier with color formatting
func (s *TerminalReporter) colorTier(tier string) string {
	if !s.UseColor {
		return strings.ToUpper(tier)
	}

	switch tier {
	case string(TierElite):
		return s.green("🌟 ELITE")
	case string(TierHigh):
		return s.cyan("✓ HIGH")
	case string(TierMedium):
		return s.yellow("○ MEDIUM")
	case string(TierLow):
		return s.red("✗ LOW")
	default:
		return tier
	}
}

// interpretRCs provides guidance on RC counts
func (s *TerminalReporter) interpretRCs(median float64) string {
	if median <= 2 {
		return "Excellent - Predictable releases"
	} else if median <= 4 {
		return "Good - Reasonable iteration"
	} else {
		return "Consider improving testing earlier in cycle"
	}
}

// interpretBatchSize provides guidance on batch sizes
func (s *TerminalReporter) interpretBatchSize(avg float64) string {
	if avg <= 5 {
		return "Excellent - Small, focused changes"
	} else if avg <= 10 {
		return "Good - Moderate batch size"
	} else if avg <= 20 {
		return "Large batches - Consider more frequent releases"
	} else {
		return "Very large batches - High complexity and risk"
	}
}

// ANSI color codes
func (s *TerminalReporter) green(text string) string {
	if !s.UseColor {
		return text
	}
	return fmt.Sprintf("\033[32m%s\033[0m", text)
}

func (s *TerminalReporter) cyan(text string) string {
	if !s.UseColor {
		return text
	}
	return fmt.Sprintf("\033[36m%s\033[0m", text)
}

func (s *TerminalReporter) yellow(text string) string {
	if !s.UseColor {
		return text
	}
	return fmt.Sprintf("\033[33m%s\033[0m", text)
}

func (s *TerminalReporter) red(text string) string {
	if !s.UseColor {
		return text
	}
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

func (s *TerminalReporter) bold(text string) string {
	if !s.UseColor {
		return text
	}
	return fmt.Sprintf("\033[1m%s\033[0m", text)
}

func (s *TerminalReporter) dim(text string) string {
	if !s.UseColor {
		return text
	}
	return fmt.Sprintf("\033[2m%s\033[0m", text)
}

// JSONReporter outputs metrics in JSON format
type JSONReporter struct {
	Pretty bool
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(pretty bool) *JSONReporter {
	return &JSONReporter{Pretty: pretty}
}

// Report outputs metrics as JSON
func (s *JSONReporter) Report(snapshot MetricsSnapshot, writer io.Writer) error {
	// Create output structure with metadata
	output := map[string]any{
		"report_metadata": map[string]any{
			"generated_at": snapshot.CalculatedAt.Format(time.RFC3339),
			"period_start": snapshot.PeriodStart.Format(time.RFC3339),
			"period_end":   snapshot.PeriodEnd.Format(time.RFC3339),
			"version":      "1.0.0",
		},
		"dora_metrics": map[string]any{
			"deployment_frequency": map[string]any{
				"per_day":     snapshot.DeploymentFrequencyPerDay,
				"total_count": snapshot.DeploymentCount,
				"tier":        snapshot.DeploymentTier,
			},
			"lead_time_for_changes": map[string]any{
				"p50_hours": snapshot.LeadTimeP50Hours,
				"p90_hours": snapshot.LeadTimeP90Hours,
				"p95_hours": snapshot.LeadTimeP95Hours,
				"tier":      snapshot.LeadTimeTier,
			},
			"change_failure_rate": map[string]any{
				"percentage":   snapshot.ChangeFailureRate,
				"failed_count": snapshot.FailedDeploymentCount,
				"total_count":  snapshot.DeploymentCount,
				"tier":         snapshot.FailureRateTier,
			},
			"time_to_restore": map[string]any{
				"mean_hours":     snapshot.MTTRHours,
				"median_hours":   snapshot.MedianTTRHours,
				"p95_hours":      snapshot.P95TTRHours,
				"incident_count": snapshot.IncidentCount,
				"tier":           snapshot.RestoreTimeTier,
			},
		},
		"quality_metrics": map[string]any{
			"release_iterations": map[string]any{
				"average_rcs": snapshot.AverageRCsPerRelease,
				"median_rcs":  snapshot.MedianRCsPerRelease,
			},
			"batch_size": map[string]any{
				"total_commits":               snapshot.TotalCommitsInPeriod,
				"average_commits_per_release": snapshot.AverageCommitsPerRelease,
			},
		},
		"overall_tier": snapshot.OverallTier,
	}

	var encoder *json.Encoder
	if s.Pretty {
		encoder = json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
	} else {
		encoder = json.NewEncoder(writer)
	}

	return encoder.Encode(output)
}
