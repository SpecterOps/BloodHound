import { AssetGroupTagsListItem } from 'js-client-library';
import React, { FC } from 'react';
import { UseQueryResult } from 'react-query';
import DownArrow from '../../../components/AppIcon/Icons/DownArrow';
import { cn } from '../../../utils';
import { itemSkeletons } from '../utils';
import SummaryCard from './SummaryCard';

type SummaryListProps = {
    onSelect: (id: number) => void;
    listQuery: UseQueryResult<AssetGroupTagsListItem[]>;
    selected: string;
    title: 'Tiers' | 'Labels';
};

const SummaryList: FC<SummaryListProps> = ({ onSelect, listQuery, selected, title }) => {
    return (
        <div className='flex flex-col w-full h-full space-y-4'>
            <div className='flex flex-col flex-1 space-y-4'>
                <ul>
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
                            ?.sort((a, b) => {
                                return b.name.localeCompare(a.name);
                            })
                            .map((listItem) => {
                                return (
                                    <React.Fragment key={listItem.id}>
                                        <li
                                            key={listItem.id}
                                            onClick={() => {
                                                onSelect(listItem.id);
                                            }}
                                            className={cn('relative mb-4 cursor-pointer', {
                                                'border rounded-xl border-[#33318F] dark:border-[#ffffff]':
                                                    selected === listItem.id.toString(),
                                            })}>
                                            <SummaryCard
                                                title={listItem.name}
                                                type={listItem.type}
                                                id={listItem.id}
                                                selectorCount={listItem.counts?.selectors}
                                                memberCount={listItem.counts?.members}
                                            />
                                        </li>
                                        {listItem.type === 1 ? (
                                            <div key={listItem.id} className='flex justify-center mt-2 last:hidden'>
                                                <DownArrow className='w-8 h-6' />
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
