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

import userEvent from '@testing-library/user-event';
import { render } from '../../test-utils';
import { BaseColumnHeader, SortableHeader } from './ColumnHeaders';

describe('ColumnHeaders', () => {
    describe('BaseColumnHeader', () => {
        it('renders the title passed via title prop', () => {
            const expected = 'Test Column Header';
            const screen = render(<BaseColumnHeader title={expected} textAlign='left' />);

            expect(screen.getByText(expected)).toBeInTheDocument();
        });
        it('adds expected text alignment class depending on the textAlign prop', () => {
            const screen = render(<BaseColumnHeader title={'test'} textAlign='left' />);

            expect(screen.getByText('test')).toHaveClass('text-left');
        });
    });
    describe('SortableHeader', () => {
        it('renders the title passed via title prop', () => {
            const expected = 'Test Column Header';
            const screen = render(<SortableHeader title={expected} sortOrder='asc' onSort={vi.fn} />);

            expect(screen.getByText(expected)).toBeInTheDocument();
        });
        it('calls onSort when clicked', async () => {
            const mockOnSort = vi.fn();
            const screen = render(<SortableHeader title={'test'} onSort={mockOnSort} />);

            const user = userEvent.setup();
            await user.click(screen.getByText('app-icon-sort-empty'));

            expect(mockOnSort).toHaveBeenCalledTimes(1);
        });
        it('displays CaretDownOutline icon when the sortOrder is undefined', () => {
            const screen = render(<SortableHeader title={'test'} onSort={vi.fn} />);
            expect(screen.getByText('app-icon-sort-empty')).toBeInTheDocument();
        });
        it('displays CaretDown icon when the sortOrder is desc', () => {
            const screen = render(<SortableHeader title={'test'} sortOrder='desc' onSort={vi.fn} />);
            expect(screen.getByText('app-icon-sort-desc')).toBeInTheDocument();
        });
        it('displays CaretUp icon when the sortOrder is asc', () => {
            const screen = render(<SortableHeader title={'test'} sortOrder='asc' onSort={vi.fn} />);
            expect(screen.getByText('app-icon-sort-asc')).toBeInTheDocument();
        });
        it('does not call onSort when disable=true', async () => {
            const mockOnSort = vi.fn();
            const screen = render(<SortableHeader title={'test'} sortOrder={undefined} onSort={mockOnSort} disable />);

            const header = screen.getByText('app-icon-sort-empty');
            expect(header.parentElement?.parentElement?.className.includes('pointer-events-none')).toBeTruthy();
        });
    });
});
