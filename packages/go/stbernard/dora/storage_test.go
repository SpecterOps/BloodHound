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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Check that database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestStorageDeployments(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test saving a deployment
	deployment := Deployment{
		ID:           "123",
		SHA:          "abc123",
		WorkflowName: "Deploy",
		Environment:  "production",
		Status:       "completed",
		DeployedAt:   time.Now(),
		Conclusion:   "success",
	}

	if err := storage.SaveDeployments(ctx, []Deployment{deployment}); err != nil {
		t.Fatalf("Failed to save deployment: %v", err)
	}

	// Test retrieving deployments
	start := time.Now().AddDate(0, 0, -1)
	end := time.Now().AddDate(0, 0, 1)

	deployments, err := storage.GetDeployments(ctx, start, end)
	if err != nil {
		t.Fatalf("Failed to get deployments: %v", err)
	}

	if len(deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(deployments))
	}

	if deployments[0].ID != deployment.ID {
		t.Errorf("Expected deployment ID %s, got %s", deployment.ID, deployments[0].ID)
	}
}

func TestStorageCommits(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test saving a commit
	commit := Commit{
		SHA:         "abc123",
		Message:     "Fix bug",
		CommittedAt: time.Now(),
	}

	if err := storage.SaveCommits(ctx, []Commit{commit}); err != nil {
		t.Fatalf("Failed to save commit: %v", err)
	}

	// Test retrieving commits
	start := time.Now().AddDate(0, 0, -1)
	end := time.Now().AddDate(0, 0, 1)

	commits, err := storage.GetCommits(ctx, start, end)
	if err != nil {
		t.Fatalf("Failed to get commits: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}

	if commits[0].SHA != commit.SHA {
		t.Errorf("Expected commit SHA %s, got %s", commit.SHA, commits[0].SHA)
	}
}

func TestStoragePullRequests(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test saving a PR
	now := time.Now()
	pr := PullRequest{
		Number:    42,
		Title:     "Fix bug",
		State:     "merged",
		CreatedAt: now,
		MergedAt:  &now,
	}

	if err := storage.SavePullRequests(ctx, []PullRequest{pr}); err != nil {
		t.Fatalf("Failed to save PR: %v", err)
	}

	// Test retrieving PRs
	start := time.Now().AddDate(0, 0, -1)
	end := time.Now().AddDate(0, 0, 1)

	prs, err := storage.GetPullRequests(ctx, start, end)
	if err != nil {
		t.Fatalf("Failed to get PRs: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prs))
	}

	if prs[0].Number != pr.Number {
		t.Errorf("Expected PR number %d, got %d", pr.Number, prs[0].Number)
	}
}
