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
import { fireEvent, render, screen } from '@testing-library/react';
import { Pagination } from './Pagination';

describe('Pagination', () => {
    it.each([0, 3])('shows an empty result range and disables navigation with no rows on page %i', (page) => {
        const onPageChange = vi.fn();
        render(
            <Pagination
                page={page}
                rowsPerPage={25}
                count={0}
                onPageChange={onPageChange}
                onRowsPerPageChange={vi.fn()}
            />
        );

        const previousButton = screen.getByRole('button', { name: 'Go to previous page' });
        const nextButton = screen.getByRole('button', { name: 'Go to next page' });

        expect(screen.getByText('0 - 0 of 0')).toBeInTheDocument();
        expect(previousButton).toBeDisabled();
        expect(nextButton).toBeDisabled();

        fireEvent.click(previousButton);
        fireEvent.click(nextButton);
        expect(onPageChange).not.toHaveBeenCalled();
    });

    it.each([
        { name: 'first page', page: 0, count: 60, range: '1 - 25 of 60', previousDisabled: true, nextDisabled: false },
        {
            name: 'middle page',
            page: 1,
            count: 60,
            range: '26 - 50 of 60',
            previousDisabled: false,
            nextDisabled: false,
        },
        {
            name: 'partial last page',
            page: 2,
            count: 60,
            range: '51 - 60 of 60',
            previousDisabled: false,
            nextDisabled: true,
        },
        {
            name: 'full last page',
            page: 1,
            count: 50,
            range: '26 - 50 of 50',
            previousDisabled: false,
            nextDisabled: true,
        },
        { name: 'only page', page: 0, count: 10, range: '1 - 10 of 10', previousDisabled: true, nextDisabled: true },
    ])('preserves the range and navigation on the $name', ({ page, count, range, previousDisabled, nextDisabled }) => {
        const onPageChange = vi.fn();
        render(
            <Pagination
                page={page}
                rowsPerPage={25}
                count={count}
                onPageChange={onPageChange}
                onRowsPerPageChange={vi.fn()}
            />
        );

        const previousButton = screen.getByRole('button', { name: 'Go to previous page' });
        const nextButton = screen.getByRole('button', { name: 'Go to next page' });

        expect(screen.getByText(range)).toBeInTheDocument();
        expect(previousButton).toHaveProperty('disabled', previousDisabled);
        expect(nextButton).toHaveProperty('disabled', nextDisabled);

        fireEvent.click(previousButton);
        fireEvent.click(nextButton);
        expect(onPageChange.mock.calls).toEqual([
            ...(previousDisabled ? [] : [[page - 1]]),
            ...(nextDisabled ? [] : [[page + 1]]),
        ]);
    });
});
