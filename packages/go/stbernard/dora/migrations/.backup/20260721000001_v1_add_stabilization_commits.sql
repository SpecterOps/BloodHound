-- +goose Up
-- Add stabilization_commits column to track commits in RC2+
-- This measures the rework/stabilization effort between RCs
-- RC1 has 0 (no previous RC), RC2+ has commit count from previous RC

ALTER TABLE deployments ADD COLUMN stabilization_commits INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE deployments DROP COLUMN stabilization_commits;
