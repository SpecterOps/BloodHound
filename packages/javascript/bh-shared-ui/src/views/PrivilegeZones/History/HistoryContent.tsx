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

import { Card, CardHeader, CardTitle, DataTable } from 'doodle-ui';
import { useState } from 'react';
import { SearchInput } from '../../../components/SearchInput';
import { useInfiniteScroll } from '../../../hooks';
import { measureElement } from '../utils';
import { FilterDialog } from './FilterDialog';
import HistoryNote from './HistoryNote';
import { columns } from './columns';
import { useAssetGroupTagHistoryQuery } from './hooks';
import { AssetGroupTagHistoryFilters, DataTableProps } from './types';
import { DEFAULT_FILTER_VALUE } from './utils';

const tableProps: DataTableProps['TableProps'] = {
    className: 'table-fixed',
    disableDefaultOverflowAuto: true,
};

const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
    className: 'sticky top-0 z-10 shadow-sm text-base',
};

const tableHeadProps: DataTableProps['TableHeadProps'] = {
    className: 'text-center',
};

const tableCellProps: DataTableProps['TableCellProps'] = {
    className: 'truncate group relative px-8 py-0',
};

const emptyHistoryData = { pages: [{ count: 0, data: { records: [] } }] };

const HistoryContent = () => {
    const [search, setSearch] = useState('');
    const [filters, setFilters] = useState<AssetGroupTagHistoryFilters>(DEFAULT_FILTER_VALUE);

    const query = useAssetGroupTagHistoryQuery(filters, search);

    const historyData = query.data ?? emptyHistoryData;
    const records = historyData.pages.flatMap((item) => item.data.records);

    const { scrollRef, onScroll } = useInfiniteScroll({
        canFetchMore: !!query.hasNextPage,
        isFetching: query.isFetchingNextPage,
        fetchMore: query.fetchNextPage,
        loadedCount: records.length,
    });

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        count: records.length ?? 0,
        getScrollElement: () => scrollRef.current,
        measureElement,
        overscan: 2,
        estimateSize: () => 50,
    };

    return (
        <>
            <p className='mt-6'>
                The History Log provides a record of changes to your Zones and Labels, including the type of change that
                occurred, who made it, and when it happened.
                <br />
                Use the log to audit and track changes to your Zones and Labels over time. Log items past 90 days are
                cleared.
            </p>
            <div data-testid='history-wrapper' className='flex gap-6 mt-4 h-[calc(100%-5rem)]'>
                <Card className='flex flex-col'>
                    <CardHeader className='flex-row ml-3 justify-between items-center'>
                        <CardTitle>History Log</CardTitle>
                        <div className='flex items-center'>
                            <SearchInput id='search-pz-history' value={search} onInputChange={setSearch} />
                            <FilterDialog setFilters={setFilters} filters={filters} />
                        </div>
                    </CardHeader>

                    <div
                        onScroll={onScroll}
                        ref={scrollRef}
                        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                        tabIndex={0}
                        className='overflow-y-auto mb-1 min-h-32'>
                        <DataTable
                            aria-label='History Log Table'
                            data={records}
                            TableHeaderProps={tableHeaderProps}
                            TableHeadProps={tableHeadProps}
                            TableProps={tableProps}
                            TableCellProps={tableCellProps}
                            columns={columns}
                            virtualizationOptions={virtualizationOptions}
                        />
                    </div>
                </Card>
                <div className='w-[400px] min-w-[400px] overflow-y-auto mb-1'>
                    <HistoryNote />
                </div>
            </div>
        </>
    );
};

export default HistoryContent;
