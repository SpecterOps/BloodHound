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

import { Card, CardHeader, CardTitle, DataTable } from '@bloodhoundenterprise/doodleui';
import { useCallback, useEffect, useRef, useState } from 'react';
import { SearchInput } from '../../../components/SearchInput';
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

    const { data, isFetching: isHistoryFetching, fetchNextPage } = useAssetGroupTagHistoryQuery(filters, search);

    const historyData = data ?? emptyHistoryData;
    const records = historyData.pages.flatMap((item) => item.data.records);
    const totalDBRowCount = historyData.pages[0].count;
    const totalFetched = records.length;

    const scrollRef = useRef<HTMLDivElement>(null);

    const fetchMoreOnBottomReached = useCallback(
        (containerRefElement?: HTMLDivElement | null) => {
            if (!containerRefElement) return;

            const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
            // once the user has scrolled near the bottom of the table, fetch more data if we can
            if (scrollHeight - scrollTop - clientHeight < 20 && !isHistoryFetching && totalFetched < totalDBRowCount) {
                fetchNextPage();
            }
        },
        [fetchNextPage, isHistoryFetching, totalFetched, totalDBRowCount]
    );

    useEffect(() => {
        fetchMoreOnBottomReached(scrollRef.current);
    }, [fetchMoreOnBottomReached]);

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        count: totalFetched ?? 0,
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
            <div data-testid='history-wrapper' className='flex gap-8 mt-6 grow'>
                <Card className='grow'>
                    <CardHeader className='flex-row ml-3 justify-between items-center'>
                        <CardTitle>History Log</CardTitle>
                        <div className='flex items-center '>
                            <SearchInput id='search-pz-history' value={search} onInputChange={setSearch} />
                            <FilterDialog setFilters={setFilters} filters={filters} />
                        </div>
                    </CardHeader>

                    <div
                        onScroll={(e) => fetchMoreOnBottomReached(e.currentTarget)}
                        ref={scrollRef}
                        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                        tabIndex={0}
                        className='overflow-y-auto h-[68dvh]'>
                        <DataTable
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
                <HistoryNote />
            </div>
        </>
    );
};

export default HistoryContent;
