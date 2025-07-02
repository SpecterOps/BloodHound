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

import {
    Button,
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTag,
    AssetGroupTagSelector,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypeTier,
    parseTieringConfiguration,
} from 'js-client-library';
import { FC, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { AppIcon, SortableHeader } from '../../../components';
import { useGetConfiguration } from '../../../hooks';
import { SortOrder } from '../../../types';
import { cn } from '../../../utils';
import { itemSkeletons } from '../utils';
import { SelectedHighlight, getListHeight, isSelector, isTag } from './utils';

const getCountElement = (listItem: AssetGroupTag | AssetGroupTagSelector): React.ReactNode => {
    if (listItem.counts === undefined) {
        return null;
    } else if (isTag(listItem)) {
        return <span className='text-base ml-4'>{listItem.counts.selectors.toLocaleString()}</span>;
    } else if (isSelector(listItem)) {
        return <span className='text-base ml-4'>{listItem.counts.members.toLocaleString()}</span>;
    } else {
        return null;
    }
};

type DetailsListProps = {
    title: 'Selectors' | 'Tiers' | 'Labels';
    listQuery: UseQueryResult<AssetGroupTag[]> | UseQueryResult<AssetGroupTagSelector[]>;
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
 * @returns The component that displays a list of entities for the zone management page
 */
export const DetailsList: FC<DetailsListProps> = ({ title, listQuery, selected, onSelect }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>('asc');
    const { data } = useGetConfiguration();
    const tieringConfig = parseTieringConfiguration(data);
    const multiTierAnalysisEnabled = tieringConfig?.value.multi_tier_analysis_enabled;

    return (
        <div data-testid={`zone-management_details_${title.toLowerCase()}-list`}>
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
                    data-testid={`zone-management_details_${title.toLowerCase()}-list_static-order`}
                    className='p-0 relative w-full border-b-2 border-neutral-light-5 dark:border-neutral-dark-5'>
                    <div className='inline-flex items-center justify-center h-10 transition-colors text-neutral-dark-5 dark:text-neutral-light-5 pl-6 font-bold text-xl'>
                        {title}
                    </div>
                </div>
            )}
            <div
                className={cn(`overflow-y-auto`, {
                    'border-x-2 border-neutral-light-5 dark:border-neutral-dark-5': title === 'Selectors',
                    'h-[762px]': getListHeight(window.innerHeight) === 762,
                    'h-[642px]': getListHeight(window.innerHeight) === 642,
                    'h-[438px]': getListHeight(window.innerHeight) === 438,
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
                        listQuery.data
                            ?.sort((a, b) => {
                                if (isTag(a) && isTag(b) && title === 'Tiers') {
                                    // A tag can be a tier and also have position null it seems
                                    return (a.position || 0) - (b.position || 0);
                                } else {
                                    switch (sortOrder) {
                                        case 'asc':
                                            return a.name.localeCompare(b.name);
                                        case 'desc':
                                            return b.name.localeCompare(a.name);
                                        default:
                                            return b.name.localeCompare(a.name);
                                    }
                                }
                            })
                            .map((listItem) => {
                                // Filters out Tier Tags when the active tab is 'Labels'
                                if (isTag(listItem) && listItem.type === AssetGroupTagTypeTier && title !== 'Tiers') {
                                    return null;
                                }

                                // Filters out Label and Owned Tags when the active tab is 'Tiers'
                                if (
                                    isTag(listItem) &&
                                    (listItem.type === AssetGroupTagTypeLabel ||
                                        listItem.type === AssetGroupTagTypeOwned) &&
                                    title === 'Tiers'
                                ) {
                                    return null;
                                }

                                const isDisabled = isSelector(listItem) && listItem.disabled_at;

                                return (
                                    <li
                                        data-testid={`zone-management_details_${title.toLowerCase()}-list_item-${listItem.id}`}
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
                                            className='flex justify-between w-full overflow-hidden'
                                            onClick={() => {
                                                onSelect(listItem.id);
                                            }}>
                                            <div className='flex items-center'>
                                                {multiTierAnalysisEnabled &&
                                                    isTag(listItem) &&
                                                    !listItem?.analysis_enabled && (
                                                        <TooltipProvider>
                                                            <TooltipRoot>
                                                                <TooltipTrigger>
                                                                    <div className='flex flex-row items-center mb-0.5'>
                                                                        <AppIcon.DataAlert
                                                                            size={16}
                                                                            data-testid='analysis_disabled_icon'
                                                                            className='mr-2 text-[#ED8537]'
                                                                        />
                                                                    </div>
                                                                </TooltipTrigger>
                                                                <TooltipPortal>
                                                                    <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                                                                        Analysis disabled
                                                                    </TooltipContent>
                                                                </TooltipPortal>
                                                            </TooltipRoot>
                                                        </TooltipProvider>
                                                    )}
                                                <div
                                                    className={cn(
                                                        'text-base dark:text-white truncate sm:max-w-[50px] lg:max-w-[100px] xl:max-w-[150px] 2xl:max-w-[300px]',
                                                        {
                                                            'text-[#8E8C95] dark:text-[#919191]': isDisabled,
                                                        }
                                                    )}
                                                    title={isDisabled ? `Disabled: ${listItem.name}` : listItem.name}>
                                                    {listItem.name}
                                                </div>
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
