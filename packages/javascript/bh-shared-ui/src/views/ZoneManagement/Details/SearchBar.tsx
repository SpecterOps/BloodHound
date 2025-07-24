import { Button, Input } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagTypeLabel, AssetGroupTagTypeTier } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { AppIcon } from '../../../components';
import { apiClient, cn, useAppNavigate } from '../../../utils';
import { getTagUrlValue } from '../utils';

type SearchBarProps = {
    selected: string;
    selectorId: string;
    labelId: string;
};

type Item = {
    id: number;
    name: string;
};

type SectorKey = 'tags' | 'selectors' | 'members';

type Sector = 'Tiers' | 'Selectors' | 'Objects';

const sectorMap: Record<Sector, SectorKey> = {
    Tiers: 'tags',
    Selectors: 'selectors',
    Objects: 'members',
};

const SearchBar: React.FC<SearchBarProps> = ({ selected, selectorId }) => {
    const [query, setQuery] = useState('');
    const [open, setOpen] = useState(false);
    const [results, setResults] = useState<Record<SectorKey, Item[]>>({ tags: [], selectors: [], members: [] });
    const navigate = useAppNavigate();
    const { tierId, labelId } = useParams();
    const inputRef = React.useRef<HTMLDivElement>(null);

    const searchQuery = useQuery({
        queryKey: ['zone-management', 'search', query, selected],
        queryFn: async () => {
            const body = { query, tag_type: labelId ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier };
            const res = await apiClient.searchAssetGroupTags(body);
            return res.data.data;
        },
        enabled: query.length >= 3 && selected !== undefined,
    });

    useEffect(() => {
        if (query.length < 3) {
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
    }, [query, searchQuery.data]);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (inputRef.current && event.target instanceof Node && !inputRef.current.contains(event.target)) {
                setOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleClick = (sector: Sector, item: Item) => {
        setOpen(false);
        setQuery('');
        let url = '';
        const base = getTagUrlValue(labelId);

        if (sector === 'Tiers') {
            url = `/zone-management/details/${base}/${item.asset_group_tag_id}`;
        } else if (sector === 'Selectors') {
            url = `/zone-management/details/${base}/${item.asset_group_tag_id}/selector/${item.id}`;
        } else if (sector === 'Objects') {
            if (selectorId) {
                url = `/zone-management/details/${base}/${labelId}/selector/${selectorId}/member/${item.id}`;
            } else {
                url = `/zone-management/details/${base}/${item.asset_group_tag_id}/member/${item.id}`;
            }
        }
        navigate(url);
    };

    return (
        <div className='relative' ref={inputRef}>
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
                <div className='absolute min-w-lg bg-neutral-light-2 dark:bg-neutral-dark-2 border border-neutral-light-5 dark:border-neutral-dark-5 z-10'>
                    {(['Tiers', 'Selectors', 'Objects'] as Sector[]).map((sector) => (
                        <div key={sector}>
                            <p className='font-bold'>{sector}</p>
                            {results[sectorMap[sector]].length > 0 ? (
                                <ul>
                                    {results[sectorMap[sector]].map((item, index: number) => (
                                        <li
                                            className={cn('flex overflow-hidden max-w-lg', {
                                                'bg-neutral-light-4 dark:bg-neutral-dark-4': index % 2 === 0,
                                            })}
                                            key={item.id}>
                                            <Button
                                                className='max-w-full overflow-hidden'
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
