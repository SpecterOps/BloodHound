-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0
-- +goose Up
-- Initial DORA metrics schema
-- Deployments are tracked via Git tags following semver (vMAJOR.MINOR.PATCH[-rcN])
-- Quality metrics track RC count, patch count, and stabilization commits per release

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
    stabilization_commits INTEGER NOT NULL DEFAULT 0,
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
    html_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_commits_committed_at ON commits(committed_at);

-- +goose Down
DROP INDEX IF EXISTS idx_commits_committed_at;
DROP TABLE IF EXISTS commits;

DROP INDEX IF EXISTS idx_deployments_is_production;
DROP INDEX IF EXISTS idx_deployments_version;
DROP INDEX IF EXISTS idx_deployments_deployed_at;
DROP INDEX IF EXISTS idx_deployments_sha;
DROP TABLE IF EXISTS deployments;
