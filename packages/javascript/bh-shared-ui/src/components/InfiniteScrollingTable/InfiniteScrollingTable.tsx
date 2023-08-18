// Copyright 2023 Specter Ops, Inc.
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

import { List, ListItem, Tooltip, Skeleton } from '@mui/material';
import memoize from 'memoize-one';
import React, { memo, useState } from 'react';
import { areEqual, FixedSizeList, ListChildComponentProps } from 'react-window';
import InfiniteLoader from 'react-window-infinite-loader';
import NodeIcon from '../NodeIcon';

const ITEM_SIZE = 32;

const InnerElement = ({ style, ...rest }: any) => (
    <List
        component='ul'
        data-testid='infinite-scroll-table'
        disablePadding
        style={{ ...style, overflowX: 'hidden' }}
        {...rest}
    />
);

interface InfiniteScrollingTableProps {
    fetchDataCallback: ({ skip, limit }: { skip: number; limit: number }) => Promise<{
        data: any[];
        total: number;
        limit: number;
        skip: number;
    }>;
    itemCount?: number;
    onClick?: (item: any) => void;
}

const createItemData = memoize((items, onClick) => ({
    items,
    onClick,
}));

const Row = memo(({ data, index, style }: ListChildComponentProps) => {
    const { items, onClick } = data;
    const item = items[index];
    const itemClass = index % 2 ? 'odd-item' : 'even-item';

    if (item === undefined) {
        return (
            <ListItem
                className={itemClass}
                style={{ ...style, whiteSpace: 'nowrap', padding: '0 8px' }}
                data-testid='entity-row'>
                <Skeleton variant='text' width='100%' />
            </ListItem>
        );
    }

    const normalizedItem = {
        id: item.objectID || item.props?.objectid || '',
        name: item.name || item.objectID || item.props?.name || item.props?.objectid || 'Unknown',
        type: item.label || item.kind || '',
    };

    return (
        <ListItem
            button
            className={itemClass}
            onClick={() => {
                onClick(normalizedItem);
            }}
            style={{
                ...style,
                padding: '0 8px',
            }}
            data-testid='entity-row'>
            <NodeIcon nodeType={normalizedItem.type} />
            <Tooltip title={normalizedItem.name}>
                <div style={{ minWidth: '0', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {normalizedItem.name}
                </div>
            </Tooltip>
        </ListItem>
    );
}, areEqual);

const InfiniteScrollingTable: React.FC<InfiniteScrollingTableProps> = ({
    fetchDataCallback,
    itemCount = 1000,
    onClick = () => {},
}) => {
    const [isFetching, setIsFetching] = useState(false);
    const [items, setItems] = useState<Record<number, any>>({});
    const itemData = createItemData(items, onClick);
    const isItemLoaded = (index: number) => !!items[index];

    const loadMoreItems = async (startIndex: number, stopIndex: number) => {
        if (isFetching) return;

        setIsFetching(true);
        const limit = stopIndex - startIndex + 1;
        return fetchDataCallback({ skip: startIndex, limit: limit })
            .then((data) => {
                const itemMap: Record<string, any> = {};
                for (let i = 0; i < limit; i++) {
                    itemMap[i + startIndex] = data.data[i];
                }
                setItems(Object.assign({}, items, itemMap));
            })
            .finally(() => {
                setIsFetching(false);
            });
    };

    return (
        <InfiniteLoader
            threshold={32}
            minimumBatchSize={128}
            isItemLoaded={isItemLoaded}
            itemCount={itemCount}
            loadMoreItems={loadMoreItems}>
            {({ onItemsRendered, ref }) => (
                <FixedSizeList
                    height={Math.min(itemCount, 16) * ITEM_SIZE}
                    itemCount={itemCount}
                    itemData={itemData}
                    itemSize={ITEM_SIZE}
                    onItemsRendered={onItemsRendered}
                    innerElementType={InnerElement}
                    ref={ref}
                    width={'100%'}
                    initialScrollOffset={0}
                    style={{ borderRadius: 4 }}>
                    {Row}
                </FixedSizeList>
            )}
        </InfiniteLoader>
    );
};

export default InfiniteScrollingTable;
