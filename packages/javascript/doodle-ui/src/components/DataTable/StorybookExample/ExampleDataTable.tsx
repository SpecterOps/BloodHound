import React, { useState } from 'react';
import { DataTable } from '../DataTable';
import { Pagination } from 'components/Pagination';
import { getColumns, getData } from './utils';

const DATA = getData(50);

const ExampleDataTable: React.FC = () => {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);
    const [filter, setFilter] = useState<string | null>(null);
    const [sortBy, setSortBy] = useState<keyof ReturnType<typeof getData>[0] | null>(null);
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc' | null>(null);

    const handleRowsPerPageChange = (rowsPerPage: number) => {
        setPage(0);
        setRowsPerPage(rowsPerPage);
    };

    const handleSort = (sortByColumn: string) => {
        if (sortBy === sortByColumn) {
            switch (sortOrder) {
                case null:
                    setSortOrder('desc');
                    break;
                case 'desc':
                    setSortOrder('asc');
                    break;
                case 'asc':
                    setSortBy(null);
                    setSortOrder(null);
                    break;
            }
        } else {
            setSortBy(sortByColumn as keyof ReturnType<typeof getData>[0]);
            setSortOrder('desc');
        }
    };

    const columns = getColumns(sortOrder, handleSort);
    let data = DATA;

    if (filter) {
        data = data.filter((d) => d?.nonTierZeroPrincipal?.toLowerCase()?.includes(filter.toLowerCase()));
    }

    if (sortBy) {
        if (sortOrder === 'asc') {
            data = data.slice().sort((a, b) => {
                return a[sortBy] < b[sortBy] ? 1 : -1;
            });
        } else {
            data = data.slice().sort((a, b) => {
                return a[sortBy] < b[sortBy] ? -1 : 1;
            });
        }
    }

    const filteredDataCount = data.length;

    data = data.slice(page * rowsPerPage, (page + 1) * rowsPerPage);

    return (
        <>
            <div className='flex p-6 items-center'>
                <h2 className='mr-auto text-3xl'>Attack Paths</h2>
                <input
                    placeholder='Search Non Tier Zero Principal'
                    className='border-b-2 border-black pl-3 w-1/4'
                    id='test-search'
                    value={filter ?? ''}
                    onChange={(e) => {
                        setFilter(e.target.value);
                        setPage(0);
                    }}
                />
            </div>
            <DataTable
                TableHeadProps={{ className: 'font-bold text-base' }}
                TableBodyProps={{ className: 'text-xs font-roboto' }}
                onRowClick={console.log}
                selectedRow={data[0].id}
                columns={columns}
                data={data}
            />
            <Pagination
                count={filteredDataCount}
                rowsPerPage={rowsPerPage}
                page={page}
                onPageChange={setPage}
                onRowsPerPageChange={handleRowsPerPageChange}
                className='justify-start'
            />
        </>
    );
};

export default ExampleDataTable;
