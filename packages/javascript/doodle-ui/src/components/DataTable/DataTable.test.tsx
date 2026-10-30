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
import '@testing-library/jest-dom';
import matchers from '@testing-library/jest-dom/matchers';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { DataTable } from './DataTable';

expect.extend(matchers);

vi.mock('@tanstack/react-virtual', () => ({
    useVirtualizer: () => ({
        getTotalSize: () => 57,
        getVirtualItems: () => [{ index: 0, size: 57, start: 0 }],
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
    it('uses outlined canvas styling for the table surface, header, and rows', () => {
        const { container } = render(<DataTable columns={columns} data={[{ name: 'Ada Lovelace' }]} />);

        expect(container.firstChild).toHaveClass(
            'border',
            'border-solid',
            'border-neutral-light-4',
            'dark:border-neutral-900',
            'bg-neutral-1',
            'shadow-none'
        );

        expect(screen.getByRole('columnheader', { name: 'Name' })).toHaveClass('bg-neutral-2');

        expect(screen.getByRole('row', { name: 'Ada Lovelace' })).toHaveClass(
            'bg-neutral-1',
            'hover:bg-data-table-row-hover-fill'
        );

        expect(screen.getByRole('row', { name: 'Ada Lovelace' })).not.toHaveClass(
            'bg-data-table-row-even-fill',
            'bg-data-table-row-odd-fill'
        );
    });
});
