import React, { ForwardedRef, useCallback, useMemo, useRef } from 'react';
import { UseInfiniteQueryResult } from 'react-query';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { useMeasure } from '../../hooks/useMeasure';

type PageParam = {
    skip: number;
    limit: number;
};

type PageWithItems<T> = {
    items: T[];
    nextPageParam?: PageParam;
};

export type InfiniteQueryFixedListProps<T> = {
    itemSize: number;
    queryResult: UseInfiniteQueryResult<PageWithItems<T>>;
    renderRow: (item: T, index: number, style: React.CSSProperties, isScrolling?: boolean) => React.ReactNode;
    overscanCount?: number;
    thresholdCount?: number;
    listRef?: ForwardedRef<FixedSizeList<T[]>>;
};

export const InfiniteQueryFixedList = <T,>({
    itemSize,
    queryResult,
    renderRow,
    overscanCount = 5,
    thresholdCount = 5,
    listRef,
}: InfiniteQueryFixedListProps<T>) => {
    const containerRef = useRef<HTMLDivElement>(null);

    const [width, height] = useMeasure(containerRef);

    const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = queryResult;

    const items = useMemo(() => data?.pages.flatMap((page) => page.items) ?? [], [data]);

    const isItemLoaded = useCallback((index: number) => index < items.length, [items.length]);

    const loadMoreItems = useCallback(async () => {
        if (!isFetchingNextPage) await fetchNextPage();
    }, [isFetchingNextPage, fetchNextPage]);

    const Row = useCallback(
        ({ index, style, data: itemData, isScrolling }: ListChildComponentProps<T[]>) => {
            if (!isItemLoaded(index)) {
                return <div style={style}>Loading...</div>;
            }
            return renderRow(itemData[index], index, style, isScrolling);
        },
        [isItemLoaded, renderRow]
    );

    return (
        <div ref={containerRef} className='h-full w-full'>
            {height > 0 && width > 0 && (
                <FixedSizeList<T[]>
                    ref={listRef}
                    height={height}
                    itemSize={itemSize}
                    itemCount={hasNextPage ? items.length + 1 : items.length}
                    overscanCount={overscanCount}
                    width={width}
                    onItemsRendered={({ visibleStopIndex }) => {
                        console.log(hasNextPage, visibleStopIndex);
                        if (hasNextPage && !isFetchingNextPage && visibleStopIndex >= items.length - thresholdCount) {
                            loadMoreItems();
                        }
                    }}
                    itemData={items}>
                    {Row}
                </FixedSizeList>
            )}
        </div>
    );
};
