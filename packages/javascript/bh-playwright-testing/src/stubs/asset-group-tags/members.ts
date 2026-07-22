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

type StubMemberSelector = {
    id: number;
    asset_group_tag_id: number;
    name: string;
};

type StubMember = {
    asset_group_tag_id: number;
    id: number;
    primary_kind: string;
    object_id: string;
    name: string;
    source: number;
    properties: Record<string, unknown>;
    selectors: StubMemberSelector[];
};

export type AssetGroupTagMemberStubData = {
    member?: Partial<StubMember>;
    /** Entity kinds returned by the node-properties endpoint (e.g. `GET /api/v2/computers/:id`). */
    entityKinds?: string[];
    /** Entity properties returned by the node-properties endpoint. */
    entityProperties?: Record<string, unknown>;
};

const buildMember = (overrides: Partial<StubMember> = {}): StubMember => ({
    asset_group_tag_id: 1,
    id: 4001,
    primary_kind: 'Computer',
    object_id: 'S-1-5-21-PW-4001',
    name: 'PLAYWRIGHT_COMPUTER_1',
    source: 1,
    properties: {
        name: 'PLAYWRIGHT_COMPUTER_1',
        objectid: 'S-1-5-21-PW-4001',
    },
    selectors: [{ id: 3003, asset_group_tag_id: 1, name: 'PLAYWRIGHT_CUSTOM_RULE_1' }],
    ...overrides,
});

export type AssetGroupTagMemberStubOptions = {
    data?: AssetGroupTagMemberStubData;
};

/**
 * Stubs the single-member endpoint (`GET /api/v2/asset-group-tags/:tagId/members/:memberId`) plus
 * the node-properties endpoint the shared Entity Info panel resolves for the member's kind (a
 * Computer by default, i.e. `GET /api/v2/computers/:objectId`). Together these let the Object side
 * panel render deterministic member info and its "Rules" section without hitting the real backend.
 * Non-GET traffic falls through to any lower-priority route handlers.
 */
export async function installAssetGroupTagMemberStub(
    page: Page,
    opts: AssetGroupTagMemberStubOptions = {}
): Promise<void> {
    const member = buildMember(opts.data?.member);
    const kinds = opts.data?.entityKinds ?? [member.primary_kind, 'Base'];
    const properties = opts.data?.entityProperties ?? member.properties;

    await page.route(/\/api\/v2\/asset-group-tags\/\d+\/members\/\d+(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { member } } });
    });

    await page.route(/\/api\/v2\/computers\/[^/?]+(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { kinds, props: properties } } });
    });
}
