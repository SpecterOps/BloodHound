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
    AssetGroupTagCertifiedValues,
    AssetGroupTagSelector,
    AssetGroupTagSelectorNode,
    AssetGroupTagSelectorSeed,
    AssetGroupTagTypeValues,
    SeedTypeValues,
} from 'js-client-library';

export const createAssetGroupLabels = (count: number = 10) => {
    const data: AssetGroupTag[] = [];

    for (let i = 1; i < count; i++) {
        data.push({
            id: i,
            name: `Tier-${i - 1}`,
            kind_id: faker.datatype.number(),
            type: faker.datatype.number({ min: 1, max: 2 }) as AssetGroupTagTypeValues,
            position: faker.datatype.number({ min: 0, max: 10 }),
            description: faker.random.words(),
            created_at: faker.date.past().toISOString(),
            created_by: faker.internet.email(),
            updated_at: faker.date.past().toISOString(),
            updated_by: faker.internet.email(),
            deleted_at: faker.date.past().toISOString(),
            deleted_by: faker.internet.email(),
            requireCertify: faker.datatype.boolean(),
            count: faker.datatype.number(),
        });
    }

    return data;
};

export const createSelectors = (count: number = 10, tierId: number = 0) => {
    const data: AssetGroupTagSelector[] = [];

    for (let i = 0; i < count; i++) {
        data.push({
            id: i,
            asset_group_tag_id: i,
            name: `tier-${tierId - 1}-selector-${i}`,
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
            count: faker.datatype.number(),
            seeds: createSelectorSeeds(10, i),
        });
    }

    return data;
};

export const createSelectorSeeds = (count: number = 10, selectorId: number = 0) => {
    const data: AssetGroupTagSelectorSeed[] = [];
    const seedType = faker.datatype.number({ min: 1, max: 2 }) as SeedTypeValues;

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
    const data: AssetGroupTagSelectorNode[] = [];

    for (let i = skip; i < skip + limit; i++) {
        if (i === count) break;

        const name = Number.isNaN(selectorId)
            ? `tier-${assetGroupId - 1}-object-${i}`
            : `tier-${assetGroupId - 1}-selector-${selectorId}-object-${i}`;

        data.push({
            selector_id: selectorId || 0,
            node_id: i.toString(),
            id: i,
            certified: faker.datatype.number({ min: -1, max: 2 }) as AssetGroupTagCertifiedValues,
            certified_by: faker.internet.email(),
            name: name,
        });
    }

    return data;
};

export const createAssetGroupMembersCount = (selectorId: number = 0) => {
    const data = {
        id: selectorId,
        total_count: faker.datatype.number(),
        counts: {
            User: faker.datatype.number(),
            Computer: faker.datatype.number(),
            Container: faker.datatype.number(),
        },
    };

    return data;
};

export const createAssetGroupMemberInfo = (assetGroupId: number, memberId: number | undefined) => {
    const data = [];

    for (let i = 0; i < 10; i++) {
        const name = `selector-${memberId && memberId + i}`;

        data.push({
            id: assetGroupId,
            member_id: memberId || 0,
            name: name,
        });
    }

    return data;
};
