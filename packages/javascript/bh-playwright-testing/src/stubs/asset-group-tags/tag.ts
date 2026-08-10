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

type StubTag = {
    id: number;
    name: string;
    kind_id: number;
    type: number;
    position: number;
    require_certify: boolean;
    description: string;
    analysis_enabled: boolean;
    glyph: string | null;
    created_at: string;
    created_by: string;
    updated_at: string;
    updated_by: string;
    deleted_at: string | null;
    deleted_by: string | null;
};

export type AssetGroupTagStubData = {
    tag?: Partial<StubTag>;
};

const buildTag = (overrides: Partial<StubTag> = {}): StubTag => ({
    id: 1,
    name: 'PLAYWRIGHT_ZONE',
    kind_id: 10,
    type: 1,
    position: 1,
    require_certify: false,
    description: 'Playwright stubbed zone',
    analysis_enabled: true,
    glyph: null,
    created_at: '2024-01-01T00:00:00Z',
    created_by: 'playwright',
    updated_at: '2024-01-02T00:00:00Z',
    updated_by: 'playwright',
    deleted_at: null,
    deleted_by: null,
    ...overrides,
});

export type AssetGroupTagStubOptions = {
    data?: AssetGroupTagStubData;
};

/**
 * Stubs the single-tag endpoint (`GET /api/v2/asset-group-tags/:tagId`) so components that call
 * `useAssetGroupTagInfo` (the Zone side panel, the Edit Zone form, and both Rule forms) render
 * deterministic tag data instead of hitting the real backend. The route intentionally matches only
 * a trailing numeric id so it does not collide with the tag-list route or nested sub-resource
 * routes (e.g. `/selectors`, `/members`). Non-GET traffic falls through to any lower-priority route
 * handlers.
 */
export async function installAssetGroupTagStub(page: Page, opts: AssetGroupTagStubOptions = {}): Promise<void> {
    const tag = buildTag(opts.data?.tag);

    await page.route(/\/api\/v2\/asset-group-tags\/\d+(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { tag } } });
    });
}
