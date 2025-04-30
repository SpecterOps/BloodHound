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

import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTag, AssetGroupTagSelector, SeedTypes } from 'js-client-library';
import { CSSProperties, FC } from 'react';

export const ItemSkeleton = (title: string, key: number, style?: CSSProperties) => {
    return (
        <li
            key={key}
            data-testid={`tier-management_details_${title.toLowerCase()}-list_loading-skeleton`}
            style={style}
            className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-full w-full'>
            <Skeleton className='h-10 rounded-none min-h-10' />
        </li>
    );
};

export const itemSkeletons = [ItemSkeleton, ItemSkeleton, ItemSkeleton];

const isActive = (selected: string | undefined, itemId: string | number) => {
    if (typeof itemId === 'number') return selected === itemId.toString();
    else return selected === itemId;
};

export const SelectedHighlight: FC<{ selected: string | undefined; itemId: string | number; title: string }> = ({
    selected,
    itemId,
    title,
}) => {
    return isActive(selected, itemId) ? (
        <div
            data-testid={`tier-management_details_${title.toLowerCase()}-list_active-${title.toLowerCase()}-item-${selected}`}
            className='h-full bg-primary pr-1 absolute'></div>
    ) : null;
};

export const isTag = (data: any): data is AssetGroupTag => {
    return 'kind_id' in data;
};

export const isSelector = (data: any): data is AssetGroupTagSelector => {
    return 'seeds' in data;
};

export const getSelectorSeedType = (selector: AssetGroupTagSelector): SeedTypes => {
    const firstSeed = selector.seeds[0];

    return firstSeed.type;
};
