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
-- Build the new partitioned table under a staging name, backfill it from the
-- existing audit_logs, drop the original, then rename the staging table back to
-- audit_logs. This leaves a single table under the original name so existing
-- audit logging continues to work without maintaining two tables.
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

-- Child partition names are chosen to match the final audit_logs parent name so
-- the GC daemon's partition maintenance (audit_logs_YYYY_MM) stays aligned after
-- the rename below.
-- +goose StatementBegin
DO $$
DECLARE
    start_date DATE := '2024-01-01';
    end_date DATE := '2026-08-01';
    current_month DATE;
BEGIN
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
END $$;
-- +goose StatementEnd

CREATE TABLE audit_logs_default PARTITION OF audit_logs_partitioned DEFAULT;

INSERT INTO audit_logs_partitioned (
    id, created_at, action, actor_id, actor_name, actor_email,
    request_id, source_ip_address, status, commit_id, fields, source
)
SELECT
    id,
    COALESCE(created_at, '2020-01-01'::timestamptz) AS created_at,
    action,
    actor_id,
    actor_name,
    actor_email,
    request_id,
    source_ip_address,
    status,
    commit_id,
    fields,
    'legacy' AS source
FROM audit_logs;

DROP TABLE audit_logs;

ALTER TABLE audit_logs_partitioned RENAME TO audit_logs;

ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;

-- Advance the sequence past the largest copied id so the next insert does not
-- collide with a backfilled row.
SELECT setval('audit_logs_id_seq', COALESCE((SELECT MAX(id) FROM audit_logs), 1));

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_actor_email ON audit_logs(actor_email);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_source_ip_address ON audit_logs(source_ip_address);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);
CREATE INDEX idx_audit_logs_source ON audit_logs(source);

-- +goose Down
DROP TABLE IF EXISTS audit_logs CASCADE;

CREATE TABLE audit_logs (
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

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_actor_email ON audit_logs(actor_email);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_source_ip_address ON audit_logs(source_ip_address);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);