// Copyright 2026 Specter Ops, Inc.
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

import type { Page } from '@playwright/test';

type StubHistoryRecord = {
    id: number;
    created_at: string;
    actor: string;
    email: string | null;
    action: string;
    target: string;
    asset_group_tag_id: number;
    environment_id: string | null;
    note: string | null;
};

type StubHistoryTag = {
    id: number;
    name: string;
};

export type AssetGroupTagsHistoryStubData = {
    records?: StubHistoryRecord[];
    tags?: StubHistoryTag[];
};

const DEFAULT_TAGS: StubHistoryTag[] = [
    { id: 1, name: 'PLAYWRIGHT_ZONE' },
    { id: 2, name: 'PLAYWRIGHT_LABEL' },
];

const DEFAULT_RECORDS: StubHistoryRecord[] = [
    {
        id: 5001,
        created_at: '2024-01-03T12:00:00Z',
        actor: 'playwright',
        email: 'playwright@specterops.io',
        action: 'CreateTag',
        target: 'PLAYWRIGHT_ZONE',
        asset_group_tag_id: 1,
        environment_id: null,
        note: 'Created the Playwright zone for accessibility testing.',
    },
    {
        id: 5002,
        created_at: '2024-01-04T09:30:00Z',
        actor: 'playwright',
        email: 'playwright@specterops.io',
        action: 'UpdateSelector',
        target: 'PLAYWRIGHT_ADMIN_RULE',
        asset_group_tag_id: 1,
        environment_id: null,
        note: null,
    },
    {
        id: 5003,
        created_at: '2024-01-05T15:45:00Z',
        actor: 'playwright',
        email: 'playwright@specterops.io',
        action: 'CreateSelector',
        target: 'PLAYWRIGHT_LABEL_RULE',
        asset_group_tag_id: 2,
        environment_id: null,
        note: 'Added a rule to the Playwright label.',
    },
];

export type AssetGroupTagsHistoryStubOptions = {
    data?: AssetGroupTagsHistoryStubData;
};

/**
 * Stubs the Privilege Zones History Log endpoint (`GET` and `POST /api/v2/asset-group-tags-history`)
 * so the History tab renders deterministic history records without touching real data. The route is
 * anchored to the `-history` suffix so it does not collide with the tag-list route
 * (`/api/v2/asset-group-tags`). Both GET (no search) and POST (search) return the same records; the
 * search input is not filtered here so tests control the rendered rows entirely through `records`.
 *
 * The tag-list route (`GET /api/v2/asset-group-tags`) is also stubbed so the History table can
 * resolve each record's Zone/Label name (`tagName`) and so the filter dialog's Zone/Label select
 * renders deterministic options. Pass `data.records: []` to exercise the empty state, or override
 * `data.records`/`data.tags` to shape the loaded state.
 *
 * Non-matching HTTP methods fall through to any lower-priority route handlers.
 */
export async function installAssetGroupTagsHistoryStub(
    page: Page,
    opts: AssetGroupTagsHistoryStubOptions = {}
): Promise<void> {
    const records = opts.data?.records ?? DEFAULT_RECORDS;
    const tags = opts.data?.tags ?? DEFAULT_TAGS;

    await page.route(/\/api\/v2\/asset-group-tags-history(\?.*)?$/, async (route) => {
        const method = route.request().method();
        if (method !== 'GET' && method !== 'POST') return route.fallback();

        return route.fulfill({
            json: {
                data: { records },
                count: records.length,
                limit: records.length,
                skip: 0,
            },
        });
    });

    await page.route(/\/api\/v2\/asset-group-tags(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { tags } } });
    });
}
