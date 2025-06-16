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

import { AssetGroupTag, AssetGroupTagSelector, SeedTypes } from 'js-client-library';
import { FC } from 'react';

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

export const getListHeight = (windoHeight: number) => {
    if (windoHeight > 1080) return 762;
    if (1080 >= windoHeight && windoHeight > 900) return 642;
    if (900 >= windoHeight) return 438;
    return 438;
};
