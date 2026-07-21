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
	"math"
	"sort"
	"strings"
	"time"
)

// Calculator computes DORA metrics from collected data
type Calculator struct {
	storage *Storage
}

// NewCalculator creates a new metrics calculator
func NewCalculator(storage *Storage) *Calculator {
	return &Calculator{storage: storage}
}

// CalculateMetrics computes all DORA metrics for a time period
func (s *Calculator) CalculateMetrics(ctx context.Context, startTime, endTime time.Time) (MetricsSnapshot, error) {
	var snapshot MetricsSnapshot

	snapshot.PeriodStart = startTime
	snapshot.PeriodEnd = endTime
	snapshot.CalculatedAt = time.Now()

	// Fetch data
	deployments, err := s.storage.GetDeployments(ctx, startTime, endTime)
	if err != nil {
		return snapshot, fmt.Errorf("getting deployments: %w", err)
	}

	commits, err := s.storage.GetCommits(ctx, startTime, endTime)
	if err != nil {
		return snapshot, fmt.Errorf("getting commits: %w", err)
	}

	// Calculate Deployment Frequency
	if err := s.calculateDeploymentFrequency(&snapshot, deployments, startTime, endTime); err != nil {
		return snapshot, fmt.Errorf("calculating deployment frequency: %w", err)
	}

	// Calculate Lead Time for Changes
	if err := s.calculateLeadTime(&snapshot, deployments, commits); err != nil {
		return snapshot, fmt.Errorf("calculating lead time: %w", err)
	}

	// Calculate Change Failure Rate
	if err := s.calculateChangeFailureRate(&snapshot, deployments); err != nil {
		return snapshot, fmt.Errorf("calculating change failure rate: %w", err)
	}

	// Calculate Time to Restore Service (MTTR)
	if err := s.calculateTimeToRestore(&snapshot, deployments); err != nil {
		return snapshot, fmt.Errorf("calculating time to restore: %w", err)
	}

	// Calculate Quality Metrics (RC counts, batch sizes)
	s.calculateQualityMetrics(&snapshot, deployments, commits)

	// Determine overall tier
	snapshot.OverallTier = s.determineOverallTier(snapshot)

	return snapshot, nil
}

// calculateDeploymentFrequency calculates how often deployments occur
// Only counts feature releases (minor/major versions), not hotfixes/patches.
// Hotfixes are counted separately in Change Failure Rate and MTTR metrics.
// Examples:
//   v9.2.0 (patch=0) → Counts as deployment ✓
//   v9.3.0 (patch=0) → Counts as deployment ✓
//   v9.2.1 (patch=1) → Does NOT count (hotfix) ✗
//   v9.2.2 (patch=2) → Does NOT count (hotfix) ✗
func (s *Calculator) calculateDeploymentFrequency(
	snapshot *MetricsSnapshot,
	deployments []Deployment,
	startTime, endTime time.Time,
) error {
	// Count only production deployments that are NOT patches
	// Patches/hotfixes are excluded because they represent failure resolution, not new deployments
	productionCount := 0
	for _, d := range deployments {
		if d.IsProduction && !d.IsPatch {
			productionCount++
		}
	}

	snapshot.DeploymentCount = productionCount

	// Calculate frequency per day
	days := endTime.Sub(startTime).Hours() / 24
	if days > 0 {
		snapshot.DeploymentFrequencyPerDay = float64(productionCount) / days
	}

	// Classify into tier
	snapshot.DeploymentTier = classifyDeploymentFrequency(snapshot.DeploymentFrequencyPerDay)

	return nil
}

