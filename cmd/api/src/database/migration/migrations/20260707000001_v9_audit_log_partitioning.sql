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
-- Convert audit_logs to a range-partitioned table. NO TRANSACTION only prevents
-- goose from wrapping this file in its own transaction; it does not make the
-- conversion resumable. Because goose records a NO TRANSACTION migration as
-- applied only after every statement succeeds, a failure re-runs this file
-- top-to-bottom; every statement below is therefore written to be idempotent so
-- a re-run converges instead of erroring.
--
-- The work is: build the partitioned table under a staging name, backfill it in
-- bounded id batches, verify the row counts, then swap it into place under the
-- original audit_logs name so existing audit logging keeps working against a
-- single table.
--
-- Note on resumability: the backfill is a single DO block and therefore runs as
-- one transaction, so a pod restart mid-backfill rolls it back entirely -- there
-- is no committed partial progress. The migration does not attempt to resume;
-- instead the contract is a single idempotent full pass that repeats from the
-- start on any failure, converging via idempotency (staging is dropped and
-- rebuilt each run, ON CONFLICT DO NOTHING on the backfill, IF NOT EXISTS on
-- partitions/indexes, and completion is detected by audit_logs already being
-- partitioned).
--
-- Note on batching: this deployment drives migrations through pgx v5 stdlib in
-- the default (extended) query protocol, which rejects COMMIT inside a DO block.
-- The backfill loop therefore cannot commit per batch; it runs as one implicit
-- transaction. This is the underlying reason partial progress cannot be
-- preserved. Batching is retained solely to bound per-iteration memory and plan
-- cost. For a tenant so large that even a single-pass copy exceeds the startup
-- window, run this out-of-band with DisableMigrations set rather than at pod
-- startup.

-- Phase 1: staging table. Rebuilt from empty on every run (no-resume contract):
-- drop any leftover audit_logs_partitioned from a prior interrupted run, then
-- create it fresh. Skipped once a prior run already swapped the partitioned table
-- into place under the audit_logs name (detected by audit_logs being range
-- partitioned), so there is nothing left to stage.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_partitioned_table pt
        JOIN pg_class c ON c.oid = pt.partrelid
        WHERE c.relname = 'audit_logs' AND pt.partstrat = 'r'
    ) THEN
        RETURN;
    END IF;

    DROP TABLE IF EXISTS audit_logs_partitioned;
    CREATE TABLE audit_logs_partitioned (
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
END $$;
-- +goose StatementEnd

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

-- Phase 3: batched backfill. Copy rows in bounded id windows, always starting
-- from the beginning (id 0): staging is freshly empty each run (Phase 1), so this
-- is a full pass, and ON CONFLICT DO NOTHING is a defensive idempotency guard.
-- Skipped once the swap has happened (staging no longer exists).
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
    window_lo := 0;

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

-- Phase 4: verify counts, then swap staging into place with a rename-aside.
-- Guarded so a re-run after a completed swap is a no-op (staging is gone). The
-- count-equality check assumes writes to audit_logs are quiesced during the
-- migration (startup, HA-leader gated); a mismatch aborts and leaves the staging
-- rows for the next run to rebuild rather than swapping in an incomplete copy.
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

    -- Detach the sequence from the current audit_logs.id before renaming the table
    -- aside. The sequence is OWNED BY audit_logs.id (see the init migration);
    -- ownership follows the rename, so we re-attach it to the swapped-in table's
    -- column below to keep it from being dropped alongside audit_logs_old.
    EXECUTE 'ALTER SEQUENCE audit_logs_id_seq OWNED BY NONE';
    -- Rename-aside swap: retain the original as audit_logs_old (not dropped) so the
    -- swap stays reversible while indexes are built; a later step drops it once the
    -- new table is verified.
    EXECUTE 'ALTER TABLE audit_logs RENAME TO audit_logs_old';
    EXECUTE 'ALTER TABLE audit_logs_partitioned RENAME TO audit_logs';
    EXECUTE 'ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id';

    -- No table grants exist on audit_logs in this schema (single owner role; no
    -- GRANT statements in any migration), so there are no grants to re-apply from
    -- audit_logs_old. If a non-owner role is ever granted access out-of-band, copy
    -- its grants onto the swapped-in table here.

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

-- Phase 6: drop the retained original now that the swapped-in table is verified
-- and indexed. IF EXISTS so a re-run after the drop (or a run that never created
-- it) is a no-op. The sequence was re-owned to the new audit_logs.id above, so
-- dropping audit_logs_old does not take the sequence with it.
DROP TABLE IF EXISTS audit_logs_old;

-- +goose Down
-- The Up block re-attaches audit_logs_id_seq to the partitioned audit_logs.id,
-- so dropping the table with CASCADE also drops the owned sequence. Also drop the
-- staging and rename-aside tables in case Down runs against a half-completed Up
-- (before the swap, or after the swap but before audit_logs_old is dropped).
-- Recreate the sequence before the table that references it, then re-own it to
-- the new column. Guards are idempotent because Down also runs NO TRANSACTION.
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS audit_logs_partitioned CASCADE;
DROP TABLE IF EXISTS audit_logs_old CASCADE;

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