-- +goose Up
-- Initial DORA metrics schema with tag-based deployments
-- Deployments are tracked via Git tags following semver (vMAJOR.MINOR.PATCH[-rcN])
-- Quality metrics track RC count and patch count per release

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

-- +goose Down
DROP INDEX IF EXISTS idx_prs_state;
DROP INDEX IF EXISTS idx_prs_merged_at;
DROP TABLE IF EXISTS pull_requests;

DROP INDEX IF EXISTS idx_commits_pr_number;
DROP INDEX IF EXISTS idx_commits_committed_at;
DROP TABLE IF EXISTS commits;

DROP INDEX IF EXISTS idx_deployments_is_production;
DROP INDEX IF EXISTS idx_deployments_version;
DROP INDEX IF EXISTS idx_deployments_deployed_at;
DROP INDEX IF EXISTS idx_deployments_sha;
DROP TABLE IF EXISTS deployments;
