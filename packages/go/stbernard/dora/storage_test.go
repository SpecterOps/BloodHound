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
	"strings"
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

	// Test saving a production deployment
	now := time.Now()
	rcNum := 3
	deployment := Deployment{
		Tag:          "v9.4.0",
		SHA:          "abc123",
		Version:      "9.4.0",
		DeployedAt:   now,
		IsProduction: true,
		IsRC:         false,
		RCNumber:     nil,
		IsPatch:      false,
		PatchNumber:  0,
		TotalRCs:     rcNum,
		TotalPatches: 0,
		HTMLURL:      "https://github.com/SpecterOps/bloodhound-enterprise/releases/tag/v9.4.0",
	}

	if err := storage.SaveDeployments(ctx, []Deployment{deployment}); err != nil {
		t.Fatalf("Failed to save deployment: %v", err)
	}

	// Test retrieving deployments
	start := now.AddDate(0, 0, -1)
	end := now.AddDate(0, 0, 1)

	deployments, err := storage.GetDeployments(ctx, start, end)
	if err != nil {
		t.Fatalf("Failed to get deployments: %v", err)
	}

	if len(deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(deployments))
	}

	if deployments[0].Tag != deployment.Tag {
		t.Errorf("Expected deployment tag %s, got %s", deployment.Tag, deployments[0].Tag)
	}

	if deployments[0].TotalRCs != rcNum {
		t.Errorf("Expected %d RCs, got %d", rcNum, deployments[0].TotalRCs)
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

func TestGooseMigration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create storage which should run goose migrations
	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Verify deployments table exists with correct schema
	var tableInfo string
	err = storage.db.QueryRow(`
		SELECT sql FROM sqlite_master
		WHERE type='table' AND name='deployments'
	`).Scan(&tableInfo)
	if err != nil {
		t.Fatalf("Failed to query table schema: %v", err)
	}

	// Check that new schema has 'tag' column
	if !strings.Contains(tableInfo, "tag") {
		t.Error("Schema should have 'tag' column")
	}

	// Verify goose version table exists
	var version int64
	err = storage.db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to query goose version: %v", err)
	}

	if version < 1 {
		t.Errorf("Expected at least version 1, got %d", version)
	}

	// Test idempotency - running migrations again should be safe
	storage2, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen storage: %v", err)
	}
	defer storage2.Close()
}
