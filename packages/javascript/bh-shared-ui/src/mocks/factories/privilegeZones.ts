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

import { faker } from '@faker-js/faker/locale/en';
import {
    AssetGroupTag,
    AssetGroupTagMemberInfo,
    AssetGroupTagMemberListItem,
    AssetGroupTagSelector,
    AssetGroupTagSelectorAutoCertifyType,
    AssetGroupTagSelectorSeed,
    AssetGroupTagType,
    AssetGroupTagTypeZone,
    CustomRulesKey,
    DefaultRulesKey,
    DisabledRulesKey,
    NodeSourceChild,
    SeedTypes,
} from 'js-client-library';

export const createAssetGroupTag = (tagId: number = 0, name?: string, type?: AssetGroupTagType): AssetGroupTag => {
    return {
        id: tagId,
        name: name ? name : `tag-${tagId - 1}`,
        kind_id: faker.datatype.number(),
        glyph: null,
        type: type ?? AssetGroupTagTypeZone,
        position: tagId,
        description: faker.random.words(10),
        created_at: faker.date.past().toISOString(),
        created_by: faker.internet.email(),
        updated_at: faker.date.past().toISOString(),
        updated_by: faker.internet.email(),
        deleted_at: faker.date.past().toISOString(),
        deleted_by: faker.internet.email(),
        require_certify: faker.datatype.boolean(),
        analysis_enabled: faker.datatype.boolean(),
    };
};

export const createAssetGroupTagWithCounts = (tagId: number = 0): AssetGroupTag => {
    return {
        ...createAssetGroupTag(tagId),
        counts: {
            selectors: faker.datatype.number(),
            members: faker.datatype.number(),
            [CustomRulesKey]: faker.datatype.number(),
            [DefaultRulesKey]: faker.datatype.number(),
            [DisabledRulesKey]: faker.datatype.number(),
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

export const createRule = (tagId: number = 0, ruleId: number = 0) => {
    const data: AssetGroupTagSelector = {
        id: ruleId,
        asset_group_tag_id: tagId,
        name: `tag-${tagId - 1}-rule-${ruleId}`,
        allow_disable: faker.datatype.boolean(),
        description: faker.random.words(),
        is_default: faker.datatype.boolean(),
        auto_certify: faker.datatype.number({ min: 0, max: 2 }) as AssetGroupTagSelectorAutoCertifyType,
        created_at: faker.date.past().toISOString(),
        created_by: faker.internet.email(),
        updated_at: faker.date.past().toISOString(),
        updated_by: faker.internet.email(),
        disabled_at: null,
        disabled_by: '',
        seeds: createRuleSeeds(10, ruleId),
    };

    return data;
};

export const createRuleWithCounts = (tagId: number = 0, ruleId: number = 0) => {
    const data: AssetGroupTagSelector = {
        ...createRule(tagId, ruleId),
        counts: { members: faker.datatype.number() },
    };

    return data;
};

export const createRuleWithCypher = (tagId: number = 0, ruleId: number = 0) => {
    const data: AssetGroupTagSelector = {
        ...createRule(tagId, ruleId),
        seeds: [{ selector_id: 2, type: 2, value: '9a092ad2-3114-40e7-9bb6-6e47944ad83c' }],
    };

    return data;
};

export const createRules = (count: number = 10, tagId: number = 0) => {
    const data: AssetGroupTagSelector[] = [];

    for (let i = 0; i < count; i++) {
        data.push(createRuleWithCounts(tagId, i));
    }

    return data;
};

export const createRuleSeeds = (count: number = 10, ruleId: number = 0) => {
    const data: AssetGroupTagSelectorSeed[] = [];
    const seedType: SeedTypes = faker.datatype.number({ min: 1, max: 2 }) as SeedTypes;

    for (let i = 0; i < count; i++) {
        data.push({
            selector_id: ruleId,
            type: seedType,
            value: faker.datatype.uuid(),
        });
    }

    return data;
};

export const createObjects = (
    assetGroupId: number,
    ruleId: number | undefined,
    skip: number,
    limit: number,
    count: number
) => {
    const data: AssetGroupTagMemberListItem[] = [];

    for (let i = skip; i < skip + limit; i++) {
        if (i === count) break;

        const name = Number.isNaN(ruleId)
            ? `tag-${assetGroupId - 1}-object-${i}`
            : `tag-${assetGroupId - 1}-rule-${ruleId}-object-${i}`;

        data.push({
            id: i,
            asset_group_tag_id: assetGroupId,
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
        asset_group_tag_id: parseInt(tagId, 10),
        name: 'member',
        primary_kind: 'User',
        object_id: faker.datatype.uuid(),
        selectors: createRules(10, parseInt(tagId)),
        properties: JSON.parse(faker.datatype.json()),
        source: 1,
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
