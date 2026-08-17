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

type StubSearchMember = {
    asset_group_tag_id: number;
    id: number;
    primary_kind: string;
    object_id: string;
    name: string;
    source: number;
};

type StubSearchSelector = {
    id: number;
    asset_group_tag_id: number;
    name: string;
};

type StubSearchTag = {
    id: number;
    name: string;
};

export type AssetGroupTagsSearchStubData = {
    tags?: StubSearchTag[];
    selectors?: StubSearchSelector[];
    members?: StubSearchMember[];
};

const DEFAULT_MEMBERS: StubSearchMember[] = [
    {
        asset_group_tag_id: 1,
        id: 1001,
        primary_kind: 'User',
        object_id: 'S-1-5-21-PLAYWRIGHT-1001',
        name: 'PLAYWRIGHT_ADMIN_USER',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 1002,
        primary_kind: 'Group',
        object_id: 'S-1-5-21-PLAYWRIGHT-1002',
        name: 'PLAYWRIGHT_ADMIN_GROUP',
        source: 1,
    },
    {
        asset_group_tag_id: 1,
        id: 1003,
        primary_kind: 'Computer',
        object_id: 'S-1-5-21-PLAYWRIGHT-1003',
        name: 'PLAYWRIGHT_WORKSTATION',
        source: 1,
    },
];

const DEFAULT_SELECTORS: StubSearchSelector[] = [{ id: 2001, asset_group_tag_id: 1, name: 'PLAYWRIGHT_ADMIN_RULE' }];

const DEFAULT_TAGS: StubSearchTag[] = [{ id: 1, name: 'PLAYWRIGHT_ZONE' }];

const DEFAULT_DATA: Required<AssetGroupTagsSearchStubData> = {
    tags: DEFAULT_TAGS,
    selectors: DEFAULT_SELECTORS,
    members: DEFAULT_MEMBERS,
};

const matchesQuery = (name: string, query: string): boolean => name.toLowerCase().includes(query.toLowerCase());

export type AssetGroupTagsSearchStubOptions = {
    data?: AssetGroupTagsSearchStubData;
};

/**
 * Stubs the Privilege Zones search endpoint (`POST /api/v2/asset-group-tags/search`) so the detail
 * page search bar can load Objects (members), Rules (selectors), and tags without touching real
 * data. The handler reads the `query` from the request body and returns only the entries whose
 * name contains that query (case-insensitive), so the stubbed response reflects the search input.
 * Non-POST traffic falls through to any lower-priority route handlers.
 */
export async function installAssetGroupTagsSearchStub(
    page: Page,
    opts: AssetGroupTagsSearchStubOptions = {}
): Promise<void> {
    const tags = opts.data?.tags ?? DEFAULT_DATA.tags;
    const selectors = opts.data?.selectors ?? DEFAULT_DATA.selectors;
    const members = opts.data?.members ?? DEFAULT_DATA.members;

    await page.route(/\/api\/v2\/asset-group-tags\/search(\?|$)/, async (route) => {
        if (route.request().method() !== 'POST') {
            return route.fallback();
        }

        const query: string = route.request().postDataJSON()?.query ?? '';

        return route.fulfill({
            json: {
                data: {
                    tags: tags.filter((tag) => matchesQuery(tag.name, query)),
                    selectors: selectors.filter((selector) => matchesQuery(selector.name, query)),
                    members: members.filter((member) => matchesQuery(member.name, query)),
                },
            },
        });
    });
}
