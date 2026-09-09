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
	"strings"
	"testing"
)

// TestInterpretRCs tests release candidate guidance
func TestInterpretRCs(t *testing.T) {
	reporter := &TerminalReporter{UseColor: false}

	tests := []struct {
		name     string
		median   float64
		expected string
	}{
		{"excellent_1rc", 1.0, "Excellent - Predictable releases"},
		{"excellent_2rc", 2.0, "Excellent - Predictable releases"},
		{"good_3rc", 3.0, "Good - Reasonable iteration"},
		{"good_4rc", 4.0, "Good - Reasonable iteration"},
		{"needs_improvement_5rc", 5.0, "Consider improving testing earlier in cycle"},
		{"needs_improvement_10rc", 10.0, "Consider improving testing earlier in cycle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reporter.interpretRCs(tt.median)
			if result != tt.expected {
				t.Errorf("interpretRCs(%.1f) = %s, expected %s", tt.median, result, tt.expected)
			}
		})
	}
}

// TestInterpretBatchSize tests batch size guidance
func TestInterpretBatchSize(t *testing.T) {
	reporter := &TerminalReporter{UseColor: false}

	tests := []struct {
		name     string
		avg      float64
		expected string
	}{
		{"excellent_3commits", 3.0, "Excellent - Small, focused changes"},
		{"excellent_5commits", 5.0, "Excellent - Small, focused changes"},
		{"good_7commits", 7.0, "Good - Moderate batch size"},
		{"good_10commits", 10.0, "Good - Moderate batch size"},
		{"large_15commits", 15.0, "Large batches - Consider more frequent releases"},
		{"large_20commits", 20.0, "Large batches - Consider more frequent releases"},
		{"very_large_25commits", 25.0, "Very large batches - High complexity and risk"},
		{"very_large_50commits", 50.0, "Very large batches - High complexity and risk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reporter.interpretBatchSize(tt.avg)
			if result != tt.expected {
				t.Errorf("interpretBatchSize(%.1f) = %s, expected %s", tt.avg, result, tt.expected)
			}
		})
	}
}

// TestAssessStabilizationCommits tests stabilization commit assessment
func TestAssessStabilizationCommits(t *testing.T) {
	reporter := &TerminalReporter{UseColor: false}

	tests := []struct {
		name         string
		median       float64
		expectedText string
	}{
		{"minimal_1", 1.0, "Minimal fixes"},
		{"minimal_2", 2.0, "Minimal fixes"},
		{"some_3", 3.0, "Some fixes"},
		{"some_5", 5.0, "Some fixes"},
		{"many_7", 7.0, "Many fixes"},
		{"many_10", 10.0, "Many fixes"},
		{"extensive_15", 15.0, "Extensive rework"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reporter.assessStabilizationCommits(tt.median)
			// assessMetric includes icons, so check if text is contained
			if !strings.Contains(result, tt.expectedText) {
				t.Errorf("assessStabilizationCommits(%.1f) = %s, expected to contain %s", tt.median, result, tt.expectedText)
			}
		})
	}
}

// TestColorTier tests tier color formatting
func TestColorTier(t *testing.T) {
	tests := []struct {
		name      string
		tier      string
		useColor  bool
		wantPlain string
		wantEmoji string
	}{
		{"elite_no_color", string(TierElite), false, "ELITE", ""},
		{"high_no_color", string(TierHigh), false, "HIGH", ""},
		{"medium_no_color", string(TierMedium), false, "MEDIUM", ""},
		{"low_no_color", string(TierLow), false, "LOW", ""},
		{"elite_color", string(TierElite), true, "ELITE", "🌟"},
		{"high_color", string(TierHigh), true, "HIGH", "✓"},
		{"medium_color", string(TierMedium), true, "MEDIUM", "○"},
		{"low_color", string(TierLow), true, "LOW", "✗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &TerminalReporter{UseColor: tt.useColor}
			result := reporter.colorTier(tt.tier)

			if !tt.useColor {
				if result != tt.wantPlain {
					t.Errorf("colorTier(%s, color=false) = %s, expected %s", tt.tier, result, tt.wantPlain)
				}
			} else {
				if !strings.Contains(result, tt.wantEmoji) {
					t.Errorf("colorTier(%s, color=true) missing emoji %s in %s", tt.tier, tt.wantEmoji, result)
				}
				if !strings.Contains(result, tt.wantPlain) {
					t.Errorf("colorTier(%s, color=true) missing text %s in %s", tt.tier, tt.wantPlain, result)
				}
			}
		})
	}
}

// TestColorHelpers tests ANSI color helper functions
func TestColorHelpers(t *testing.T) {
	reporter := &TerminalReporter{UseColor: true}

	// Test that colors produce ANSI codes
	if !strings.Contains(reporter.green("test"), "\033[32m") {
		t.Error("green() should produce ANSI green code")
	}
	if !strings.Contains(reporter.cyan("test"), "\033[36m") {
		t.Error("cyan() should produce ANSI cyan code")
	}
	if !strings.Contains(reporter.yellow("test"), "\033[33m") {
		t.Error("yellow() should produce ANSI yellow code")
	}
	if !strings.Contains(reporter.red("test"), "\033[31m") {
		t.Error("red() should produce ANSI red code")
	}

	// Test no-color mode
	noColorReporter := &TerminalReporter{UseColor: false}
	if noColorReporter.green("test") != "test" {
		t.Error("green() with UseColor=false should return plain text")
	}
}
