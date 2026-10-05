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
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/specterops/bloodhound/packages/go/chow/payload"
)

type durationSummary struct {
	Avg time.Duration
	Min time.Duration
	Max time.Duration
}

type benchmarkResult struct {
	File             string
	Bytes            int64
	Runs             int
	Status           string
	Error            string
	CriticalErrors   int
	ValidationErrors int
	Durations        durationSummary
}

func main() {
	var (
		runs   int
		warmup int
		strict bool
	)

	flag.IntVar(&runs, "runs", 3, "number of measured validation runs per file")
	flag.IntVar(&warmup, "warmup", 1, "number of unmeasured warmup validation runs per file")
	flag.BoolVar(&strict, "strict", false, "exit non-zero when a file fails validation")
	flag.Parse()

	if err := run(os.Stdout, flag.Args(), runs, warmup, strict); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(w io.Writer, files []string, runs int, warmup int, strict bool) error {
	if runs < 1 {
		return fmt.Errorf("-runs must be greater than 0")
	}
	if warmup < 0 {
		return fmt.Errorf("-warmup must be 0 or greater")
	}
	if len(files) == 0 {
		return fmt.Errorf("usage: chowbench [-runs N] [-warmup N] [-strict] file [file...]")
	}

	schema, err := payload.LoadSchema()
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}

	results := make([]benchmarkResult, 0, len(files))
	for _, file := range files {
		result := benchmarkFile(file, schema, runs, warmup)
		results = append(results, result)
	}

	if err := writeResults(w, results); err != nil {
		return err
	}
	return exitErrorForResults(results, strict)
}

func benchmarkFile(file string, schema payload.Schema, runs int, warmup int) benchmarkResult {
	result := benchmarkResult{
		File: file,
		Runs: runs,
	}

	if stat, err := os.Stat(file); err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	} else {
		result.Bytes = stat.Size()
	}

	for i := 0; i < warmup; i++ {
		_, _ = validateFile(file, schema)
	}

	durations := make([]time.Duration, 0, runs)
	for i := 0; i < runs; i++ {
		start := time.Now()
		report, err := validateFile(file, schema)
		durations = append(durations, time.Since(start))

		result.Status, result.Error = statusForValidationResult(report, err)
		result.CriticalErrors = len(report.CriticalErrors)
		result.ValidationErrors = len(report.ValidationErrors)
	}

	result.Durations = summarizeDurations(durations)
	return result
}

func validateFile(file string, schema payload.Schema) (payload.ValidationReport, error) {
	reader, err := os.Open(file)
	if err != nil {
		return payload.ValidationReport{}, err
	}
	defer reader.Close()

	validator := payload.NewValidator(reader, schema)
	_, report, err := validator.ParseAndValidate()
	return report, err
}

func summarizeDurations(durations []time.Duration) durationSummary {
	if len(durations) == 0 {
		return durationSummary{}
	}

	var total time.Duration
	summary := durationSummary{
		Min: durations[0],
		Max: durations[0],
	}

	for _, duration := range durations {
		total += duration
		if duration < summary.Min {
			summary.Min = duration
		}
		if duration > summary.Max {
			summary.Max = duration
		}
	}

	summary.Avg = total / time.Duration(len(durations))
	return summary
}

func statusForValidationResult(report payload.ValidationReport, err error) (string, string) {
	if err == nil {
		return "ok", ""
	}

	if len(report.CriticalErrors) > 0 {
		return "critical_error", err.Error()
	}

	if len(report.ValidationErrors) > 0 ||
		errors.Is(err, payload.ErrValidationErrors) ||
		errors.Is(err, payload.ErrMaxValidationErrors) {
		return "validation_error", err.Error()
	}

	return "error", err.Error()
}

func exitErrorForResults(results []benchmarkResult, strict bool) error {
	var hasValidationFailure bool
	for _, result := range results {
		switch result.Status {
		case "error":
			return fmt.Errorf("one or more files could not be benchmarked")
		case "validation_error", "critical_error":
			hasValidationFailure = true
		}
	}

	if strict && hasValidationFailure {
		return fmt.Errorf("one or more files failed validation")
	}

	return nil
}

func writeResults(w io.Writer, results []benchmarkResult) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "file\tbytes\truns\tstatus\tavg\tmin\tmax\tcritical\tvalidation\terror"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%d\t%d\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
			result.File,
			result.Bytes,
			result.Runs,
			result.Status,
			result.Durations.Avg,
			result.Durations.Min,
			result.Durations.Max,
			result.CriticalErrors,
			result.ValidationErrors,
			result.Error,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
