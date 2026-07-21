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
		cmd        = flag.NewFlagSet("dora report", flag.ExitOnError)
		daysFlag   int
		formatFlag string
		outputFlag string
		noColorFlag bool
	)

	cmd.IntVar(&daysFlag, "days", defaultDays, fmt.Sprintf("Number of days to report on (default: %s from config)", config.Metrics.DefaultPeriod))
	cmd.StringVar(&formatFlag, "format", "terminal", "Output format (terminal, json)")
	cmd.StringVar(&outputFlag, "output", "", "Output file (default: stdout)")
	cmd.BoolVar(&noColorFlag, "no-color", false, "Disable color output for terminal format")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Generate DORA metrics report\n\n")
		fmt.Fprintf(w, "Usage: %s dora report [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "Options:\n")
		cmd.PrintDefaults()
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Generate terminal report for last 30 days\n")
		fmt.Fprintf(w, "  %s dora report\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Generate report for last 90 days\n")
		fmt.Fprintf(w, "  %s dora report -days 90\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Export as JSON\n")
		fmt.Fprintf(w, "  %s dora report -format json -output metrics.json\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Terminal report without colors (CI-friendly)\n")
		fmt.Fprintf(w, "  %s dora report -no-color\n\n", filepath.Base(os.Args[0]))
	}

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing report flags: %w", err)
		}
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

	// Calculate time range
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -daysFlag)

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
