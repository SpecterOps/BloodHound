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

import { DndContext, KeyboardSensor, useDndContext, useSensor, useSensors } from '@dnd-kit/core';
import { horizontalListSortingStrategy, SortableContext } from '@dnd-kit/sortable';
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './Table';

const data = [{ name: 'First row' }, { name: 'Second row' }];
const columnHelper = createColumnHelper<(typeof data)[number]>();
const columns = [columnHelper.accessor('name', { header: 'Name' })];
const twoColumnData = [
    { first: 'First A', second: 'First B' },
    { first: 'Second A', second: 'Second B' },
];
const twoColumnHelper = createColumnHelper<(typeof twoColumnData)[number]>();
const twoColumns = [
    twoColumnHelper.accessor('first', { header: 'A' }),
    twoColumnHelper.accessor('second', { header: 'B' }),
];

const ContextObserver = ({ observe }: { observe: (context: ReturnType<typeof useDndContext>) => void }) => {
    observe(useDndContext());
    return null;
};

const DraggableTable = ({ observe }: { observe: (context: ReturnType<typeof useDndContext>) => void }) => {
    const table = useReactTable({ data, columns, getCoreRowModel: getCoreRowModel() });
    const sensors = useSensors(useSensor(KeyboardSensor));

    return (
        <DndContext sensors={sensors}>
            <ContextObserver observe={observe} />
            <SortableContext items={['name']} strategy={horizontalListSortingStrategy}>
                <Table>
                    <TableHeader>
                        <TableRow>
                            {table.getFlatHeaders().map((header) => (
                                <TableHead key={header.id} header={header}>
                                    {flexRender(header.column.columnDef.header, header.getContext())}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows.map((row) => (
                            <TableRow key={row.id}>
                                {row.getVisibleCells().map((cell) => (
                                    <TableCell key={cell.id} cell={cell}>
                                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                    </TableCell>
                                ))}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </SortableContext>
        </DndContext>
    );
};

const TableWithDisabledCell = ({
    data = twoColumnData,
    observe,
}: {
    data?: typeof twoColumnData;
    observe: (context: ReturnType<typeof useDndContext>) => void;
}) => {
    const table = useReactTable({ data, columns: twoColumns, getCoreRowModel: getCoreRowModel() });

    return (
        <DndContext sensors={useSensors(useSensor(KeyboardSensor))}>
            <ContextObserver observe={observe} />
            <SortableContext items={['first', 'second']} strategy={horizontalListSortingStrategy}>
                <Table>
                    <TableHeader>
                        <TableRow>
                            {table.getFlatHeaders().map((header) => (
                                <TableHead
                                    key={header.id}
                                    header={header}
                                    enableDragging={header.column.id !== 'second'}>
                                    {flexRender(header.column.columnDef.header, header.getContext())}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows.map((row) => (
                            <TableRow key={row.id}>
                                {row.getVisibleCells().map((cell) => (
                                    <TableCell key={cell.id} cell={cell} enableDragging={cell.column.id !== 'second'}>
                                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                    </TableCell>
                                ))}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </SortableContext>
        </DndContext>
    );
};

describe('Table column dragging', () => {
    afterEach(() => vi.restoreAllMocks());

    it('keeps the header registered as the active node when keyboard dragging with body rows', async () => {
        vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 0, 200, 40));
        const observe = vi.fn();
        const user = userEvent.setup();
        render(<DraggableTable observe={observe} />);

        const header = screen.getByRole('columnheader', { name: /Name/ });
        const context = () => observe.mock.lastCall![0] as ReturnType<typeof useDndContext>;

        expect(context().draggableNodes.get('name')?.node.current).toBe(header);
        expect(context().droppableContainers.get('name')?.node.current).toBe(header);

        await user.tab();
        expect(screen.getByRole('button', { name: 'Drag to reorder' })).toHaveFocus();
        await user.keyboard('[Space]');

        await waitFor(() => expect(context().activeNode).toBe(header));
        expect(context().active?.id).toBe('name');
        fireEvent.keyDown(document, { code: 'ArrowRight' });
        await waitFor(() => {
            expect(screen.getByRole('cell', { name: 'First row' }).style.transform).toBe(header.style.transform);
            expect(header.style.transform).toContain('25px');
        });
        fireEvent.keyDown(document, { code: 'Escape' });
        await waitFor(() => expect(context().active).toBeNull());
    });

    it('keeps disabled and newly mounted cells aligned while dragging a column', async () => {
        vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
            return new DOMRect(this.closest('th')?.textContent?.includes('B') ? 200 : 0, 0, 200, 40);
        });
        const observe = vi.fn();
        const user = userEvent.setup();
        const { rerender } = render(<TableWithDisabledCell observe={observe} />);
        const context = () => observe.mock.lastCall![0] as ReturnType<typeof useDndContext>;

        await user.tab();
        await user.keyboard('[Space]');
        for (let index = 0; index < 5; index++) {
            fireEvent.keyDown(document, { code: 'ArrowRight' });
        }

        await waitFor(() => {
            const header = screen.getByRole('columnheader', { name: 'B' });
            const cell = screen.getByRole('cell', { name: 'First B' });
            expect(context().over?.id).toBe('second');
            expect(header.style.transform).not.toBe('');
            expect(cell.style.transform).toBe(header.style.transform);
        });

        const newRow = { first: 'Third A', second: 'Third B' };
        rerender(<TableWithDisabledCell data={[...twoColumnData, newRow]} observe={observe} />);

        await waitFor(() => {
            const header = screen.getByRole('columnheader', { name: /A/ });
            const newCell = screen.getByRole('cell', { name: 'Third A' });
            expect(newCell.style.transform).toBe(header.style.transform);
        });
    });

    it('renders plain cells without a drag context', () => {
        render(
            <Table>
                <TableBody>
                    <TableRow>
                        <TableCell>Plain cell</TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        );

        expect(screen.getByRole('cell', { name: 'Plain cell' })).toBeVisible();
    });
});
