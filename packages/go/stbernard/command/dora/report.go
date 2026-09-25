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

// parseDefaultPeriod converts a period string to number of days
// Supports:
//   - Days: "30d", "90d", "30", "90days"
//   - Months: "3m", "6mo", "6months" (assumes 30 days/month)
//   - Years: "1y", "3yr", "3years" (assumes 365 days/year)
//
// Falls back to 90 days if parsing fails.
func parseDefaultPeriod(period string) int {
	period = strings.TrimSpace(period)
	if period == "" {
		return 90
	}

	// Detect suffix and multiplier
	var (
		multiplier int
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
		return 90 // Fallback to 90 days
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

	// Validate format flag early
	switch formatFlag {
	case "terminal", "json":
		// Valid formats
	default:
		return fmt.Errorf("unsupported format: %s (supported: terminal, json)", formatFlag)
	}

	startTime, endTime, err := resolveTimeRange(
		daysFlag,
		startFlag,
		endFlag,
		lastQuarterFlag,
		fiscalStartFlag,
		time.Now(),
	)
	if err != nil {
		return err
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

	// Generate report based on format (validated earlier)
	var reporter dora.Reporter
	if formatFlag == "json" {
		reporter = dora.NewJSONReporter(true) // Pretty print by default
	} else {
		reporter = dora.NewTerminalReporter(!noColorFlag)
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
