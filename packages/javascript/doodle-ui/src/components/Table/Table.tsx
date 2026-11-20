// Copyright 2026 Specter Ops, Inc.
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

import * as React from 'react';

import { useDndContext, useDndMonitor } from '@dnd-kit/core';
import { hasSortableData, horizontalListSortingStrategy, useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { faGripVertical } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Cell, Header } from '@tanstack/react-table';
import { IconButton } from '../IconButton';
import { cn } from '../utils';
import {
    tableBodyClasses,
    tableCaptionClasses,
    tableCellClasses,
    tableClasses,
    tableContainerClasses,
    tableContainerOverflowClasses,
    tableDragHandleClasses,
    tableDragHandleCursorClasses,
    tableDraggingCellClasses,
    tableFooterClasses,
    tableHeadClasses,
    tableHeadContentClasses,
    tableHeaderClasses,
    tableRowClasses,
} from './Table.styles';

type TableDragDelta = { x: number; y: number };
const TableDragDeltasContext = React.createContext<React.MutableRefObject<Map<string, TableDragDelta>> | null>(null);

const TableDragMonitor = ({ dragDeltas }: { dragDeltas: Map<string, TableDragDelta> }) => {
    useDndMonitor({
        onDragMove: ({ active, delta }) => {
            if (typeof active.id === 'string') {
                dragDeltas.set(active.id, delta);
            }
        },
        onDragEnd: () => dragDeltas.clear(),
        onDragCancel: () => dragDeltas.clear(),
    });
    return null;
};

const Table = React.forwardRef<
    HTMLTableElement,
    React.HTMLAttributes<HTMLTableElement> & { disableDefaultOverflowAuto?: boolean; tableContainerClassName?: string }
>(({ className, disableDefaultOverflowAuto, tableContainerClassName, ...props }, ref) => {
    const { active } = useDndContext();
    const dragDeltas = React.useRef(new Map<string, TableDragDelta>());

    return (
        <TableDragDeltasContext.Provider value={dragDeltas}>
            {active && hasSortableData(active) && <TableDragMonitor dragDeltas={dragDeltas.current} />}
            <div
                className={cn(
                    tableContainerClasses,
                    !disableDefaultOverflowAuto && tableContainerOverflowClasses,
                    tableContainerClassName
                )}>
                <table ref={ref} className={cn(tableClasses, className)} {...props} />
            </div>
        </TableDragDeltasContext.Provider>
    );
});
Table.displayName = 'Table';

const TableHeader = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => <thead ref={ref} className={cn(tableHeaderClasses, className)} {...props} />
);
TableHeader.displayName = 'TableHeader';

const TableBody = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => <tbody ref={ref} className={cn(tableBodyClasses, className)} {...props} />
);
TableBody.displayName = 'TableBody';

const TableFooter = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => <tfoot ref={ref} className={cn(tableFooterClasses, className)} {...props} />
);
TableFooter.displayName = 'TableFooter';

const TableRow = React.forwardRef<HTMLTableRowElement, React.HTMLAttributes<HTMLTableRowElement>>(
    ({ className, ...props }, ref) => <tr ref={ref} className={cn(tableRowClasses, className)} {...props} />
);
TableRow.displayName = 'TableRow';

interface TableHeadProps<TData, TValue> extends React.ThHTMLAttributes<HTMLTableCellElement> {
    header?: Header<TData, TValue>;
    // True when table level and column level drag and drop enabled.
    enableDragging?: boolean;
}

const TableHead = React.forwardRef(function TableHead<TData, TValue>(
    { className, header, enableDragging = true, style: propsStyle, ...props }: TableHeadProps<TData, TValue>,
    ref: React.ForwardedRef<HTMLTableCellElement>
) {
    const { attributes, isDragging, listeners, setNodeRef, transform } = useSortable({
        id: header?.column.id ?? '',
        disabled: !enableDragging,
    });

    const headerStyle: React.CSSProperties = {
        opacity: isDragging ? 0.8 : 1,
        position: 'relative',
        transform: CSS.Translate.toString(transform),
        borderRadius: isDragging ? '4px 4px 0 0' : '',
    };

    const zIndex = isDragging ? 40 : header?.column.getIsPinned() ? 30 : 1;
    const leftValue = isDragging && header?.column.getIsPinned() ? 'initial' : '';
    return (
        <th
            ref={setNodeRef}
            className={cn(tableHeadClasses, className, isDragging ? tableDraggingCellClasses : '')}
            {...props}
            style={{
                ...headerStyle,
                ...propsStyle,
                zIndex: zIndex,
                left: leftValue,
            }}>
            <div ref={ref} className={tableHeadContentClasses}>
                {enableDragging && header?.id !== 'empty-column' && (
                    <IconButton
                        aria-label='Drag to reorder'
                        className={cn(
                            tableDragHandleClasses,
                            tableDragHandleCursorClasses[isDragging ? 'dragging' : 'idle']
                        )}
                        {...attributes}
                        {...listeners}>
                        <FontAwesomeIcon icon={faGripVertical} />
                    </IconButton>
                )}
                {props.children}
            </div>
        </th>
    );
}) as (<TData, TValue>(
    props: TableHeadProps<TData, TValue> & React.RefAttributes<HTMLTableCellElement>
) => React.ReactElement | null) & { displayName?: string };
TableHead.displayName = 'TableHead';

