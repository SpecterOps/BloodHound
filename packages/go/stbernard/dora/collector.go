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
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidTimeRange = errors.New("start time must be before end time")
	ErrNoToken          = errors.New("no GitHub token available")
)

// GitHubCollector collects data from GitHub API
type GitHubCollector struct {
	config *Config
	env    environment.Environment
	client *github.Client
}

// NewGitHubCollector creates a new GitHub data collector
func NewGitHubCollector(config *Config, env environment.Environment) *GitHubCollector {
	return &GitHubCollector{
		config: config,
		env:    env,
	}
}

// getClient creates an authenticated GitHub client
func (s *GitHubCollector) getClient(ctx context.Context) (*github.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	// Try environment variable first
	token := GetTokenFromEnv()
	if token == nil {
		// Try gh CLI
		var err error
		token, err = GetTokenFromGHCLI(s.env)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrNoToken, err)
		}
	}

	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(ctx, ts)
	s.client = github.NewClient(tc)

	return s.client, nil
}

// CollectDeployments collects deployment data from GitHub Actions
func (s *GitHubCollector) CollectDeployments(ctx context.Context, startTime, endTime time.Time) ([]Deployment, error) {
	if startTime.After(endTime) {
		return nil, ErrInvalidTimeRange
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting GitHub client: %w", err)
	}

	var (
		deployments []Deployment
		page        = 1
	)

	for {
		opts := &github.ListWorkflowRunsOptions{
			Status: "completed",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		}

		runs, resp, err := client.Actions.ListRepositoryWorkflowRuns(
			ctx,
			s.config.GitHub.Owner,
			s.config.GitHub.Repo,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("listing workflow runs: %w", err)
		}

		for _, run := range runs.WorkflowRuns {
			// Filter by production workflow if configured
			if s.config.GitHub.Production.Workflow != "" {
				if run.GetName() != s.config.GitHub.Production.Workflow {
					continue
				}
			}

			// Filter by time range
			createdAt := run.GetCreatedAt().Time
			if createdAt.Before(startTime) || createdAt.After(endTime) {
				continue
			}

			deployment := Deployment{
				ID:           fmt.Sprintf("%d", run.GetID()),
				SHA:          run.GetHeadSHA(),
				WorkflowName: run.GetName(),
				WorkflowFile: run.GetPath(),
				Environment:  s.config.GitHub.Production.Environment,
				Status:       run.GetStatus(),
				DeployedAt:   run.GetCreatedAt().Time,
				TriggeredBy:  run.GetActor().GetLogin(),
				Conclusion:   run.GetConclusion(),
				HTMLURL:      run.GetHTMLURL(),
			}

			// Calculate duration if run is complete
			if !run.GetRunStartedAt().IsZero() && !run.GetUpdatedAt().IsZero() {
				started := run.GetRunStartedAt().Time
				updated := run.GetUpdatedAt().Time
				deployment.DurationSecs = int(updated.Sub(started).Seconds())
			}

			deployments = append(deployments, deployment)
		}

		// Stop if we've gone past our start time
		if len(runs.WorkflowRuns) > 0 {
			lastRun := runs.WorkflowRuns[len(runs.WorkflowRuns)-1]
			if lastRun.CreatedAt != nil && lastRun.CreatedAt.Time.Before(startTime) {
				break
			}
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return deployments, nil
}

// CollectCommits collects commit data from GitHub
func (s *GitHubCollector) CollectCommits(ctx context.Context, startTime, endTime time.Time) ([]Commit, error) {
	if startTime.After(endTime) {
		return nil, ErrInvalidTimeRange
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting GitHub client: %w", err)
	}

	var (
		commits []Commit
		page    = 1
	)

	for {
		opts := &github.CommitsListOptions{
			Since: startTime,
			Until: endTime,
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		}

		githubCommits, resp, err := client.Repositories.ListCommits(
			ctx,
			s.config.GitHub.Owner,
			s.config.GitHub.Repo,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("listing commits: %w", err)
		}

		for _, gc := range githubCommits {
			commit := Commit{
				SHA:         gc.GetSHA(),
				Message:     gc.GetCommit().GetMessage(),
				CommittedAt: gc.GetCommit().GetCommitter().GetDate().Time,
				HTMLURL:     gc.GetHTMLURL(),
			}

			// Try to extract PR number from commit message
			// GitHub adds "Merge pull request #123" or "(#123)" to merge commits
			if prNum := extractPRNumber(commit.Message); prNum != nil {
				commit.PRNumber = prNum
			}

			commits = append(commits, commit)
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return commits, nil
}

// CollectPullRequests collects PR data from GitHub
func (s *GitHubCollector) CollectPullRequests(ctx context.Context, startTime, endTime time.Time) ([]PullRequest, error) {
	if startTime.After(endTime) {
		return nil, ErrInvalidTimeRange
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting GitHub client: %w", err)
	}

	var (
		prs  []PullRequest
		page = 1
	)

	for {
		opts := &github.PullRequestListOptions{
			State:     "all", // Get both open and closed
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		}

		githubPRs, resp, err := client.PullRequests.List(
			ctx,
			s.config.GitHub.Owner,
			s.config.GitHub.Repo,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("listing pull requests: %w", err)
		}

		for _, gpr := range githubPRs {
			// Filter by merge time (if merged)
			if gpr.MergedAt != nil {
				mergedTime := gpr.MergedAt.Time
				if mergedTime.Before(startTime) || mergedTime.After(endTime) {
					continue
				}
			}

			pr := PullRequest{
				Number:    gpr.GetNumber(),
				Title:     gpr.GetTitle(),
				State:     gpr.GetState(),
				CreatedAt: gpr.GetCreatedAt().Time,
				BaseRef:   gpr.GetBase().GetRef(),
				HeadRef:   gpr.GetHead().GetRef(),
				HTMLURL:   gpr.GetHTMLURL(),
			}

			if gpr.MergedAt != nil {
				mergedAt := gpr.MergedAt.Time
				pr.MergedAt = &mergedAt
			}

			if gpr.ClosedAt != nil {
				closedAt := gpr.ClosedAt.Time
				pr.ClosedAt = &closedAt
			}

			if gpr.MergeCommitSHA != nil {
				pr.MergeCommitSHA = gpr.MergeCommitSHA
			}

			prs = append(prs, pr)
		}

		// Stop if we've gone past our start time (since we're sorting by updated desc)
		if len(githubPRs) > 0 {
			lastPR := githubPRs[len(githubPRs)-1]
			if lastPR.UpdatedAt != nil && lastPR.UpdatedAt.Time.Before(startTime) {
				break
			}
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return prs, nil
}

// extractPRNumber extracts PR number from commit message
// Looks for patterns like "Merge pull request #123" or "(#123)"
func extractPRNumber(message string) *int {
	// TODO: Implement regex-based PR number extraction
	return nil
}
