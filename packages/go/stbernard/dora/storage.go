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
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// Storage handles database operations for DORA metrics
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance and initializes the database
func NewStorage(dbPath string) (*Storage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	storage := &Storage{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// initSchema creates the database tables if they don't exist
func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS deployments (
		id TEXT PRIMARY KEY,
		sha TEXT NOT NULL,
		workflow_name TEXT,
		workflow_file TEXT,
		environment TEXT,
		status TEXT NOT NULL,
		deployed_at TIMESTAMP NOT NULL,
		duration_seconds INTEGER,
		triggered_by TEXT,
		conclusion TEXT,
		html_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_deployments_sha ON deployments(sha);
	CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at ON deployments(deployed_at);
	CREATE INDEX IF NOT EXISTS idx_deployments_environment ON deployments(environment);

	CREATE TABLE IF NOT EXISTS commits (
		sha TEXT PRIMARY KEY,
		message TEXT,
		committed_at TIMESTAMP NOT NULL,
		pr_number INTEGER,
		html_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_commits_committed_at ON commits(committed_at);
	CREATE INDEX IF NOT EXISTS idx_commits_pr_number ON commits(pr_number);

	CREATE TABLE IF NOT EXISTS pull_requests (
		number INTEGER PRIMARY KEY,
		title TEXT,
		state TEXT,
		created_at TIMESTAMP NOT NULL,
		merged_at TIMESTAMP,
		closed_at TIMESTAMP,
		merge_commit_sha TEXT,
		base_ref TEXT,
		head_ref TEXT,
		html_url TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_prs_merged_at ON pull_requests(merged_at);
	CREATE INDEX IF NOT EXISTS idx_prs_state ON pull_requests(state);

	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		description TEXT
	);

	INSERT OR IGNORE INTO schema_version (version, description)
	VALUES (1, 'Initial schema');
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveDeployments saves multiple deployments to the database
func (s *Storage) SaveDeployments(ctx context.Context, deployments []Deployment) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO deployments
		(id, sha, workflow_name, workflow_file, environment, status, deployed_at,
		 duration_seconds, triggered_by, conclusion, html_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range deployments {
		_, err = stmt.ExecContext(ctx,
			d.ID, d.SHA, d.WorkflowName, d.WorkflowFile, d.Environment,
			d.Status, d.DeployedAt, d.DurationSecs, d.TriggeredBy,
			d.Conclusion, d.HTMLURL,
		)
		if err != nil {
			return fmt.Errorf("inserting deployment %s: %w", d.ID, err)
		}
	}

	return tx.Commit()
}

// GetDeployments retrieves deployments within a time range
func (s *Storage) GetDeployments(ctx context.Context, start, end time.Time) ([]Deployment, error) {
	query := `
		SELECT id, sha, workflow_name, workflow_file, environment, status,
		       deployed_at, duration_seconds, triggered_by, conclusion, html_url
		FROM deployments
		WHERE deployed_at BETWEEN ? AND ?
		ORDER BY deployed_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("querying deployments: %w", err)
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var d Deployment
		err := rows.Scan(
			&d.ID, &d.SHA, &d.WorkflowName, &d.WorkflowFile, &d.Environment,
			&d.Status, &d.DeployedAt, &d.DurationSecs, &d.TriggeredBy,
			&d.Conclusion, &d.HTMLURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning deployment: %w", err)
		}
		deployments = append(deployments, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating deployments: %w", err)
	}

	return deployments, nil
}

// SaveCommits saves multiple commits to the database
func (s *Storage) SaveCommits(ctx context.Context, commits []Commit) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO commits (sha, message, committed_at, pr_number, html_url)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range commits {
		_, err = stmt.ExecContext(ctx, c.SHA, c.Message, c.CommittedAt, c.PRNumber, c.HTMLURL)
		if err != nil {
			return fmt.Errorf("inserting commit %s: %w", c.SHA, err)
		}
	}

	return tx.Commit()
}

// GetCommits retrieves commits within a time range
func (s *Storage) GetCommits(ctx context.Context, start, end time.Time) ([]Commit, error) {
	query := `
		SELECT sha, message, committed_at, pr_number, html_url
		FROM commits
		WHERE committed_at BETWEEN ? AND ?
		ORDER BY committed_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("querying commits: %w", err)
	}
	defer rows.Close()

	var commits []Commit
	for rows.Next() {
		var c Commit
		err := rows.Scan(&c.SHA, &c.Message, &c.CommittedAt, &c.PRNumber, &c.HTMLURL)
		if err != nil {
			return nil, fmt.Errorf("scanning commit: %w", err)
		}
		commits = append(commits, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating commits: %w", err)
	}

	return commits, nil
}

// SavePullRequests saves multiple PRs to the database
func (s *Storage) SavePullRequests(ctx context.Context, prs []PullRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO pull_requests
		(number, title, state, created_at, merged_at, closed_at,
		 merge_commit_sha, base_ref, head_ref, html_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, pr := range prs {
		_, err = stmt.ExecContext(ctx,
			pr.Number, pr.Title, pr.State, pr.CreatedAt, pr.MergedAt,
			pr.ClosedAt, pr.MergeCommitSHA, pr.BaseRef, pr.HeadRef, pr.HTMLURL,
		)
		if err != nil {
			return fmt.Errorf("inserting PR %d: %w", pr.Number, err)
		}
	}

	return tx.Commit()
}

// GetPullRequests retrieves PRs within a time range (by merge time)
func (s *Storage) GetPullRequests(ctx context.Context, start, end time.Time) ([]PullRequest, error) {
	query := `
		SELECT number, title, state, created_at, merged_at, closed_at,
		       merge_commit_sha, base_ref, head_ref, html_url
		FROM pull_requests
		WHERE merged_at BETWEEN ? AND ?
		ORDER BY merged_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("querying pull requests: %w", err)
	}
	defer rows.Close()

	var prs []PullRequest
	for rows.Next() {
		var pr PullRequest
		err := rows.Scan(
			&pr.Number, &pr.Title, &pr.State, &pr.CreatedAt, &pr.MergedAt,
			&pr.ClosedAt, &pr.MergeCommitSHA, &pr.BaseRef, &pr.HeadRef, &pr.HTMLURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning pull request: %w", err)
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pull requests: %w", err)
	}

	return prs, nil
}
