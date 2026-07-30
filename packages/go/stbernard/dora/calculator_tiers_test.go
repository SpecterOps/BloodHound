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
	"path/filepath"
	"testing"
)

// TestClassifyLeadTime tests DORA lead time tier classification
// Based on official DORA thresholds
func TestClassifyLeadTime(t *testing.T) {
	tests := []struct {
		name         string
		p50Hours     float64
		expectedTier PerformanceTier
	}{
		// Elite: < 24 hours (1 day)
		{"elite_1hour", 1.0, TierElite},
		{"elite_12hours", 12.0, TierElite},
		{"elite_boundary", 23.99, TierElite},

		// High: 24-168 hours (1 day to 1 week)
		{"high_1day", 24.0, TierHigh},
		{"high_3days", 72.0, TierHigh},
		{"high_boundary", 167.99, TierHigh},

		// Medium: 168-720 hours (1 week to 1 month)
		{"medium_1week", 168.0, TierMedium},
		{"medium_2weeks", 336.0, TierMedium},
		{"medium_boundary", 719.99, TierMedium},

		// Low: >= 720 hours (1 month+)
		{"low_1month", 720.0, TierLow},
		{"low_2months", 1440.0, TierLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := classifyLeadTime(tt.p50Hours)
			if tier != string(tt.expectedTier) {
				t.Errorf("classifyLeadTime(%.2f): expected %s, got %s", tt.p50Hours, tt.expectedTier, tier)
			}
		})
	}
}

// TestClassifyDeploymentFrequency tests DORA deployment frequency tier classification
func TestClassifyDeploymentFrequency(t *testing.T) {
	tests := []struct {
		name         string
		perDay       float64
		expectedTier PerformanceTier
	}{
		// Elite: >= 1 per day
		{"elite_multiple", 5.0, TierElite},
		{"elite_boundary", 1.0, TierElite},

		// High: >= 1/7 per day (weekly) to < 1/day
		{"high_every2days", 0.5, TierHigh},
		{"high_weekly", 1.0 / 7.0, TierHigh},
		{"high_boundary", 0.142858, TierHigh},

		// Medium: >= 1/30 per day (monthly) to < 1/week
		{"medium_biweekly", 1.0 / 14.0, TierMedium},
		{"medium_monthly", 1.0 / 30.0, TierMedium},
		{"medium_boundary", 0.0334, TierMedium},

		// Low: < 1/30 per day (less than monthly)
		{"low_quarterly", 1.0 / 90.0, TierLow},
		{"low_rarely", 0.01, TierLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := classifyDeploymentFrequency(tt.perDay)
			if tier != string(tt.expectedTier) {
				t.Errorf("classifyDeploymentFrequency(%.6f): expected %s, got %s", tt.perDay, tt.expectedTier, tier)
			}
		})
	}
}

// TestClassifyFailureRate tests DORA change failure rate tier classification
func TestClassifyFailureRate(t *testing.T) {
	tests := []struct {
		name         string
		percentage   float64
		expectedTier PerformanceTier
	}{
		// Elite: < 5%
		{"elite_zero", 0.0, TierElite},
		{"elite_low", 2.5, TierElite},
		{"elite_boundary", 4.99, TierElite},

		// High: 5% to < 10%
		{"high_min", 5.0, TierHigh},
		{"high_mid", 7.5, TierHigh},
		{"high_boundary", 9.99, TierHigh},

		// Medium: 10% to < 15%
		{"medium_min", 10.0, TierMedium},
		{"medium_mid", 12.5, TierMedium},
		{"medium_boundary", 14.99, TierMedium},

		// Low: >= 15%
		{"low_min", 15.0, TierLow},
		{"low_high", 25.0, TierLow},
		{"low_very_high", 50.0, TierLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := classifyFailureRate(tt.percentage)
			if tier != string(tt.expectedTier) {
				t.Errorf("classifyFailureRate(%.2f): expected %s, got %s", tt.percentage, tt.expectedTier, tier)
			}
		})
	}
}

// TestClassifyRestoreTime tests DORA time to restore tier classification
func TestClassifyRestoreTime(t *testing.T) {
	tests := []struct {
		name         string
		medianHours  float64
		expectedTier PerformanceTier
	}{
		// Elite: < 1 hour
		{"elite_15min", 0.25, TierElite},
		{"elite_30min", 0.5, TierElite},
		{"elite_boundary", 0.99, TierElite},

		// High: 1 to < 24 hours
		{"high_1hour", 1.0, TierHigh},
		{"high_12hours", 12.0, TierHigh},
		{"high_boundary", 23.99, TierHigh},

		// Medium: 24 to < 168 hours (1 week)
		{"medium_1day", 24.0, TierMedium},
		{"medium_3days", 72.0, TierMedium},
		{"medium_boundary", 167.99, TierMedium},

		// Low: >= 168 hours (1 week+)
		{"low_1week", 168.0, TierLow},
		{"low_2weeks", 336.0, TierLow},
		{"low_1month", 720.0, TierLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := classifyRestoreTime(tt.medianHours)
			if tier != string(tt.expectedTier) {
				t.Errorf("classifyRestoreTime(%.2f): expected %s, got %s", tt.medianHours, tt.expectedTier, tier)
			}
		})
	}
}



// TestDetermineOverallTier tests overall tier determination logic
func TestDetermineOverallTier(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)

	tests := []struct {
		name                string
		deploymentTier      string
		leadTimeTier        string
		failureRateTier     string
		restoreTimeTier     string
		expectedOverallTier string
		description         string
	}{
		{
			name:                "all_elite",
			deploymentTier:      string(TierElite),
			leadTimeTier:        string(TierElite),
			failureRateTier:     string(TierElite),
			restoreTimeTier:     string(TierElite),
			expectedOverallTier: string(TierElite),
			description:         "Perfect scores across all metrics",
		},
		{
			name:                "mostly_high_one_medium",
			deploymentTier:      string(TierHigh),
			leadTimeTier:        string(TierHigh),
			failureRateTier:     string(TierMedium),
			restoreTimeTier:     string(TierHigh),
			expectedOverallTier: string(TierMedium),
			description:         "Conservative: lowest tier wins",
		},
		{
			name:                "two_low_metrics",
			deploymentTier:      string(TierLow),
			leadTimeTier:        string(TierLow),
			failureRateTier:     string(TierHigh),
			restoreTimeTier:     string(TierElite),
			expectedOverallTier: string(TierLow),
			description:         "Two low metrics result in low overall",
		},
		{
			name:                "mixed_tiers",
			deploymentTier:      string(TierElite),
			leadTimeTier:        string(TierMedium),
			failureRateTier:     string(TierHigh),
			restoreTimeTier:     "", // Not included unless we have incidents
			expectedOverallTier: string(TierMedium),
			description:         "Mixed scores: medium is lowest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := MetricsSnapshot{
				DeploymentTier:  tt.deploymentTier,
				LeadTimeTier:    tt.leadTimeTier,
				FailureRateTier: tt.failureRateTier,
				RestoreTimeTier: tt.restoreTimeTier,
			}
			overall := calc.determineOverallTier(snapshot)
			if overall != tt.expectedOverallTier {
				t.Errorf("%s: expected %s, got %s", tt.description, tt.expectedOverallTier, overall)
			}
		})
	}
}
