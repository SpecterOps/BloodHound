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
import { AssetGroupTag, AssetGroupTagSelector } from 'js-client-library';
import { FC } from 'react';
import { UseInfiniteQueryResult } from 'react-query';
import {
    InfiniteQueryFixedList,
    InfiniteQueryFixedListProps,
} from '../../../components/InfiniteQueryFixedList/InfiniteQueryFixedList';
import { cn } from '../../../utils';
import { itemSkeletons } from '../utils';
import { SelectedHighlight, getListHeight, isSelector, isTag } from './utils';

const getCountElement = (listItem: AssetGroupTag | AssetGroupTagSelector): React.ReactNode => {
    if (listItem.counts === undefined) {
        return null;
    } else if (isTag(listItem)) {
        return <span className='text-base'>{listItem.counts.selectors.toLocaleString()}</span>;
    } else if (isSelector(listItem)) {
        return <span className='text-base'>{listItem.counts.members.toLocaleString()}</span>;
    } else {
        return null;
    }
};

type SelectorsListProps = {
    listQuery: UseInfiniteQueryResult<{
        items: AssetGroupTagSelector[];
        nextPageParam?: { skip: number; limit: number };
    }>;
    selected: string | undefined;
    onSelect: (id: number) => void;
};

const SelectorsListWrapper = ({ children }: { children: React.ReactNode }) => {
    return (
        <div data-testid={`zone-management_details_selectors-list`}>
            <div
                data-testid={`zone-management_details_selectors-list_static-order`}
                className='p-0 relative w-full border-b-2 border-neutral-light-5 dark:border-neutral-dark-5'>
                <div className='inline-flex items-center justify-center h-10 transition-colors text-neutral-dark-5 dark:text-neutral-light-5 pl-6 font-bold text-xl'>
                    Selectors
                </div>
            </div>

            <div
                className={cn(`overflow-y-auto border-x-2 border-neutral-light-5 dark:border-neutral-dark-5`, {
                    'h-[762px]': getListHeight(window.innerHeight) === 762,
                    'h-[642px]': getListHeight(window.innerHeight) === 642,
                    'h-[438px]': getListHeight(window.innerHeight) === 438,
                })}>
                {children}
            </div>
        </div>
    );
};

/**
 * @description This component is meant to display the lists for either Tiers, Labels, or Selectors but not the Members list since that is a paginated list that loads more data as a user scrolls.
 * @param {object} props
 * @param {UseQueryResult} props.listQuery The endpoint call result wrapper from react query that allows us to hook into different states that the fetched data could be in
 * @param {selected} props.selected The id of the particular entity that is selected for the list. It is used for selected item rendering
 * @param {(id:number) => void} props.onSelect The click handler that should be called when an item from this list is selected. This is primarily being used to set the selected id state in the parent Details component
 * @returns The component that displays a list of entities for the zone management page
 */
export const SelectorsList: FC<SelectorsListProps> = ({ listQuery, selected, onSelect }) => {
    if (listQuery.isLoading) {
        return (
            <SelectorsListWrapper>
                <ul>
                    {itemSkeletons.map((skeleton, index) => {
                        return skeleton('Selectors', index);
                    })}
                </ul>
            </SelectorsListWrapper>
        );
    }

    if (listQuery.isError) {
        return (
            <SelectorsListWrapper>
                <ul>
                    <li className='border-y border-neutral-light-3 dark:border-neutral-dark-3 relative h-10 pl-2'>
                        <span className='text-base'>There was an error fetching this data</span>
                    </li>
                </ul>
            </SelectorsListWrapper>
        );
    }

    const Row: InfiniteQueryFixedListProps<AssetGroupTagSelector>['renderRow'] = (item, index, style) => {
        const isDisabled = isSelector(item) && item.disabled_at;

        return (
            <div
                style={style}
                key={item.id}
                className={cn('border-y border-neutral-light-3 dark:border-neutral-dark-3 relative h-10', {
                    'bg-neutral-light-4 dark:bg-neutral-dark-4': selected === item.id.toString(),
                })}>
                <SelectedHighlight selected={selected} itemId={item.id} title={'Selectors'} />
                <Button
                    variant={'text'}
                    className='flex justify-between w-full overflow-hidden'
                    onClick={() => {
                        onSelect(item.id);
                    }}>
                    <div className='flex items-center'>
                        <div
                            className={cn(
                                'text-base dark:text-white truncate sm:max-w-[50px] lg:max-w-[100px] xl:max-w-[150px] 2xl:max-w-[300px]',
                                {
                                    'text-[#8E8C95] dark:text-[#919191]': isDisabled,
                                }
                            )}
                            title={isDisabled ? `Disabled: ${item.name}` : item.name}>
                            {item.name}
                        </div>
                    </div>
                    {getCountElement(item)}
                </Button>
            </div>
        );
    };

    return (
        <SelectorsListWrapper>
            <InfiniteQueryFixedList<AssetGroupTagSelector> itemSize={40} queryResult={listQuery} renderRow={Row} />
        </SelectorsListWrapper>
    );
};
