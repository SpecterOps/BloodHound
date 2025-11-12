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

import { Button, Skeleton } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagSelector } from 'js-client-library';
import { FC } from 'react';
import { UseInfiniteQueryResult } from 'react-query';
import { SortableHeader } from '../../../components';
import { InfiniteQueryFixedList, InfiniteQueryFixedListProps } from '../../../components/InfiniteQueryFixedList';
import { SortOrder } from '../../../types';
import { cn } from '../../../utils';
import { SelectedHighlight } from './SelectedHighlight';
import { getListHeight } from './utils';

const LoadingRow = (_: number, style: React.CSSProperties) => (
    <div
        data-testid={`privilege-zones_selectors-list_loading-skeleton`}
        style={style}
        className='border-y border-neutral-3 relative w-full p-2'>
        <Skeleton className={`h-full`} />
    </div>
);

type SelectorsListProps = {
    listQuery: UseInfiniteQueryResult<{
        items: AssetGroupTagSelector[];
        nextPageParam?: { skip: number; limit: number };
    }>;
    selected: string | undefined;
    onSelect: (id: number) => void;
    sortOrder: SortOrder;
    onChangeSortOrder: (sort: SortOrder) => void;
};

const SelectorsListWrapper = ({
    children,
    onChangeSortOrder,
    sortOrder,
}: {
    children: React.ReactNode;
    onChangeSortOrder: (sort: SortOrder) => void;
    sortOrder: SortOrder;
}) => {
    return (
        <div className='min-w-0 w-1/3' data-testid={`privilege-zones_details_selectors-list`}>
            <SortableHeader
                title={'Selectors'}
                onSort={() => {
                    onChangeSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
                }}
                sortOrder={sortOrder}
                classes={{
                    container: 'border-b-2 border-neutral-5',
                    button: 'pl-6 font-bold text-xl',
                }}
            />
            <div
                className={cn(`border-x-2 border-neutral-5`, {
                    'h-[760px]': getListHeight(window.innerHeight) === 760,
                    'h-[640px]': getListHeight(window.innerHeight) === 640,
                    'h-[436px]': getListHeight(window.innerHeight) === 436,
                })}>
                {children}
            </div>
        </div>
    );
};

/**
 * @description This component displays an infinitely scrolling list of Selectors
 * @param {object} props
 * @param {UseInfiniteQueryResult} props.listQuery The endpoint call result wrapper from react query that allows us to hook into different states that the fetched data could be in
 * @param {(string|undefined)} props.selected The id of the particular entity that is selected for the list. It is used for selected item rendering
 * @param {(id:number) => void} props.onSelect The click handler that should be called when an item from this list is selected. This is primarily being used to set the selected id state in the parent Details component
 * @returns The component that displays a list of selectors for the zone management page
 */
export const SelectorsList: FC<SelectorsListProps> = ({
    listQuery,
    onChangeSortOrder,
    onSelect,
    selected,
    sortOrder,
}) => {
    if (listQuery.isError) {
        return (
            <SelectorsListWrapper sortOrder={sortOrder} onChangeSortOrder={onChangeSortOrder}>
                <ul>
                    <li className='border-y border-neutral-3 relative h-10 pl-2'>
                        <span className='text-base'>There was an error fetching this data</span>
                    </li>
                </ul>
            </SelectorsListWrapper>
        );
    }

    const Row: InfiniteQueryFixedListProps<AssetGroupTagSelector>['renderRow'] = (item, index, style) => {
        const isDisabled = item.disabled_at;

        return (
            <div
                style={style}
                role='listitem'
                key={item.id}
                className={cn('border-y border-neutral-3 relative', {
                    'bg-neutral-4': selected === item.id.toString(),
                })}>
                <SelectedHighlight itemId={item.id} type='selector' />
                <Button
                    variant='text'
                    className='flex justify-between w-full overflow-hidden'
                    onClick={() => onSelect(item.id)}>
                    <span
                        className={cn('text-base dark:text-white truncate', {
                            'text-[#8E8C95] dark:text-[#919191]': isDisabled,
                        })}
                        title={isDisabled ? `Disabled: ${item.name}` : item.name}>
                        {item.name}
                    </span>
                    {item.counts && <span className='text-base pl-2'>{item.counts.members.toLocaleString()}</span>}
                </Button>
            </div>
        );
    };

    return (
        <SelectorsListWrapper sortOrder={sortOrder} onChangeSortOrder={onChangeSortOrder}>
            <InfiniteQueryFixedList<AssetGroupTagSelector>
                itemSize={40}
                queryResult={listQuery}
                renderRow={Row}
                renderLoadingRow={LoadingRow}
            />
        </SelectorsListWrapper>
    );
};
