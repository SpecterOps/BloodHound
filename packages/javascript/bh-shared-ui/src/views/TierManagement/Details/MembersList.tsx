import { Button } from '@bloodhoundenterprise/doodleui';
import { useCallback, useEffect, useRef, useState } from 'react';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import InfiniteLoader from 'react-window-infinite-loader';
import { NodeIcon, SortableHeader } from '../../../components';
import { SortOrder } from '../../../types';
import { apiClient, cn } from '../../../utils';
import { ItemSkeleton, SelectedHighlight } from './utils';

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
    selectedTier: number;
    selectedSelector: number | null;
    selected: number | null;
    onClick: (id: number) => void;
    itemCount: number;
}

export const MembersList: React.FC<MembersListProps> = ({
    selectedTier,
    selectedSelector,
    selected,
    onClick,
    itemCount,
}) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>();
    const [isFetching, setIsFetching] = useState(false);
    const [items, setItems] = useState<any>({});
    const infiniteLoaderRef = useRef<any>(null);
    const listRef = useRef<any>(null);
    const previousSelector = usePreviousValue<number | null>(selectedSelector);
    const previousTier = usePreviousValue<number>(selectedTier);

    const itemData = { onClick, selected, items, title: 'Members' };

    const isItemLoaded = (index: number) => {
        return !!items[index];
    };

    const loadMoreItems = useCallback(
        async (startIndex: number, stopIndex: number) => {
            if (isFetching) return;

            setIsFetching(true);

            const limit = stopIndex - startIndex + 1;

            const fetchData = getFetchCallback(selectedTier, selectedSelector);

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
        [items, isFetching, selectedSelector, selectedTier]
    );

    useEffect(() => {
        if (previousSelector !== selectedSelector || previousTier !== selectedTier) {
            if (listRef?.current?.scrollTo) listRef.current.scrollTo(0);
            if (infiniteLoaderRef?.current?.resetLoadMoreItemsCache)
                infiniteLoaderRef?.current?.resetLoadMoreItemsCache(true);
            loadMoreItems(0, 128);
        }
    }, [selectedSelector, selectedTier, loadMoreItems, previousSelector, previousTier]);

    return (
        <div data-testid={`tier-management_details_members-list`}>
            <SortableHeader
                title={'Objects'}
                onSort={() => {
                    if (sortOrder === undefined) {
                        // first click
                        setSortOrder('desc');
                    } else if (sortOrder === 'desc') {
                        // second click
                        setSortOrder('asc');
                    } else if (sortOrder === 'asc') {
                        // third click
                        setSortOrder(undefined);
                    }
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
                {({ onItemsRendered }) => (
                    <FixedSizeList
                        height={500}
                        itemCount={itemCount}
                        itemData={itemData}
                        itemSize={ITEM_SIZE}
                        onItemsRendered={onItemsRendered}
                        innerElementType={InnerElement}
                        ref={listRef}
                        width={'100%'}>
                        {Row}
                    </FixedSizeList>
                )}
            </InfiniteLoader>
        </div>
    );
};