// calculateLeadTime calculates time from previous release to hotfix/patch release
// Lead time for changes measures how long it takes to create and deploy a fix.
// For patches/hotfixes: time from previous release (when issue was introduced) to patch release.
// For regular releases: time from previous release (represents the development cycle).
// Uses tag timestamps directly - no commit data needed.
func (s *Calculator) calculateLeadTime(
	snapshot *MetricsSnapshot,
	deployments []Deployment,
	commits []Commit,
) error {
	var (
		leadTimes          []float64
		previousProduction *Deployment
	)

	// Sort deployments by time (oldest first) to process chronologically
	sortedDeps := make([]Deployment, len(deployments))
	copy(sortedDeps, deployments)
	sort.Slice(sortedDeps, func(i, j int) bool {
		return sortedDeps[i].DeployedAt.Before(sortedDeps[j].DeployedAt)
	})

	// Process deployments chronologically
	for i := range sortedDeps {
		d := &sortedDeps[i]
		if !d.IsProduction {
			continue // Skip RCs
		}

		// Calculate lead time from previous production release to this one
		if previousProduction != nil {
			leadTimeHours := d.DeployedAt.Sub(previousProduction.DeployedAt).Hours()
			if leadTimeHours > 0 {
				leadTimes = append(leadTimes, leadTimeHours)
			}
		}

		// Update previous production release for next iteration
		previousProduction = d
	}

	// Calculate percentiles
	if len(leadTimes) > 0 {
		sort.Float64s(leadTimes)
		snapshot.LeadTimeP50Hours = percentile(leadTimes, 0.50)
		snapshot.LeadTimeP90Hours = percentile(leadTimes, 0.90)
		snapshot.LeadTimeP95Hours = percentile(leadTimes, 0.95)
		snapshot.LeadTimeTier = classifyLeadTime(snapshot.LeadTimeP50Hours)
	}

	return nil
}

// calculateChangeFailureRate calculates percentage of deployments requiring patches
// Only counts feature releases (patch=0) as deployments. Patches are fixes, not deployments.
// A deployment "fails" if it required any hotfixes/patches.
// Examples:
//   v9.2.0 with patches v9.2.1, v9.2.2 → Failed deployment (TotalPatches=2)
//   v9.3.0 with no patches → Successful deployment (TotalPatches=0)
//   v9.2.1 (patch itself) → Not counted as a deployment
func (s *Calculator) calculateChangeFailureRate(
	snapshot *MetricsSnapshot,
	deployments []Deployment,
) error {
	productionCount := 0
	failedCount := 0

	for _, d := range deployments {
		// Only count feature releases (not patches) as deployments
		if d.IsProduction && !d.IsPatch {
			productionCount++
			// A deployment is considered "failed" if it required patches (hotfixes)
			if d.TotalPatches > 0 {
				failedCount++
			}
		}
	}

	snapshot.FailedDeploymentCount = failedCount

	if productionCount > 0 {
		snapshot.ChangeFailureRate = float64(failedCount) / float64(productionCount) * 100
	}

	snapshot.FailureRateTier = classifyFailureRate(snapshot.ChangeFailureRate)

	return nil
}

