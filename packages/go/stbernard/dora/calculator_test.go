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
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestCalculateDeploymentFrequency(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	// Create test data: 10 production deployments over 30 days
	now := time.Now()
	var deployments []Deployment
	for i := 0; i < 10; i++ {
		deployments = append(deployments, Deployment{
			Tag:          "v9." + string(rune('0'+i)) + ".0",
			SHA:          "sha" + string(rune('0'+i)),
			Version:      "9." + string(rune('0'+i)) + ".0",
			DeployedAt:   now.AddDate(0, 0, -i*3), // Every 3 days
			IsProduction: true,
			IsRC:         false,
			IsPatch:      false,
			PatchNumber:  0,
		})
	}

	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	startTime := now.AddDate(0, 0, -30)
	endTime := now
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Verify deployment frequency
	if snapshot.DeploymentCount != 10 {
		t.Errorf("Expected 10 deployments, got %d", snapshot.DeploymentCount)
	}

	expectedFreq := 10.0 / 30.0 // ~0.33 per day
	if snapshot.DeploymentFrequencyPerDay < expectedFreq-0.01 || snapshot.DeploymentFrequencyPerDay > expectedFreq+0.01 {
		t.Errorf("Expected frequency ~%.2f, got %.2f", expectedFreq, snapshot.DeploymentFrequencyPerDay)
	}

	// Should be "high" tier (between once per week and once per day)
	if snapshot.DeploymentTier != string(TierHigh) {
		t.Errorf("Expected tier %s, got %s", TierHigh, snapshot.DeploymentTier)
	}
}

func TestCalculateLeadTime(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	now := time.Now()

	// Create commits (2 hours before deployment)
	commits := []Commit{
		{SHA: "sha1", CommittedAt: now.Add(-2 * time.Hour), Message: "Fix bug"},
		{SHA: "sha2", CommittedAt: now.Add(-4 * time.Hour), Message: "Add feature"},
	}

	// Create deployments
	deployments := []Deployment{
		{
			Tag:          "v9.1.0",
			SHA:          "sha1",
			Version:      "9.1.0",
			DeployedAt:   now,
			IsProduction: true,
		},
		{
			Tag:          "v9.2.0",
			SHA:          "sha2",
			Version:      "9.2.0",
			DeployedAt:   now,
			IsProduction: true,
		},
	}

	if err := storage.SaveCommits(ctx, commits); err != nil {
		t.Fatalf("Failed to save commits: %v", err)
	}

	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	startTime := now.AddDate(0, 0, -1)
	endTime := now.AddDate(0, 0, 1)
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Lead times should be 2 and 4 hours, median = 3 hours
	if snapshot.LeadTimeP50Hours < 2.9 || snapshot.LeadTimeP50Hours > 3.1 {
		t.Errorf("Expected P50 ~3 hours, got %.2f", snapshot.LeadTimeP50Hours)
	}

	// Should be "elite" tier (< 24 hours)
	if snapshot.LeadTimeTier != string(TierElite) {
		t.Errorf("Expected tier %s, got %s", TierElite, snapshot.LeadTimeTier)
	}
}

func TestCalculateChangeFailureRate(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	now := time.Now()

	// Create deployments: 10 total, 2 required patches (20% failure rate)
	var deployments []Deployment
	for i := 0; i < 10; i++ {
		patches := 0
		if i < 2 {
			patches = 1 // First 2 deployments needed patches
		}
		deployments = append(deployments, Deployment{
			Tag:          "v9." + string(rune('0'+i)) + ".0",
			SHA:          "sha" + string(rune('0'+i)),
			Version:      "9." + string(rune('0'+i)) + ".0",
			DeployedAt:   now.AddDate(0, 0, -i),
			IsProduction: true,
			TotalPatches: patches,
		})
	}

	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	startTime := now.AddDate(0, 0, -11)
	endTime := now
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Should have 20% failure rate
	if snapshot.ChangeFailureRate < 19.0 || snapshot.ChangeFailureRate > 21.0 {
		t.Errorf("Expected ~20%% failure rate, got %.2f%%", snapshot.ChangeFailureRate)
	}

	// Should be "low" tier (> 15%)
	if snapshot.FailureRateTier != string(TierLow) {
		t.Errorf("Expected tier %s, got %s", TierLow, snapshot.FailureRateTier)
	}
}

