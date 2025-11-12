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
import { FC } from 'react';
import { usePZPathParams } from '../../../hooks';

export const SelectedHighlight: FC<{
    itemId: string | number;
    type: 'tag' | 'selector' | 'member';
}> = ({ itemId, type }) => {
    const { tagId, selectorId, memberId } = usePZPathParams();

    const itemIdStr = itemId.toString();
    const activeType = memberId ? 'member' : selectorId ? 'selector' : 'tag';

    if (activeType !== type) {
        return null;
    }

    const isActive =
        (type === 'tag' && tagId === itemIdStr) ||
        (type === 'selector' && selectorId === itemIdStr) ||
        (type === 'member' && memberId === itemIdStr);

    if (!isActive) return null;

    return (
        <div
            className='h-full bg-primary pr-1 absolute'
            data-testid={`privilege-zones_details_${type}s-list_active-${type}s-item-${itemId}`}
        />
    );
};
