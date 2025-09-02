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

import { Card, CardHeader, CardTitle, DataTable, Skeleton } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagHistoryRecord } from 'js-client-library';
import { DateTime } from 'luxon';
import React, { useRef, useState } from 'react';
import { useInfiniteQuery } from 'react-query';
import { AppIcon } from '../../../components';
import { SearchInput } from '../../../components/SearchInput';
import { useTagsQuery } from '../../../hooks';
import { apiClient } from '../../../utils';
import HistoryNotes from './HistoryNotes';
import { useHistoryTableContext } from './HistoryTableContext';

const BASE_COLUMNS = [
    {
        header: () => <div className='pl-8 text-left'>Name</div>,
        id: 'name',
    },
    {
        header: () => <div className='pl-8 text-left'>Action</div>,
        id: 'action',
    },
    {
        header: () => <div className='pl-8 text-left'>Date</div>,
        id: 'date',
    },
    {
        header: () => <div className='pl-8 text-left'>Tier/Label</div>,
        id: 'tier',
    },
    {
        header: () => <div className='pl-8 text-left'>Made by</div>,
        id: 'author',
    },
    {
        header: () => <div className='pl-8 text-center'>Notes</div>,
        id: 'notes',
    },
];

const LOADING_COLS = [
    () => <Skeleton className='h-4' />,
    () => <Skeleton className='h-4' />,
    () => <Skeleton className='h-4' />,
    () => <Skeleton className='h-4' />,
    () => <Skeleton className='h-4' />,
];

const NoteComponent = ({ row }: any) => {
    const { setCurrentNote, setShowNoteDetails } = useHistoryTableContext();
    const author = row.original.email;

    const handleOnClick = () => {
        const selectedNote = {
            note: row.original.note,
            createdBy: row.original.email,
            timestamp: row.original.date,
        };
        setShowNoteDetails(true);
        setCurrentNote(selectedNote);
    };

    return (
        <div className='text-center'>
            {author === 'System' ? (
                <p className='pl-4'>-</p>
            ) : (
                <button
                    className='disabled:opacity-25 pl-4'
                    onClick={() => handleOnClick()}
                    disabled={!row.original.note}>
                    <AppIcon.LinedPaper size={17} />
                </button>
            )}
        </div>
    );
};

const HISTORY_COLS = [
    ({ row }: any) => <div>{row.original.target}</div>,
    ({ row }: any) => <div>{row.original.action.replace(/([A-Z])/g, ' $1').trim()}</div>,
    ({ row }: any) => <div>{row.original.date}</div>,
    ({ row }: any) => <div>{row.original.tagName}</div>,
    ({ row }: any) => <div>{row.original.email || row.original.actor}</div>,
    ({ row }: any) => <NoteComponent row={row} />,
];

const PAGE_SIZE = 15;

/** Generates an array of column data in the success or loading states */
const getColumns = (isLoading: boolean) => {
    return BASE_COLUMNS.map((item, index) => ({
        ...item,
        cell: isLoading ? LOADING_COLS[index] : HISTORY_COLS[index],
    }));
};

const useAssetGroupTagHistoryQuery = (query?: string) => {
    const doSearch = query && query.length >= 3;
    const queryKey = doSearch ? query : 'static';

    return useInfiniteQuery<{
        count: number;
        data: { records: AssetGroupTagHistoryRecord[] };
        limit: number;
        skip: number;
    }>({
        queryKey: ['asset-group-tag-history', queryKey],
        queryFn: async ({ pageParam = 1 }) => {
            const skip = (pageParam - 1) * PAGE_SIZE;

            const result = doSearch
                ? await apiClient.searchAssetGroupTagHistory(PAGE_SIZE, skip, query)
                : await apiClient.getAssetGroupTagHistory(PAGE_SIZE, skip);

            return result.data;
        },
        getNextPageParam: (lastPage) => {
            const nextPage = lastPage.skip / PAGE_SIZE + 2;

            if ((nextPage - 1) * PAGE_SIZE >= lastPage.count) {
                return undefined;
            }

            return nextPage;
        },
        getPreviousPageParam: (firstPage) => {
            if (firstPage.skip === 0) {
                return undefined;
            }

            return firstPage.skip / PAGE_SIZE - 1;
        },
    });
};

