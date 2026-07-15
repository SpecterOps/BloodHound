// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

import type { Page } from '@playwright/test';
import { installAssetGroupTagMemberStub } from './members';
import { installAssetGroupTagsSearchStub } from './search';
import { installAssetGroupTagSelectorStub } from './selectors';
import { installAssetGroupTagStub } from './tag';
import { installAssetGroupTagsZoneDetailsStub } from './zone-details';

const labelTag = {
    id: 2,
    name: 'PLAYWRIGHT_LABEL',
    type: 2,
    position: 2,
    description: 'Playwright stubbed label',
    analysis_enabled: false,
    require_certify: false,
};

/** Installs deterministic Label data for a single Label request. */
export async function installAssetGroupLabelStub(page: Page): Promise<void> {
    await installAssetGroupTagStub(page, { data: { tag: labelTag } });
}

/** Installs a deterministic Label rule for a single rule request. */
export async function installAssetGroupLabelSelectorStub(page: Page): Promise<void> {
    await installAssetGroupTagSelectorStub(page, {
        data: {
            selector: {
                id: 3303,
                asset_group_tag_id: labelTag.id,
                name: 'PLAYWRIGHT_LABEL_RULE_1',
                description: 'Playwright stubbed label rule',
                seeds: [{ selector_id: 3303, type: 1, value: 'S-1-5-21-PW-4401' }],
            },
        },
    });
}

/** Installs deterministic Label member and node data for a single member request. */
export async function installAssetGroupLabelMemberStub(page: Page): Promise<void> {
    await installAssetGroupTagMemberStub(page, {
        data: {
            member: {
                asset_group_tag_id: labelTag.id,
                id: 4401,
                name: 'PLAYWRIGHT_LABEL_COMPUTER_1',
                object_id: 'S-1-5-21-PW-4401',
                properties: {
                    name: 'PLAYWRIGHT_LABEL_COMPUTER_1',
                    objectid: 'S-1-5-21-PW-4401',
                },
                selectors: [{ id: 3303, asset_group_tag_id: labelTag.id, name: 'PLAYWRIGHT_LABEL_RULE_1' }],
            },
        },
    });

    await page.route(/\/api\/v2\/nodes\/\d+(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({
            json: {
                data: {
                    node_id: 4401,
                    kinds: [{ name: 'Computer', node_kind_id: null }],
                    properties: {
                        name: 'PLAYWRIGHT_LABEL_COMPUTER_1',
                        objectid: 'S-1-5-21-PW-4401',
                    },
                },
            },
        });
    });

    // The shared Object Information panel loads count data for each Computer relationship section.
    // Return empty results for all of them so the panel does not render request-error indicators.
    await page.route(/\/api\/v2\/computers\/[^/?]+\/[^/?]+(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: [], count: 0, limit: 128, skip: 0 } });
    });
}

/** Installs Label-specific search results. */
export async function installAssetGroupLabelsSearchStub(page: Page): Promise<void> {
    await installAssetGroupTagsSearchStub(page, {
        data: {
            tags: [{ id: labelTag.id, name: labelTag.name }],
            selectors: [{ id: 3303, asset_group_tag_id: labelTag.id, name: 'PLAYWRIGHT_LABEL_ADMIN_RULE' }],
            members: [
                {
                    asset_group_tag_id: labelTag.id,
                    id: 4401,
                    primary_kind: 'User',
                    object_id: 'S-1-5-21-PW-4401',
                    name: 'PLAYWRIGHT_LABEL_ADMIN_USER',
                    source: 1,
                },
            ],
        },
    });
}

/**
 * Installs deterministic Label detail data. Labels have custom rules and objects, but never
 * Default Rules, unlike Privilege Zones.
 */
export async function installAssetGroupLabelDetailsStub(page: Page): Promise<void> {
    const rules = [
        {
            id: 3303,
            asset_group_tag_id: labelTag.id,
            name: 'PLAYWRIGHT_LABEL_RULE_1',
            description: 'Playwright stubbed label rule',
            is_default: false,
            allow_disable: true,
            auto_certify: false,
            seeds: [],
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-02T00:00:00Z',
            disabled_at: null,
        },
        {
            id: 3304,
            asset_group_tag_id: labelTag.id,
            name: 'PLAYWRIGHT_LABEL_RULE_2',
            description: 'Playwright stubbed label rule',
            is_default: false,
            allow_disable: true,
            auto_certify: false,
            seeds: [],
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-02T00:00:00Z',
            disabled_at: null,
        },
    ];
    const objects = [
        {
            asset_group_tag_id: labelTag.id,
            id: 4401,
            primary_kind: 'Computer',
            object_id: 'S-1-5-21-PW-4401',
            name: 'PLAYWRIGHT_LABEL_COMPUTER_1',
            source: 1,
        },
        {
            asset_group_tag_id: labelTag.id,
            id: 4402,
            primary_kind: 'Computer',
            object_id: 'S-1-5-21-PW-4402',
            name: 'PLAYWRIGHT_LABEL_COMPUTER_2',
            source: 1,
        },
        {
            asset_group_tag_id: labelTag.id,
            id: 4403,
            primary_kind: 'Group',
            object_id: 'S-1-5-21-PW-4403',
            name: 'PLAYWRIGHT_LABEL_GROUP_1',
            source: 1,
        },
    ];

    await installAssetGroupTagsZoneDetailsStub(page, {
        data: {
            tag: {
                ...labelTag,
                kind_id: 10,
                glyph: null,
                created_at: '2024-01-01T00:00:00Z',
                created_by: 'playwright',
                updated_at: '2024-01-02T00:00:00Z',
                updated_by: 'playwright',
                deleted_at: null,
                deleted_by: null,
                counts: {
                    selectors: rules.length,
                    custom_selectors: rules.length,
                    default_selectors: 0,
                    disabled_selectors: 0,
                    members: objects.length,
                },
            },
            rules,
            objects,
        },
    });
}
