import { List, ListItemButton, Tooltip } from '@mui/material';
import { FixedSizeList, ListChildComponentProps } from 'react-window';
import { NodeIcon } from '.';

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
