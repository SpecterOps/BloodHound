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
	"regexp"
	"sort"
	"strconv"
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

// CollectDeployments collects deployment data from Git tags
// Tags follow semver format: vMAJOR.MINOR.PATCH[-rcN]
// Production deployments are tags without -rc suffix (e.g., v9.4.0)
// Release candidates have -rcN suffix (e.g., v9.4.0-rc1)
// Patch releases have PATCH > 0 (e.g., v9.4.1 is a hotfix)
func (s *GitHubCollector) CollectDeployments(ctx context.Context, startTime, endTime time.Time) ([]Deployment, error) {
	if startTime.After(endTime) {
		return nil, ErrInvalidTimeRange
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting GitHub client: %w", err)
	}

	// Fetch all tags from the repository
	var (
		allTags []*github.RepositoryTag
		page    = 1
	)

	for {
		opts := &github.ListOptions{
			Page:    page,
			PerPage: 100,
		}

		tags, resp, err := client.Repositories.ListTags(
			ctx,
			s.config.GitHub.Owner,
			s.config.GitHub.Repo,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("listing tags: %w", err)
		}

		allTags = append(allTags, tags...)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	// Parse tags and build deployments with quality metrics
	deployments, err := s.parseTagsToDeployments(ctx, client, allTags, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("parsing tags: %w", err)
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

// semverTag represents a parsed semantic version tag
type semverTag struct {
	Tag          string
	SHA          string
	Version      string // e.g., "9.4.0"
	Major        int
	Minor        int
	Patch        int
	IsRC         bool
	RCNumber     int
	Timestamp    time.Time
	HTMLURL      string
}

// parseTagsToDeployments parses repository tags into deployments with quality metrics
func (s *GitHubCollector) parseTagsToDeployments(
	ctx context.Context,
	client *github.Client,
	tags []*github.RepositoryTag,
	startTime, endTime time.Time,
) ([]Deployment, error) {
	var (
		parsedTags  []semverTag
		semverRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-rc(\d+))?$`)
	)

	// Parse all tags
	for _, tag := range tags {
		tagName := tag.GetName()
		matches := semverRegex.FindStringSubmatch(tagName)
		if matches == nil {
			// Skip non-semver tags
			continue
		}

		var (
			major, _    = strconv.Atoi(matches[1])
			minor, _    = strconv.Atoi(matches[2])
			patch, _    = strconv.Atoi(matches[3])
			isRC        = matches[4] != ""
			rcNumber    = 0
			version     = fmt.Sprintf("%d.%d.%d", major, minor, patch)
		)

		if isRC {
			rcNumber, _ = strconv.Atoi(matches[4])
		}

		// Get commit timestamp
		commit, _, err := client.Repositories.GetCommit(
			ctx,
			s.config.GitHub.Owner,
			s.config.GitHub.Repo,
			tag.GetCommit().GetSHA(),
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("getting commit for tag %s: %w", tagName, err)
		}

		timestamp := commit.GetCommit().GetCommitter().GetDate().Time

		parsedTags = append(parsedTags, semverTag{
			Tag:       tagName,
			SHA:       tag.GetCommit().GetSHA(),
			Version:   version,
			Major:     major,
			Minor:     minor,
			Patch:     patch,
			IsRC:      isRC,
			RCNumber:  rcNumber,
			Timestamp: timestamp,
			HTMLURL:   fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", s.config.GitHub.Owner, s.config.GitHub.Repo, tagName),
		})
	}

	// Sort tags by timestamp (oldest first) for processing
	sort.Slice(parsedTags, func(i, j int) bool {
		return parsedTags[i].Timestamp.Before(parsedTags[j].Timestamp)
	})

	// Calculate quality metrics and build deployments
	deployments := s.calculateQualityMetrics(parsedTags, startTime, endTime)

	return deployments, nil
}

// calculateQualityMetrics calculates RC and patch counts for each deployment
func (s *GitHubCollector) calculateQualityMetrics(
	tags []semverTag,
	startTime, endTime time.Time,
) []Deployment {
	var (
		deployments []Deployment
		rcCounts    = make(map[string]int) // version -> RC count
		patchCounts = make(map[string]int) // "major.minor" -> patch count
	)

	// First pass: count RCs and patches
	for _, tag := range tags {
		if tag.IsRC {
			rcCounts[tag.Version]++
		} else if tag.Patch > 0 {
			minorVersion := fmt.Sprintf("%d.%d", tag.Major, tag.Minor)
			patchCounts[minorVersion]++
		}
	}

	// Second pass: build deployments with metrics
	for _, tag := range tags {
		// Filter by time range
		if tag.Timestamp.Before(startTime) || tag.Timestamp.After(endTime) {
			continue
		}

		var (
			rcNumber     *int
			totalRCs     = rcCounts[tag.Version]
			minorVersion = fmt.Sprintf("%d.%d", tag.Major, tag.Minor)
			totalPatches = patchCounts[minorVersion]
		)

		if tag.IsRC {
			rcNumber = &tag.RCNumber
		}

		deployment := Deployment{
			Tag:          tag.Tag,
			SHA:          tag.SHA,
			Version:      tag.Version,
			DeployedAt:   tag.Timestamp,
			IsProduction: !tag.IsRC,
			IsRC:         tag.IsRC,
			RCNumber:     rcNumber,
			IsPatch:      tag.Patch > 0,
			PatchNumber:  tag.Patch,
			TotalRCs:     totalRCs,
			TotalPatches: totalPatches,
			HTMLURL:      tag.HTMLURL,
		}

		deployments = append(deployments, deployment)
	}

	return deployments
}
