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
	"strings"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// periodResult holds the results for a single period
type periodResult struct {
	Name               string
	StartTime          time.Time
	EndTime            time.Time
	DeploymentCount    int
	FrequencyPerDay    float64
	LeadTimeP50        float64
	FailureRate        float64
	MTTR               float64
	IncidentCount      int
	DeploymentTier     string
	LeadTimeTier       string
	FailureRateTier    string
	RestoreTimeTier    string
	OverallTier        string
	HasData            bool
}

// runTrends handles the trends subcommand
func (s *command) runTrends() error {
	var (
		cmd        = flag.NewFlagSet("dora trends", flag.ExitOnError)
		yearsFlag  string
		periodFlag string
		outputFlag string
	)

	cmd.StringVar(&yearsFlag, "years", fmt.Sprintf("%d", time.Now().Year()), "Years to generate reports for (single year, comma-separated, or 'all' for all available data)")
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
		fmt.Fprintf(w, "\nYear Specification:\n")
		fmt.Fprintf(w, "  Single:   -years 2024\n")
		fmt.Fprintf(w, "  Multiple: -years 2024,2025,2026\n")
		fmt.Fprintf(w, "  Range:    -years 2024-2026\n")
		fmt.Fprintf(w, "  All data: -years all\n")
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Generate quarterly reports for single year\n")
		fmt.Fprintf(w, "  %s dora trends -years 2024 -period quarters\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Generate for multiple years (one table)\n")
		fmt.Fprintf(w, "  %s dora trends -years 2024,2025,2026 -period quarters\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Generate for year range\n")
		fmt.Fprintf(w, "  %s dora trends -years 2024-2026 -period quarters\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Generate for all available data\n")
		fmt.Fprintf(w, "  %s dora trends -years all -period quarters\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Monthly reports for multiple years\n")
		fmt.Fprintf(w, "  %s dora trends -years 2024,2025 -period months\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "\nOutput:\n")
		fmt.Fprintf(w, "  Creates JSON files in the output directory:\n")
		fmt.Fprintf(w, "    quarters: 2024-Q1.json, 2024-Q2.json, 2025-Q1.json, ...\n")
		fmt.Fprintf(w, "    months:   2024-01.json, 2024-02.json, ..., 2025-12.json\n")
		fmt.Fprintf(w, "  Multiple years shown in single comparison table\n")
		fmt.Fprintf(w, "\nAnalysis:\n")
		fmt.Fprintf(w, "  Use jq to compare metrics across periods:\n")
		fmt.Fprintf(w, "    jq '.dora_metrics.deployment_frequency.per_day' %s/*.json\n", outputFlag)
		fmt.Fprintf(w, "    jq '.dora_metrics.change_failure_rate.percentage' %s/*.json\n", outputFlag)
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

	// Parse years
	years, err := s.parseYears(yearsFlag, ctx, storage)
	if err != nil {
		return fmt.Errorf("parsing years: %w", err)
	}

	if len(years) == 0 {
		return fmt.Errorf("no years specified or found in data")
	}

	// Generate reports based on period type
	switch periodFlag {
	case "quarters":
		return s.generateQuarterlyReportsMultiYear(ctx, calculator, years, outputDir)
	case "months":
		return s.generateMonthlyReportsMultiYear(ctx, calculator, years, outputDir)
	default:
		return fmt.Errorf("invalid period type: %s (use 'quarters' or 'months')", periodFlag)
	}
}

// parseYears parses the years flag and returns a list of years to process
func (s *command) parseYears(yearsFlag string, ctx context.Context, storage *dora.Storage) ([]int, error) {
	yearsFlag = strings.TrimSpace(yearsFlag)

	// Special case: "all" means all years with data
	if yearsFlag == "all" {
		return s.getAllYearsWithData(ctx, storage)
	}

	var years []int

	// Check for range (e.g., "2024-2026")
	if strings.Contains(yearsFlag, "-") {
		parts := strings.Split(yearsFlag, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid year range: %s (use format: 2024-2026)", yearsFlag)
		}

		startYear, err := time.Parse("2006", strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start year: %s", parts[0])
		}

		endYear, err := time.Parse("2006", strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end year: %s", parts[1])
		}

		for year := startYear.Year(); year <= endYear.Year(); year++ {
			years = append(years, year)
		}

		return years, nil
	}

	// Check for comma-separated list
	if strings.Contains(yearsFlag, ",") {
		parts := strings.Split(yearsFlag, ",")
		for _, part := range parts {
			year, err := time.Parse("2006", strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("invalid year: %s", part)
			}
			years = append(years, year.Year())
		}
		return years, nil
	}

	// Single year
	year, err := time.Parse("2006", yearsFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid year: %s", yearsFlag)
	}

	return []int{year.Year()}, nil
}

// getAllYearsWithData finds all years that have deployment data
func (s *command) getAllYearsWithData(ctx context.Context, storage *dora.Storage) ([]int, error) {
	// Get all deployments to find the year range
	deployments, err := storage.GetDeployments(ctx, time.Time{}, time.Now())
	if err != nil {
		return nil, fmt.Errorf("getting deployments: %w", err)
	}

	if len(deployments) == 0 {
		return nil, fmt.Errorf("no deployments found in database")
	}

	// Find min and max years
	yearSet := make(map[int]bool)
	for _, d := range deployments {
		if d.IsProduction {
			yearSet[d.DeployedAt.Year()] = true
		}
	}

	var years []int
	for year := range yearSet {
		years = append(years, year)
	}

	// Sort years
	for i := 0; i < len(years)-1; i++ {
		for j := i + 1; j < len(years); j++ {
			if years[i] > years[j] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	return years, nil
}

// generateQuarterlyReportsMultiYear generates quarterly reports for multiple years
func (s *command) generateQuarterlyReportsMultiYear(ctx context.Context, calc *dora.Calculator, years []int, outputDir string) error {
	if len(years) == 1 {
		fmt.Printf("Generating quarterly reports for %d...\n\n", years[0])
	} else {
		fmt.Printf("Generating quarterly reports for %d-%d...\n\n", years[0], years[len(years)-1])
	}

	var allResults []periodResult

	for _, year := range years {
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

		for _, q := range quarters {
			result, err := s.generatePeriodReport(ctx, calc, fmt.Sprintf("%d-%s", year, q.name), q.start, q.end, outputDir)
			if err != nil {
				return err
			}

			// Only include periods with data
			if result.HasData {
				allResults = append(allResults, result)
			}
		}
	}

	// Print summary table
	if len(allResults) > 0 {
		s.printResultsTable(allResults, "Quarterly")
		fmt.Printf("\n✅ Generated %d quarterly reports in %s/\n", len(allResults), outputDir)
	} else {
		fmt.Printf("\n⚠️  No data found for specified years\n")
	}

	return nil
}

// generateMonthlyReportsMultiYear generates monthly reports for multiple years
func (s *command) generateMonthlyReportsMultiYear(ctx context.Context, calc *dora.Calculator, years []int, outputDir string) error {
	if len(years) == 1 {
		fmt.Printf("Generating monthly reports for %d...\n\n", years[0])
	} else {
		fmt.Printf("Generating monthly reports for %d-%d...\n\n", years[0], years[len(years)-1])
	}

	var allResults []periodResult

	for _, year := range years {
		for month := 1; month <= 12; month++ {
			startTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			// Last day of month
			endTime := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

			periodName := fmt.Sprintf("%d-%02d", year, month)
			result, err := s.generatePeriodReport(ctx, calc, periodName, startTime, endTime, outputDir)
			if err != nil {
				return err
			}

			// Only include months with data
			if result.HasData {
				allResults = append(allResults, result)
			}
		}
	}

	// Print summary table
	if len(allResults) > 0 {
		s.printResultsTable(allResults, "Monthly")
		fmt.Printf("\n✅ Generated %d monthly reports in %s/\n", len(allResults), outputDir)
	} else {
		fmt.Printf("\n⚠️  No data found for specified years\n")
	}

	return nil
}

// generatePeriodReport generates a single report for a time period
func (s *command) generatePeriodReport(ctx context.Context, calc *dora.Calculator, periodName string, startTime, endTime time.Time, outputDir string) (periodResult, error) {
	result := periodResult{
		Name:      periodName,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Calculate metrics
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		return result, fmt.Errorf("calculating metrics for %s: %w", periodName, err)
	}

	// Check if there's any data in this period
	if snapshot.DeploymentCount == 0 {
		fmt.Printf("⊘  %s: No data (skipped)\n", periodName)
		return result, nil
	}

	result.HasData = true
	result.DeploymentCount = snapshot.DeploymentCount
	result.FrequencyPerDay = snapshot.DeploymentFrequencyPerDay
	result.LeadTimeP50 = snapshot.LeadTimeP50Hours
	result.FailureRate = snapshot.ChangeFailureRate
	result.MTTR = snapshot.MedianTTRHours
	result.IncidentCount = snapshot.IncidentCount
	result.DeploymentTier = snapshot.DeploymentTier
	result.LeadTimeTier = snapshot.LeadTimeTier
	result.FailureRateTier = snapshot.FailureRateTier
	result.RestoreTimeTier = snapshot.RestoreTimeTier
	result.OverallTier = snapshot.OverallTier

	outputFile := filepath.Join(outputDir, periodName+".json")
	fmt.Printf("📊 %s: %s to %s\n", periodName, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

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
		return result, fmt.Errorf("creating output file %s: %w", outputFile, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return result, fmt.Errorf("encoding JSON for %s: %w", periodName, err)
	}

	// Print summary
	fmt.Printf("   ✓ %d deployments, %.2f/day, %.1f%% failure, %.1fh MTTR\n",
		snapshot.DeploymentCount,
		snapshot.DeploymentFrequencyPerDay,
		snapshot.ChangeFailureRate,
		snapshot.MedianTTRHours,
	)
	fmt.Printf("   📁 Saved to %s\n\n", outputFile)

	return result, nil
}

// printResultsTable prints a comparative table of all period results
func (s *command) printResultsTable(results []periodResult, periodType string) {
	if len(results) == 0 {
		return
	}

	fmt.Println()
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%s Comparison\n", periodType)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println()

	// Table header
	fmt.Printf("%-10s │ %8s │ %9s │ %13s │ %12s │ %13s │ %8s\n",
		"Period", "Deploys", "Freq/Day", "Lead (P50)", "Failure %", "MTTR (Med)", "Tier")
	fmt.Println(strings.Repeat("─", 130))

	// Table rows
	for _, r := range results {
		fmt.Printf("%-10s │ %8d │ %9.2f │ %10.1fh │ %11.1f%% │ %10.1fh │ %8s\n",
			r.Name,
			r.DeploymentCount,
			r.FrequencyPerDay,
			r.LeadTimeP50,
			r.FailureRate,
			r.MTTR,
			r.OverallTier,
		)
	}

	fmt.Println()

	// Summary statistics
	var (
		totalDeploys   int
		avgFreq        float64
		avgLeadTime    float64
		avgFailureRate float64
		avgMTTR        float64
	)

	for _, r := range results {
		totalDeploys += r.DeploymentCount
		avgFreq += r.FrequencyPerDay
		avgLeadTime += r.LeadTimeP50
		avgFailureRate += r.FailureRate
		if r.IncidentCount > 0 {
			avgMTTR += r.MTTR
		}
	}

	n := float64(len(results))
	avgFreq /= n
	avgLeadTime /= n
	avgFailureRate /= n
	avgMTTR /= n

	fmt.Println("Summary:")
	fmt.Printf("  Total Deployments: %d across %d periods\n", totalDeploys, len(results))
	fmt.Printf("  Average Frequency: %.2f deployments/day\n", avgFreq)
	fmt.Printf("  Average Lead Time: %.1f hours (%.1f days)\n", avgLeadTime, avgLeadTime/24)
	fmt.Printf("  Average Failure Rate: %.1f%%\n", avgFailureRate)
	fmt.Printf("  Average MTTR: %.1f hours (%.1f days)\n", avgMTTR, avgMTTR/24)
}
