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
import { AssetGroupTag } from 'js-client-library';
import { FC, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { SortableHeader } from '../../../components';
import { useHighestPrivilegeTagId, usePZPathParams } from '../../../hooks';
import { SortOrder } from '../../../types';
import { cn } from '../../../utils';
import { ZoneAnalysisIcon } from '../ZoneAnalysisIcon';
import { itemSkeletons } from '../utils';
import { SelectedHighlight } from './SelectedHighlight';
import { isTag } from './utils';

type TagListProps = {
    title: 'Zones' | 'Labels';
    listQuery: UseQueryResult<AssetGroupTag[]>;
    selected?: string;
    onSelect: (id: number) => void;
};
/**
 * @description This component is meant to display the lists for either Zones or Labels but not the Members list since that is a paginated list that loads more data as a user scrolls.
 * @param {object} props
 * @param {title} props.title Limited to 'Zones' | 'Labels' as that is what this component is built for
 * @param {UseQueryResult} props.listQuery The endpoint call result wrapper from react query that allows us to hook into different states that the fetched data could be in
 * @param {selected} props.selected The id of the particular entity that is selected for the list. It is used for selected item rendering
 * @param {(id:number) => void} props.onSelect The click handler that should be called when an item from this list is selected. This is primarily being used to set the selected id state in the parent Details component
 * @returns The component that displays a list of entities for the zone management page
 */
export const TagList: FC<TagListProps> = ({ title, listQuery, selected, onSelect }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>('asc');
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { isLabelPage, isZonePage } = usePZPathParams();

    return (
        <div className='min-w-0 w-1/3' data-testid={`privilege-zones_details_${title.toLowerCase()}-list`}>
            {isLabelPage ? (
                <SortableHeader
                    title={title}
                    onSort={() => {
                        sortOrder === 'desc' ? setSortOrder('asc') : setSortOrder('desc');
                    }}
                    sortOrder={sortOrder}
                    classes={{
                        container: 'border-b-2 border-neutral-5',
                        button: 'pl-6 font-bold text-xl',
                    }}
                />
            ) : (
                <div
                    data-testid={`privilege-zones_details_${title.toLowerCase()}-list_static-order`}
                    className='p-0 relative w-full border-b-2 border-neutral-5'>
                    <div className='inline-flex items-center justify-center h-10 transition-colors pl-6 font-bold text-xl'>
                        {title}
                    </div>
                </div>
            )}
            <ul>
                {listQuery.isLoading ? (
                    itemSkeletons.map((skeleton, index) => {
                        return skeleton(title, index);
                    })
                ) : listQuery.isError ? (
                    <li className='border-y border-neutral-3 relative h-10 pl-2'>
                        <span className='text-base'>There was an error fetching this data</span>
                    </li>
                ) : listQuery.isSuccess ? (
                    listQuery.data
                        ?.sort((a, b) => {
                            if (isTag(a) && isTag(b) && isZonePage) {
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
                            return (
                                <li
                                    key={listItem.id}
                                    data-testid={`privilege-zones_details_${title.toLowerCase()}-list_item-${listItem.id}`}
                                    className={cn('border-y border-neutral-3 relative h-10', {
                                        'bg-neutral-4': selected === listItem.id.toString(),
                                    })}>
                                    <SelectedHighlight itemId={listItem.id} type='tag' />
                                    <Button
                                        variant='text'
                                        className='flex justify-between w-full'
                                        onClick={() => onSelect(listItem.id)}>
                                        <div className='flex items-center overflow-hidden'>
                                            {isZonePage && listItem.id !== topTagId && (
                                                <ZoneAnalysisIcon
                                                    size={18}
                                                    tooltip
                                                    analysisEnabled={listItem?.analysis_enabled}
                                                />
                                            )}
                                            <span className='text-base dark:text-white truncate' title={listItem.name}>
                                                {listItem.name}
                                            </span>
                                        </div>
                                        {listItem.counts && (
                                            <span className='text-base ml-4'>
                                                {listItem.counts.selectors.toLocaleString()}
                                            </span>
                                        )}
                                    </Button>
                                </li>
                            );
                        })
                ) : null}
            </ul>
        </div>
    );
};
