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

const columnHelper = createColumnHelper<any>();

// TODO: Populate name and version with data from query
export const columns = [
    columnHelper.accessor('name', {
        id: 'name',
        header: () => <span className='pl-6'>Name</span>,
        cell: (/*{ row }*/) => <span className='pl-6'>Name goes here</span>,
    }),
    columnHelper.accessor('version', {
        id: 'version',
        header: () => <span className=''>Version</span>,
        cell: (/*{ row }*/) => <span>v1.2.3</span>,
    }),
    columnHelper.accessor('delete', {
        id: 'delete-item',
        header: () => <span className='sr-only'>Delete</span>,
        cell: ({ row }) => (
            <button aria-label={`Delete ${row.original.name}`}>
                <Trash size={18} />
            </button>
        ),
        size: 0,
    }),
];

export const ActiveExtensionsCard = () => {
    const [search, setSearch] = useState('');

    // TODO: Replace with useQuery to fetch active extensions
    const data: { name: string; version: string }[] = [
        // { name: 'test', version: '1.0.0' }
    ];
    const hasData = data.length > 0;

    return (
        <Card className='flex flex-col gap-4 overflow-hidden'>
            <header className='flex justify-between pt-6 px-6 gap-3'>
                <CardTitle className='text-base'>Active Extensions</CardTitle>
                <SearchInput
                    className='self-start w-80'
                    disabled={!hasData}
                    id='search-active-extensions'
                    onInputChange={setSearch}
                    value={search}
                />
            </header>

            <div className='min-h-48'>
                <DataTable
                    data={data}
                    noResultsFallback={
                        <TableRow>
                            <TableCell colSpan={3} className='h-36 text-center'>
                                There are currently no active extensions
                            </TableCell>
                        </TableRow>
                    }
                    columns={columns}
                />
            </div>
        </Card>
    );
};
