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

import { Card, CardHeader, CardTitle, DataTable } from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { useCallback, useEffect, useRef, useState } from 'react';
import { SearchInput } from '../../../components/SearchInput';
import { useTagsQuery } from '../../../hooks';
import { LuxonFormat } from '../../../utils';
import { DEFAULT_FILTER_VALUE, FilterDialog, type AssetGroupTagHistoryFilters } from './FilterDialog';
import HistoryNote from './HistoryNote';
import { columns } from './columns';
import { useAssetGroupTagHistoryQuery } from './hooks';
import { DataTableProps, HistoryItem } from './types';
import { measureElement } from './utils';

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

    const {
        data: logHistory,
        isFetching: isHistoryFetching,
        isSuccess: isHistorySuccess,
        fetchNextPage,
    } = useAssetGroupTagHistoryQuery(filters, search);
    const { data: tags, isSuccess: isTagsSuccess } = useTagsQuery();

    const historyData = logHistory ?? emptyHistoryData;
    const totalDBRowCount = historyData.pages[0].count;
    const historyItemsRaw = historyData.pages.flatMap((item) => item.data.records);
    const totalFetched = historyItemsRaw.length;

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

    const isSuccess = isHistorySuccess && isTagsSuccess;

    const historyItems: HistoryItem[] = isSuccess
        ? historyItemsRaw.map((item) => {
              const tagName = tags?.find((tag) => tag.id === item.asset_group_tag_id)?.name;

              return {
                  ...item,
                  tagName,
                  date: DateTime.fromISO(item.created_at).toFormat(LuxonFormat.ISO_8601),
              };
          })
        : [];

    return (
        <div data-testid='history-wrapper' className='flex gap-8 mt-6 grow'>
            <Card className='grow'>
                <CardHeader className='flex-row ml-3 justify-between items-center'>
                    <CardTitle>History Log</CardTitle>
                    <div className='flex items-center '>
                        <SearchInput value={search} onInputChange={setSearch} />
                        <FilterDialog setFilters={setFilters} filters={filters} />
                    </div>
                </CardHeader>

                <div
                    onScroll={(e) => fetchMoreOnBottomReached(e.currentTarget)}
                    ref={scrollRef}
                    className={`overflow-y-auto h-[calc(90vh_-_255px)] `}>
                    <DataTable
                        data={historyItems}
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
    );
};

export default HistoryContent;
