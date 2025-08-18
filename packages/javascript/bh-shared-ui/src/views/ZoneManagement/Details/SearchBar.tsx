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
import { Button, Input } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagSearchResponse,
    AssetGroupTag,
    AssetGroupTagSelector,
    AssetGroupTagMember,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeTier,
} from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { AppIcon } from '../../../components';
import { useDebouncedValue } from '../../../hooks';
import { apiClient, cn, useAppNavigate } from '../../../utils';
import { getTagUrlValue } from '../utils';

type SearchBarProps = {
    selected: string | undefined;
};

type SectorKey = 'tags' | 'selectors' | 'members';

type Sector = 'Tiers' | 'Selectors' | 'Objects';

type SearchItem = AssetGroupTag | AssetGroupTagSelector | AssetGroupTagMember;

const sectorMap: Record<Sector, SectorKey> = {
    Tiers: 'tags',
    Selectors: 'selectors',
    Objects: 'members',
};

const SearchBar: React.FC<SearchBarProps> = ({ selected }) => {
    const [query, setQuery] = useState('');
    const [open, setOpen] = useState(false);
    const [results, setResults] = useState<AssetGroupTagSearchResponse>({
        tags: [],
        selectors: [],
        members: [],
    });
    const debouncedInputValue = useDebouncedValue(query, 300);
    const navigate = useAppNavigate();
    const { labelId } = useParams();
    const inputRef = React.useRef<HTMLDivElement>(null);

    const scope = labelId ? 'label' : 'tier';
    const searchQuery = useQuery({
        queryKey: ['zone-management', 'search', debouncedInputValue, selected, scope],
        queryFn: async () => {
            const body = {
                query: debouncedInputValue,
                tag_type: scope ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier,
            };
            const res = await apiClient.searchAssetGroupTags(body);
            return res.data.data;
        },
        enabled: debouncedInputValue.length >= 3 && selected !== undefined,
    });

    useEffect(() => {
        if (debouncedInputValue.length < 3) {
            setResults({ tags: [], selectors: [], members: [] });
            setOpen(false);
            return;
        }

        if (searchQuery.data) {
            const { tags, selectors, members } = searchQuery.data;

            setResults({
                tags: tags ?? [],
                selectors: selectors ?? [],
                members: members ?? [],
            });

            setOpen(true);
        }
    }, [debouncedInputValue, searchQuery.data]);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (inputRef.current && event.target instanceof Node && !inputRef.current.contains(event.target)) {
                setOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleClick = (sector: Sector, item: SearchItem) => {
        setOpen(false);
        setQuery('');
        let url = '';
        const base = getTagUrlValue(labelId);

        if (sector === 'Tiers') {
            const tag = item as AssetGroupTag;
            url = `/zone-management/details/${base}/${tag.id}`;
        } else if (sector === 'Selectors') {
            const selector = item as AssetGroupTagSelector;
            url = `/zone-management/details/${base}/${selector.asset_group_tag_id}/selector/${selector.id}`;
        } else if (sector === 'Objects') {
            const member = item as AssetGroupTagMember;
            url = `/zone-management/details/${base}/${member.asset_group_tag_id}/member/${member.id}`;
        }
        navigate(url);
    };

    return (
        <div className='relative w-4/6' ref={inputRef}>
            <div className='flex items-center border-b-2 border-neutral-dark-1 dark:border-neutral-light-1'>
                <AppIcon.MagnifyingGlass />
                <Input
                    placeholder='Search'
                    className='border-none bg-transparent dark:bg-transparent placeholder-neutral-dark-1 dark:placeholder-neutral-light-1 focus-visible:ring-0 focus-visible:ring-offset-0 pl-3'
                    onChange={(e) => {
                        setQuery(e.target.value);
                    }}
                    value={query}
                    onFocus={() => {
                        if (query.length >= 3) setOpen(true);
                    }}
                />
            </div>
            {open && (
                <div className='absolute w-[512px] bg-neutral-light-2 dark:bg-neutral-dark-2 border border-neutral-light-5 dark:border-neutral-dark-5 z-10'>
                    {(['Tiers', 'Selectors', 'Objects'] as Sector[]).map((sector) => (
                        <div key={sector}>
                            <p className='font-bold'>{sector}</p>
                            {results[sectorMap[sector]].length > 0 ? (
                                <ul>
                                    {[...results[sectorMap[sector]]]
                                        .sort((a, b) =>
                                            a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })
                                        )
                                        .map((item: SearchItem) => (
                                            <li
                                                className={cn('flex max-w-lg min-w-0', {
                                                    'bg-neutral-light-4 dark:bg-neutral-dark-4':
                                                        (item.id as number) % 2 === 0,
                                                })}
                                                key={item.id}>
                                                <Button
                                                    className='overflow-hidden'
                                                    variant={'text'}
                                                    onClick={() => handleClick(sector, item)}>
                                                    <span className='truncate'>{item.name}</span>
                                                </Button>
                                            </li>
                                        ))}
                                </ul>
                            ) : (
                                <p>No results</p>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default SearchBar;
