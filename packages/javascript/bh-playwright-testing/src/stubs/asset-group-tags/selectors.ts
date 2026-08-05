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

type StubSeed = {
    selector_id: number;
    type: number;
    value: string;
};

type StubSelector = {
    id: number;
    asset_group_tag_id: number;
    name: string;
    description: string;
    is_default: boolean;
    allow_disable: boolean;
    auto_certify: number;
    seeds: StubSeed[];
    created_at: string;
    created_by: string;
    updated_at: string;
    updated_by: string;
    disabled_at: string | null;
    disabled_by: string | null;
};

export type AssetGroupTagSelectorStubData = {
    selector?: Partial<StubSelector>;
};

const buildSelector = (overrides: Partial<StubSelector> = {}): StubSelector => ({
    id: 3003,
    asset_group_tag_id: 1,
    name: 'PLAYWRIGHT_CUSTOM_RULE_1',
    description: 'Playwright stubbed rule',
    is_default: false,
    allow_disable: true,
    auto_certify: 0,
    seeds: [{ selector_id: 3003, type: 1, value: 'S-1-5-21-PW-4001' }],
    created_at: '2024-01-01T00:00:00Z',
    created_by: 'playwright',
    updated_at: '2024-01-02T00:00:00Z',
    updated_by: 'playwright',
    disabled_at: null,
    disabled_by: null,
    ...overrides,
});

export type AssetGroupTagSelectorStubOptions = {
    data?: AssetGroupTagSelectorStubData;
};

/**
 * Stubs the single-selector endpoint (`GET /api/v2/asset-group-tags/:tagId/selectors/:ruleId`) so
 * the Rule side panel (`useRuleInfo`) and the Edit Rule form render deterministic rule data instead
 * of hitting the real backend. The stubbed seed uses an object-id seed (`type: 1`) so the details
 * card renders without mounting the cypher editor. Non-GET traffic falls through to any
 * lower-priority route handlers.
 */
export async function installAssetGroupTagSelectorStub(
    page: Page,
    opts: AssetGroupTagSelectorStubOptions = {}
): Promise<void> {
    const selector = buildSelector(opts.data?.selector);

    await page.route(/\/api\/v2\/asset-group-tags\/\d+\/selectors\/\d+(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { selector } } });
    });
}