func TestCalculateTimeToRestore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	now := time.Now()

	// Scenario: v9.1.0 released, then v9.1.1 patch 2 hours later (incident)
	//           v9.2.0 released, then v9.2.1 patch 6 hours later (incident)
	//           Median restore time = 4 hours = Elite tier
	deployments := []Deployment{
		{
			Tag:          "v9.1.0",
			SHA:          "sha1",
			Version:      "9.1.0",
			DeployedAt:   now.Add(-10 * time.Hour),
			IsProduction: true,
			IsPatch:      false,
			PatchNumber:  0,
		},
		{
			Tag:          "v9.1.1",
			SHA:          "sha2",
			Version:      "9.1.0", // Same base version
			DeployedAt:   now.Add(-8 * time.Hour), // 2 hours after v9.1.0
			IsProduction: true,
			IsPatch:      true,
			PatchNumber:  1,
		},
		{
			Tag:          "v9.2.0",
			SHA:          "sha3",
			Version:      "9.2.0",
			DeployedAt:   now.Add(-7 * time.Hour),
			IsProduction: true,
			IsPatch:      false,
			PatchNumber:  0,
		},
		{
			Tag:          "v9.2.1",
			SHA:          "sha4",
			Version:      "9.2.0", // Same base version
			DeployedAt:   now.Add(-1 * time.Hour), // 6 hours after v9.2.0
			IsProduction: true,
			IsPatch:      true,
			PatchNumber:  1,
		},
	}

	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	startTime := now.Add(-12 * time.Hour)
	endTime := now
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Should have 2 incidents
	if snapshot.IncidentCount != 2 {
		t.Errorf("Expected 2 incidents, got %d", snapshot.IncidentCount)
	}

	// Median restore time should be 4 hours (between 2 and 6)
	if snapshot.MedianTTRHours < 3.5 || snapshot.MedianTTRHours > 4.5 {
		t.Errorf("Expected median ~4 hours, got %.2f", snapshot.MedianTTRHours)
	}

	// Should be "elite" tier (< 1 hour is elite, but we're using median which is 4h = high tier)
	// Wait, let me recalculate: 2h and 6h → median = 4h → should be "high" tier
	if snapshot.RestoreTimeTier != string(TierHigh) {
		t.Errorf("Expected tier %s, got %s", TierHigh, snapshot.RestoreTimeTier)
	}
}

func TestCalculateTimeToRestoreNoIncidents(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	now := time.Now()

	// Only initial releases, no patches (no incidents)
	deployments := []Deployment{
		{
			Tag:          "v9.1.0",
			SHA:          "sha1",
			Version:      "9.1.0",
			DeployedAt:   now.Add(-2 * time.Hour),
			IsProduction: true,
			IsPatch:      false,
		},
		{
			Tag:          "v9.2.0",
			SHA:          "sha2",
			Version:      "9.2.0",
			DeployedAt:   now.Add(-1 * time.Hour),
			IsProduction: true,
			IsPatch:      false,
		},
	}

	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	startTime := now.Add(-3 * time.Hour)
	endTime := now
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Should have 0 incidents (no patches)
	if snapshot.IncidentCount != 0 {
		t.Errorf("Expected 0 incidents, got %d", snapshot.IncidentCount)
	}

	// MTTR should be 0 (no incidents)
	if snapshot.MTTRHours != 0 {
		t.Errorf("Expected MTTR 0, got %.2f", snapshot.MTTRHours)
	}

	// Restore time tier should be empty or not affect overall tier
	if snapshot.RestoreTimeTier != "" {
		t.Logf("Restore time tier (no incidents): %s", snapshot.RestoreTimeTier)
	}
}

