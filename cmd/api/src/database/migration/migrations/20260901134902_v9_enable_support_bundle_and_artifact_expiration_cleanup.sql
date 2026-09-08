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
UPDATE feature_flags
SET enabled = TRUE
WHERE
  key = 'collector_support_bundle'
  OR key = 'artifact_expiration_cleanup';

-- +goose Down
UPDATE feature_flags
SET enabled = FALSE
WHERE
  key = 'collector_support_bundle'
  OR key = 'artifact_expiration_cleanup';