// calculateTimeToRestore calculates mean time to restore service (MTTR)
// Every hotfix/patch (patch number > 0) represents an incident that required a fix.
// Time to restore = time from previous release IN THE SAME MINOR VERSION to the hotfix.
// This correctly handles sequential patches like v9.0.1 → v9.0.2 → v9.0.3
// But EXCLUDES cross-version jumps like v9.3.0 → v9.4.0 (different release cycles)
// Examples:
//   v9.2.0 (May 26) → v9.2.2 (May 29) = 3 days to restore ✓
//   v9.0.1 (Apr 14) → v9.0.2 (Apr 20) = 6 days to restore ✓
//   v9.3.0 (Jun 15) → v9.4.0 (Jul 2)  = NOT tracked (different minor versions)
func (s *Calculator) calculateTimeToRestore(
	snapshot *MetricsSnapshot,
	deployments []Deployment,
) error {
	var (
		restoreTimes []float64
		// Map: "major.minor" -> most recent production release in that series
		minorVersionToRelease = make(map[string]*Deployment)
	)

	// Sort deployments by time (oldest first) to process chronologically
	sortedDeps := make([]Deployment, len(deployments))
	copy(sortedDeps, deployments)
	sort.Slice(sortedDeps, func(i, j int) bool {
		return sortedDeps[i].DeployedAt.Before(sortedDeps[j].DeployedAt)
	})

	// Process deployments chronologically
	for i := range sortedDeps {
		d := &sortedDeps[i]
		if !d.IsProduction {
			continue // Skip RCs
		}

		// Parse version to get major.minor
		// Version format is "major.minor.patch" (e.g., "9.2.0" or "9.2.2")
		minorVersion := getMinorVersion(d.Version)

		// Every patch/hotfix is an incident (patch > 0)
		if d.IsPatch && d.PatchNumber > 0 {
			// Find the previous release in this minor version series
			if previousRelease, exists := minorVersionToRelease[minorVersion]; exists {
				// Calculate restore time: time from previous release to this hotfix
				// Previous could be the .0 release or another patch
				restoreHours := d.DeployedAt.Sub(previousRelease.DeployedAt).Hours()
				if restoreHours > 0 {
					restoreTimes = append(restoreTimes, restoreHours)
				}
			}
		}

		// Update the most recent release for this minor version
		// This could be .0 (initial release) or a patch (for sequential hotfixes)
		minorVersionToRelease[minorVersion] = d
	}

	snapshot.IncidentCount = len(restoreTimes)

	// Calculate statistics if we have restore times
	if len(restoreTimes) > 0 {
		// Calculate mean (MTTR)
		var sum float64
		for _, t := range restoreTimes {
			sum += t
		}
		snapshot.MTTRHours = sum / float64(len(restoreTimes))

		// Calculate median and P95
		sort.Float64s(restoreTimes)
		snapshot.MedianTTRHours = percentile(restoreTimes, 0.50)
		snapshot.P95TTRHours = percentile(restoreTimes, 0.95)

		// Classify into tier (use median for classification)
		snapshot.RestoreTimeTier = classifyRestoreTime(snapshot.MedianTTRHours)
	}

	return nil
}

// calculateQualityMetrics calculates non-DORA quality indicators
// These metrics help understand development practices and release complexity
func (s *Calculator) calculateQualityMetrics(
	snapshot *MetricsSnapshot,
	deployments []Deployment,
	commits []Commit,
) {
	var (
		rcCounts              []int
		stabilizationCommits  []int
	)

	// Collect RC counts and stabilization commits from production releases
	for _, d := range deployments {
		if d.IsProduction && !d.IsRC {
			rcCounts = append(rcCounts, d.TotalRCs)
		}
		// Collect stabilization commits from RC2+ (RC1 always has 0)
		if d.IsRC && d.StabilizationCommits > 0 {
			stabilizationCommits = append(stabilizationCommits, d.StabilizationCommits)
		}
	}

	// Calculate RC statistics
	if len(rcCounts) > 0 {
		// Mean RCs per release
		var sum int
		for _, count := range rcCounts {
			sum += count
		}
		snapshot.AverageRCsPerRelease = float64(sum) / float64(len(rcCounts))

		// Median RCs per release (more robust to outliers)
		sort.Ints(rcCounts)
		medianIdx := len(rcCounts) / 2
		if len(rcCounts)%2 == 0 && len(rcCounts) > 1 {
			snapshot.MedianRCsPerRelease = float64(rcCounts[medianIdx-1]+rcCounts[medianIdx]) / 2.0
		} else {
			snapshot.MedianRCsPerRelease = float64(rcCounts[medianIdx])
		}
	}

	// Calculate stabilization commit statistics (RC2+ only)
	if len(stabilizationCommits) > 0 {
		// Mean stabilization commits
		var sum int
		for _, count := range stabilizationCommits {
			sum += count
		}
		snapshot.AverageStabilizationCommits = float64(sum) / float64(len(stabilizationCommits))

		// Median stabilization commits (more robust to outliers)
		sort.Ints(stabilizationCommits)
		medianIdx := len(stabilizationCommits) / 2
		if len(stabilizationCommits)%2 == 0 && len(stabilizationCommits) > 1 {
			snapshot.MedianStabilizationCommits = float64(stabilizationCommits[medianIdx-1]+stabilizationCommits[medianIdx]) / 2.0
		} else {
			snapshot.MedianStabilizationCommits = float64(stabilizationCommits[medianIdx])
		}
	}

	// Total commits in period
	snapshot.TotalCommitsInPeriod = len(commits)

	// Average commits per release (batch size indicator)
	productionCount := 0
	for _, d := range deployments {
		if d.IsProduction && !d.IsRC {
			productionCount++
		}
	}
	if productionCount > 0 {
		snapshot.AverageCommitsPerRelease = float64(len(commits)) / float64(productionCount)
	}
}

