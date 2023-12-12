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

import { List, ListItemButton, Tooltip } from '@mui/material';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import NodeIcon from './NodeIcon';

export type VirtualizedNodeListItem = {
    name: string;
    objectId: string;
    kind: string;
    onClick?: (index: number) => void;
};

interface VirtualizedNodeListProps {
    nodes: VirtualizedNodeListItem[];
    itemSize?: number;
}

const normalizeItem = (item: VirtualizedNodeListItem): VirtualizedNodeListItem => ({
    name: item.name || item.objectId || 'Unknown',
    objectId: item.objectId,
    kind: item.kind || '',
    onClick: item.onClick,
});

const InnerElement = ({ style, ...rest }: any) => (
    <List component='ul' disablePadding style={{ ...style, overflowX: 'hidden' }} {...rest} />
);

const Row = ({ data, index, style }: ListChildComponentProps<VirtualizedNodeListItem[]>) => {
    const items = data;
    const item = items[index];
    const normalizedItem = normalizeItem(item);
    const itemClass = index % 2 ? 'odd-item' : 'even-item';

    return (
        <ListItemButton
            className={itemClass}
            style={{
                ...style,
                padding: '0 8px',
            }}
            onClick={() => normalizedItem.onClick?.(index)}
            data-testid='entity-row'>
            <NodeIcon nodeType={normalizedItem.kind} />
            <Tooltip title={normalizedItem.name}>
                <div style={{ minWidth: '0', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {normalizedItem.name}
                </div>
            </Tooltip>
        </ListItemButton>
    );
};

const VirtualizedNodeList = ({ nodes, itemSize = 32 }: VirtualizedNodeListProps) => {
    return (
        <FixedSizeList
            height={Math.min(nodes.length, 16) * itemSize}
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
