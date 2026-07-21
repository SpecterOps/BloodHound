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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// calculateLastFiscalQuarter returns the start and end dates of the most recent complete fiscal quarter
// based on the fiscal year start month (1=Jan, 2=Feb, etc.)
func calculateLastFiscalQuarter(fiscalStartMonth int) (time.Time, time.Time) {
	var (
		now          = time.Now()
		currentMonth = int(now.Month())
		currentYear  = now.Year()
	)

	// Calculate which quarter we're currently in based on fiscal start
	// and determine the previous complete quarter
	var quarterStartMonth int
	var startYear, endYear int

	// Normalize current month relative to fiscal year start
	// e.g., if fiscal starts in Feb (2), and we're in Apr (4), we're 2 months into FY
	monthsIntoFY := currentMonth - fiscalStartMonth
	if monthsIntoFY < 0 {
		monthsIntoFY += 12
	}

	// Determine which quarter just completed
	// Quarters are 3 months each: Q1 (0-2), Q2 (3-5), Q3 (6-8), Q4 (9-11)
	completedQuarter := (monthsIntoFY - 1) / 3 // -1 because we want the *completed* quarter

	if completedQuarter < 0 {
		// We're in the first quarter of the FY, so last complete quarter is Q4 of previous FY
		completedQuarter = 3
		if fiscalStartMonth == 1 {
			startYear = currentYear - 1
		} else if currentMonth < fiscalStartMonth {
			startYear = currentYear - 1
		} else {
			startYear = currentYear
		}
	} else {
		// We're past Q1, so the completed quarter is in the current FY
		if currentMonth < fiscalStartMonth {
			startYear = currentYear - 1
		} else {
			startYear = currentYear
		}
	}

	// Calculate the start month of the completed quarter
	quarterStartMonth = fiscalStartMonth + (completedQuarter * 3)
	if quarterStartMonth > 12 {
		quarterStartMonth -= 12
		startYear++
	}

	// Start of quarter
	start := time.Date(startYear, time.Month(quarterStartMonth), 1, 0, 0, 0, 0, time.UTC)

	// End of quarter (last day of third month at 23:59:59)
	endMonth := quarterStartMonth + 2
	endYear = startYear
	if endMonth > 12 {
		endMonth -= 12
		endYear++
	}

	// Last second of the last day of the month
	endMonthStart := time.Date(endYear, time.Month(endMonth)+1, 1, 0, 0, 0, 0, time.UTC)
	end := endMonthStart.Add(-time.Second)

	return start, end
}

// parseDefaultPeriod converts a period string to number of days
// Supports:
//   - Days: "30d", "90d", "30", "90days"
//   - Months: "3m", "6mo", "6months" (assumes 30 days/month)
//   - Years: "1y", "3yr", "3years" (assumes 365 days/year)
// Falls back to 30 days if parsing fails.
func parseDefaultPeriod(period string) int {
	period = strings.TrimSpace(period)
	if period == "" {
		return 30
	}

	// Detect suffix and multiplier
	var (
		multiplier = 1 // Default: days
		value      string
	)

	periodLower := strings.ToLower(period)

	// Check for year suffixes
	if strings.HasSuffix(periodLower, "yr") || strings.HasSuffix(periodLower, "years") {
		multiplier = 365
		value = strings.TrimSuffix(strings.TrimSuffix(periodLower, "yr"), "years")
	} else if strings.HasSuffix(periodLower, "y") {
		multiplier = 365
		value = strings.TrimSuffix(periodLower, "y")
	} else if strings.HasSuffix(periodLower, "mo") || strings.HasSuffix(periodLower, "months") {
		multiplier = 30
		value = strings.TrimSuffix(strings.TrimSuffix(periodLower, "mo"), "months")
	} else if strings.HasSuffix(periodLower, "m") {
		multiplier = 30
		value = strings.TrimSuffix(periodLower, "m")
	} else {
		// Default: assume days, remove 'd' suffix if present
		value = strings.TrimSuffix(strings.TrimSuffix(periodLower, "d"), "days")
		multiplier = 1
	}

	// Parse the numeric value
	numValue, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || numValue <= 0 {
		return 30 // Fallback to 30 days
	}

	return numValue * multiplier
}

