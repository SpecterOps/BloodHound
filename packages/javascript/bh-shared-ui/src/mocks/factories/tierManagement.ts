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
    AssetLabel,
    AssetSelector,
    CertifiedValues,
    SeedTypeValues,
    SelectorNode,
    SelectorSeed,
} from 'js-client-library';

export const createAssetGroupLabels = (count: number = 10) => {
    const data: AssetLabel[] = [];

    for (let i = 1; i < count; i++) {
        data.push({
            id: i,
            name: `Tier-${i - 1}`,
            kind_id: faker.datatype.number(),
            asset_group_tier_id: i,
            description: faker.random.words(),
            created_at: faker.date.past().toISOString(),
            created_by: faker.internet.email(),
            updated_at: faker.date.past().toISOString(),
            updated_by: faker.internet.email(),
            deleted_at: faker.date.past().toISOString(),
            deleted_by: faker.internet.email(),
            count: faker.datatype.number(),
        });
    }

    return data;
};

export const createSelectors = (count: number = 10, tierId: number = 0) => {
    const data: AssetSelector[] = [];

    for (let i = 0; i < count; i++) {
        data.push({
            id: i,
            asset_group_label_id: i,
            name: `tier-${tierId - 1}-selector-${i}`,
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
    const data: SelectorSeed[] = [];
    const seedType = faker.datatype.number({ min: 0, max: 1 }) as SeedTypeValues;

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
    const data: SelectorNode[] = [];

    for (let i = skip; i < skip + limit; i++) {
        if (i === count) break;

        const name = Number.isNaN(selectorId)
            ? `tier-${assetGroupId - 1}-object-${i}`
            : `tier-${assetGroupId - 1}-selector-${selectorId}-object-${i}`;

        data.push({
            selector_id: selectorId || 0,
            node_id: i,
            id: i,
            certified: faker.datatype.number({ min: -1, max: 2 }) as CertifiedValues,
            certified_by: faker.internet.email(),
            name: name,
        });
    }

    return data;
};
