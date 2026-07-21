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

func TestDeployment(t *testing.T) {
	now := time.Now()
	deployment := Deployment{
		ID:           "123",
		SHA:          "abc123",
		WorkflowName: "Deploy Production",
		WorkflowFile: ".github/workflows/deploy.yml",
		Environment:  "production",
		Status:       "success",
		DeployedAt:   now,
		DurationSecs: 300,
		TriggeredBy:  "user",
		Conclusion:   "success",
		HTMLURL:      "https://github.com/org/repo/actions/runs/123",
	}

	if deployment.ID != "123" {
		t.Errorf("Expected ID to be '123', got '%s'", deployment.ID)
	}
	if deployment.SHA != "abc123" {
		t.Errorf("Expected SHA to be 'abc123', got '%s'", deployment.SHA)
	}
}

func TestCommit(t *testing.T) {
	now := time.Now()
	prNumber := 42
	commit := Commit{
		SHA:         "abc123",
		Author:      "John Doe",
		AuthorEmail: "john@example.com",
		Committer:   "John Doe",
		Message:     "Fix bug",
		CommittedAt: now,
		PRNumber:    &prNumber,
		HTMLURL:     "https://github.com/org/repo/commit/abc123",
	}

	if commit.SHA != "abc123" {
		t.Errorf("Expected SHA to be 'abc123', got '%s'", commit.SHA)
	}
	if commit.PRNumber == nil || *commit.PRNumber != 42 {
		t.Errorf("Expected PR number to be 42, got %v", commit.PRNumber)
	}
}

func TestPullRequest(t *testing.T) {
	now := time.Now()
	mergeCommit := "abc123"
	pr := PullRequest{
		Number:         42,
		Title:          "Fix bug",
		State:          "merged",
		CreatedAt:      now,
		MergedAt:       &now,
		Author:         "John Doe",
		MergeCommitSHA: &mergeCommit,
		BaseRef:        "main",
		HeadRef:        "feature-branch",
		HTMLURL:        "https://github.com/org/repo/pull/42",
	}

	if pr.Number != 42 {
		t.Errorf("Expected PR number to be 42, got %d", pr.Number)
	}
	if pr.MergeCommitSHA == nil || *pr.MergeCommitSHA != "abc123" {
		t.Errorf("Expected merge commit SHA to be 'abc123', got %v", pr.MergeCommitSHA)
	}
}

func TestPerformanceTier(t *testing.T) {
	tests := []struct {
		tier     PerformanceTier
		expected string
	}{
		{TierElite, "elite"},
		{TierHigh, "high"},
		{TierMedium, "medium"},
		{TierLow, "low"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.tier.String() != tc.expected {
				t.Errorf("Expected tier string to be '%s', got '%s'", tc.expected, tc.tier.String())
			}
		})
	}
}

func TestMetricsSnapshot(t *testing.T) {
	now := time.Now()
	snapshot := MetricsSnapshot{
		ID:                        1,
		PeriodStart:               now.AddDate(0, 0, -30),
		PeriodEnd:                 now,
		DeploymentCount:           100,
		DeploymentFrequencyPerDay: 3.33,
		DeploymentTier:            string(TierElite),
		LeadTimeP50Hours:          12.5,
		LeadTimeP90Hours:          24.0,
		LeadTimeP95Hours:          36.0,
		LeadTimeTier:              string(TierHigh),
		FailedDeploymentCount:     5,
		ChangeFailureRate:         5.0,
		FailureRateTier:           string(TierElite),
		IncidentCount:             2,
		MTTRHours:                 0.5,
		MedianTTRHours:            0.4,
		P95TTRHours:               1.2,
		RestoreTimeTier:           string(TierElite),
		OverallTier:               string(TierElite),
		CalculatedAt:              now,
	}

	if snapshot.DeploymentCount != 100 {
		t.Errorf("Expected deployment count to be 100, got %d", snapshot.DeploymentCount)
	}
	if snapshot.OverallTier != string(TierElite) {
		t.Errorf("Expected overall tier to be 'elite', got '%s'", snapshot.OverallTier)
	}
}
