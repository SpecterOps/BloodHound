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
	"testing"
	"time"
)

func TestParseDefaultPeriod(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// Days
		{"30 days", "30d", 30},
		{"90 days", "90d", 90},
		{"days suffix uppercase", "30D", 30},
		{"days long form", "30days", 30},
		{"days long form uppercase", "30DAYS", 30},
		{"just number", "45", 45},

		// Months (30 days/month)
		{"3 months", "3mo", 90},
		{"6 months", "6mo", 180},
		{"months short", "3m", 90},
		{"months long", "6months", 180},
		{"months uppercase", "3MO", 90},

		// Years (365 days/year)
		{"1 year", "1y", 365},
		{"3 years", "3yr", 1095},
		{"years long", "2years", 730},
		{"years uppercase", "3YR", 1095},

		// Edge cases
		{"empty string", "", 90},
		{"whitespace", "  90d  ", 90},
		{"invalid", "invalid", 90},
		{"negative", "-5d", 90},
		{"zero", "0d", 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDefaultPeriod(tt.input)
			if result != tt.expected {
				t.Errorf("parseDefaultPeriod(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDefaultPeriodRealWorldExamples(t *testing.T) {
	// Real-world usage examples from .dora.yaml
	examples := map[string]struct {
		input    string
		expected int
		desc     string
	}{
		"standard 90 days": {
			input:    "90d",
			expected: 90,
			desc:     "Common quarterly reporting period",
		},
		"3 years for historical": {
			input:    "3yr",
			expected: 1095,
			desc:     "Collect 3 years of data once, report on subsets",
		},
		"6 months trend": {
			input:    "6mo",
			expected: 180,
			desc:     "Semi-annual trend analysis",
		},
		"1 year baseline": {
			input:    "1y",
			expected: 365,
			desc:     "Annual performance baseline",
		},
	}

	for name, ex := range examples {
		t.Run(name, func(t *testing.T) {
			result := parseDefaultPeriod(ex.input)
			if result != ex.expected {
				t.Errorf("%s: parseDefaultPeriod(%q) = %d, want %d",
					ex.desc, ex.input, result, ex.expected)
			}
			t.Logf("✓ %s: %q → %d days", ex.desc, ex.input, result)
		})
	}
}

func TestCalculateLastFiscalQuarter(t *testing.T) {
	tests := []struct {
		name             string
		fiscalStartMonth int
		referenceTime    string // YYYY-MM-DD format
		expectedStart    string // YYYY-MM-DD format
		expectedEnd      string // YYYY-MM-DD 23:59:59 format
	}{
		{
			name:             "January FY, reference in Q2 (April)",
			fiscalStartMonth: 1, // January
			referenceTime:    "2026-04-15",
			expectedStart:    "2026-01-01 00:00:00",
			expectedEnd:      "2026-03-31 23:59:59",
		},
		{
			name:             "January FY, reference in Q3 (September)",
			fiscalStartMonth: 1,
			referenceTime:    "2026-09-30",
			expectedStart:    "2026-04-01 00:00:00",
			expectedEnd:      "2026-06-30 23:59:59",
		},
		{
			name:             "February FY, reference in Q2 (July)",
			fiscalStartMonth: 2,
			referenceTime:    "2026-07-31",
			expectedStart:    "2026-02-01 00:00:00",
			expectedEnd:      "2026-04-30 23:59:59",
		},
		{
			name:             "October FY, reference in Q1 (November)",
			fiscalStartMonth: 10,
			referenceTime:    "2025-11-15",
			expectedStart:    "2025-07-01 00:00:00",
			expectedEnd:      "2025-09-30 23:59:59",
		},
		{
			name:             "October FY, reference in Q2 (February)",
			fiscalStartMonth: 10,
			referenceTime:    "2026-02-28",
			expectedStart:    "2025-10-01 00:00:00",
			expectedEnd:      "2025-12-31 23:59:59",
		},
		{
			name:             "January FY, reference in Q1 (January)",
			fiscalStartMonth: 1,
			referenceTime:    "2026-01-15",
			expectedStart:    "2025-10-01 00:00:00",
			expectedEnd:      "2025-12-31 23:59:59",
		},
		{
			name:             "February FY, reference in Q1 (February)",
			fiscalStartMonth: 2,
			referenceTime:    "2026-02-20",
			expectedStart:    "2025-11-01 00:00:00",
			expectedEnd:      "2026-01-31 23:59:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse reference time
			refTime, err := time.Parse("2006-01-02", tt.referenceTime)
			if err != nil {
				t.Fatalf("Failed to parse reference time: %v", err)
			}

			// Calculate quarter
			start, end := calculateLastFiscalQuarterAt(tt.fiscalStartMonth, refTime)

			// Parse expected times
			expectedStart, err := time.Parse("2006-01-02 15:04:05", tt.expectedStart)
			if err != nil {
				t.Fatalf("Failed to parse expected start: %v", err)
			}
			expectedStart = expectedStart.UTC()

			expectedEnd, err := time.Parse("2006-01-02 15:04:05", tt.expectedEnd)
			if err != nil {
				t.Fatalf("Failed to parse expected end: %v", err)
			}
			expectedEnd = expectedEnd.UTC()

			// Verify start matches
			if !start.Equal(expectedStart) {
				t.Errorf("Start mismatch:\n  got:  %v\n  want: %v",
					start.Format("2006-01-02 15:04:05"), expectedStart.Format("2006-01-02 15:04:05"))
			}

			// Verify end matches
			if !end.Equal(expectedEnd) {
				t.Errorf("End mismatch:\n  got:  %v\n  want: %v",
					end.Format("2006-01-02 15:04:05"), expectedEnd.Format("2006-01-02 15:04:05"))
			}

			// Verify UTC timezone
			if start.Location() != time.UTC {
				t.Errorf("Start time not in UTC: %v", start.Location())
			}
			if end.Location() != time.UTC {
				t.Errorf("End time not in UTC: %v", end.Location())
			}

			// Verify start < end
			if !start.Before(end) {
				t.Errorf("Start (%v) should be before end (%v)", start, end)
			}

			t.Logf("✓ Fiscal start=%s, ref=%s: Q = %v to %v",
				time.Month(tt.fiscalStartMonth), tt.referenceTime,
				start.Format("2006-01-02"), end.Format("2006-01-02"))
		})
	}
}

func TestResolveTimeRange(t *testing.T) {
	referenceTime, err := time.Parse(time.RFC3339, "2026-08-20T10:15:30Z")
	if err != nil {
		t.Fatalf("parsing reference time: %v", err)
	}

	testCases := []struct {
		name             string
		days             int
		startDate        string
		endDate          string
		lastQuarter      bool
		fiscalStartMonth int
		expectedStart    string
		expectedEnd      string
		expectedError    string
	}{
		{
			name:             "days range",
			days:             90,
			fiscalStartMonth: 2,
			expectedStart:    "2026-05-22T10:15:30Z",
			expectedEnd:      "2026-08-20T10:15:30Z",
		},
		{
			name:             "custom range includes the end date",
			startDate:        "2026-02-01",
			endDate:          "2026-04-30",
			fiscalStartMonth: 2,
			expectedStart:    "2026-02-01T00:00:00Z",
			expectedEnd:      "2026-04-30T23:59:59Z",
		},
		{
			name:             "last complete fiscal quarter",
			lastQuarter:      true,
			fiscalStartMonth: 2,
			expectedStart:    "2026-05-01T00:00:00Z",
			expectedEnd:      "2026-07-31T23:59:59Z",
		},
		{
			name:             "last quarter takes precedence over explicit dates",
			startDate:        "2026-01-01",
			endDate:          "2026-01-31",
			lastQuarter:      true,
			fiscalStartMonth: 2,
			expectedStart:    "2026-05-01T00:00:00Z",
			expectedEnd:      "2026-07-31T23:59:59Z",
		},
		{
			name:             "invalid fiscal start month",
			lastQuarter:      true,
			fiscalStartMonth: 13,
			expectedError:    "fiscal start month must be between 1 and 12",
		},
		{
			name:             "invalid custom range",
			startDate:        "2026-05-01",
			endDate:          "2026-04-30",
			fiscalStartMonth: 2,
			expectedError:    "start date (2026-05-01) cannot be after end date (2026-04-30)",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualStart, actualEnd, err := resolveTimeRange(
				testCase.days,
				testCase.startDate,
				testCase.endDate,
				testCase.lastQuarter,
				testCase.fiscalStartMonth,
				referenceTime,
			)
			if testCase.expectedError != "" {
				if err == nil || err.Error() != testCase.expectedError {
					t.Fatalf("resolveTimeRange() error = %v, want %q", err, testCase.expectedError)
				}

				return
			}
			if err != nil {
				t.Fatalf("resolveTimeRange() error = %v", err)
			}

			if actualStart.Format(time.RFC3339) != testCase.expectedStart {
				t.Errorf("start = %s, want %s", actualStart.Format(time.RFC3339), testCase.expectedStart)
			}
			if actualEnd.Format(time.RFC3339) != testCase.expectedEnd {
				t.Errorf("end = %s, want %s", actualEnd.Format(time.RFC3339), testCase.expectedEnd)
			}
		})
	}
}
