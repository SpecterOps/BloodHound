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

import { Card, CardTitle, createColumnHelper, DataTable, TableCell, TableRow } from '@bloodhoundenterprise/doodleui';
import { Trash } from 'lucide-react';
import { useState } from 'react';
import { SearchInput } from '../../components';
import { useExtensionsQuery } from '../../hooks';

const columnHelper = createColumnHelper<any>();

export const columns = [
    columnHelper.accessor('name', {
        id: 'name',
        header: () => <span className='pl-6'>Name</span>,
        cell: ({ row }) => <span className='pl-6'>{row.original.name}</span>,
    }),
    columnHelper.accessor('version', {
        id: 'version',
        header: () => <span className=''>Version</span>,
        cell: ({ row }) => <span>{row.original.version}</span>,
    }),
    columnHelper.accessor('delete', {
        id: 'delete-item',
        // This is a hack to make the column header not visible but still accessible for screen readers
        // Also keeps last column consistent width when no rows are rendered
        header: () => <span className='opacity-0'>Delete</span>,
        cell: ({ row }) => (
            <button aria-label={`Delete ${row.original.name}`}>
                <Trash size={18} />
            </button>
        ),
        size: 0,
    }),
];

export const ERROR_MESSAGE = 'There was an error fetching extensions';
export const LOADING_MESSAGE = 'Loading extensions...';
export const NO_DATA_MESSAGE = 'There are currently no active extensions';
export const NO_SEARCH_RESULTS_MESSAGE = 'No extensions match your search terms';

const TABLE_CELL_HEIGHT = 57;
const TABLE_HEADER_HEIGHT = 52;
const EMPTY_STATE_HEIGHT = `${TABLE_HEADER_HEIGHT + TABLE_CELL_HEIGHT * 2}px`;

export const ActiveExtensionsCard = () => {
    const [search, setSearch] = useState('');
    const { data = [], isError, isLoading, isSuccess } = useExtensionsQuery();

    const hasData = !isLoading && isSuccess && data.length > 0;
    const filteredData = data.filter((extension) => extension.name.toLowerCase().includes(search.toLowerCase()));
    const isEmptySearch = hasData && filteredData.length === 0;

    let fallbackMessage = LOADING_MESSAGE;

    if (isError) {
        fallbackMessage = ERROR_MESSAGE;
    } else if (isSuccess && !hasData) {
        fallbackMessage = NO_DATA_MESSAGE;
    } else if (isEmptySearch) {
        fallbackMessage = NO_SEARCH_RESULTS_MESSAGE;
    }

    return (
        <Card className='flex flex-col gap-4 overflow-hidden'>
            <header className='flex justify-between pt-6 px-6 gap-3'>
                <CardTitle className='text-base'>Active Extensions</CardTitle>
                <SearchInput
                    className='self-start w-80'
                    id='search-active-extensions'
                    onInputChange={setSearch}
                    value={search}
                />
            </header>

            <div
                // DataTable currently has some issues with table and cell height within a Card element
                // Tailwind doesn't have a way to calculate dynamic heights, so inline styles are used
                style={{
                    minHeight:
                        !hasData || isEmptySearch
                            ? EMPTY_STATE_HEIGHT
                            : `${TABLE_HEADER_HEIGHT + TABLE_CELL_HEIGHT * filteredData.length}px`,
                }}>
                <DataTable
                    data={filteredData}
                    noResultsFallback={
                        <TableRow>
                            <TableCell colSpan={3} className='h-28 text-center'>
                                {fallbackMessage}
                            </TableCell>
                        </TableRow>
                    }
                    columns={columns}
                />
            </div>
        </Card>
    );
};
