// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { Flag } from 'bh-shared-ui';
import { uniqBy } from 'lodash';

type FeatureFlagOption = Partial<Flag> & Pick<Flag, 'key' | 'enabled'>;
export const createFeatureFlags = (featureFlags: FeatureFlagOption[]) => {
    const now = new Date();
    const defaults = [
        {
            key: 'enable_saml_sso',
            enabled: true,
        },
        {
            key: 'scope_collection_by_ou',
            enabled: true,
        },
        {
            key: 'azure_support',
            enabled: true,
        },
        {
            key: 'reconciliation',
            enabled: true,
        },
        {
            key: 'entity_panel_cache',
            enabled: true,
        },
        {
            key: 'pg_migration_dual_ingest',
            enabled: false,
        },
        {
            key: 'clear_graph_data',
            enabled: true,
        },
        {
            key: 'risk_exposure_new_calculation',
            enabled: false,
        },
        {
            key: 'fedramp_eula',
            enabled: false,
        },
        {
            key: 'auto_tag_t0_parent_objects',
            enabled: false,
        },
        {
            key: 'dark_mode',
            enabled: true,
        },
        {
            key: 'adcs',
            enabled: false,
        },
    ];
    const flags = uniqBy([...featureFlags, ...defaults], 'key');
    return flags.map((flag, i) => ({
        id: i,
        created_at: now.toISOString(),
        updated_at: now.toISOString(),
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
        key: flag.key,
        name: (flag as any).name ?? '',
        description: (flag as any).description ?? '',
        enabled: flag.enabled,
        user_updatable: (flag as any).user_updatable ?? false,
    }));
};
