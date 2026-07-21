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

// parseDefaultPeriod converts a period string (e.g., "30d", "90d") to number of days
func parseDefaultPeriod(period string) int {
	period = strings.TrimSpace(period)
	if period == "" {
		return 30
	}

	// Remove 'd' suffix if present
	period = strings.TrimSuffix(period, "d")
	period = strings.TrimSuffix(period, "D")

	days, err := strconv.Atoi(period)
	if err != nil || days <= 0 {
		return 30 // Fallback to default
	}

	return days
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
		cmd         = flag.NewFlagSet("dora report", flag.ExitOnError)
		daysFlag    int
		startFlag   string
		endFlag     string
		formatFlag  string
		outputFlag  string
		noColorFlag bool
	)

	cmd.IntVar(&daysFlag, "days", defaultDays, fmt.Sprintf("Number of days back from now (default: %s from config)", config.Metrics.DefaultPeriod))
	cmd.StringVar(&startFlag, "start", "", "Start date (YYYY-MM-DD) - overrides -days")
	cmd.StringVar(&endFlag, "end", "", "End date (YYYY-MM-DD) - defaults to now")
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
		fmt.Fprintf(w, "  1. Use -days for recent period (default)\n")
		fmt.Fprintf(w, "  2. Use -start and -end for specific date range\n")
		fmt.Fprintf(w, "  3. Use -start alone for period from date to now\n")
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Last 90 days (simple)\n")
		fmt.Fprintf(w, "  %s dora report -days 90\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Specific quarter (Q1 2024)\n")
		fmt.Fprintf(w, "  %s dora report -start 2024-01-01 -end 2024-03-31\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Last 90 days of 2023\n")
		fmt.Fprintf(w, "  %s dora report -start 2023-10-03 -end 2023-12-31\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # From specific date to now\n")
		fmt.Fprintf(w, "  %s dora report -start 2024-01-01\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Export Q4 2023 as JSON\n")
		fmt.Fprintf(w, "  %s dora report -start 2023-10-01 -end 2023-12-31 -format json -output q4-2023.json\n\n", filepath.Base(os.Args[0]))
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

	if startFlag != "" {
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
