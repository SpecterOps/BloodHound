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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// runTrends handles the trends subcommand
func (s *command) runTrends() error {
	var (
		cmd        = flag.NewFlagSet("dora trends", flag.ExitOnError)
		yearFlag   int
		periodFlag string
		outputFlag string
	)

	cmd.IntVar(&yearFlag, "year", time.Now().Year(), "Year to generate reports for")
	cmd.StringVar(&periodFlag, "period", "quarters", "Period type (quarters, months)")
	cmd.StringVar(&outputFlag, "output", "dora-trends", "Output directory for reports")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Generate DORA metrics trend reports for comparative analysis\n\n")
		fmt.Fprintf(w, "Usage: %s dora trends [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "Options:\n")
		cmd.PrintDefaults()
		fmt.Fprintf(w, "\nPeriod Types:\n")
		fmt.Fprintf(w, "  quarters - Generate Q1, Q2, Q3, Q4 reports\n")
		fmt.Fprintf(w, "  months   - Generate reports for each month\n")
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Generate quarterly reports for 2024\n")
		fmt.Fprintf(w, "  %s dora trends -year 2024 -period quarters\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Generate monthly reports for 2025\n")
		fmt.Fprintf(w, "  %s dora trends -year 2025 -period months\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Custom output directory\n")
		fmt.Fprintf(w, "  %s dora trends -year 2024 -output reports/2024-quarters\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "\nOutput:\n")
		fmt.Fprintf(w, "  Creates JSON files in the output directory:\n")
		fmt.Fprintf(w, "    quarters: 2024-Q1.json, 2024-Q2.json, 2024-Q3.json, 2024-Q4.json\n")
		fmt.Fprintf(w, "    months:   2024-01.json, 2024-02.json, ..., 2024-12.json\n")
		fmt.Fprintf(w, "\nAnalysis:\n")
		fmt.Fprintf(w, "  Use jq to compare metrics across periods:\n")
		fmt.Fprintf(w, "    jq '.dora_metrics.deployment_frequency.per_day' %s/2024-*.json\n", outputFlag)
		fmt.Fprintf(w, "    jq '.dora_metrics.change_failure_rate.percentage' %s/2024-*.json\n", outputFlag)
	}

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing trends flags: %w", err)
		}
	}

	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	// Load configuration
	config, err := dora.LoadConfig(paths.Root)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Create storage
	storagePath := config.Storage.Path
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(paths.Root, storagePath)
	}

	storage, err := dora.NewStorage(storagePath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	calculator := dora.NewCalculator(storage)
	ctx := context.Background()

	// Create output directory
	outputDir := outputFlag
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(paths.Root, outputDir)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Generate reports based on period type
	switch periodFlag {
	case "quarters":
		return s.generateQuarterlyReports(ctx, calculator, yearFlag, outputDir)
	case "months":
		return s.generateMonthlyReports(ctx, calculator, yearFlag, outputDir)
	default:
		return fmt.Errorf("invalid period type: %s (use 'quarters' or 'months')", periodFlag)
	}
}

// generateQuarterlyReports generates Q1-Q4 reports for a given year
func (s *command) generateQuarterlyReports(ctx context.Context, calc *dora.Calculator, year int, outputDir string) error {
	quarters := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{"Q1", time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"Q2", time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 6, 30, 23, 59, 59, 0, time.UTC)},
		{"Q3", time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 9, 30, 23, 59, 59, 0, time.UTC)},
		{"Q4", time.Date(year, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)},
	}

	fmt.Printf("Generating quarterly reports for %d...\n\n", year)

	for _, q := range quarters {
		if err := s.generatePeriodReport(ctx, calc, fmt.Sprintf("%d-%s", year, q.name), q.start, q.end, outputDir); err != nil {
			return err
		}
	}

	fmt.Printf("\n✅ All quarterly reports generated in %s/\n", outputDir)
	return nil
}

// generateMonthlyReports generates monthly reports for a given year
func (s *command) generateMonthlyReports(ctx context.Context, calc *dora.Calculator, year int, outputDir string) error {
	fmt.Printf("Generating monthly reports for %d...\n\n", year)

	for month := 1; month <= 12; month++ {
		startTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		// Last day of month
		endTime := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

		periodName := fmt.Sprintf("%d-%02d", year, month)
		if err := s.generatePeriodReport(ctx, calc, periodName, startTime, endTime, outputDir); err != nil {
			return err
		}
	}

	fmt.Printf("\n✅ All monthly reports generated in %s/\n", outputDir)
	return nil
}

// generatePeriodReport generates a single report for a time period
func (s *command) generatePeriodReport(ctx context.Context, calc *dora.Calculator, periodName string, startTime, endTime time.Time, outputDir string) error {
	outputFile := filepath.Join(outputDir, periodName+".json")

	fmt.Printf("📊 %s: %s to %s\n", periodName, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	// Calculate metrics
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		return fmt.Errorf("calculating metrics for %s: %w", periodName, err)
	}

	// Create JSON output structure
	output := map[string]any{
		"report_metadata": map[string]any{
			"period":       periodName,
			"generated_at": time.Now().Format(time.RFC3339),
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
				"median_hours":   snapshot.MedianTTRHours,
				"mean_hours":     snapshot.MTTRHours,
				"p95_hours":      snapshot.P95TTRHours,
				"incident_count": snapshot.IncidentCount,
				"tier":           snapshot.RestoreTimeTier,
			},
		},
		"quality_metrics": map[string]any{
			"average_rcs_per_release": snapshot.AverageRCsPerRelease,
			"median_rcs_per_release":  snapshot.MedianRCsPerRelease,
			"average_stabilization":   snapshot.AverageStabilizationCommits,
			"median_stabilization":    snapshot.MedianStabilizationCommits,
			"average_commits":         snapshot.AverageCommitsPerRelease,
			"total_commits":           snapshot.TotalCommitsInPeriod,
		},
		"overall_tier": snapshot.OverallTier,
	}

	// Write JSON file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputFile, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON for %s: %w", periodName, err)
	}

	// Print summary
	fmt.Printf("   ✓ %d deployments, %.2f/day, %.1f%% failure, %.1fh MTTR\n",
		snapshot.DeploymentCount,
		snapshot.DeploymentFrequencyPerDay,
		snapshot.ChangeFailureRate,
		snapshot.MedianTTRHours,
	)
	fmt.Printf("   📁 Saved to %s\n\n", outputFile)

	return nil
}
