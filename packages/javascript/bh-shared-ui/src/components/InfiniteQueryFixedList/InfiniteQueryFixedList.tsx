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

import React, { ForwardedRef, useCallback, useMemo, useRef } from 'react';
import { UseInfiniteQueryResult } from 'react-query';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { useMeasure } from '../../hooks/useMeasure';
import { PaginatedResult } from '../../utils/paginatedFetcher';

export type InfiniteQueryFixedListProps<T> = {
    itemSize: number;
    queryResult: UseInfiniteQueryResult<PaginatedResult<T>>;
    renderRow: (item: T, index: number, style: React.CSSProperties, isScrolling?: boolean) => React.ReactNode;
    renderLoadingRow?: (index: number, style: React.CSSProperties) => React.ReactNode;
    overscanCount?: number;
    thresholdCount?: number;
    placeholderCount?: number;
    listRef?: ForwardedRef<FixedSizeList<T[]>>;
};

export const InfiniteQueryFixedList = <T,>({
    itemSize,
    queryResult,
    renderRow,
    renderLoadingRow,
    overscanCount = 5,
    thresholdCount = 5,
    placeholderCount = 1,
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
                if (renderLoadingRow) return renderLoadingRow(index, style);
                return <div style={style}>Loading {index}...</div>;
            }
            return renderRow(itemData[index], index, style, isScrolling);
        },
        [isItemLoaded, renderRow, renderLoadingRow]
    );

    const isInitialLoading = queryResult.isLoading && !isFetchingNextPage;

    const totalRows = hasNextPage ? items.length + 1 : items.length;

    const itemCount = isInitialLoading ? placeholderCount : totalRows;

    return (
        <div ref={containerRef} className='h-full w-full'>
            {height > 0 && width > 0 && (
                <FixedSizeList<T[]>
                    ref={listRef}
                    height={height}
                    itemSize={itemSize}
                    itemCount={itemCount}
                    overscanCount={overscanCount}
                    width={width}
                    onItemsRendered={({ visibleStopIndex }) => {
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
