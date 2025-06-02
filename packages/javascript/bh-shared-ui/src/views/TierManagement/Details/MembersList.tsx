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
import { AssetGroupTagMemberListItem } from 'js-client-library';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import InfiniteLoader from 'react-window-infinite-loader';
import { NodeIcon, SortableHeader } from '../../../components';
import { usePreviousValue } from '../../../hooks';
import { SortOrder } from '../../../types';
import { apiClient, cn } from '../../../utils';
import { ItemSkeleton, SelectedHighlight, getListHeight } from './utils';

const ITEM_SIZE = 40;

const Row = ({
    data,
    index,
    style,
}: ListChildComponentProps<{
    selected: string | undefined;
    title: string;
    onClick: (id: string) => void;
    items: Record<number, AssetGroupTagMemberListItem>;
}>) => {
    const { items, onClick, selected, title } = data;
    const listItem = items[index];

    if (listItem === undefined) {
        return ItemSkeleton(title, index, { ...style, whiteSpace: 'nowrap', padding: '0 8px' });
    }

    return (
        <li
            key={index}
            className={cn('border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative', {
                'bg-neutral-light-4 dark:bg-neutral-dark-4': selected === listItem.id.toString(),
            })}
            style={style}>
            <SelectedHighlight selected={selected} itemId={listItem.id} title={title} />
            <Button
                variant={'text'}
                className='flex justify-start w-full'
                title={`Type: ${listItem.primary_kind}; Name: ${listItem.name}`}
                onClick={() => {
                    onClick(listItem.id?.toString());
                }}>
                <NodeIcon nodeType={listItem.primary_kind} />
                <span className='text-base dark:text-white ml-2 truncate'>{listItem.name}</span>
            </Button>
        </li>
    );
};

const InnerElement = ({ style, ...rest }: any) => (
    <ul
        style={{ ...style, overflowX: 'hidden' }}
        className={'border-neutral-light-5 dark:border-neutral-dark-5'}
        {...rest}></ul>
);

const getFetchCallback = (
    selectedTag: string | undefined,
    selectedSelector: string | undefined,
    sortOrder: SortOrder
) => {
    const tag = selectedTag || 1;

    const sort_by = sortOrder === 'asc' ? 'name' : '-name';

    if (selectedSelector) {
        return ({ skip, limit }: { skip: number; limit: number }) => {
            return apiClient
                .getAssetGroupTagSelectorMembers(tag, selectedSelector, skip, limit, sort_by)
                .then((res) => {
                    const response = {
                        data: res.data.data['members'],
                        skip: res.data.skip,
                        limit: res.data.limit,
                        total: res.data.count,
                    };
                    return response;
                });
        };
    } else {
        return ({ skip, limit }: { skip: number; limit: number }) => {
            return apiClient.getAssetGroupTagMembers(tag, skip, limit, sort_by).then((res) => {
                const response = {
                    data: res.data.data['members'],
                    skip: res.data.skip,
                    limit: res.data.limit,
                    total: res.data.count,
                };
                return response;
            });
        };
    }
};

interface MembersListProps {
    selected: string | undefined;
    onClick: (id: string) => void;
    itemCount?: number;
}

/**
 * @description This component is used to render the Objects/Members list for a given Tier, Label, or Selector. It is specifically built with both a fixed render window and a scroll loader as it is expected that the number of entities that this list may display would be large enough that trying to load all of these DOM nodes at once would cause the page to be sluggish and result in a poor user experience.
 * @param props
 * @param {selected} props.selected The currently selected Object/Member. This selection can be null.
 * @param {onClick} props.onClick The click handler for when a particular member is selected. This is primarily used for setting the selected entity in the parent component.
 * @param {itemCount} props.itemCount The total item count for the list that is to be rendered. This informs the `InfiniteLoader` component as to when the list will end since the data is being fetched in pages as opposed to all at once.
 * @returns The MembersList component for rendering in the Tier Management page.
 */
export const MembersList: React.FC<MembersListProps> = ({ selected, onClick, itemCount = 0 }) => {
    const { tierId, labelId, selectorId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const [sortOrder, setSortOrder] = useState<SortOrder>('asc');
    const [isFetching, setIsFetching] = useState(false);
    const [items, setItems] = useState<Record<number, AssetGroupTagMemberListItem>>({});

    const infiniteLoaderRef = useRef<InfiniteLoader | null>(null);
    const heightRef = useRef(getListHeight(window.innerHeight));

    const previousSelector = usePreviousValue<string | undefined>(selectorId);
    const previousTag = usePreviousValue<string | undefined>(tagId);
    const previousSortOrder = usePreviousValue<SortOrder>(sortOrder);

    const clearCache =
        (selectorId && selectorId !== previousSelector) ||
        (tagId && tagId !== previousTag) ||
        sortOrder !== previousSortOrder;

    const itemData = { onClick, selected, items, title: 'Members' };

    useEffect(() => {
        const updateListHeight = () => {
            heightRef.current = getListHeight(window.innerHeight);
        };

        window.addEventListener('resize', updateListHeight);

        return () => window.removeEventListener('resize', updateListHeight);
    }, []);

    const isItemLoaded = useCallback(
        (index: number) => {
            if (clearCache) return false;
            return !!items[index];
        },
        [items, clearCache]
    );

    const fetchData = useMemo(() => getFetchCallback(tagId, selectorId, sortOrder), [tagId, selectorId, sortOrder]);

    const loadMoreItems = async (startIndex: number, stopIndex: number) => {
        if (isFetching) return;

        setIsFetching(true);

        const limit = stopIndex - startIndex + 1;

        return fetchData({ skip: startIndex, limit: limit })
            .then((data) => {
                const itemMap: any = {};

                for (let i = 0; i < limit; i++) {
                    itemMap[i + startIndex] = data.data[i];
                }

                setItems(Object.assign({}, items, itemMap));
            })
            .finally(() => {
                setIsFetching(false);
            });
    };

    // Because the endpoint that needs to be used to fetch the list of members is dynamic based on whether
    // a selector is selected or not, this useEffect is used so that the cache of the `InfiniteLoader`
    // component is reset which triggers calling the appropriate endpoint when a selected tier or selected
    // selector changes. Without this useEffect, the list of objects/members does not clear when new data
    // is fetched.
    useEffect(() => {
        if (!clearCache) return;

        if (infiniteLoaderRef?.current?.resetloadMoreItemsCache) {
            infiniteLoaderRef?.current?.resetloadMoreItemsCache(true);
        }
    }, [clearCache, isItemLoaded]);

    return (
        <div data-testid={`tier-management_details_members-list`}>
            <SortableHeader
                title={'Objects'}
                onSort={() => {
                    setSortOrder((prev) => {
                        if (prev === 'asc') {
                            return 'desc';
                        } else {
                            return 'asc';
                        }
                    });
                }}
                sortOrder={sortOrder}
                classes={{
                    container: 'border-b-2 border-neutral-light-5 dark:border-neutral-dark-5',
                    button: 'pl-6 font-bold text-xl',
                }}
            />
            <InfiniteLoader
                ref={infiniteLoaderRef}
                threshold={64}
                minimumBatchSize={128}
                isItemLoaded={isItemLoaded}
                itemCount={itemCount}
                loadMoreItems={loadMoreItems}>
                {({ onItemsRendered, ref }) => (
                    <FixedSizeList
                        height={heightRef.current}
                        itemCount={itemCount}
                        itemData={itemData}
                        itemSize={ITEM_SIZE}
                        ref={ref}
                        onItemsRendered={onItemsRendered}
                        innerElementType={InnerElement}
                        width={'100%'}>
                        {Row}
                    </FixedSizeList>
                )}
            </InfiniteLoader>
        </div>
    );
};