func TestCalculateQualityMetrics(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	now := time.Now()

	// Create scenario:
	// - v9.1.0 with 2 RCs (v9.1.0-rc1, v9.1.0-rc2, then v9.1.0)
	// - v9.2.0 with 3 RCs (v9.2.0-rc1, v9.2.0-rc2, v9.2.0-rc3, then v9.2.0)
	// - v9.3.0 with 1 RC (v9.3.0-rc1, then v9.3.0)
	// Expected: Average RCs = (2+3+1)/3 = 2.0, Median = 2
	deployments := []Deployment{
		// v9.1.0 series (2 RCs)
		{Tag: "v9.1.0", SHA: "sha1", Version: "9.1.0", DeployedAt: now.Add(-10 * time.Hour), IsProduction: true, TotalRCs: 2},
		// v9.2.0 series (3 RCs)
		{Tag: "v9.2.0", SHA: "sha2", Version: "9.2.0", DeployedAt: now.Add(-5 * time.Hour), IsProduction: true, TotalRCs: 3},
		// v9.3.0 series (1 RC)
		{Tag: "v9.3.0", SHA: "sha3", Version: "9.3.0", DeployedAt: now.Add(-1 * time.Hour), IsProduction: true, TotalRCs: 1},
	}

	// Create 15 commits (5 commits per release on average)
	// Make sure they're all within the time range we'll query
	var commits []Commit
	for i := 0; i < 15; i++ {
		commits = append(commits, Commit{
			SHA:         fmt.Sprintf("commit-sha-%d", i),
			CommittedAt: now.Add(-time.Duration(11-i/2) * time.Hour), // Spread within 11 hours
			Message:     fmt.Sprintf("Commit %d", i),
		})
	}

	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	if err := storage.SaveCommits(ctx, commits); err != nil {
		t.Fatalf("Failed to save commits: %v", err)
	}

	// Calculate metrics
	startTime := now.Add(-12 * time.Hour)
	endTime := now
	snapshot, err := calc.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Verify RC metrics
	expectedAvgRCs := 2.0 // (2 + 3 + 1) / 3
	if snapshot.AverageRCsPerRelease < expectedAvgRCs-0.1 || snapshot.AverageRCsPerRelease > expectedAvgRCs+0.1 {
		t.Errorf("Expected average RCs ~%.1f, got %.2f", expectedAvgRCs, snapshot.AverageRCsPerRelease)
	}

	expectedMedianRCs := 2.0 // sorted: [1, 2, 3] → median = 2
	if snapshot.MedianRCsPerRelease != expectedMedianRCs {
		t.Errorf("Expected median RCs %.1f, got %.2f", expectedMedianRCs, snapshot.MedianRCsPerRelease)
	}

	// Verify commit metrics
	if snapshot.TotalCommitsInPeriod != 15 {
		t.Errorf("Expected 15 total commits, got %d", snapshot.TotalCommitsInPeriod)
	}

	expectedAvgCommits := 5.0 // 15 commits / 3 releases
	if snapshot.AverageCommitsPerRelease < expectedAvgCommits-0.1 || snapshot.AverageCommitsPerRelease > expectedAvgCommits+0.1 {
		t.Errorf("Expected average commits ~%.1f, got %.2f", expectedAvgCommits, snapshot.AverageCommitsPerRelease)
	}

	// Log insights
	t.Logf("Quality Metrics:")
	t.Logf("  Average RCs per release: %.2f", snapshot.AverageRCsPerRelease)
	t.Logf("  Median RCs per release: %.2f", snapshot.MedianRCsPerRelease)
	t.Logf("  Total commits: %d", snapshot.TotalCommitsInPeriod)
	t.Logf("  Average commits per release: %.2f", snapshot.AverageCommitsPerRelease)
}

