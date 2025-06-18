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
import { CSSProperties } from 'react';

export const TIER_ZERO_ID = '1';
export const OWNED_ID = '2';

export const getTagUrlValue = (labelId: string | undefined) => {
    return labelId === undefined ? 'tier' : 'label';
};

export const ItemSkeleton = (title: string, key: number, height?: string, style?: CSSProperties) => {
    return (
        <li
            key={key}
            data-testid={`zone-management_${title.toLowerCase()}-list_loading-skeleton`}
            style={style}
            className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative w-full'>
            <Skeleton className={`${height ?? 'min-h-10'} rounded-none`} />
        </li>
    );
};

export const itemSkeletons = [ItemSkeleton, ItemSkeleton, ItemSkeleton];
