-- +goose Up
-- Remove pull request tracking (never used in DORA metric calculations)
-- PRs were collected but not used for any of the 4 DORA metrics

-- Drop pull_requests table and indices
DROP INDEX IF EXISTS idx_prs_state;
DROP INDEX IF EXISTS idx_prs_merged_at;
DROP TABLE IF EXISTS pull_requests;

-- Remove pr_number column from commits table
-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
CREATE TABLE commits_new (
    sha TEXT PRIMARY KEY,
    message TEXT,
    committed_at TIMESTAMP NOT NULL,
    html_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table
INSERT INTO commits_new (sha, message, committed_at, html_url, created_at)
SELECT sha, message, committed_at, html_url, created_at FROM commits;

-- Drop old table and rename new one
DROP INDEX IF EXISTS idx_commits_pr_number;
DROP INDEX IF EXISTS idx_commits_committed_at;
DROP TABLE commits;
ALTER TABLE commits_new RENAME TO commits;

-- Recreate index (pr_number index is gone)
CREATE INDEX IF NOT EXISTS idx_commits_committed_at ON commits(committed_at);

-- +goose Down
-- Recreate pull_requests table
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

-- Add pr_number back to commits
CREATE TABLE commits_new (
    sha TEXT PRIMARY KEY,
    message TEXT,
    committed_at TIMESTAMP NOT NULL,
    pr_number INTEGER,
    html_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO commits_new (sha, message, committed_at, html_url, created_at)
SELECT sha, message, committed_at, html_url, created_at FROM commits;

DROP INDEX IF EXISTS idx_commits_committed_at;
DROP TABLE commits;
ALTER TABLE commits_new RENAME TO commits;

CREATE INDEX IF NOT EXISTS idx_commits_committed_at ON commits(committed_at);
CREATE INDEX IF NOT EXISTS idx_commits_pr_number ON commits(pr_number);
