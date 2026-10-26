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

-- Existing records remain in audit_logs_old during the three-month
-- coexistence period. New records are written only to partitioned audit_logs.

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_log_status') THEN
        CREATE TYPE audit_log_status AS ENUM ('intent', 'success', 'failure');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_log_source') THEN
        CREATE TYPE audit_log_source AS ENUM ('middleware', 'legacy');
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_partitioned_table partitioned_table
        JOIN pg_class table_class ON table_class.oid = partitioned_table.partrelid
        WHERE table_class.relname = 'audit_logs' AND partitioned_table.partstrat = 'r'
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
        actor_email         TEXT DEFAULT NULL,
        request_id          TEXT,
        source_ip_address   TEXT,
        status              audit_log_status DEFAULT 'intent',
        commit_id           TEXT,
        fields              JSONB,
        source              audit_log_source DEFAULT 'middleware',
        PRIMARY KEY (id, created_at)
    ) PARTITION BY RANGE (created_at);
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
    current_month DATE := date_trunc('month', CURRENT_TIMESTAMP)::date;
    partition_end_date DATE := (date_trunc('month', CURRENT_TIMESTAMP) + interval '2 months')::date;
BEGIN
    IF to_regclass('audit_logs_partitioned') IS NULL THEN
        RETURN;
    END IF;

    WHILE current_month < partition_end_date LOOP
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

-- +goose StatementBegin
DO $$
DECLARE
    max_id BIGINT;
    sequence_value BIGINT;
BEGIN
    IF to_regclass('audit_logs_partitioned') IS NULL THEN
        RETURN;
    END IF;

    SELECT MAX(id) INTO max_id FROM audit_logs;
    SELECT last_value INTO sequence_value FROM audit_logs_id_seq;

    ALTER SEQUENCE audit_logs_id_seq OWNED BY NONE;
    ALTER TABLE audit_logs RENAME TO audit_logs_old;

    -- Free the canonical index names for the replacement table.
    ALTER INDEX IF EXISTS idx_audit_logs_created_at RENAME TO idx_audit_logs_old_created_at;
    ALTER INDEX IF EXISTS idx_audit_logs_actor_id RENAME TO idx_audit_logs_old_actor_id;
    ALTER INDEX IF EXISTS idx_audit_logs_actor_email RENAME TO idx_audit_logs_old_actor_email;
    ALTER INDEX IF EXISTS idx_audit_logs_action RENAME TO idx_audit_logs_old_action;
    ALTER INDEX IF EXISTS idx_audit_logs_source_ip_address RENAME TO idx_audit_logs_old_source_ip_address;
    ALTER INDEX IF EXISTS idx_audit_logs_status RENAME TO idx_audit_logs_old_status;

    ALTER TABLE audit_logs_partitioned RENAME TO audit_logs;
    ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;

    IF max_id IS NOT NULL AND max_id > sequence_value THEN
        PERFORM setval('audit_logs_id_seq', max_id, true);
    END IF;
END $$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_email ON audit_logs(actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source_ip_address ON audit_logs(source_ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source ON audit_logs(source);

