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
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // SQLite driver
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

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

	// Run goose migrations
	if err := storage.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
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

// runMigrations executes goose migrations using embedded SQL files
func (s *Storage) runMigrations() error {
	goose.SetBaseFS(migrationFiles)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	if err := goose.Up(s.db, "migrations"); err != nil {
		return fmt.Errorf("running goose migrations: %w", err)
	}

	return nil
}

// SaveDeployments saves multiple deployments to the database
func (s *Storage) SaveDeployments(ctx context.Context, deployments []Deployment) error {
	if len(deployments) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, d := range deployments {
		ib := sqlbuilder.NewInsertBuilder()
		ib.InsertInto("deployments")
		ib.Cols(
			"tag", "sha", "version", "deployed_at", "is_production", "is_rc",
			"rc_number", "is_patch", "patch_number", "total_rcs", "total_patches", "stabilization_commits", "html_url",
		)
		ib.Values(
			d.Tag, d.SHA, d.Version, d.DeployedAt, d.IsProduction, d.IsRC,
			d.RCNumber, d.IsPatch, d.PatchNumber, d.TotalRCs, d.TotalPatches, d.StabilizationCommits, d.HTMLURL,
		)

		// SQLite uses INSERT OR REPLACE for upsert
		query, args := ib.Build()
		query = "INSERT OR REPLACE INTO deployments " + query[len("INSERT INTO deployments "):]

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("inserting deployment %s: %w", d.Tag, err)
		}
	}

	return tx.Commit()
}

// GetDeployments retrieves deployments within a time range
func (s *Storage) GetDeployments(ctx context.Context, start, end time.Time) ([]Deployment, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"tag", "sha", "version", "deployed_at", "is_production", "is_rc", "rc_number",
		"is_patch", "patch_number", "total_rcs", "total_patches", "stabilization_commits", "html_url",
	)
	sb.From("deployments")
	sb.Where(
		sb.Between("deployed_at", start, end),
	)
	sb.OrderBy("deployed_at").Desc()

	query, args := sb.Build()
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying deployments: %w", err)
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var d Deployment
		err := rows.Scan(
			&d.Tag, &d.SHA, &d.Version, &d.DeployedAt, &d.IsProduction, &d.IsRC,
			&d.RCNumber, &d.IsPatch, &d.PatchNumber, &d.TotalRCs, &d.TotalPatches, &d.StabilizationCommits,
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
	if len(commits) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, c := range commits {
		ib := sqlbuilder.NewInsertBuilder()
		ib.InsertInto("commits")
		ib.Cols("sha", "message", "committed_at", "pr_number", "html_url")
		ib.Values(c.SHA, c.Message, c.CommittedAt, c.PRNumber, c.HTMLURL)

		query, args := ib.Build()
		query = "INSERT OR REPLACE INTO commits " + query[len("INSERT INTO commits "):]

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("inserting commit %s: %w", c.SHA, err)
		}
	}

	return tx.Commit()
}

// GetCommits retrieves commits within a time range
func (s *Storage) GetCommits(ctx context.Context, start, end time.Time) ([]Commit, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("sha", "message", "committed_at", "pr_number", "html_url")
	sb.From("commits")
	sb.Where(sb.Between("committed_at", start, end))
	sb.OrderBy("committed_at").Desc()

	query, args := sb.Build()
	rows, err := s.db.QueryContext(ctx, query, args...)
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
	if len(prs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, pr := range prs {
		ib := sqlbuilder.NewInsertBuilder()
		ib.InsertInto("pull_requests")
		ib.Cols(
			"number", "title", "state", "created_at", "merged_at", "closed_at",
			"merge_commit_sha", "base_ref", "head_ref", "html_url",
		)
		ib.Values(
			pr.Number, pr.Title, pr.State, pr.CreatedAt, pr.MergedAt,
			pr.ClosedAt, pr.MergeCommitSHA, pr.BaseRef, pr.HeadRef, pr.HTMLURL,
		)

		query, args := ib.Build()
		query = "INSERT OR REPLACE INTO pull_requests " + query[len("INSERT INTO pull_requests "):]

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("inserting PR %d: %w", pr.Number, err)
		}
	}

	return tx.Commit()
}

// GetPullRequests retrieves PRs within a time range (by merge time)
func (s *Storage) GetPullRequests(ctx context.Context, start, end time.Time) ([]PullRequest, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"number", "title", "state", "created_at", "merged_at", "closed_at",
		"merge_commit_sha", "base_ref", "head_ref", "html_url",
	)
	sb.From("pull_requests")
	sb.Where(sb.Between("merged_at", start, end))
	sb.OrderBy("merged_at").Desc()

	query, args := sb.Build()
	rows, err := s.db.QueryContext(ctx, query, args...)
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
