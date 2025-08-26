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
import { AssetGroupTagsHistory } from 'js-client-library';
import { DateTime } from 'luxon';
import { useRef, useState } from 'react';
import { useQuery } from 'react-query';
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
    ({ row }: any) => <div className='text-primary'>{row.original.target}</div>,
    ({ row }: any) => <div>{row.original.action.replace(/([A-Z])/g, ' $1').trim()}</div>,
    ({ row }: any) => <div>{row.original.date}</div>,
    ({ row }: any) => <div>{row.original.tagName}</div>,
    ({ row }: any) => <div>{row.original.email || row.original.actor}</div>,
    ({ row }: any) => <NoteComponent row={row} />,
];

// The height of the tabs and page description; where history log table starts
const TABLE_Y_OFFSET = '255px';

const TABLE_HEIGHT_OFFSET = '326px';

const QUERY_LIMIT = 1000;
const QUERY_SKIP = 0;

/** Generates an array of column data in the success or loading states */
const getColumns = (isLoading: boolean) => {
    return BASE_COLUMNS.map((item, index) => ({
        ...item,
        cell: isLoading ? LOADING_COLS[index] : HISTORY_COLS[index],
    }));
};

const useAssetGroupTagHistoryQuery = (query: string) => {
    const doSearch = query.length >= 3;
    const queryKey = doSearch ? query : 'static';

    return useQuery<AssetGroupTagsHistory>({
        queryKey: ['asset-group-tag-history', queryKey],
        queryFn: async () => {
            const result = doSearch
                ? await apiClient.searchAssetGroupTagHistory(QUERY_LIMIT, QUERY_SKIP, query)
                : await apiClient.getAssetGroupTagHistory(QUERY_LIMIT, QUERY_SKIP);
            return result.data;
        },
    });
};

const HistoryContent = () => {
    const [search, setSearch] = useState('');

    const {
        data: history,
        isLoading: isHistoryLoading,
        isSuccess: isHistorySuccess,
    } = useAssetGroupTagHistoryQuery(search);
    const { data: tags, isLoading: isTagsLoading, isSuccess: isTagsSuccess } = useTagsQuery();

    const scrollRef = useRef<HTMLDivElement>(null);

    const isLoading = isHistoryLoading || isTagsLoading;
    const isSuccess = isHistorySuccess && isTagsSuccess;

    const historyItems = isSuccess
        ? history?.data.records.map((item) => {
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
        <div id='history-wrapper' className={`flex gap-8 mt-6 h-[calc(100vh_-_${TABLE_Y_OFFSET})]`}>
            <Card id='has-grid'>
                <CardHeader className='flex-row ml-3 justify-between items-center'>
                    <CardTitle>History Log</CardTitle>
                    <SearchInput value={search} onInputChange={setSearch} />
                </CardHeader>

                <div ref={scrollRef} className={`overflow-y-auto h-[calc(100vh_-_${TABLE_HEIGHT_OFFSET})]`}>
                    <DataTable
                        data={historyItems ?? []}
                        TableHeaderProps={tableHeaderProps}
                        TableHeadProps={tableHeadProps}
                        TableProps={tableProps}
                        TableCellProps={tableCellProps}
                        columns={getColumns(isLoading)}
                        virtualizationOptions={{
                            rangeExtractor: (range) => {
                                return new Array(range.count).fill(0).map((_, index) => {
                                    return range.startIndex + index;
                                });
                            },
                            estimateSize: () => 17,
                        }}
                    />
                </div>
            </Card>
            <HistoryNotes />
        </div>
    );
};

export default HistoryContent;
