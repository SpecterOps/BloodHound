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
    AssetGroupTagSearchResponse,
    AssetGroupTagSelector,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeTier,
} from 'js-client-library';
import React, { useState } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { AppIcon } from '../../../components';
import { useDebouncedValue } from '../../../hooks';
import { apiClient, useAppNavigate } from '../../../utils';
import { getTagUrlValue } from '../utils';

type SearchBarProps = {
    selected: string | undefined;
};

type Sector = 'Tiers' | 'Selectors' | 'Objects';

type SearchItem = AssetGroupTag | AssetGroupTagSelector | AssetGroupTagMember;

type SearchResults = AssetGroupTagSearchResponse['data'];

type FlattenedItem = {
    sector: Sector;
    item: SearchItem;
};

const SearchBar: React.FC<SearchBarProps> = ({ selected }) => {
    const [query, setQuery] = useState('');
    const debouncedInputValue = useDebouncedValue(query, 300);
    const navigate = useAppNavigate();
    const { labelId } = useParams();

    const scope = labelId ? 'label' : 'tier';
    const searchQuery = useQuery({
        queryKey: ['zone-management', 'search', debouncedInputValue, selected, scope],
        queryFn: async () => {
            const body = {
                query: debouncedInputValue,
                tag_type: scope === 'label' ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier,
            };
            const res = await apiClient.searchAssetGroupTags(body);
            return res.data.data;
        },
        enabled: debouncedInputValue.length >= 3 && selected !== undefined,
        select: (data) => data ?? { tags: [], selectors: [], members: [] },
    });

    const results: SearchResults = searchQuery.data ?? { tags: [], selectors: [], members: [] };

    const handleClick = (sector: Sector, item: SearchItem) => {
        setQuery('');
        const base = getTagUrlValue(labelId);
        if (sector === 'Tiers') {
            navigate(`/zone-management/details/${base}/${(item as AssetGroupTag).id}`);
        } else if (sector === 'Selectors') {
            navigate(
                `/zone-management/details/${base}/${(item as AssetGroupTagSelector).asset_group_tag_id}/selector/${item.id}`
            );
        } else if (sector === 'Objects') {
            navigate(
                `/zone-management/details/${base}/${(item as AssetGroupTagMember).asset_group_tag_id}/member/${item.id}`
            );
        }
    };

    // Flatten results with sector since useCombobox requires one flattened array of items
    const items: FlattenedItem[] = [
        ...results.tags.map((tag) => ({ sector: 'Tiers' as const, item: tag })),
        ...results.selectors.map((selector) => ({ sector: 'Selectors' as const, item: selector })),
        ...results.members.map((member) => ({ sector: 'Objects' as const, item: member })),
    ];

    const { getMenuProps, getInputProps, getComboboxProps, getItemProps, isOpen, openMenu } =
        useCombobox<FlattenedItem>({
            items,
            inputValue: query,
            itemToString: (flattened) => (flattened ? flattened.item.name : ''),
            onInputValueChange: ({ inputValue }) => {
                setQuery(inputValue ?? '');
            },
            onSelectedItemChange: ({ selectedItem }) => {
                if (selectedItem) handleClick(selectedItem.sector, selectedItem.item);
            },
        });

    // group items back into separate sectors for rendering
    const groupedSectors: Record<Sector, FlattenedItem[]> = {
        Tiers: items.filter((i) => i.sector === 'Tiers'),
        Selectors: items.filter((i) => i.sector === 'Selectors'),
        Objects: items.filter((i) => i.sector === 'Objects'),
    };

    return (
        <div {...getComboboxProps()} className='relative w-4/6'>
            <Popover open={isOpen && query.length >= 3}>
                <PopoverTrigger asChild>
                    <div className='flex items-center border-b-2 border-neutral-dark-1 dark:border-neutral-light-1'>
                        <AppIcon.MagnifyingGlass />
                        <Input
                            placeholder='Search'
                            className='border-none bg-transparent dark:bg-transparent placeholder-neutral-dark-1 dark:placeholder-neutral-light-1 focus-visible:ring-0 focus-visible:ring-offset-0 pl-3'
                            {...getInputProps({ onFocus: openMenu })}
                        />
                    </div>
                </PopoverTrigger>
                <PopoverContent
                    className='w-[448px] max-h-[400px] overflow-y-auto'
                    onOpenAutoFocus={(e) => e.preventDefault()}>
                    <ul {...getMenuProps()} className='space-y-4'>
                        {isOpen &&
                            (['Tiers', 'Selectors', 'Objects'] as Sector[]).map((sector) => (
                                <li key={sector}>
                                    <p className='font-bold mb-1'>{sector}</p>
                                    {groupedSectors[sector].length > 0 ? (
                                        <ul>
                                            {groupedSectors[sector].map((wrappedItem) => {
                                                //global index for all items so we have unique indices with no overlap
                                                const globalIndex = items.indexOf(wrappedItem);
                                                return (
                                                    <li
                                                        key={wrappedItem.item.id}
                                                        {...getItemProps({
                                                            item: wrappedItem,
                                                            index: globalIndex,
                                                        })}
                                                        className='flex max-w-lg min-w-0'>
                                                        <Button
                                                            className='overflow-hidden justify-start w-full'
                                                            variant='text'>
                                                            <span className='truncate'>{wrappedItem.item.name}</span>
                                                        </Button>
                                                    </li>
                                                );
                                            })}
                                        </ul>
                                    ) : (
                                        <p className='text-sm text-neutral-500'>No results</p>
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
