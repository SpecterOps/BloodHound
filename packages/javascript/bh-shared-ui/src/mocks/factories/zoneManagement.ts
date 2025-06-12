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

import { faker } from '@faker-js/faker';
import {
    AssetGroupTag,
    AssetGroupTagMemberInfo,
    AssetGroupTagMemberListItem,
    AssetGroupTagSelector,
    AssetGroupTagSelectorSeed,
    AssetGroupTagTypeTier,
    NodeSourceChild,
    SeedTypes,
} from 'js-client-library';

export const createAssetGroupTag = (tagId: number = 0): AssetGroupTag => {
    return {
        id: tagId,
        name: `Tier-${tagId - 1}`,
        kind_id: faker.datatype.number(),
        type: AssetGroupTagTypeTier,
        position: faker.datatype.number({ min: 0, max: 10 }),
        description: faker.random.words(1000),
        created_at: faker.date.past().toISOString(),
        created_by: faker.internet.email(),
        updated_at: faker.date.past().toISOString(),
        updated_by: faker.internet.email(),
        deleted_at: faker.date.past().toISOString(),
        deleted_by: faker.internet.email(),
        requireCertify: faker.datatype.boolean(),
    };
};

export const createAssetGroupTagWithCounts = (tagId: number = 0): AssetGroupTag => {
    return {
        ...createAssetGroupTag(tagId),
        counts: {
            selectors: faker.datatype.number(),
            members: faker.datatype.number(),
        },
    };
};

export const createAssetGroupTags = (count: number = 1) => {
    const data: AssetGroupTag[] = [];

    for (let i = 1; i <= count; i++) {
        const tag = createAssetGroupTagWithCounts(i);
        data.push(tag);
    }

    return data;
};

export const createSelector = (tagId: number = 0, selectorId: number = 0) => {
    const data: AssetGroupTagSelector = {
        id: selectorId,
        asset_group_tag_id: selectorId,
        name: `tier-${tagId - 1}-selector-${selectorId}`,
        allow_disable: faker.datatype.boolean(),
        description: faker.random.words(),
        is_default: faker.datatype.boolean(),
        auto_certify: faker.datatype.boolean(),
        created_at: faker.date.past().toISOString(),
        created_by: faker.internet.email(),
        updated_at: faker.date.past().toISOString(),
        updated_by: faker.internet.email(),
        disabled_at: faker.date.past().toISOString(),
        disabled_by: faker.internet.email(),
        seeds: createSelectorSeeds(10, selectorId),
    };

    return data;
};

export const createSelectorWithCounts = (tagId: number = 0, selectorId: number = 0) => {
    const data: AssetGroupTagSelector = {
        ...createSelector(tagId, selectorId),
        counts: { members: faker.datatype.number() },
    };

    return data;
};

export const createSelectors = (count: number = 10, tagId: number = 0) => {
    const data: AssetGroupTagSelector[] = [];

    for (let i = 0; i < count; i++) {
        data.push(createSelectorWithCounts(tagId, i));
    }

    return data;
};

export const createSelectorSeeds = (count: number = 10, selectorId: number = 0) => {
    const data: AssetGroupTagSelectorSeed[] = [];
    const seedType: SeedTypes = faker.datatype.number({ min: 1, max: 2 }) as SeedTypes;

    for (let i = 0; i < count; i++) {
        data.push({
            selector_id: selectorId,
            type: seedType,
            value: faker.datatype.uuid(),
        });
    }

    return data;
};

export const createSelectorNodes = (
    assetGroupId: number,
    selectorId: number | undefined,
    skip: number,
    limit: number,
    count: number
) => {
    const data: AssetGroupTagMemberListItem[] = [];

    for (let i = skip; i < skip + limit; i++) {
        if (i === count) break;

        const name = Number.isNaN(selectorId)
            ? `tier-${assetGroupId - 1}-object-${i}`
            : `tier-${assetGroupId - 1}-selector-${selectorId}-object-${i}`;

        data.push({
            id: i,
            primary_kind: 'User',
            object_id: faker.datatype.uuid(),
            name: name,
            source: NodeSourceChild,
        });
    }

    return data;
};

export const createAssetGroupMemberInfo = (tagId: string, memberId: string) => {
    const data: AssetGroupTagMemberInfo = {
        id: parseInt(memberId),
        name: 'member',
        primary_kind: 'User',
        object_id: faker.datatype.uuid(),
        selectors: createSelectors(10, parseInt(tagId)),
        properties: JSON.parse(faker.datatype.json()),
    };

    return data;
};

export const createAssetGroupMembersCount = () => {
    const data = {
        total_count: faker.datatype.number(),
        counts: {
            User: faker.datatype.number(),
            Computer: faker.datatype.number(),
            Container: faker.datatype.number(),
        },
    };

    return data;
};
