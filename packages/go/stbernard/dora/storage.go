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

// initSchema creates the database tables if they don't exist and runs migrations
func (s *Storage) initSchema() error {
	// Create schema_version table first
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("creating schema_version table: %w", err)
	}

	// Get current schema version
	var currentVersion int
	err = s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("getting current schema version: %w", err)
	}

	// Run migrations
	if currentVersion == 0 {
		// Initial schema - tag-based deployments
		if err := s.migrateToV1(); err != nil {
			return fmt.Errorf("migrating to v1: %w", err)
		}
	}

	return nil
}

// migrateToV1 creates the initial schema with tag-based deployments
// If an old workflow-based deployments table exists, data is preserved in commits/PRs
// but the deployments table is recreated for tag-based tracking
func (s *Storage) migrateToV1() error {
	// Check if old deployments table exists with old schema
	var hasOldDeployments bool
	err := s.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type='table' AND name='deployments'
	`).Scan(&hasOldDeployments)
	if err != nil {
		return fmt.Errorf("checking for existing deployments table: %w", err)
	}

	if hasOldDeployments {
		// Check if it's the old schema by looking for workflow_name column
		var hasOldSchema bool
		err = s.db.QueryRow(`
			SELECT COUNT(*) > 0
			FROM pragma_table_info('deployments')
			WHERE name='workflow_name'
		`).Scan(&hasOldSchema)
		if err != nil {
			return fmt.Errorf("checking deployments schema: %w", err)
		}

		if hasOldSchema {
			// Rename old table to preserve data temporarily
			_, err = s.db.Exec(`ALTER TABLE deployments RENAME TO deployments_old`)
			if err != nil {
				return fmt.Errorf("renaming old deployments table: %w", err)
			}
		}
	}

	// Create new schema
	schema := `
	CREATE TABLE IF NOT EXISTS deployments (
		tag TEXT PRIMARY KEY,
		sha TEXT NOT NULL,
		version TEXT NOT NULL,
		deployed_at TIMESTAMP NOT NULL,
		is_production BOOLEAN NOT NULL,
		is_rc BOOLEAN NOT NULL,
		rc_number INTEGER,
		is_patch BOOLEAN NOT NULL,
		patch_number INTEGER NOT NULL,
		total_rcs INTEGER NOT NULL DEFAULT 0,
		total_patches INTEGER NOT NULL DEFAULT 0,
		html_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_deployments_sha ON deployments(sha);
	CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at ON deployments(deployed_at);
	CREATE INDEX IF NOT EXISTS idx_deployments_version ON deployments(version);
	CREATE INDEX IF NOT EXISTS idx_deployments_is_production ON deployments(is_production);

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
	`

	_, err = s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("creating new schema: %w", err)
	}

	// Check if we have old data to inform the user
	var oldRowCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM deployments_old`).Scan(&oldRowCount)
	if err == nil && oldRowCount > 0 {
		fmt.Printf("\n⚠️  Migration Notice:\n")
		fmt.Printf("   The deployment data model has changed from workflow runs to Git tags.\n")
		fmt.Printf("   Your %d old deployment records (workflow-based) cannot be automatically migrated.\n", oldRowCount)
		fmt.Printf("   Please run 'stbernard dora collect' to re-collect deployment data from Git tags.\n\n")
	}

	// Drop old deployments table
	_, err = s.db.Exec(`DROP TABLE IF EXISTS deployments_old`)
	if err != nil {
		return fmt.Errorf("dropping old deployments table: %w", err)
	}

	// Record migration
	_, err = s.db.Exec(`
		INSERT INTO schema_version (version, description)
		VALUES (1, 'Initial schema with tag-based deployments')
	`)
	if err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	return nil
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
		(tag, sha, version, deployed_at, is_production, is_rc, rc_number,
		 is_patch, patch_number, total_rcs, total_patches, html_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range deployments {
		_, err = stmt.ExecContext(ctx,
			d.Tag, d.SHA, d.Version, d.DeployedAt, d.IsProduction, d.IsRC,
			d.RCNumber, d.IsPatch, d.PatchNumber, d.TotalRCs, d.TotalPatches,
			d.HTMLURL,
		)
		if err != nil {
			return fmt.Errorf("inserting deployment %s: %w", d.Tag, err)
		}
	}

	return tx.Commit()
}

// GetDeployments retrieves deployments within a time range
func (s *Storage) GetDeployments(ctx context.Context, start, end time.Time) ([]Deployment, error) {
	query := `
		SELECT tag, sha, version, deployed_at, is_production, is_rc, rc_number,
		       is_patch, patch_number, total_rcs, total_patches, html_url
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
			&d.Tag, &d.SHA, &d.Version, &d.DeployedAt, &d.IsProduction, &d.IsRC,
			&d.RCNumber, &d.IsPatch, &d.PatchNumber, &d.TotalRCs, &d.TotalPatches,
			&d.HTMLURL,
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
