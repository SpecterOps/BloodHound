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
	sb.WriteString(s.bold("\n╔═════════════════════════════════════════════════════════════════════════════╗\n"))
	sb.WriteString(s.bold("║                          DORA Metrics Report                                ║\n"))
	sb.WriteString(s.bold("╚═════════════════════════════════════════════════════════════════════════════╝\n\n"))

	// Period and metadata
	sb.WriteString(s.dim(fmt.Sprintf("  Period: %s to %s  |  Generated: %s\n\n",
		snapshot.PeriodStart.Format("2006-01-02"),
		snapshot.PeriodEnd.Format("2006-01-02"),
		snapshot.CalculatedAt.Format("2006-01-02 15:04:05"))))

	// Overall Performance
	sb.WriteString(s.bold("  Overall Performance: "))
	sb.WriteString(s.colorTier(snapshot.OverallTier))
	sb.WriteString("\n\n")

	// DORA Metrics Table
	sb.WriteString(s.bold("━━━ DORA Metrics ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))
	sb.WriteString(s.renderDORATable(snapshot))

	// Quality Metrics Table
	if snapshot.TotalCommitsInPeriod > 0 {
		sb.WriteString("\n")
		sb.WriteString(s.bold("━━━ Quality Indicators ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))
		sb.WriteString(s.renderQualityTable(snapshot))
	}

	sb.WriteString("\n")

	// Write to output
	_, err := writer.Write([]byte(sb.String()))
	return err
}

// renderDORATable creates a formatted table for DORA metrics
func (s *TerminalReporter) renderDORATable(snapshot MetricsSnapshot) string {
	var sb strings.Builder

	// Table header
	sb.WriteString("  ┌────────────────────────────────┬──────────────────────────────────┬──────────┐\n")
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %-8s │\n",
		s.bold("Metric"),
		s.bold("Value"),
		s.bold("Tier")))
	sb.WriteString("  ├────────────────────────────────┼──────────────────────────────────┼──────────┘\n")

	// Row 1: Deployment Frequency
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Deployment Frequency",
		fmt.Sprintf("%.2f per day (%d total)", snapshot.DeploymentFrequencyPerDay, snapshot.DeploymentCount),
		s.colorTier(snapshot.DeploymentTier)))

	// Row 2: Lead Time
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Lead Time for Changes",
		fmt.Sprintf("P50: %.1fh  P90: %.1fh  P95: %.1fh", snapshot.LeadTimeP50Hours, snapshot.LeadTimeP90Hours, snapshot.LeadTimeP95Hours),
		s.colorTier(snapshot.LeadTimeTier)))

	// Row 3: Change Failure Rate
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Change Failure Rate",
		fmt.Sprintf("%.1f%% (%d/%d deployments)", snapshot.ChangeFailureRate, snapshot.FailedDeploymentCount, snapshot.DeploymentCount),
		s.colorTier(snapshot.FailureRateTier)))

	// Row 4: Time to Restore
	var mttrValue string
	var mttrTier string
	if snapshot.IncidentCount > 0 {
		mttrValue = fmt.Sprintf("Median: %.1fh  Mean: %.1fh", snapshot.MedianTTRHours, snapshot.MTTRHours)
		mttrTier = s.colorTier(snapshot.RestoreTimeTier)
	} else {
		mttrValue = "No incidents 🎉"
		mttrTier = s.dim("N/A")
	}
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Time to Restore Service",
		mttrValue,
		mttrTier))

	// Table footer
	sb.WriteString("  └────────────────────────────────┴──────────────────────────────────┴──────────┘\n")

	return sb.String()
}

// renderQualityTable creates a formatted table for quality metrics
func (s *TerminalReporter) renderQualityTable(snapshot MetricsSnapshot) string {
	var sb strings.Builder

	// Table header
	sb.WriteString("  ┌────────────────────────────────┬──────────────────────────────────┬──────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %-20s │\n",
		s.bold("Indicator"),
		s.bold("Value"),
		s.bold("Assessment")))
	sb.WriteString("  ├────────────────────────────────┼──────────────────────────────────┼──────────────────────┘\n")

	// Row 1: Release Iterations (RCs)
	rcValue := fmt.Sprintf("Avg: %.1f  Median: %.1f", snapshot.AverageRCsPerRelease, snapshot.MedianRCsPerRelease)
	rcAssessment := s.assessRCs(snapshot.MedianRCsPerRelease)
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Release Iterations (RCs)",
		rcValue,
		rcAssessment))

	// Row 2: RC Stabilization (commits in RC2+)
	if snapshot.AverageStabilizationCommits > 0 {
		stabValue := fmt.Sprintf("Avg: %.1f  Median: %.1f per RC",
			snapshot.AverageStabilizationCommits, snapshot.MedianStabilizationCommits)
		stabAssessment := s.assessStabilizationCommits(snapshot.MedianStabilizationCommits)
		sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
			"RC Stabilization (RC2+)",
			stabValue,
			stabAssessment))
	}

	// Row 3: Batch Size (total commits per release)
	batchValue := fmt.Sprintf("%.1f commits/release", snapshot.AverageCommitsPerRelease)
	batchAssessment := s.assessBatchSize(snapshot.AverageCommitsPerRelease)
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Batch Size (total)",
		batchValue,
		batchAssessment))

	// Row 4: Total Activity
	sb.WriteString(fmt.Sprintf("  │ %-30s │ %-32s │ %s\n",
		"Total Commits",
		fmt.Sprintf("%d in period", snapshot.TotalCommitsInPeriod),
		s.dim("—")))

	// Table footer
	sb.WriteString("  └────────────────────────────────┴──────────────────────────────────┴──────────────────────┘\n")

	return sb.String()
}

// assessRCs provides short assessment of RC counts for table display
// RC1 is expected (proposed release), RC2+ indicates stabilization/rework needed
func (s *TerminalReporter) assessRCs(median float64) string {
	if median <= 1 {
		return s.green("✓ Elite (minimal rework)")
	} else if median <= 2 {
		return s.cyan("○ Excellent")
	} else if median <= 4 {
		return s.yellow("△ Good (some rework)")
	} else {
		return s.red("✗ High rework needed")
	}
}

// assessStabilizationCommits provides short assessment of RC stabilization effort
func (s *TerminalReporter) assessStabilizationCommits(median float64) string {
	if median <= 2 {
		return s.green("✓ Minimal fixes")
	} else if median <= 5 {
		return s.cyan("○ Some fixes")
	} else if median <= 10 {
		return s.yellow("△ Many fixes")
	} else {
		return s.red("✗ Extensive rework")
	}
}

// assessBatchSize provides short assessment of batch sizes for table display
func (s *TerminalReporter) assessBatchSize(avg float64) string {
	if avg <= 5 {
		return s.green("✓ Excellent")
	} else if avg <= 10 {
		return s.cyan("○ Good")
	} else if avg <= 20 {
		return s.yellow("△ Large")
	} else {
		return s.red("✗ Very large")
	}
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
