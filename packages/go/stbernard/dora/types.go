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

import "time"

// Deployment represents a production deployment tracked via Git tags
// Tags follow semver format: vMAJOR.MINOR.PATCH[-rcN]
// Examples: v9.4.0 (production), v9.4.0-rc1 (release candidate), v9.4.1 (hotfix patch)
type Deployment struct {
	// Core deployment information
	Tag        string    `db:"tag"`         // Full tag name (e.g., v9.4.0, v9.4.0-rc1)
	SHA        string    `db:"sha"`         // Git commit SHA
	Version    string    `db:"version"`     // Semantic version (e.g., 9.4.0)
	DeployedAt time.Time `db:"deployed_at"` // Tag creation timestamp

	// Release type classification
	IsProduction bool `db:"is_production"` // true for v9.4.0, false for v9.4.0-rc1
	IsRC         bool `db:"is_rc"`         // true for release candidates (-rcN)
	RCNumber     *int `db:"rc_number"`     // Release candidate number (e.g., 1 for -rc1)
	IsPatch      bool `db:"is_patch"`      // true for patch releases (PATCH > 0)
	PatchNumber  int  `db:"patch_number"`  // Patch version number

	// Quality metrics (calculated during collection)
	TotalRCs     int `db:"total_rcs"`     // Total RCs before this production release
	TotalPatches int `db:"total_patches"` // Total patches for this minor version

	// GitHub metadata
	HTMLURL string `db:"html_url"` // GitHub tag URL
}

// Commit represents a Git commit
// Note: Author information is intentionally excluded to maintain team-level focus
// and prevent misuse of DORA metrics for individual performance evaluation
type Commit struct {
	SHA         string    `db:"sha"`
	Message     string    `db:"message"`
	CommittedAt time.Time `db:"committed_at"`
	PRNumber    *int      `db:"pr_number"`
	HTMLURL     string    `db:"html_url"`
}

// PullRequest represents a GitHub pull request
// Note: Author information is intentionally excluded to maintain team-level focus
// and prevent misuse of DORA metrics for individual performance evaluation
type PullRequest struct {
	Number         int        `db:"number"`
	Title          string     `db:"title"`
	State          string     `db:"state"`
	CreatedAt      time.Time  `db:"created_at"`
	MergedAt       *time.Time `db:"merged_at"`
	ClosedAt       *time.Time `db:"closed_at"`
	MergeCommitSHA *string    `db:"merge_commit_sha"`
	BaseRef        string     `db:"base_ref"`
	HeadRef        string     `db:"head_ref"`
	HTMLURL        string     `db:"html_url"`
}

// Issue represents a JIRA issue
// Note: Assignee/reporter information is intentionally excluded to maintain team-level focus
// and prevent misuse of DORA metrics for individual performance evaluation
type Issue struct {
	Key        string     `db:"key"`
	Summary    string     `db:"summary"`
	Type       string     `db:"type"`
	Status     string     `db:"status"`
	Priority   string     `db:"priority"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	ResolvedAt *time.Time `db:"resolved_at"`
	Resolution *string    `db:"resolution"`
	Labels     []string   `db:"labels"` // Stored as JSON in DB
	IsIncident bool       `db:"is_incident"`
	HTMLURL    string     `db:"html_url"`
}

// IssueTransition represents a status change for a JIRA issue
// Note: Author information is intentionally excluded to maintain team-level focus
type IssueTransition struct {
	ID             int       `db:"id"`
	IssueKey       string    `db:"issue_key"`
	FromStatus     *string   `db:"from_status"`
	ToStatus       string    `db:"to_status"`
	TransitionedAt time.Time `db:"transitioned_at"`
}

// MetricsSnapshot represents calculated metrics for a time period
type MetricsSnapshot struct {
	ID                        int       `db:"id"`
	PeriodStart               time.Time `db:"period_start"`
	PeriodEnd                 time.Time `db:"period_end"`
	DeploymentCount           int       `db:"deployment_count"`
	DeploymentFrequencyPerDay float64   `db:"deployment_frequency_per_day"`
	DeploymentTier            string    `db:"deployment_tier"`
	LeadTimeP50Hours          float64   `db:"lead_time_p50_hours"`
	LeadTimeP90Hours          float64   `db:"lead_time_p90_hours"`
	LeadTimeP95Hours          float64   `db:"lead_time_p95_hours"`
	LeadTimeTier              string    `db:"lead_time_tier"`
	FailedDeploymentCount     int       `db:"failed_deployment_count"`
	ChangeFailureRate         float64   `db:"change_failure_rate"`
	FailureRateTier           string    `db:"failure_rate_tier"`
	IncidentCount             int       `db:"incident_count"`
	MTTRHours                 float64   `db:"mttr_hours"`
	MedianTTRHours            float64   `db:"median_ttr_hours"`
	P95TTRHours               float64   `db:"p95_ttr_hours"`
	RestoreTimeTier           string    `db:"restore_time_tier"`
	OverallTier               string    `db:"overall_tier"`
	CalculatedAt              time.Time `db:"calculated_at"`

	// Quality metrics (non-DORA but useful for understanding practices)
	AverageRCsPerRelease     float64 `db:"average_rcs_per_release"`     // Mean RCs before production
	MedianRCsPerRelease      float64 `db:"median_rcs_per_release"`      // Median RCs (more robust)
	TotalCommitsInPeriod     int     `db:"total_commits_in_period"`     // All commits in time range
	AverageCommitsPerRelease float64 `db:"average_commits_per_release"` // Batch size indicator
}

// PerformanceTier represents the DORA performance classification
type PerformanceTier string

const (
	TierElite  PerformanceTier = "elite"
	TierHigh   PerformanceTier = "high"
	TierMedium PerformanceTier = "medium"
	TierLow    PerformanceTier = "low"
)

// String returns the string representation of the performance tier
func (s PerformanceTier) String() string {
	return string(s)
}