// percentile calculates the nth percentile of a sorted slice
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// classifyDeploymentFrequency classifies deployment frequency into DORA tiers
// Based on: https://dora.dev/guides/dora-metrics-four-keys/#deployment-frequency
func classifyDeploymentFrequency(perDay float64) string {
	if perDay >= 1.0 {
		return string(TierElite) // Multiple deploys per day (on-demand)
	} else if perDay >= 1.0/7.0 {
		return string(TierHigh) // Between once per day and once per week
	} else if perDay >= 1.0/30.0 {
		return string(TierMedium) // Between once per week and once per month
	}
	return string(TierLow) // Less than once per month
}

// classifyLeadTime classifies lead time into DORA tiers
// Based on: https://dora.dev/guides/dora-metrics-four-keys/#lead-time-for-changes
func classifyLeadTime(hoursP50 float64) string {
	if hoursP50 < 24 {
		return string(TierElite) // Less than one day
	} else if hoursP50 < 24*7 {
		return string(TierHigh) // Between one day and one week
	} else if hoursP50 < 24*30 {
		return string(TierMedium) // Between one week and one month
	}
	return string(TierLow) // More than one month
}

// classifyFailureRate classifies change failure rate into DORA tiers
// Based on: https://dora.dev/guides/dora-metrics-four-keys/#change-failure-rate
func classifyFailureRate(rate float64) string {
	if rate < 5.0 {
		return string(TierElite) // 0-5%
	} else if rate < 10.0 {
		return string(TierHigh) // 5-10%
	} else if rate < 15.0 {
		return string(TierMedium) // 10-15%
	}
	return string(TierLow) // Over 15%
}

// classifyRestoreTime classifies time to restore service into DORA tiers
// Based on: https://dora.dev/guides/dora-metrics-four-keys/#time-to-restore-service
func classifyRestoreTime(hoursMedian float64) string {
	if hoursMedian < 1.0 {
		return string(TierElite) // Less than one hour
	} else if hoursMedian < 24.0 {
		return string(TierHigh) // Less than one day
	} else if hoursMedian < 24.0*7.0 {
		return string(TierMedium) // Less than one week
	}
	return string(TierLow) // More than one week
}

// getMinorVersion extracts the major.minor version from a semver string
// Example: "9.2.0" -> "9.2", "9.0.3" -> "9.0"
func getMinorVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}

// determineOverallTier determines the overall performance tier
// Uses the lowest tier among all metrics (conservative approach)
func (s *Calculator) determineOverallTier(snapshot MetricsSnapshot) string {
	tiers := []string{
		snapshot.DeploymentTier,
		snapshot.LeadTimeTier,
		snapshot.FailureRateTier,
	}

	// Only include restore time tier if we have incident data
	if snapshot.IncidentCount > 0 && snapshot.RestoreTimeTier != "" {
		tiers = append(tiers, snapshot.RestoreTimeTier)
	}

	// Return the lowest tier (most conservative)
	for _, tier := range []string{string(TierLow), string(TierMedium), string(TierHigh), string(TierElite)} {
		for _, t := range tiers {
			if t == tier {
				return tier
			}
		}
	}

	return string(TierLow)
}
