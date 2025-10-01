-- Copyright 2025 Specter Ops, Inc.
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

-- Graph Operations Replay Log
-- This table stores a linear history of all graph modification operations (node/edge creates, updates, deletes)
-- for replay and audit purposes. The timestamp is authoritative and operations are treated as sequential
-- with no branching.
CREATE TABLE IF NOT EXISTS graph_operations_replay_log (
    id SERIAL PRIMARY KEY,

    -- Type of change operation
    change_type VARCHAR(10) NOT NULL CHECK (change_type IN ('create', 'update', 'delete')),

    -- Type of graph object being modified
    object_type VARCHAR(10) NOT NULL CHECK (object_type IN ('node', 'edge')),

    -- Identifier of the object (for nodes: objectid, for edges: composite key)
    object_id VARCHAR(255) NOT NULL,

    -- For nodes: the node labels/kinds (e.g., ["User", "Base"])
    -- For edges: the edge kind (e.g., "MemberOf")
    labels JSONB,

    -- For edges: source node object_id
    source_object_id VARCHAR(255),

    -- For edges: target node object_id
    target_object_id VARCHAR(255),

    -- All properties associated with the object at the time of change
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Timestamps (required by GORM Serial/Basic embedding)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Index for efficient retrieval of recent changes
    CONSTRAINT chk_edge_requires_endpoints CHECK (
        object_type != 'edge' OR (source_object_id IS NOT NULL AND target_object_id IS NOT NULL)
    )
);

-- Index for retrieving recent replay log entries efficiently
CREATE INDEX IF NOT EXISTS idx_graph_operations_replay_log_created_at
    ON graph_operations_replay_log(created_at DESC);

-- Index for finding changes by object
CREATE INDEX IF NOT EXISTS idx_graph_operations_replay_log_object
    ON graph_operations_replay_log(object_type, object_id);

-- Index for finding edge changes
CREATE INDEX IF NOT EXISTS idx_graph_operations_replay_log_edge
    ON graph_operations_replay_log(source_object_id, target_object_id)
    WHERE object_type = 'edge';