func TestCalculateTimeToRestoreMinorVersionBoundary(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	// Create test scenario matching real-world example:
	// v9.0.0, v9.0.1, v9.0.2 (same minor version - should track)
	// v9.2.0, v9.2.2 (same minor version - should track)
	// v9.3.0, v9.4.0 (different minor versions - should NOT track)

	baseTime := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	deployments := []Deployment{
		// v9.0 series
		{Tag: "v9.0.0", Version: "9.0.0", DeployedAt: baseTime, IsProduction: true, IsPatch: false, PatchNumber: 0},
		{Tag: "v9.0.1", Version: "9.0.1", DeployedAt: baseTime.Add(14 * 24 * time.Hour), IsProduction: true, IsPatch: true, PatchNumber: 1}, // 14 days later
		{Tag: "v9.0.2", Version: "9.0.2", DeployedAt: baseTime.Add(20 * 24 * time.Hour), IsProduction: true, IsPatch: true, PatchNumber: 2}, // 6 days after .1

		// v9.2 series
		{Tag: "v9.2.0", Version: "9.2.0", DeployedAt: baseTime.Add(30 * 24 * time.Hour), IsProduction: true, IsPatch: false, PatchNumber: 0},
		{Tag: "v9.2.2", Version: "9.2.2", DeployedAt: baseTime.Add(33 * 24 * time.Hour), IsProduction: true, IsPatch: true, PatchNumber: 2}, // 3 days later

		// v9.3 and v9.4 (different minor versions - should NOT track between them)
		{Tag: "v9.3.0", Version: "9.3.0", DeployedAt: baseTime.Add(45 * 24 * time.Hour), IsProduction: true, IsPatch: false, PatchNumber: 0},
		{Tag: "v9.4.0", Version: "9.4.0", DeployedAt: baseTime.Add(60 * 24 * time.Hour), IsProduction: true, IsPatch: false, PatchNumber: 0},
	}

	// Save deployments
	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	snapshot, err := calc.CalculateMetrics(ctx, baseTime.Add(-1*time.Hour), baseTime.Add(61*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Expected MTTR data points:
	// v9.0.0 → v9.0.1: 14 days = 336 hours
	// v9.0.1 → v9.0.2: 6 days = 144 hours
	// v9.2.0 → v9.2.2: 3 days = 72 hours
	// NOT COUNTED: v9.3.0 → v9.4.0 (different minor versions)

	expectedIncidents := 3 // v9.0.1, v9.0.2, v9.2.2
	if snapshot.IncidentCount != expectedIncidents {
		t.Errorf("Expected %d incidents, got %d", expectedIncidents, snapshot.IncidentCount)
	}

	// Median of [72, 144, 336] = 144
	expectedMedian := 144.0
	if snapshot.MedianTTRHours < expectedMedian-1 || snapshot.MedianTTRHours > expectedMedian+1 {
		t.Errorf("Expected median MTTR ~%.1f hours, got %.2f", expectedMedian, snapshot.MedianTTRHours)
	}

	// Mean of [72, 144, 336] = 184
	expectedMean := 184.0
	if snapshot.MTTRHours < expectedMean-1 || snapshot.MTTRHours > expectedMean+1 {
		t.Errorf("Expected mean MTTR ~%.1f hours, got %.2f", expectedMean, snapshot.MTTRHours)
	}

	t.Logf("MTTR Metrics (respects minor version boundaries):")
	t.Logf("  Incidents tracked: %d (expected: %d)", snapshot.IncidentCount, expectedIncidents)
	t.Logf("  Median MTTR: %.2f hours (%.1f days)", snapshot.MedianTTRHours, snapshot.MedianTTRHours/24)
	t.Logf("  Mean MTTR: %.2f hours (%.1f days)", snapshot.MTTRHours, snapshot.MTTRHours/24)
	t.Logf("  ✓ Correctly excluded v9.3.0 → v9.4.0 (different minor versions)")
}

func TestDeploymentFrequencyExcludesPatches(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calc := NewCalculator(storage)
	ctx := context.Background()

	// Create test scenario:
	// v9.1.0 (feature release, no patches)
	// v9.2.0 (feature release, has patches)
	// v9.2.2 (patch - should NOT count as deployment)
	// v9.3.0 (feature release, no patches)
	// v9.4.0 (feature release, no patches)

	baseTime := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	deployments := []Deployment{
		{Tag: "v9.1.0", Version: "9.1.0", DeployedAt: baseTime, IsProduction: true, IsPatch: false, PatchNumber: 0, TotalPatches: 0},
		{Tag: "v9.2.0", Version: "9.2.0", DeployedAt: baseTime.Add(7 * 24 * time.Hour), IsProduction: true, IsPatch: false, PatchNumber: 0, TotalPatches: 1},
		{Tag: "v9.2.2", Version: "9.2.2", DeployedAt: baseTime.Add(10 * 24 * time.Hour), IsProduction: true, IsPatch: true, PatchNumber: 2, TotalPatches: 0},
		{Tag: "v9.3.0", Version: "9.3.0", DeployedAt: baseTime.Add(20 * 24 * time.Hour), IsProduction: true, IsPatch: false, PatchNumber: 0, TotalPatches: 0},
		{Tag: "v9.4.0", Version: "9.4.0", DeployedAt: baseTime.Add(35 * 24 * time.Hour), IsProduction: true, IsPatch: false, PatchNumber: 0, TotalPatches: 0},
	}

	// Save deployments
	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	snapshot, err := calc.CalculateMetrics(ctx, baseTime.Add(-1*time.Hour), baseTime.Add(36*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Expected: 4 deployments (excludes v9.2.2 patch)
	expectedDeployments := 4
	if snapshot.DeploymentCount != expectedDeployments {
		t.Errorf("Expected %d deployments, got %d (should exclude patches)", expectedDeployments, snapshot.DeploymentCount)
	}

	// Expected: 1 failure (v9.2.0 had patches)
	expectedFailures := 1
	if snapshot.FailedDeploymentCount != expectedFailures {
		t.Errorf("Expected %d failed deployments, got %d", expectedFailures, snapshot.FailedDeploymentCount)
	}

	// Expected: 25% failure rate (1/4)
	expectedRate := 25.0
	if snapshot.ChangeFailureRate < expectedRate-0.1 || snapshot.ChangeFailureRate > expectedRate+0.1 {
		t.Errorf("Expected failure rate ~%.1f%%, got %.2f%%", expectedRate, snapshot.ChangeFailureRate)
	}

	t.Logf("Deployment Frequency (excludes patches):")
	t.Logf("  Total deployments: %d (expected: %d)", snapshot.DeploymentCount, expectedDeployments)
	t.Logf("  Failed deployments: %d (expected: %d)", snapshot.FailedDeploymentCount, expectedFailures)
	t.Logf("  Failure rate: %.1f%% (expected: %.1f%%)", snapshot.ChangeFailureRate, expectedRate)
	t.Logf("  ✓ Correctly excluded v9.2.2 patch from deployment count")
	t.Logf("  ✓ Counted v9.2.0 as failed (TotalPatches=1)")
}
