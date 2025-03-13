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
import { FC, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { NodeIcon, SortableHeader } from '../../components';
import { SortOrder } from '../../types';
import { EntityKinds, cn } from '../../utils';

type DetailsListItem = {
    name: string;
    id: number;
    count?: number;
    kind?: EntityKinds;
};

type DetailsListProps = {
    title: string;
    listQuery: UseQueryResult<DetailsListItem[], unknown>;
    selected: number | null;
    onSelect: (id: number) => void;
    sortable?: boolean | undefined;
    nodeIcon?: boolean | undefined;
};

const ItemSkeleton = (title: string, key: number) => {
    return (
        <li
            key={key}
            data-testid={`tier-management_details_${title.toLowerCase()}-list_loading-skeleton`}
            className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-full w-full'>
            <Skeleton className='h-10 rounded-none' />
        </li>
    );
};

const itemSkeletons = [ItemSkeleton, ItemSkeleton, ItemSkeleton];

const isActive = (selected: number | null, itemId: number) => {
    return selected === itemId;
};

export const DetailsList: FC<DetailsListProps> = ({ title, listQuery, selected, onSelect, nodeIcon, sortable }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>();

    return (
        <div data-testid={`tier-management_details_${title.toLowerCase()}-list`}>
            {sortable ? (
                <SortableHeader
                    title={title}
                    onSort={() => {
                        if (sortOrder === undefined) {
                            // first click
                            setSortOrder('desc');
                        } else if (sortOrder === 'desc') {
                            // second click
                            setSortOrder('asc');
                        } else if (sortOrder === 'asc') {
                            // third click
                            setSortOrder(undefined);
                        }
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
            <ul
                className={cn({
                    'border-x-[1px] border-neutral-light-5 dark:border-neutral-dark-5': title === 'Selectors',
                })}>
                {listQuery.isLoading ? (
                    itemSkeletons.map((skeleton, index) => {
                        return skeleton(title, index);
                    })
                ) : listQuery.isError ? (
                    <li>There was an error fetching this data</li>
                ) : listQuery.isSuccess ? (
                    listQuery.data
                        .sort((a, b) => {
                            switch (sortOrder) {
                                case 'asc':
                                    return a.name.localeCompare(b.name);
                                case 'desc':
                                    return b.name.localeCompare(a.name);
                                default:
                                    return b.name.localeCompare(a.name);
                            }
                        })
                        .map((listItem, index) => {
                            return (
                                <li
                                    key={index}
                                    className={cn(
                                        'border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-full',
                                        {
                                            'bg-neutral-light-4 dark:bg-neutral-dark-4': selected === listItem.id,
                                        }
                                    )}>
                                    {isActive(selected, listItem.id) && (
                                        <div
                                            data-testid={`tier-management_details_${title.toLowerCase()}-list_active-${title.toLowerCase()}-item-${selected}`}
                                            className='h-full bg-primary pr-1 absolute'></div>
                                    )}
                                    {nodeIcon ? (
                                        <Button
                                            variant={'text'}
                                            className='flex justify-start w-full'
                                            onClick={() => {
                                                onSelect(listItem.id);
                                            }}>
                                            <NodeIcon nodeType={listItem.kind || 'Unknown'} />
                                            <span className='text-base ml-2'>{listItem.name}</span>
                                        </Button>
                                    ) : (
                                        <Button
                                            variant={'text'}
                                            className='flex justify-between w-full'
                                            onClick={() => {
                                                onSelect(listItem.id);
                                            }}>
                                            <span className='text-base'>{listItem.name}</span>
                                            {Object.hasOwn(listItem, 'count') && (
                                                <span className='text-base'>{listItem.count!.toLocaleString()}</span>
                                            )}
                                        </Button>
                                    )}
                                </li>
                            );
                        })
                ) : null}
            </ul>
        </div>
    );
};
