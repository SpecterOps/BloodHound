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

-- Set `back_button_support` feature flag as user updatable
UPDATE feature_flags SET user_updatable = true WHERE key = 'back_button_support';

-- Specify the `back_button_support` feature flag is currently only for BHCE users
UPDATE feature_flags SET description = 'Enable users to quickly navigate between views in a wider range of scenarios by utilizing the browser navigation buttons. Currently for BloodHound Community Edition users only.' WHERE key = 'back_button_support';