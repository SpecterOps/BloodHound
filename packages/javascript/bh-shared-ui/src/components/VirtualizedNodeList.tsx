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

import { Tooltip } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagNode, GraphNode } from 'js-client-library';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { cn } from '../utils';
import NodeIcon from './NodeIcon';

export type NormalizedNodeItem = {
    name: string;
    objectId: string;
    kind: string;
    onClick?: (index: number) => void;
    graphId?: string;
};

const isGraphNode = (node: unknown): node is GraphNode => {
    return 'label' in (node as GraphNode);
};

const isAssetGroupTagNode = (node: unknown): node is AssetGroupTagNode => {
    return 'object_id' in (node as AssetGroupTagNode);
};

const isNormalizedNodeItem = (node: unknown): node is NormalizedNodeItem => {
    const castedNode = node as NormalizedNodeItem;
    return 'name' in castedNode && 'objectId' in castedNode && 'kind' in castedNode;
};

const normalizeItem = <T,>(item: T): NormalizedNodeItem => {
    const defaultName = 'NO NAME';

    if (isGraphNode(item)) {
        return {
            ...item,
            name: item.label || item.objectId || defaultName,
        };
    } else if (isAssetGroupTagNode(item)) {
        return {
            ...item,
            name: item.name || item.object_id || defaultName,
            objectId: item.object_id,
            kind: item.primary_kind,
        };
    } else if (isNormalizedNodeItem(item)) {
        return {
            ...item,
            name: item.name || item.objectId || defaultName,
        };
    } else {
        throw new Error('item type is unknown');
    }
};

const InnerElement = ({ style, ...rest }: any) => (
    // Top margin is adjusted to account for FixedSizeList's default of 'overflow: auto'
    // causing the scrollbar to render even for a single node
    <ul style={{ ...style, overflowX: 'hidden', marginTop: 0, overflowY: 'auto' }} {...rest}></ul>
);

const Row = <T,>({ data, index, style }: ListChildComponentProps<NodeList<T>>) => {
    const items = data;
    const item = items[index];
    const normalizedItem = normalizeItem(item);

    return (
        <li
            className={cn(
                'bg-neutral-light-2 dark:bg-neutral-dark-2 flex items-center pl-2 border-y border-y-neutral-light-5 dark:border-y-neutral-dark-5',
                {
                    'bg-neutral-light-3 dark:bg-neutral-dark-3': index % 2 !== 0,
                }
            )}
            style={{
                ...style,
                padding: '0 8px',
            }}
            onClick={() => normalizedItem.onClick?.(index)}
            data-testid='entity-row'>
            <NodeIcon nodeType={normalizedItem.kind} />
            <Tooltip tooltip={normalizedItem.name}>
                <div className='truncate text-ellipsis ml-10'>{normalizedItem.name}</div>
            </Tooltip>
        </li>
    );
};

type NodeList<T> = Array<T>;

interface VirtualizedNodeListProps<T> {
    nodes: NodeList<T>;
    itemSize?: number;
    heightScalar?: number;
}

const VirtualizedNodeList = <T,>({ nodes, itemSize = 32, heightScalar = 16 }: VirtualizedNodeListProps<T>) => {
    return (
        <FixedSizeList
            height={Math.min(nodes.length, heightScalar) * itemSize}
            itemCount={nodes.length}
            itemData={nodes}
            itemSize={itemSize}
            innerElementType={InnerElement}
            width={'100%'}
            initialScrollOffset={0}
            style={{ borderRadius: 4 }}>
            {Row}
        </FixedSizeList>
    );
};

export default VirtualizedNodeList;
