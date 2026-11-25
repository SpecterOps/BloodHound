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
import {
    closestCenter,
    DndContext,
    type DragEndEvent,
    KeyboardSensor,
    MouseSensor,
    TouchSensor,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import { restrictToHorizontalAxis } from '@dnd-kit/modifiers';
import {
    arrayMove,
    horizontalListSortingStrategy,
    SortableContext,
    sortableKeyboardCoordinates,
} from '@dnd-kit/sortable';
import { faker } from '@faker-js/faker';
import type { Meta, StoryObj } from '@storybook/react';
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table';
import { useState } from 'react';
import { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow } from './Table';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Table',
    component: Table,
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component: `
Use Table to display data in rows and columns so users can scan records and compare values. Use column headers to label each field, a caption to describe the data, and an optional footer for totals or summaries.

When users need to change the column order, add drag handles to the headers. The Enable Dragging example shows how to reorder columns with a mouse, touch, or keyboard.

### **[View the Table design in Figma](https://www.figma.com/design/pqlfSCxXVB5MjX5Gq50rW8/Alpha-%F0%9F%90%95?node-id=1155-189143&m=dev)
`,
            },
        },
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof Table>;

export default meta;
type Story = StoryObj<typeof meta>;

const invoices = [
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'incomplete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'incomplete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'incomplete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
];

export const Default: Story = {
    render: () => (
        <Table className='w-full border-b-[length:--table-row-border-width] [&_th]:px-4 [&_td]:pl-4'>
            <TableCaption>A list of your recent TODOs.</TableCaption>
            <TableHeader>
                <TableRow>
                    <TableHead className='w-48' enableDragging={false}>
                        Todo
                    </TableHead>
                    <TableHead className='w-48' enableDragging={false}>
                        Description
                    </TableHead>
                    <TableHead className='w-32' enableDragging={false}>
                        Status
                    </TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {invoices.map((invoice) => (
                    <TableRow key={invoice.id}>
                        <TableCell className='font-medium'>{invoice.todo}</TableCell>
                        <TableCell>{invoice.description}</TableCell>
                        <TableCell>{invoice.status}</TableCell>
                    </TableRow>
                ))}
            </TableBody>
            <TableFooter>
                <TableRow>
                    <TableCell colSpan={2}>Completed</TableCell>
                    <TableCell className='text-right'>
                        {invoices.filter((s) => s.status === 'complete').length}
                    </TableCell>
                </TableRow>
            </TableFooter>
        </Table>
    ),
};

const columnHelper = createColumnHelper<(typeof invoices)[number]>();
const draggableColumns = [
    columnHelper.accessor('todo', { header: 'Todo' }),
    columnHelper.accessor('description', { header: 'Description' }),
    columnHelper.accessor('status', { header: 'Status' }),
];

const DraggableTableExample = () => {
    const [columnOrder, setColumnOrder] = useState(['todo', 'description', 'status']);
    const sensors = useSensors(
        useSensor(MouseSensor, { activationConstraint: { distance: 4 } }),
        useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 5 } }),
        useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
    );
    const table = useReactTable({
        data: invoices,
        columns: draggableColumns,
        state: { columnOrder },
        onColumnOrderChange: setColumnOrder,
        getCoreRowModel: getCoreRowModel(),
    });

    const handleDragEnd = ({ active, over }: DragEndEvent) => {
        if (!over || active.id === over.id) return;

        setColumnOrder((order) => {
            const oldIndex = order.indexOf(String(active.id));
            const newIndex = order.indexOf(String(over.id));

            return oldIndex < 0 || newIndex < 0 ? order : arrayMove(order, oldIndex, newIndex);
        });
    };

    return (
        <DndContext
            collisionDetection={closestCenter}
            modifiers={[restrictToHorizontalAxis]}
            sensors={sensors}
            onDragEnd={handleDragEnd}>
            <Table className='w-full border-b-[length:var(--table-row-border-width)] [&_th]:px-4 [&_td]:pl-4'>
                <TableCaption>Drag a column handle to reorder the headers and their cells.</TableCaption>
                <TableHeader>
                    {table.getHeaderGroups().map((headerGroup) => (
                        <TableRow key={headerGroup.id}>
                            <SortableContext items={columnOrder} strategy={horizontalListSortingStrategy}>
                                {headerGroup.headers.map((header) => (
                                    <TableHead key={header.id} header={header} enableDragging scope='col'>
                                        {flexRender(header.column.columnDef.header, header.getContext())}
                                    </TableHead>
                                ))}
                            </SortableContext>
                        </TableRow>
                    ))}
                </TableHeader>
                <TableBody>
                    {table.getRowModel().rows.map((row) => (
                        <TableRow key={row.id}>
                            <SortableContext items={columnOrder} strategy={horizontalListSortingStrategy}>
                                {row.getVisibleCells().map((cell) => (
                                    <TableCell key={cell.id} cell={cell} enableDragging>
                                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                    </TableCell>
                                ))}
                            </SortableContext>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </DndContext>
    );
};

export const EnableDragging: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Enables column drag handles with sortable headers and cells. Drag a handle with the mouse or touch, or focus it and press Space, Left/Right, then Space to drop. Escape cancels the drag.',
            },
        },
    },
    render: () => <DraggableTableExample />,
};
