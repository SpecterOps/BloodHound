import React, { useCallback, useMemo, useRef } from 'react';
import { UseInfiniteQueryResult } from 'react-query';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import InfiniteLoader from 'react-window-infinite-loader';
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
};

export const InfiniteQueryFixedList = <T,>({ itemSize, queryResult, renderRow }: InfiniteQueryFixedListProps<T>) => {
    const containerRef = useRef<HTMLDivElement>(null);

    const [width, height] = useMeasure(containerRef);

    const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = queryResult;

    const items = useMemo(() => data?.pages.flatMap((page) => page.items) ?? [], [data?.pages]);

    const itemCount = hasNextPage ? items.length + 1 : items.length;

    const isItemLoaded = useCallback(
        (index: number) => !hasNextPage || index < items.length,
        [hasNextPage, items.length]
    );

    const loadMoreItems = async () => {
        if (!isFetchingNextPage) await fetchNextPage();
    };

    const Row = useCallback(
        ({ index, style, data, isScrolling }: ListChildComponentProps<T[]>) => {
            if (!isItemLoaded(index)) {
                return <div style={style}>Loading...</div>;
            }
            return renderRow(data[index], index, style, isScrolling);
        },
        [isItemLoaded, renderRow]
    );

    return (
        <div ref={containerRef} className='h-full w-full'>
            {height > 0 && width > 0 && (
                <InfiniteLoader isItemLoaded={isItemLoaded} itemCount={itemCount} loadMoreItems={loadMoreItems}>
                    {({ onItemsRendered, ref }) => (
                        <FixedSizeList<T[]>
                            height={height}
                            itemSize={itemSize}
                            itemCount={itemCount}
                            width={width}
                            onItemsRendered={onItemsRendered}
                            itemData={items}
                            ref={ref}>
                            {Row}
                        </FixedSizeList>
                    )}
                </InfiniteLoader>
            )}
        </div>
    );
};