const HistoryContent = () => {
    const [search, setSearch] = useState('');

    const {
        data: logHistory,
        isLoading: isHistoryLoading,
        isFetching: isHistoryFetching,
        isSuccess: isHistorySuccess,
        fetchNextPage,
    } = useAssetGroupTagHistoryQuery(search);
    const { data: tags, isLoading: isTagsLoading, isSuccess: isTagsSuccess } = useTagsQuery();

    const scrollRef = useRef<HTMLDivElement>(null);

    const historyData = logHistory ?? { pages: [{ count: 0, data: { records: [] } }] };
    const totalDBRowCount = historyData.pages[0].count;
    const historyItemsRaw = historyData.pages.flatMap((item) => item.data.records);
    const totalFetched = historyItemsRaw.length;

    const fetchMoreOnBottomReached = React.useCallback(
        (containerRefElement?: HTMLDivElement | null) => {
            if (containerRefElement) {
                const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
                //once the user has scrolled within 500px of the bottom of the table, fetch more data if we can
                if (
                    scrollHeight - scrollTop - clientHeight < 500 &&
                    !isHistoryFetching &&
                    totalFetched < totalDBRowCount
                ) {
                    fetchNextPage();
                }
            }
        },
        [fetchNextPage, isHistoryFetching, totalFetched, totalDBRowCount]
    );

    React.useEffect(() => {
        fetchMoreOnBottomReached(scrollRef.current);
    }, [fetchMoreOnBottomReached]);

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        count: totalFetched,
        getScrollElement: () => scrollRef.current,
        estimateSize: () => 55,
        measureElement:
            typeof window !== 'undefined' && navigator.userAgent.indexOf('Firefox') === -1
                ? (element) => element?.getBoundingClientRect().height
                : undefined,
        overscan: 5,
    };

    const isLoading = isHistoryLoading || isTagsLoading;
    const isSuccess = isHistorySuccess && isTagsSuccess;

    const historyItems = isSuccess
        ? historyItemsRaw.map((item) => {
              const tagName = tags?.find((tag) => tag.id === item.asset_group_tag_id)?.name;

              return {
                  ...item,
                  tagName,
                  date: DateTime.fromISO(item.created_at).toFormat('MM-dd-yyyy'),
              };
          })
        : [];

    type DataTableProps = React.ComponentProps<typeof DataTable>;

    const tableProps: DataTableProps['TableProps'] = {
        className: 'table-fixed',
        disableDefaultOverflowAuto: true,
    };

    const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
        className: 'sticky top-0 z-10 shadow-sm text-base',
    };

    const tableHeadProps: DataTableProps['TableHeadProps'] = {
        className: 'pr-2 text-center',
    };

    const tableCellProps: DataTableProps['TableCellProps'] = {
        className: 'truncate group relative pl-8',
    };

    return (
        <div id='history-wrapper' className={`flex gap-8 mt-6 grow`}>
            <Card className='grow' id='has-grid'>
                <CardHeader className='flex-row ml-3 justify-between items-center'>
                    <CardTitle>History Log</CardTitle>
                    <SearchInput value={search} onInputChange={setSearch} />
                </CardHeader>

                <div ref={scrollRef} className={`overflow-y-auto h-[calc(90vh_-_255px)] `}>
                    <DataTable
                        data={historyItems ?? []}
                        TableHeaderProps={tableHeaderProps}
                        TableHeadProps={tableHeadProps}
                        TableProps={tableProps}
                        TableCellProps={tableCellProps}
                        columns={getColumns(isLoading)}
                        virtualizationOptions={virtualizationOptions}
                    />
                </div>
            </Card>
            <HistoryNotes />
        </div>
    );
};

export default HistoryContent;
