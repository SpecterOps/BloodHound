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

import { Cell, Header, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table';
import React, { CSSProperties, useEffect, useMemo } from 'react';

// needed for table body level scope DnD setup

// needed for table body level scope DnD setup

// needed for row & cell level scope DnD setup
import { CSS } from '@dnd-kit/utilities';
// needed for table body level scope DnD setup
import {
    DndContext,
    DragEndEvent,
    KeyboardSensor,
    MouseSensor,
    TouchSensor,
    closestCenter,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import { restrictToHorizontalAxis } from '@dnd-kit/modifiers';
import { SortableContext, arrayMove, horizontalListSortingStrategy, useSortable } from '@dnd-kit/sortable';

const DraggableTableHeader = ({ header }: { header: Header<any, unknown> }) => {
    const { attributes, isDragging, listeners, setNodeRef, transform } = useSortable({
        id: header.column.id,
    });

    const style: CSSProperties = {
        opacity: isDragging ? 0.8 : 1,
        position: 'relative',
        transform: CSS.Translate.toString(transform), // translate instead of transform to avoid squishing
        transition: 'width transform 0.2s ease-in-out',
        whiteSpace: 'nowrap',
        width: header.column.getSize(),
        zIndex: isDragging ? 1 : 0,
    };

    return (
        <th colSpan={header.colSpan} ref={setNodeRef} style={style}>
            {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
            <button {...attributes} {...listeners}>
                🟰
            </button>
        </th>
    );
};

const DragAlongCell = ({ cell }: { cell: Cell<any, unknown> }) => {
    const { isDragging, setNodeRef, transform } = useSortable({
        id: cell.column.id,
    });

    const style: CSSProperties = {
        opacity: isDragging ? 0.8 : 1,
        position: 'relative',
        transform: CSS.Translate.toString(transform), // translate instead of transform to avoid squishing
        transition: 'width transform 0.2s ease-in-out',
        width: cell.column.getSize(),
        zIndex: isDragging ? 1 : 0,
    };

    return (
        <td style={style} ref={setNodeRef}>
            {flexRender(cell.column.columnDef.cell, cell.getContext())}
        </td>
    );
};
interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    items?: any;
}

const ExploreTable: React.FC<ExploreTableProps> = ({ items, open, onClose }) => {
    const mungedData = useMemo(
        () =>
            items &&
            Object.keys(items)
                .map((id) => ({ ...items[id]?.data, id }))
                .slice(0, 40),
        [items]
    );

    useEffect(
        () => () => {
            if (typeof onClose === 'function') onClose();
        },
        [onClose]
    );

    console.log({ mungedData });
    const columns = useMemo(
        () =>
            // If column order exists in redux/localStorage, use that
            Object.keys(mungedData?.[0])
                .slice(0, 10)
                .map((key: any) => {
                    return {
                        accessorKey: key,
                        cell: (info: any) => info.getValue(),
                        id: key,
                        size: 150,
                    };
                }),
        [mungedData]
    );

    const [columnOrder, setColumnOrder] = React.useState<string[]>(() => columns.map((c: any) => c.id!));

    const table = useReactTable({
        data: mungedData,
        columns,
        getCoreRowModel: getCoreRowModel(),
        state: {
            columnOrder,
        },
        onColumnOrderChange: setColumnOrder,
        debugTable: true,
        debugHeaders: true,
        debugColumns: true,
    });

    function handleDragEnd(event: DragEndEvent) {
        const { active, over } = event;
        if (active && over && active.id !== over.id) {
            setColumnOrder((columnOrder) => {
                const oldIndex = columnOrder.indexOf(active.id as string);
                const newIndex = columnOrder.indexOf(over.id as string);
                alert(
                    'Time to save column order\n' +
                        JSON.stringify(arrayMove(columnOrder, oldIndex, newIndex), () => {}, 2)
                );
                return arrayMove(columnOrder, oldIndex, newIndex); //this is just a splice util
            });
        }
    }

    const sensors = useSensors(useSensor(MouseSensor, {}), useSensor(TouchSensor, {}), useSensor(KeyboardSensor, {}));

    if (!open) return null;

    return (
        <div className='border-2 p-10 border-violet-700 absolute bottom-4 left-4 right-4 h-2/3 bg-pink-400 flex justify-center items-center'>
            <DndContext
                collisionDetection={closestCenter}
                modifiers={[restrictToHorizontalAxis]}
                onDragEnd={handleDragEnd}
                sensors={sensors}>
                <table className='block border-black relative border-2 bg-cyan-500 overflow-auto max-h-full'>
                    <thead className='grid sticky bg-yellow top-0 z-10'>
                        {table.getHeaderGroups().map((headerGroup) => (
                            <tr key={headerGroup.id}>
                                <SortableContext items={columnOrder} strategy={horizontalListSortingStrategy}>
                                    {headerGroup.headers.map((header) => (
                                        <DraggableTableHeader key={header.id} header={header} />
                                    ))}
                                </SortableContext>
                            </tr>
                        ))}
                    </thead>
                    <tbody>
                        {table.getRowModel().rows.map((row) => (
                            <tr key={row.id}>
                                {row.getVisibleCells().map((cell) => (
                                    <SortableContext
                                        key={cell.id}
                                        items={columnOrder}
                                        strategy={horizontalListSortingStrategy}>
                                        <DragAlongCell key={cell.id} cell={cell} />
                                    </SortableContext>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </DndContext>
        </div>
    );
};

export default ExploreTable;
