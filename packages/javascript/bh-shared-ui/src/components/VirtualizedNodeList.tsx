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
import { GraphNode } from 'js-client-library';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { cn } from '../utils';
import NodeIcon from './NodeIcon';

export type VirtualizedNodeListItem = {
    name: string;
    objectId: string;
    graphId?: string;
    kind: string;
    onClick?: (index: number) => void;
};

const isGraphNode = (node: any): node is GraphNode => {
    return node.isTierZero !== undefined;
};

const normalizeItem = (item: VirtualizedNodeListItem | GraphNode): VirtualizedNodeListItem => {
    if (isGraphNode(item))
        return {
            ...item,
            name: item.label || item.objectId || 'NO NAME',
        };
    else
        return {
            ...item,
            name: item.name || item.objectId || 'NO NAME',
        };
};

const InnerElement = ({ style, ...rest }: any) => (
    // Top margin is adjusted to account for FixedSizeList's default of 'overflow: auto'
    // causing the scrollbar to render even for a single node
    <ul style={{ ...style, overflowX: 'hidden', marginTop: 0 }} {...rest}></ul>
);

const Row = ({ data, index, style }: ListChildComponentProps<VirtualizableNodes>) => {
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

type NodeList<T> = T[];

export type VirtualizableNodes = NodeList<VirtualizedNodeListItem | GraphNode>;

interface VirtualizedNodeListProps {
    nodes: VirtualizableNodes;
    itemSize?: number;
    heightScalar?: number;
}

const VirtualizedNodeList = ({ nodes, itemSize = 32, heightScalar = 16 }: VirtualizedNodeListProps) => {
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
