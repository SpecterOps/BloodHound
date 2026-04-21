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

-- Webhooks subsystem tables

-- events carries the domain events that may be delivered to subscribed webhooks.
-- id is stored as text holding the canonical string form of a UUIDv7 so events
-- remain naturally time-ordered. Event ingestion is owned by a separate changeset.
CREATE TABLE IF NOT EXISTS events
(
    id           text                     NOT NULL,
    type         text                     NOT NULL,
    message      text                     NOT NULL DEFAULT '',
    data         jsonb                    NOT NULL DEFAULT '{}'::jsonb,
    created_at   timestamp with time zone NOT NULL DEFAULT current_timestamp,
    processed_at timestamp with time zone,
    PRIMARY KEY (id)
);

-- webhooks is the top-level registration of an outbound delivery target.
-- id is stored as text holding the canonical string form of a UUIDv4.
CREATE TABLE IF NOT EXISTS webhooks
(
    id          text                     NOT NULL,
    type        text                     NOT NULL,
    name        text                     NOT NULL,
    description text                     NOT NULL DEFAULT '',
    url         text                     NOT NULL,
    created_at  timestamp with time zone NOT NULL DEFAULT current_timestamp,
    created_by  text                     NOT NULL,
    updated_at  timestamp with time zone NOT NULL DEFAULT current_timestamp,
    updated_by  text                     NOT NULL,
    disabled_at timestamp with time zone,
    disabled_by text,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS webhooks_url_unique_index ON webhooks (url);

-- webhook_secrets stores the HMAC signing secret for a webhook (1:1 with webhooks).
-- Kept in a separate table so the secret can be protected independently of the
-- record that is regularly selected by the UI.
CREATE TABLE IF NOT EXISTS webhook_secrets
(
    webhook_id  text                     NOT NULL,
    hmac_secret text                     NOT NULL,
    created_at  timestamp with time zone NOT NULL DEFAULT current_timestamp,
    PRIMARY KEY (webhook_id),
    CONSTRAINT fk_webhook_secrets_webhooks
        FOREIGN KEY (webhook_id) REFERENCES webhooks (id) ON DELETE CASCADE
);

-- webhook_metadata tracks rolling delivery health per webhook (1:1 with webhooks).
CREATE TABLE IF NOT EXISTS webhook_metadata
(
    webhook_id        text                     NOT NULL,
    health            double precision         NOT NULL DEFAULT 1.0,
    attempts          integer                  NOT NULL DEFAULT 0,
    failures          integer                  NOT NULL DEFAULT 0,
    last_error        text,
    last_errored_at   timestamp with time zone,
    last_succeeded_at timestamp with time zone,
    PRIMARY KEY (webhook_id),
    CONSTRAINT fk_webhook_metadata_webhooks
        FOREIGN KEY (webhook_id) REFERENCES webhooks (id) ON DELETE CASCADE
);

-- webhook_subscriptions maps a webhook to the event types it consumes.
-- version is a payload attribute, not part of identity: a webhook can subscribe
-- to any given event_type exactly once.
CREATE TABLE IF NOT EXISTS webhook_subscriptions
(
    webhook_id text NOT NULL,
    event_type text NOT NULL,
    version    text NOT NULL,
    PRIMARY KEY (webhook_id, event_type),
    CONSTRAINT fk_webhook_subscriptions_webhooks
        FOREIGN KEY (webhook_id) REFERENCES webhooks (id) ON DELETE CASCADE
);

-- webhook_events records the delivery state of a single event to a single webhook.
CREATE TABLE IF NOT EXISTS webhook_events
(
    webhook_id       text                     NOT NULL,
    event_id         text                     NOT NULL,
    created_at       timestamp with time zone NOT NULL DEFAULT current_timestamp,
    last_status_code integer                  NOT NULL DEFAULT 0,
    last_error       text,
    attempts         integer                  NOT NULL DEFAULT 0,
    PRIMARY KEY (webhook_id, event_id),
    CONSTRAINT fk_webhook_events_webhooks
        FOREIGN KEY (webhook_id) REFERENCES webhooks (id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_events_events
        FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE
);

