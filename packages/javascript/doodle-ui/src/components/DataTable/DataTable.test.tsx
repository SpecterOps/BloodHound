// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { ColumnDef } from '@tanstack/react-table';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { DataTable } from './DataTable';

vi.mock('@tanstack/react-virtual', () => ({
    useVirtualizer: ({ count }: { count: number }) => ({
        getTotalSize: () => count * 57,
        getVirtualItems: () => Array.from({ length: count }, (_, index) => ({ index, size: 57, start: index * 57 })),
        scrollToIndex: vi.fn(),
    }),
}));

interface Person {
    name: string;
}

const columns: ColumnDef<Person>[] = [
    {
        accessorKey: 'name',
        header: 'Name',
    },
];

describe('DataTable', () => {
    it('keeps the default empty table content-sized', () => {
        render(<DataTable columns={columns} data={[]} />);

        const table = screen.getByRole('table');
        expect(screen.getByText('No results.')).toBeVisible();
        expect(table).not.toHaveClass('h-full', 'after:hidden');
        expect(table.parentElement?.parentElement?.style.height).toBe('var(--data-table-content-height)');
        expect(
            table.parentElement?.parentElement?.parentElement?.style.getPropertyValue('--data-table-content-height')
        ).toBe('0px');
    });

    it('restores virtualized sizing after a full-height empty table receives rows', () => {
        const { container, rerender } = render(<DataTable shouldFillHeight columns={columns} data={[]} />);

        const table = screen.getByRole('table');
        const heightContainer = table.parentElement?.parentElement;
        expect(container.firstChild).toHaveClass('overflow-auto');
        expect(table).toHaveClass('h-full', 'after:hidden');
        expect(heightContainer).toHaveStyle({ height: '100%' });
        expect(table.querySelector('thead')).toHaveClass('h-px');

        rerender(<DataTable shouldFillHeight columns={columns} data={[{ name: 'Ada Lovelace' }]} />);

        expect(screen.queryByText('No results.')).not.toBeInTheDocument();
        expect(screen.getByRole('row', { name: 'Ada Lovelace' })).toBeVisible();
        expect(table).not.toHaveClass('h-full', 'after:hidden');
        expect(heightContainer?.style.minHeight).toBe('var(--data-table-content-height)');
        expect(heightContainer?.style.height).toBe('');
        expect((container.firstChild as HTMLElement).style.getPropertyValue('--data-table-content-height')).toBe(
            '57px'
        );
        expect(table.querySelector('thead')).not.toHaveClass('h-px');
        expect(container.firstChild).toHaveClass('overflow-auto');

        rerender(<DataTable shouldFillHeight columns={columns} data={[]} />);

        expect(heightContainer).toHaveStyle({ height: '100%' });
        expect(heightContainer?.style.minHeight).toBe('');
        expect(table).toHaveClass('h-full', 'after:hidden');
        expect(screen.getByText('No results.')).toBeVisible();
    });

    it('preserves caller-owned scrolling while transitioning between populated and empty states', () => {
        const scrollElement = document.createElement('div');
        const virtualizationOptions = { getScrollElement: () => scrollElement };
        const { container, rerender } = render(
            <DataTable
                shouldFillHeight
                columns={columns}
                data={[{ name: 'Ada Lovelace' }]}
                virtualizationOptions={virtualizationOptions}
            />
        );

        expect(container.firstChild).toHaveClass('flex-auto');
        const heightContainer = screen.getByRole('table').parentElement?.parentElement;
        expect(heightContainer?.style.minHeight).toBe('var(--data-table-content-height)');
        expect(heightContainer?.style.height).toBe('');
        expect(container.firstChild).not.toHaveClass('overflow-auto');
        expect(container.firstChild).not.toHaveClass('min-h-0');

        rerender(
            <DataTable shouldFillHeight columns={columns} data={[]} virtualizationOptions={virtualizationOptions} />
        );

        expect(screen.getByRole('table')).toHaveClass('h-full', 'after:hidden');
        expect(container.firstChild).not.toHaveClass('overflow-auto');
        expect(container.firstChild).not.toHaveClass('min-h-0');
        expect(screen.getByText('No results.')).toBeVisible();
        expect(heightContainer).toHaveStyle({ height: '100%' });
        expect(heightContainer?.style.minHeight).toBe('');
    });

    it('uses outlined canvas styling for the table surface, header, and rows', () => {
        const { container } = render(<DataTable columns={columns} data={[{ name: 'Ada Lovelace' }]} />);

        expect(container.firstChild).toHaveClass(
            'border',
            'border-solid',
            'border-[color:var(--data-table-wrapper-border)]',
            'bg-neutral-1',
            'shadow-none'
        );

        expect(screen.getByRole('columnheader', { name: 'Name' })).toHaveClass('bg-neutral-2');

        expect(screen.getByRole('row', { name: 'Ada Lovelace' })).toHaveClass(
            'bg-neutral-1',
            'hover:bg-[color:var(--data-table-row-hover-fill)]'
        );

        expect(screen.getByRole('row', { name: 'Ada Lovelace' })).not.toHaveClass(
            'bg-[color:var(--data-table-row-even-fill)]',
            'bg-[color:var(--data-table-row-odd-fill)]'
        );
    });
});
