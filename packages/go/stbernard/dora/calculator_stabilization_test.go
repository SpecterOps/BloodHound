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
	"testing"
	"time"
)

func TestCalculateStabilizationMetrics(t *testing.T) {
	var (
		tempDir    = t.TempDir()
		dbPath     = "file:" + tempDir + "/test.db?cache=shared&mode=memory"
		now        = time.Now()
		startTime  = now.Add(-30 * 24 * time.Hour)
		endTime    = now
		ctx        = context.Background()
	)

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	calculator := NewCalculator(storage)

	// Create test deployments with RC stabilization data
	deployments := []Deployment{
		// v9.3.0 release: RC1 -> RC2 (3 commits) -> RC3 (2 commits) -> RC4 (1 commit) -> prod
		{Tag: "v9.3.0-rc1", Version: "9.3.0", DeployedAt: now.Add(-20 * 24 * time.Hour), IsRC: true, RCNumber: intPtr(1), StabilizationCommits: 0},
		{Tag: "v9.3.0-rc2", Version: "9.3.0", DeployedAt: now.Add(-19 * 24 * time.Hour), IsRC: true, RCNumber: intPtr(2), StabilizationCommits: 3},
		{Tag: "v9.3.0-rc3", Version: "9.3.0", DeployedAt: now.Add(-18 * 24 * time.Hour), IsRC: true, RCNumber: intPtr(3), StabilizationCommits: 2},
		{Tag: "v9.3.0-rc4", Version: "9.3.0", DeployedAt: now.Add(-17 * 24 * time.Hour), IsRC: true, RCNumber: intPtr(4), StabilizationCommits: 1},
		{Tag: "v9.3.0", Version: "9.3.0", DeployedAt: now.Add(-16 * 24 * time.Hour), IsProduction: true, TotalRCs: 4},
		// v9.4.0 release: RC1 -> RC2 (5 commits) -> prod
		{Tag: "v9.4.0-rc1", Version: "9.4.0", DeployedAt: now.Add(-10 * 24 * time.Hour), IsRC: true, RCNumber: intPtr(1), StabilizationCommits: 0},
		{Tag: "v9.4.0-rc2", Version: "9.4.0", DeployedAt: now.Add(-9 * 24 * time.Hour), IsRC: true, RCNumber: intPtr(2), StabilizationCommits: 5},
		{Tag: "v9.4.0", Version: "9.4.0", DeployedAt: now.Add(-8 * 24 * time.Hour), IsProduction: true, TotalRCs: 2},
	}

	// Save deployments
	if err := storage.SaveDeployments(ctx, deployments); err != nil {
		t.Fatalf("Failed to save deployments: %v", err)
	}

	// Calculate metrics
	snapshot, err := calculator.CalculateMetrics(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Verify stabilization metrics
	// RC2+ commits: 3, 2, 1, 5 = total 11, count 4
	// Average: 11 / 4 = 2.75
	// Sorted: 1, 2, 3, 5 -> Median: (2+3)/2 = 2.5
	
	if snapshot.AverageStabilizationCommits < 2.7 || snapshot.AverageStabilizationCommits > 2.8 {
		t.Errorf("Expected average stabilization commits ~2.75, got %.2f", snapshot.AverageStabilizationCommits)
	}

	if snapshot.MedianStabilizationCommits != 2.5 {
		t.Errorf("Expected median stabilization commits 2.5, got %.1f", snapshot.MedianStabilizationCommits)
	}

	t.Logf("Stabilization metrics: Avg=%.1f, Median=%.1f", 
		snapshot.AverageStabilizationCommits, snapshot.MedianStabilizationCommits)
}

func intPtr(v int) *int {
	return &v
}
