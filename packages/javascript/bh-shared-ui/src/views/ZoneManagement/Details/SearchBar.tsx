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
import { Button, Input, Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { useCombobox } from 'downshift';
import {
    AssetGroupTag,
    AssetGroupTagMember,
    AssetGroupTagSelector,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeTier,
} from 'js-client-library';
import React, { useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../components';
import { useDebouncedValue, useZonePathParams } from '../../../hooks';
import { apiClient, cn, useAppNavigate } from '../../../utils';
import { isSelector, isTag } from './utils';

type SectorMap =
    | { Tiers: 'tags'; Selectors: 'selectors'; Members: 'members' }
    | { Labels: 'tags'; Selectors: 'selectors'; Members: 'members' };

type SearchItem = AssetGroupTag | AssetGroupTagSelector | AssetGroupTagMember;

const SearchBar: React.FC = () => {
    const [query, setQuery] = useState('');
    const [isOpen, setIsOpen] = useState(false);
    const debouncedInputValue = useDebouncedValue(query, 300);
    const navigate = useAppNavigate();
    const { tagId, tagKind } = useZonePathParams();

    const searchQuery = useQuery({
        queryKey: ['zone-management', 'search', debouncedInputValue, tagId, tagKind],
        queryFn: async () => {
            const body = {
                query: debouncedInputValue,
                tag_type: tagKind === 'label' ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier,
            };
            const res = await apiClient.searchAssetGroupTags(body);
            return res.data.data;
        },
        enabled: debouncedInputValue.length >= 3 && tagId !== undefined,
    });

    const results = searchQuery.data ?? { tags: [], selectors: [], members: [] };

    const handleClick = (item: SearchItem) => {
        setQuery('');
        setIsOpen(false);

        if (isTag(item)) {
            navigate(`/zone-management/details/${tagKind}/${item.id}`);
        } else if (isSelector(item)) {
            navigate(`/zone-management/details/${tagKind}/${item.asset_group_tag_id}/selector/${item.id}`);
        } else {
            navigate(`/zone-management/details/${tagKind}/${item.asset_group_tag_id}/member/${item.id}`);
        }
    };

    // Flatten results with sector since useCombobox requires one flattened array of items
    const items: SearchItem[] = [...results.tags, ...results.selectors, ...results.members];

    const { getMenuProps, getInputProps, getComboboxProps, getItemProps, highlightedIndex } = useCombobox<SearchItem>({
        items,
        inputValue: query,
        isOpen,
        scrollIntoView: (node: HTMLElement) => node.scrollIntoView({ behavior: 'smooth', block: 'nearest' }),
        itemToString: (item) => item?.name ?? '',
        onInputValueChange: ({ inputValue = '' }) => {
            setQuery(inputValue);
            if (inputValue.length >= 3 && !isOpen) setIsOpen(true);
            if (inputValue.length < 3 && isOpen) setIsOpen(false);
        },
        onSelectedItemChange: ({ selectedItem }) => {
            if (selectedItem) handleClick(selectedItem);
        },
    });

    const sectorMap: SectorMap =
        tagKind === 'label'
            ? { Labels: 'tags', Selectors: 'selectors', Members: 'members' }
            : { Tiers: 'tags', Selectors: 'selectors', Members: 'members' };

    return (
        <div {...getComboboxProps()} className='relative w-4/6'>
            <Popover open={isOpen} onOpenChange={(open) => !open && setIsOpen(false)}>
                <PopoverTrigger asChild>
                    <div className='flex items-center'>
                        <AppIcon.MagnifyingGlass className='-mr-4' />
                        <Input variant={'underlined'} placeholder='Search' className='pl-8' {...getInputProps()} />
                    </div>
                </PopoverTrigger>
                <PopoverContent
                    className='w-[448px] max-h-[400px] overflow-y-auto'
                    onOpenAutoFocus={(e) => e.preventDefault()}>
                    {/* supressing ref error as Popover isn't mounted on initial render */}
                    {/* TODO: add comboBox component to Doodle UI and replace this usage */}
                    <ul {...getMenuProps({}, { suppressRefError: true })} className='space-y-4'>
                        {isOpen &&
                            Object.entries(sectorMap).map(([sector, key]) => (
                                <li key={sector}>
                                    <p className='font-bold mb-1'>{sector}</p>
                                    {results[key].length > 0 ? (
                                        <ul>
                                            {results[key].map((item) => {
                                                //global index for all items so we have unique indices with no overlap
                                                const globalIndex = items.indexOf(item);
                                                return (
                                                    <li
                                                        key={item.id}
                                                        {...getItemProps({
                                                            item,
                                                            index: globalIndex,
                                                        })}
                                                        className={cn('flex max-w-lg min-w-0', {
                                                            'bg-secondary text-white dark:bg-secondary-variant-2 dark:text-black':
                                                                highlightedIndex === globalIndex,
                                                        })}>
                                                        <Button
                                                            className='overflow-hidden justify-start w-full no-underline'
                                                            variant='text'>
                                                            <span
                                                                className={cn('truncate', {
                                                                    'text-white  dark:text-black':
                                                                        highlightedIndex === globalIndex,
                                                                })}>
                                                                {item.name}
                                                            </span>
                                                        </Button>
                                                    </li>
                                                );
                                            })}
                                        </ul>
                                    ) : (
                                        <p className='pl-6 text-sm text-neutral-500'>No results</p>
                                    )}
                                </li>
                            ))}
                    </ul>
                </PopoverContent>
            </Popover>
        </div>
    );
};

export default SearchBar;
