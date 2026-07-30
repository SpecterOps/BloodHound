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

-- +goose NO TRANSACTION
-- +goose Up
-- Convert audit_logs to a range-partitioned table without wrapping the whole
-- conversion in a single transaction. Running with NO TRANSACTION lets a large
-- tenant's backfill make forward progress that survives a pod restart mid-run
-- rather than rolling back the entire copy. Because goose only records a
-- NO TRANSACTION migration as applied after every statement succeeds, a failure
-- re-runs this file top-to-bottom; every statement below is therefore written to
-- be idempotent so a re-run converges instead of erroring.
--
-- The work is: build the partitioned table under a staging name, backfill it in
-- bounded id batches, verify the row counts, then swap it into place under the
-- original audit_logs name so existing audit logging keeps working against a
-- single table.
--
-- Note on batching: this deployment drives migrations through pgx v5 stdlib in
-- the default (extended) query protocol, which rejects COMMIT inside a DO block.
-- The backfill loop below therefore cannot commit per batch; it runs as one
-- implicit transaction. Batching still bounds per-iteration memory and plan cost,
-- and cross-restart resumability comes from the staging table persisting plus the
-- resume-from-MAX(id) / ON CONFLICT guards. For a tenant so large that even a
-- single-pass copy exceeds the startup window, run this out-of-band with
-- DisableMigrations set rather than at pod startup.

-- Phase 1: staging table. IF NOT EXISTS so a re-run after a partial backfill
-- reuses the already-populated staging table.
CREATE TABLE IF NOT EXISTS audit_logs_partitioned (
    id                  BIGINT NOT NULL DEFAULT nextval('audit_logs_id_seq'),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    action              TEXT NOT NULL,
    actor_id            TEXT,
    actor_name          TEXT,
    actor_email         VARCHAR(330) DEFAULT NULL::character varying,
    request_id          TEXT,
    source_ip_address   TEXT,
    status              VARCHAR(15) DEFAULT 'intent' CHECK (status IN ('intent', 'success', 'failure')),
    commit_id           TEXT,
    fields              JSONB,
    source              VARCHAR(20) DEFAULT 'middleware',
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Phase 2: monthly partitions + default. Child partition names match the final
-- audit_logs parent name so the GC daemon's partition maintenance
-- (audit_logs_YYYY_MM) stays aligned after the rename below. Skipped entirely if
-- a prior run already completed the swap (the staging table is then gone).
-- +goose StatementBegin
DO $$
DECLARE
    start_date    DATE := '2024-01-01';
    end_date      DATE := '2026-08-01';
    current_month DATE;
BEGIN
    IF to_regclass('audit_logs_partitioned') IS NULL THEN
        RETURN;
    END IF;

    current_month := start_date;
    WHILE current_month < end_date LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS audit_logs_%s PARTITION OF audit_logs_partitioned
             FOR VALUES FROM (%L) TO (%L)',
            to_char(current_month, 'YYYY_MM'),
            current_month,
            current_month + interval '1 month'
        );
        current_month := current_month + interval '1 month';
    END LOOP;

    EXECUTE 'CREATE TABLE IF NOT EXISTS audit_logs_default PARTITION OF audit_logs_partitioned DEFAULT';
END $$;
-- +goose StatementEnd

-- Phase 3: batched backfill. Copy rows in bounded id windows, resuming just past
-- the highest id already copied into staging. ON CONFLICT DO NOTHING makes an
-- overlapping window on a re-run a no-op. Skipped once the swap has happened.
-- +goose StatementBegin
DO $$
DECLARE
    batch_size CONSTANT BIGINT := 50000;
    window_lo  BIGINT;
    window_hi  BIGINT;
    source_max BIGINT;
