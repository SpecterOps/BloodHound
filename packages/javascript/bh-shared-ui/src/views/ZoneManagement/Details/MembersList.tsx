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

import { Button } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagMemberListItem } from 'js-client-library';
import { UseInfiniteQueryResult } from 'react-query';
import { NodeIcon, SortableHeader } from '../../../components';
import {
    InfiniteQueryFixedList,
    InfiniteQueryFixedListProps,
} from '../../../components/InfiniteQueryFixedList/InfiniteQueryFixedList';
import { SortOrder } from '../../../types';
import { cn } from '../../../utils';
import { SelectedHighlight, getListHeight } from './utils';

interface MembersListProps {
    listQuery: UseInfiniteQueryResult<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: { skip: number; limit: number };
    }>;
    selected: string | undefined;
    onClick: (id: string) => void;
    sortOrder: SortOrder;
    onChangeSortOrder: (sort: SortOrder) => void;
}

/**
 * @description This component is used to render the Objects/Members list for a given Tier, Label, or Selector. It is specifically built with both a fixed render window and a scroll loader as it is expected that the number of entities that this list may display would be large enough that trying to load all of these DOM nodes at once would cause the page to be sluggish and result in a poor user experience.
 * @param props
 * @param {selected} props.selected The currently selected Object/Member. This selection can be null.
 * @param {onClick} props.onClick The click handler for when a particular member is selected. This is primarily used for setting the selected entity in the parent component.
 * @returns The MembersList component for rendering in the Zone Management page.
 */
export const MembersList: React.FC<MembersListProps> = ({
    selected,
    onClick,
    listQuery,
    sortOrder,
    onChangeSortOrder,
}) => {
    const Row: InfiniteQueryFixedListProps<AssetGroupTagMemberListItem>['renderRow'] = (item, index, style) => {
        return (
            <div
                key={index}
                className={cn('border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative', {
                    'bg-neutral-light-4 dark:bg-neutral-dark-4': selected === item.id.toString(),
                })}
                style={style}>
                <SelectedHighlight selected={selected} itemId={item.id} title={'Members'} />
                <Button
                    variant={'text'}
                    className='flex justify-start w-full'
                    title={`Type: ${item.primary_kind}; Name: ${item.name}`}
                    onClick={() => {
                        onClick(item.id?.toString());
                    }}>
                    <NodeIcon nodeType={item.primary_kind} />
                    <span className='text-base dark:text-white ml-2 truncate'>{item.name}</span>
                </Button>
            </div>
        );
    };

    return (
        <div data-testid={`zone-management_details_members-list`}>
            <SortableHeader
                title={'Objects'}
                onSort={() => {
                    onChangeSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
                }}
                sortOrder={sortOrder}
                classes={{
                    container: 'border-b-2 border-neutral-light-5 dark:border-neutral-dark-5',
                    button: 'pl-6 font-bold text-xl',
                }}
            />
            <div
                className={cn(`overflow-y-auto border-x-2 border-neutral-light-5 dark:border-neutral-dark-5`, {
                    'h-[762px]': getListHeight(window.innerHeight) === 762,
                    'h-[642px]': getListHeight(window.innerHeight) === 642,
                    'h-[438px]': getListHeight(window.innerHeight) === 438,
                })}>
                <InfiniteQueryFixedList<AssetGroupTagMemberListItem>
                    itemSize={40}
                    queryResult={listQuery}
                    renderRow={Row}
                />
            </div>
        </div>
    );
};
