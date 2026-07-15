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

type StubZoneCounts = {
    selectors: number;
    custom_selectors: number;
    default_selectors: number;
    disabled_selectors: number;
    members: number;
};

type StubZoneTag = {
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
    counts: StubZoneCounts;
};

type StubRule = {
    id: number;
    asset_group_tag_id: number;
    name: string;
    description: string;
    is_default: boolean;
    allow_disable: boolean;
    auto_certify: boolean;
    seeds: unknown[];
    created_at: string;
    updated_at: string;
    disabled_at: string | null;
};

type StubObject = {
    asset_group_tag_id: number;
    id: number;
    primary_kind: string;
    object_id: string;
    name: string;
    source: number;
};

export type AssetGroupTagsZoneDetailsStubData = {
    tag?: StubZoneTag;
    rules?: StubRule[];
    objects?: StubObject[];
};

const buildRule = (overrides: Partial<StubRule> & Pick<StubRule, 'id' | 'name' | 'is_default'>): StubRule => ({
    asset_group_tag_id: 1,
    description: 'Playwright stubbed rule',
    allow_disable: true,
    auto_certify: false,
    seeds: [],
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
    disabled_at: null,
    ...overrides,
});

const DEFAULT_RULES: StubRule[] = [
    buildRule({ id: 3001, name: 'PLAYWRIGHT_DEFAULT_RULE_1', is_default: true }),
    buildRule({ id: 3002, name: 'PLAYWRIGHT_DEFAULT_RULE_2', is_default: true }),
    buildRule({ id: 3003, name: 'PLAYWRIGHT_CUSTOM_RULE_1', is_default: false }),
    buildRule({ id: 3004, name: 'PLAYWRIGHT_CUSTOM_RULE_2', is_default: false }),
];

const DEFAULT_OBJECTS: StubObject[] = [
    {
        asset_group_tag_id: 1,
        id: 4001,
        primary_kind: 'Computer',
        object_id: 'S-1-5-21-PW-4001',
        name: 'PLAYWRIGHT_COMPUTER_1',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 4002,
        primary_kind: 'Computer',
        object_id: 'S-1-5-21-PW-4002',
        name: 'PLAYWRIGHT_COMPUTER_2',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 4003,
        primary_kind: 'Computer',
        object_id: 'S-1-5-21-PW-4003',
        name: 'PLAYWRIGHT_COMPUTER_3',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 4004,
        primary_kind: 'Domain',
        object_id: 'S-1-5-21-PW-4004',
        name: 'PLAYWRIGHT_DOMAIN_1',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 4005,
        primary_kind: 'Domain',
        object_id: 'S-1-5-21-PW-4005',
        name: 'PLAYWRIGHT_DOMAIN_2',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 4006,
        primary_kind: 'Group',
        object_id: 'S-1-5-21-PW-4006',
        name: 'PLAYWRIGHT_GROUP_1',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 4007,
        primary_kind: 'Group',
        object_id: 'S-1-5-21-PW-4007',
        name: 'PLAYWRIGHT_GROUP_2',
        source: 1,
    },
];

const buildCounts = (rules: StubRule[], objects: StubObject[]): StubZoneCounts => ({
    selectors: rules.length,
    custom_selectors: rules.filter((rule) => !rule.is_default && rule.disabled_at === null).length,
    default_selectors: rules.filter((rule) => rule.is_default && rule.disabled_at === null).length,
    disabled_selectors: rules.filter((rule) => rule.disabled_at !== null).length,
    members: objects.length,
});

const buildTag = (rules: StubRule[], objects: StubObject[]): StubZoneTag => ({
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
    counts: buildCounts(rules, objects),
});

export type AssetGroupTagsZoneDetailsStubOptions = {
    data?: AssetGroupTagsZoneDetailsStubData;
};

/**
 * Stubs the Privilege Zones "Zone Details" data endpoints so the detail page renders deterministic
 * Rules (including Default Rules) and Objects (Computers, Domains, Groups) without touching real
 * data. It installs handlers for the tag list (`GET /api/v2/asset-group-tags`), object counts
 * (`.../:id/members/counts`), rules (`.../:id/selectors`), and members (`.../:id/members`). The
 * selectors handler honors the `is_default`/`disabled_at` filters and the members handler honors the
 * `primary_kind` filter so expanding a section returns the matching entries. Non-GET traffic falls
 * through to any lower-priority route handlers.
 */
export async function installAssetGroupTagsZoneDetailsStub(
    page: Page,
    opts: AssetGroupTagsZoneDetailsStubOptions = {}
): Promise<void> {
    const rules = opts.data?.rules ?? DEFAULT_RULES;
    const objects = opts.data?.objects ?? DEFAULT_OBJECTS;
    const tag = opts.data?.tag ?? buildTag(rules, objects);

    await page.route(/\/api\/v2\/asset-group-tags(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        return route.fulfill({ json: { data: { tags: [tag] } } });
    });

    await page.route(/\/api\/v2\/asset-group-tags\/\d+\/members\/counts(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        const counts = objects.reduce<Record<string, number>>((acc, object) => {
            acc[object.primary_kind] = (acc[object.primary_kind] ?? 0) + 1;
            return acc;
        }, {});
        return route.fulfill({ json: { data: { total_count: objects.length, counts } } });
    });

    await page.route(/\/api\/v2\/asset-group-tags\/\d+\/selectors(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        const params = new URL(route.request().url()).searchParams;
        const isDefault = params.get('is_default');
        const disabledAt = params.get('disabled_at');
        const filtered = rules.filter((rule) => {
            if (isDefault === 'eq:true' && !rule.is_default) return false;
            if (isDefault === 'eq:false' && rule.is_default) return false;
            if (disabledAt === 'eq:null' && rule.disabled_at !== null) return false;
            if (disabledAt === 'neq:null' && rule.disabled_at === null) return false;
            return true;
        });
        return route.fulfill({
            json: { data: { selectors: filtered }, count: filtered.length, limit: filtered.length, skip: 0 },
        });
    });

    await page.route(/\/api\/v2\/asset-group-tags\/\d+\/members(\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        const primaryKind = new URL(route.request().url()).searchParams.get('primary_kind');
        const kind = primaryKind?.startsWith('eq:') ? primaryKind.slice(3) : primaryKind;
        const filtered = kind ? objects.filter((object) => object.primary_kind === kind) : objects;
        return route.fulfill({
            json: { data: { members: filtered }, count: filtered.length, limit: filtered.length, skip: 0 },
        });
    });
}
