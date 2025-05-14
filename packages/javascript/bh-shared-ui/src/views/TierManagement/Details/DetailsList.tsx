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

import { Badge, Button } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagSelectorsListItem,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypeTier,
    AssetGroupTagsListItem,
} from 'js-client-library';
import { FC, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { SortableHeader } from '../../../components';
import { SortOrder } from '../../../types';
import { cn } from '../../../utils';
import { SelectedHighlight, itemSkeletons } from './utils';

const isTagListItem = (
    listItem: AssetGroupTagsListItem | AssetGroupTagSelectorsListItem
): listItem is AssetGroupTagsListItem => {
    if (listItem.counts === undefined) return false;
    return 'selectors' in listItem.counts;
};

const isSelectorsListItem = (
    listItem: AssetGroupTagsListItem | AssetGroupTagSelectorsListItem
): listItem is AssetGroupTagSelectorsListItem => {
    if (listItem.counts === undefined) return false;
    return !('selectors' in listItem.counts);
};

const getCountElement = (listItem: AssetGroupTagsListItem | AssetGroupTagSelectorsListItem): React.ReactNode => {
    if (listItem.counts === undefined) {
        return null;
    } else if (isTagListItem(listItem)) {
        return <span className='text-base'>{listItem.counts.selectors.toLocaleString()}</span>;
    } else if (isSelectorsListItem(listItem)) {
        return <span className='text-base'>{listItem.counts.members.toLocaleString()}</span>;
    } else {
        return null;
    }
};

type DetailsListProps = {
    title: 'Selectors' | 'Tiers' | 'Labels';
    listQuery: UseQueryResult<AssetGroupTagsListItem[]> | UseQueryResult<AssetGroupTagSelectorsListItem[]>;
    selected: string | undefined;
    onSelect: (id: number) => void;
};
/**
 * @description This component is meant to display the lists for either Tiers, Labels, or Selectors but not the Members list since that is a paginated list that loads more data as a user scrolls.
 * @param {object} props
 * @param {title} props.title Limited to 'Selectors' | 'Tiers' | 'Labels' as that is what this component is built for
 * @param {UseQueryResult} props.listQuery The endpoint call result wrapper from react query that allows us to hook into different states that the fetched data could be in
 * @param {selected} props.selected The id of the particular entity that is selected for the list. It is used for selected item rendering
 * @param {(id:number) => void} props.onSelect The click handler that should be called when an item from this list is selected. This is primarily being used to set the selected id state in the parent Details component
 * @returns The component that displays a list of entities for the tier management page
 */
export const DetailsList: FC<DetailsListProps> = ({ title, listQuery, selected, onSelect }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>('asc');

    return (
        <div data-testid={`tier-management_details_${title.toLowerCase()}-list`} className='h-full max-h-full'>
            {title !== 'Tiers' ? (
                <SortableHeader
                    title={title}
                    onSort={() => {
                        sortOrder === 'desc' ? setSortOrder('asc') : setSortOrder('desc');
                    }}
                    sortOrder={sortOrder}
                    classes={{
                        container: 'border-b-2 border-neutral-light-5 dark:border-neutral-dark-5',
                        button: 'pl-6 font-bold text-xl',
                    }}
                />
            ) : (
                <div
                    data-testid={`tier-management_details_${title.toLowerCase()}-list_static-order`}
                    className='p-0 relative w-full border-b-2 border-neutral-light-5 dark:border-neutral-dark-5'>
                    <div className='inline-flex items-center justify-center h-10 transition-colors text-neutral-dark-5 dark:text-neutral-light-5 pl-6 font-bold text-xl'>
                        {title}
                    </div>
                </div>
            )}
            <div
                className={cn('h-[calc(100%-42px)] overflow-y-auto', {
                    'border-x-2 border-neutral-light-5 dark:border-neutral-dark-5': title === 'Selectors',
                })}>
                <ul>
                    {listQuery.isLoading ? (
                        itemSkeletons.map((skeleton, index) => {
                            return skeleton(title, index);
                        })
                    ) : listQuery.isError ? (
                        <li className='border-y border-neutral-light-3 dark:border-neutral-dark-3 relative h-10 pl-2'>
                            <span className='text-base'>There was an error fetching this data</span>
                        </li>
                    ) : listQuery.isSuccess ? (
                        // getFilteredList(listQuery.data)
                        listQuery.data
                            ?.sort((a, b) => {
                                switch (sortOrder) {
                                    case 'asc':
                                        return a.name.localeCompare(b.name);
                                    case 'desc':
                                        return b.name.localeCompare(a.name);
                                    default:
                                        return b.name.localeCompare(a.name);
                                }
                            })
                            .map((listItem) => {
                                if (
                                    isTagListItem(listItem) &&
                                    listItem.type === AssetGroupTagTypeTier &&
                                    title !== 'Tiers'
                                ) {
                                    return null;
                                }
                                if (
                                    isTagListItem(listItem) &&
                                    (listItem.type === AssetGroupTagTypeLabel ||
                                        listItem.type === AssetGroupTagTypeOwned) &&
                                    title === 'Tiers'
                                ) {
                                    return null;
                                }
                                return (
                                    <li
                                        key={listItem.id}
                                        className={cn(
                                            'border-y border-neutral-light-3 dark:border-neutral-dark-3 relative h-10',
                                            {
                                                'bg-neutral-light-4 dark:bg-neutral-dark-4':
                                                    selected === listItem.id.toString(),
                                            }
                                        )}>
                                        <SelectedHighlight selected={selected} itemId={listItem.id} title={title} />
                                        <Button
                                            variant={'text'}
                                            className='flex justify-between w-full'
                                            onClick={() => {
                                                onSelect(listItem.id);
                                            }}>
                                            <div className='flex items-center'>
                                                <div className='text-base'>{listItem.name}</div>
                                                {isSelectorsListItem(listItem) && listItem.disabled_at && (
                                                    <div className='ml-2 italic'>
                                                        <Badge
                                                            label='Disabled'
                                                            className={'bg-neutral-light-5 dark:bg-neutral-dark-5'}
                                                        />
                                                    </div>
                                                )}
                                            </div>
                                            {getCountElement(listItem)}
                                        </Button>
                                    </li>
                                );
                            })
                    ) : null}
                </ul>
            </div>
        </div>
    );
};