// runReport handles the report subcommand
func (s *command) runReport() error {
	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	// Load configuration to get default period
	config, err := dora.LoadConfig(paths.Root)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Parse default period from config
	defaultDays := parseDefaultPeriod(config.Metrics.DefaultPeriod)

	var (
		cmd             = flag.NewFlagSet("dora report", flag.ExitOnError)
		daysFlag        int
		startFlag       string
		endFlag         string
		formatFlag      string
		outputFlag      string
		noColorFlag     bool
		lastQuarterFlag bool
		fiscalStartFlag int
	)

	cmd.IntVar(&daysFlag, "days", defaultDays, fmt.Sprintf("Number of days back from now (default: %s from config)", config.Metrics.DefaultPeriod))
	cmd.StringVar(&startFlag, "start", "", "Start date (YYYY-MM-DD) - overrides -days")
	cmd.StringVar(&endFlag, "end", "", "End date (YYYY-MM-DD) - defaults to now")
	cmd.BoolVar(&lastQuarterFlag, "last-quarter", false, "Report on the last complete fiscal quarter")
	cmd.IntVar(&fiscalStartFlag, "fiscal-start", 2, "Fiscal year start month (1=Jan, 2=Feb, etc.) - used with -last-quarter")
	cmd.StringVar(&formatFlag, "format", "terminal", "Output format (terminal, json)")
	cmd.StringVar(&outputFlag, "output", "", "Output file (default: stdout)")
	cmd.BoolVar(&noColorFlag, "no-color", false, "Disable color output for terminal format")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Generate DORA metrics report\n\n")
		fmt.Fprintf(w, "Usage: %s dora report [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "Options:\n")
		cmd.PrintDefaults()
		fmt.Fprintf(w, "\nTime Range Options:\n")
		fmt.Fprintf(w, "  1. Use -last-quarter for most recent complete fiscal quarter\n")
		fmt.Fprintf(w, "  2. Use -days for recent period (default)\n")
		fmt.Fprintf(w, "  3. Use -start and -end for specific date range\n")
		fmt.Fprintf(w, "  4. Use -start alone for period from date to now\n")
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Last complete fiscal quarter (automatic)\n")
		fmt.Fprintf(w, "  %s dora report -last-quarter\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Last 90 days (simple)\n")
		fmt.Fprintf(w, "  %s dora report -days 90\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Specific quarter (Q1 2024)\n")
		fmt.Fprintf(w, "  %s dora report -start 2024-01-01 -end 2024-03-31\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # From specific date to now\n")
		fmt.Fprintf(w, "  %s dora report -start 2024-01-01\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Export last quarter as JSON\n")
		fmt.Fprintf(w, "  %s dora report -last-quarter -format json -output last-quarter.json\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Terminal report without colors (CI-friendly)\n")
		fmt.Fprintf(w, "  %s dora report -no-color\n\n", filepath.Base(os.Args[0]))
	}

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing report flags: %w", err)
		}
	}

	// Calculate time range based on flags
	var startTime, endTime time.Time

	if lastQuarterFlag {
		// Calculate the last complete fiscal quarter
		startTime, endTime = calculateLastFiscalQuarter(fiscalStartFlag)
	} else if startFlag != "" {
		// Parse start date
		parsedStart, err := time.Parse("2006-01-02", startFlag)
		if err != nil {
			return fmt.Errorf("invalid start date format (use YYYY-MM-DD): %w", err)
		}
		startTime = parsedStart

		// Parse end date or default to now
		if endFlag != "" {
			parsedEnd, err := time.Parse("2006-01-02", endFlag)
			if err != nil {
				return fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
			}
			// Set to end of day (23:59:59)
			endTime = parsedEnd.Add(24*time.Hour - time.Second)
		} else {
			endTime = time.Now()
		}
	} else {
		// Use days-based calculation (default)
		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -daysFlag)
	}

	// Validate time range
	if startTime.After(endTime) {
		return fmt.Errorf("start date (%s) cannot be after end date (%s)",
			startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
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

	ctx := context.Background()

	// Calculate metrics
	calculator := dora.NewCalculator(storage)
	snapshot, err := calculator.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		return fmt.Errorf("calculating metrics: %w", err)
	}

	// Determine output writer
	writer := os.Stdout
	if outputFlag != "" {
		file, err := os.Create(outputFlag)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// Generate report based on format
	var reporter dora.Reporter
	switch formatFlag {
	case "terminal":
		reporter = dora.NewTerminalReporter(!noColorFlag)
	case "json":
		reporter = dora.NewJSONReporter(true) // Pretty print by default
	default:
		return fmt.Errorf("unsupported format: %s (supported: terminal, json)", formatFlag)
	}

	if err := reporter.Report(snapshot, writer); err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Show success message for file output
	if outputFlag != "" {
		fmt.Fprintf(os.Stderr, "✅ Report written to %s\n", outputFlag)
	}

	return nil
}
