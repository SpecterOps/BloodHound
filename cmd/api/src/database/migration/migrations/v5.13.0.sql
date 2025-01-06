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

insert into feature_flags (key, name, description, enabled, user_updatable, created_at, updated_at)
values ('dark_mode', 'Dark Mode', 'Allows users to enable or disable dark mode via a toggle in the settings menu', false, true, now(), now())
on conflict (key) do nothing;