BEGIN
    IF to_regclass('audit_logs_partitioned') IS NULL THEN
        RETURN;
    END IF;

    SELECT COALESCE(MAX(id), 0) INTO source_max FROM audit_logs;
    SELECT COALESCE(MAX(id), 0) INTO window_lo FROM audit_logs_partitioned;

    WHILE window_lo < source_max LOOP
        window_hi := window_lo + batch_size;

        INSERT INTO audit_logs_partitioned (
            id, created_at, action, actor_id, actor_name, actor_email,
            request_id, source_ip_address, status, commit_id, fields, source
        )
        SELECT
            id,
            COALESCE(created_at, '2020-01-01'::timestamptz),
            action,
            actor_id,
            actor_name,
            actor_email,
            request_id,
            source_ip_address,
            status,
            commit_id,
            fields,
            'legacy'
        FROM audit_logs
        WHERE id > window_lo AND id <= window_hi
        ON CONFLICT (id, created_at) DO NOTHING;

        window_lo := window_hi;
    END LOOP;
END $$;
-- +goose StatementEnd

-- Phase 4: verify counts, then swap staging into place. Guarded so a re-run after
-- a completed swap is a no-op. The count-equality check assumes writes to
-- audit_logs are quiesced during the migration (startup, HA-leader gated); a
-- mismatch aborts and leaves the committed staging rows for the next run to
-- finish rather than swapping in an incomplete copy.
-- +goose StatementBegin
DO $$
DECLARE
    source_count BIGINT;
    staging_count BIGINT;
BEGIN
    IF to_regclass('audit_logs_partitioned') IS NULL THEN
        RETURN;
    END IF;

    SELECT count(*) INTO source_count FROM audit_logs;
    SELECT count(*) INTO staging_count FROM audit_logs_partitioned;
    IF staging_count <> source_count THEN
        RAISE EXCEPTION 'audit_logs backfill incomplete: source=% staging=%', source_count, staging_count;
    END IF;

    -- Detach the sequence from the original audit_logs.id before dropping the
    -- table. The sequence is OWNED BY audit_logs.id (see the init migration),
    -- which would otherwise block the DROP with a dependency error. Ownership is
    -- re-attached to the renamed table's column below.
    EXECUTE 'ALTER SEQUENCE audit_logs_id_seq OWNED BY NONE';
    EXECUTE 'DROP TABLE audit_logs';
    EXECUTE 'ALTER TABLE audit_logs_partitioned RENAME TO audit_logs';
    EXECUTE 'ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id';

    -- Advance the sequence past the largest copied id so the next insert does not
    -- collide with a backfilled row.
    PERFORM setval('audit_logs_id_seq', COALESCE((SELECT MAX(id) FROM audit_logs), 1));
END $$;
-- +goose StatementEnd

-- Phase 5: indexes on the (now partitioned) audit_logs. IF NOT EXISTS so a crash
-- between the swap and index creation still converges on a re-run.
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_email ON audit_logs(actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source_ip_address ON audit_logs(source_ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source ON audit_logs(source);

-- +goose Down
-- The Up block re-attaches audit_logs_id_seq to the partitioned audit_logs.id,
-- so dropping the table with CASCADE also drops the owned sequence. Also drop the
-- staging table in case Down runs against a half-completed Up (before the swap).
-- Recreate the sequence before the table that references it, then re-own it to
-- the new column. Guards are idempotent because Down also runs NO TRANSACTION.
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS audit_logs_partitioned CASCADE;

CREATE SEQUENCE IF NOT EXISTS audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS audit_logs (
    id                  BIGINT PRIMARY KEY DEFAULT nextval('audit_logs_id_seq'),
    created_at          TIMESTAMPTZ,
    action              TEXT NOT NULL,
    actor_id            TEXT,
    actor_name          TEXT,
    actor_email         VARCHAR(330) DEFAULT NULL::character varying,
    request_id          TEXT,
    source_ip_address   TEXT,
    status              VARCHAR(15) DEFAULT 'intent' CHECK (status IN ('intent', 'success', 'failure')),
    commit_id           TEXT,
    fields              JSONB
);

ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_email ON audit_logs(actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source_ip_address ON audit_logs(source_ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);