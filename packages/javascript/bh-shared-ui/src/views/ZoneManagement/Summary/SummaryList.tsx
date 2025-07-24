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
    AssetGroupTag,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypes,
    AssetGroupTagTypeTier,
} from 'js-client-library';
import React, { FC } from 'react';
import { UseQueryResult } from 'react-query';
import DownArrow from '../../../components/AppIcon/Icons/DownArrow';
import { cn } from '../../../utils';
import { itemSkeletons } from '../utils';
import SummaryCard from './SummaryCard';

type SummaryListProps = {
    onSelect: (id: number) => void;
    listQuery: UseQueryResult<AssetGroupTag[]>;
    selected: string;
    title: 'Tiers' | 'Labels';
};

const SummaryList: FC<SummaryListProps> = ({ onSelect, listQuery, selected, title }) => {
    const targetTypes: AssetGroupTagTypes[] =
        title === 'Tiers' ? [AssetGroupTagTypeTier] : [AssetGroupTagTypeLabel, AssetGroupTagTypeOwned];

    return (
        <div className='flex flex-col w-full h-full space-y-4'>
            <div className='flex flex-col flex-1 space-y-4'>
                <ul className='overflow-y-auto h-[445px]' data-testid='zone-management_summary_tag-list'>
                    {listQuery.isLoading ? (
                        itemSkeletons.map((skeleton, index) => {
                            return skeleton(title, index, 'h-32');
                        })
                    ) : listQuery.isError ? (
                        <li className='border-y border-neutral-light-3 dark:border-neutral-dark-3 relative h-10 pl-2'>
                            <span className='text-base'>There was an error fetching this data</span>
                        </li>
                    ) : (
                        listQuery.data
                            ?.filter((listItem) => targetTypes.includes(listItem.type))
                            .sort((a, b) => {
                                const aPos = a.position ?? 1;
                                const bPos = b.position ?? 1;
                                return aPos - bPos;
                            })
                            .map((listItem) => {
                                return (
                                    <React.Fragment key={listItem.id}>
                                        <li
                                            onClick={() => {
                                                onSelect(listItem.id);
                                            }}
                                            className={cn('relative mb-4 cursor-pointer', {
                                                'border rounded-xl border-primary dark:border-secondary-variant-2':
                                                    selected === listItem.id.toString(),
                                            })}>
                                            <SummaryCard
                                                title={listItem.name}
                                                type={listItem.type}
                                                id={listItem.id}
                                                selectorCount={listItem.counts?.selectors}
                                                memberCount={listItem.counts?.members}
                                                analysisEnabled={listItem.analysis_enabled}
                                            />
                                        </li>
                                        {listItem.type === AssetGroupTagTypeTier ? (
                                            <div
                                                key={listItem.id}
                                                className='flex justify-center my-2 pl-6 last:hidden'>
                                                <DownArrow />
                                            </div>
                                        ) : null}
                                    </React.Fragment>
                                );
                            })
                    )}
                </ul>
            </div>
        </div>
    );
};

export default SummaryList;
