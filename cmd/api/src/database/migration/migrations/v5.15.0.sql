-- Copyright 2024 Specter Ops, Inc.
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

INSERT INTO feature_flags (key, name, description, enabled, user_updatable, created_at, updated_at)
VALUES ('auto_tag_t0_parent_objects', 'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero', 'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.', true, true, current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;
