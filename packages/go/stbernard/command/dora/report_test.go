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
