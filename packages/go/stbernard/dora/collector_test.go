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

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func TestNewGitHubCollector(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	config := &Config{
		GitHub: GitHubConfig{
			Owner: "SpecterOps",
			Repo:  "bloodhound-enterprise",
			Production: ProductionConfig{
				Workflow:    "cicd-distroless.yml",
				Environment: "production",
			},
		},
	}

	collector := NewGitHubCollector(config, env)
	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}
}

func TestCollectDeploymentsValidation(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	config := &Config{
		GitHub: GitHubConfig{
			Owner: "SpecterOps",
			Repo:  "bloodhound-enterprise",
		},
	}

	collector := NewGitHubCollector(config, env)

	// Test with invalid time range
	ctx := context.Background()
	endTime := time.Now()
	startTime := endTime.Add(24 * time.Hour) // Start after end

	_, err = collector.CollectDeployments(ctx, startTime, endTime)
	if err == nil {
		t.Error("Expected error for invalid time range, got nil")
	}
}

func TestCollectCommitsValidation(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	config := &Config{
		GitHub: GitHubConfig{
			Owner: "SpecterOps",
			Repo:  "bloodhound-enterprise",
		},
	}

	collector := NewGitHubCollector(config, env)

	// Test with invalid time range
	ctx := context.Background()
	endTime := time.Now()
	startTime := endTime.Add(24 * time.Hour) // Start after end

	_, err = collector.CollectCommits(ctx, startTime, endTime)
	if err == nil {
		t.Error("Expected error for invalid time range, got nil")
	}
}

func TestCollectPullRequestsValidation(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	config := &Config{
		GitHub: GitHubConfig{
			Owner: "SpecterOps",
			Repo:  "bloodhound-enterprise",
		},
	}

	collector := NewGitHubCollector(config, env)

	// Test with invalid time range
	ctx := context.Background()
	endTime := time.Now()
	startTime := endTime.Add(24 * time.Hour) // Start after end

	_, err = collector.CollectPullRequests(ctx, startTime, endTime)
	if err == nil {
		t.Error("Expected error for invalid time range, got nil")
	}
}