interface TableCellProps<TData, TValue> extends React.TdHTMLAttributes<HTMLTableCellElement> {
    cell?: Cell<TData, TValue>;
    // True when table level and column level drag and drop enabled.
    enableDragging?: boolean;
}

// Mounted only during a drag, when a DndContext is present. Plain tables do not
// require a DndContext.
const TableCellDragMonitor = ({ onMove }: { onMove: (delta: TableDragDelta) => void }) => {
    useDndMonitor({
        onDragMove: ({ delta }) => onMove(delta),
        onDragEnd: () => onMove({ x: 0, y: 0 }),
        onDragCancel: () => onMove({ x: 0, y: 0 }),
    });
    return null;
};

const TableCell = React.forwardRef(function TableCell<TData, TValue>(
    { className, cell, enableDragging = true, style: propsStyle, ...props }: TableCellProps<TData, TValue>,
    ref: React.ForwardedRef<HTMLTableCellElement>
) {
    const { active, over, activeNodeRect, droppableRects } = useDndContext();
    const sharedDragDeltas = React.useContext(TableDragDeltasContext);
    const columnId = cell?.column.id;
    const items = active && hasSortableData(active) ? active.data.current.sortable.items : [];
    const index = cell ? items.indexOf(cell.column.id) : -1;
    const activeIndex = active ? items.indexOf(active.id) : -1;
    const overIndex = over ? items.indexOf(over.id) : -1;
    const isDragging = activeIndex >= 0 && index === activeIndex;
    const isInDisplacementRange =
        (activeIndex < overIndex && index > activeIndex && index <= overIndex) ||
        (activeIndex > overIndex && index >= overIndex && index < activeIndex);
    const [dragDelta, setDragDelta] = React.useState<TableDragDelta>(
        () => (columnId ? sharedDragDeltas?.current.get(columnId) : undefined) ?? { x: 0, y: 0 }
    );
    // Only headers register sortable nodes. All cells in the active column follow its header.
    const transform =
        (enableDragging || isDragging || isInDisplacementRange) && index >= 0 && activeIndex >= 0 && overIndex >= 0
            ? isDragging
                ? { ...dragDelta, scaleX: 1, scaleY: 1 }
                : horizontalListSortingStrategy({
                      rects: items.map((id) => droppableRects.get(id)!),
                      activeNodeRect,
                      activeIndex,
                      overIndex,
                      index,
                  })
            : null;
    const cellStyle: React.CSSProperties = {
        opacity: isDragging ? 0.8 : 1,
        position: 'relative',
        transform: CSS.Translate.toString(transform),
    };
    const zIndex = isDragging ? 40 : cell?.column.getIsPinned() ? 30 : 1;
    const leftValue = isDragging && cell?.column.getIsPinned() ? 'initial' : '';
    return (
        <td
            className={cn(tableCellClasses, className, isDragging ? tableDraggingCellClasses : '')}
            {...props}
            style={{ ...cellStyle, ...propsStyle, zIndex: zIndex, left: leftValue }}>
            <div ref={ref}>{props.children}</div>
            {isDragging && <TableCellDragMonitor onMove={setDragDelta} />}
        </td>
    );
}) as (<TData, TValue>(
    props: TableCellProps<TData, TValue> & React.RefAttributes<HTMLTableCellElement>
) => React.ReactElement | null) & { displayName?: string };
TableCell.displayName = 'TableCell';

const TableCaption = React.forwardRef<HTMLTableCaptionElement, React.HTMLAttributes<HTMLTableCaptionElement>>(
    ({ className, ...props }, ref) => <caption ref={ref} className={cn(tableCaptionClasses, className)} {...props} />
);
TableCaption.displayName = 'TableCaption';

export { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow };
