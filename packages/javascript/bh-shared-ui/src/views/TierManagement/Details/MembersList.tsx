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
import { AssetGroupTagSelectorNode } from 'js-client-library';
import { useCallback, useEffect, useRef, useState } from 'react';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import InfiniteLoader from 'react-window-infinite-loader';
import { NodeIcon, SortableHeader } from '../../../components';
import { SortOrder } from '../../../types';
import { apiClient, cn } from '../../../utils';
import { ItemSkeleton, SelectedHighlight } from './utils';

// TODO: Move this out and pull from shared-ui hooks once it is brought in with in progress changes in 5231
const usePreviousValue = <T,>(value: T): T | undefined => {
    const ref = useRef<T>();
    useEffect(() => {
        ref.current = value;
    });
    return ref.current;
};

const ITEM_SIZE = 40;

const Row = ({
    data,
    index,
    style,
}: ListChildComponentProps<{
    selected: number | null;
    title: string;
    onClick: (id: number) => void;
    items: any;
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
                'bg-neutral-light-4 dark:bg-neutral-dark-4': selected === listItem.id,
            })}
            style={style}>
            <SelectedHighlight selected={selected} itemId={listItem.id} title={title} />
            <Button
                variant={'text'}
                className='flex justify-start w-full'
                onClick={() => {
                    onClick(listItem.id);
                }}>
                <NodeIcon nodeType={listItem.kind || 'Unknown'} />
                <span className='text-base ml-2'>{listItem.name}</span>
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

const getFetchCallback = (selectedTier: number, selectedSelector: number | null) => {
    if (selectedSelector !== null) {
        return ({ skip, limit }: { skip: number; limit: number }) => {
            return apiClient.getAssetGroupSelectorMembers(selectedTier, selectedSelector, skip, limit).then((res) => {
                const response = {
                    data: res.data.data.members,
                    skip: res.data.skip,
                    limit: res.data.limit,
                    total: res.data.count,
                };
                return response;
            });
        };
    } else {
        return ({ skip, limit }: { skip: number; limit: number }) => {
            return apiClient.getAssetGroupLabelMembers(selectedTier, skip, limit).then((res) => {
                const response = {
                    data: res.data.data.members,
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
    selectedTag: number;
    selectedSelector: number | null;
    selected: number | null;
    onClick: (id: number) => void;
    itemCount?: number;
}

/**
 * @description This component is used to render the Objects/Members list for a given Tier, Label, or Selector. It is specifically built with both a fixed render window and a scroll loader as it is expected that the number of entities that this list may display would be large enough that trying to load all of these DOM nodes at once would cause the page to be sluggish and result in a poor user experience.
 * @param props
 * @param {selectedTier} props.selectedTag The currently selected Tier/Label. This is used to fill in the id for the path parameter of the endpoint that is used to fetch the list of members for the given selection
 * @param {selectedSelector} props.selectedSelector The currently selected Selector. This is used to fill in the id for the path parameter of the endpoint that is used to fetch the list of members for the given selection. Unlike a selectedTier, this param can be null if there is no Selector selected.
 * @param {selected} props.selected The currently selected Object/Member. This selection can be null.
 * @param {onClick} props.onClick The click handler for when a particular member is selected. This is primarily used for setting the selected entity in the parent component.
 * @param {itemCount} props.itemCount The total item count for the list that is to be rendered. This informs the `InfiniteLoader` component as to when the list will end since the data is being fetched in pages as opposed to all at once.
 * @returns The MembersList component for rendering in the Tier Management page.
 */
export const MembersList: React.FC<MembersListProps> = ({
    selectedTag,
    selectedSelector,
    selected,
    onClick,
    itemCount = 0,
}) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>();
    const [isFetching, setIsFetching] = useState(false);
    const [items, setItems] = useState<Record<number, AssetGroupTagSelectorNode>>({});
    const infiniteLoaderRef = useRef<InfiniteLoader | null>(null);
    const previousSelector = usePreviousValue<number | null>(selectedSelector);
    const previousTier = usePreviousValue<number>(selectedTag);

    const itemData = { onClick, selected, items, title: 'Members' };

    const isItemLoaded = (index: number) => {
        return !!items[index];
    };

    const loadMoreItems = useCallback(
        async (startIndex: number, stopIndex: number) => {
            if (isFetching) return;

            setIsFetching(true);

            const limit = stopIndex - startIndex + 1;

            const fetchData = getFetchCallback(selectedTag, selectedSelector);

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
        },
        [items, isFetching, selectedSelector, selectedTag]
    );

    // Because the endpoint that needs to be used to fetch the list of members is dynamic based on whether
    // a selector is selected or not, this useEffect is used so that the cache of the `InfiniteLoader`
    // component is reset which triggers calling the appropriate endpoint when a selected tier or selected
    // selector changes. Without this useEffect, the list of objects/members does not clear when new data
    // is fetched.
    useEffect(() => {
        if (previousSelector !== selectedSelector || previousTier !== selectedTag) {
            if (infiniteLoaderRef?.current?.resetloadMoreItemsCache)
                infiniteLoaderRef?.current?.resetloadMoreItemsCache(true);
            loadMoreItems(0, 128);
        }
    }, [selectedSelector, selectedTag, loadMoreItems, previousSelector, previousTier]);

    return (
        <div data-testid={`tier-management_details_members-list`}>
            <SortableHeader
                title={'Objects'}
                onSort={() => {
                    sortOrder === 'desc' ? setSortOrder('asc') : setSortOrder('desc');
                }}
                sortOrder={sortOrder}
                classes={{
                    container: 'border-b-2 border-neutral-light-5 dark:border-neutral-dark-5',
                    button: 'pl-6 font-bold text-xl',
                }}
            />
            <InfiniteLoader
                ref={infiniteLoaderRef}
                threshold={32}
                minimumBatchSize={128}
                isItemLoaded={isItemLoaded}
                itemCount={itemCount}
                loadMoreItems={loadMoreItems}>
                {({ onItemsRendered, ref }) => (
                    <FixedSizeList
                        height={522}
                        itemCount={itemCount}
                        itemData={itemData}
                        itemSize={ITEM_SIZE}
                        onItemsRendered={onItemsRendered}
                        innerElementType={InnerElement}
                        ref={ref}
                        width={'100%'}>
                        {Row}
                    </FixedSizeList>
                )}
            </InfiniteLoader>
        </div>
    );
};
