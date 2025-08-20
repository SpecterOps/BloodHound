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
import { useState } from 'react';
import { useInfiniteQuery } from 'react-query';
import { AppIcon } from '../../../components';
import { SearchInput } from '../../../components/SearchInput';
import { useTagsQuery } from '../../../hooks';
import { apiClient } from '../../../utils';
import { PageParam, createPaginatedFetcher } from '../../../utils/paginatedFetcher';
import HistoryNotes from './HistoryNotes';
import { useHistoryTableContext } from './HistoryTableContext';

const BASE_COLUMNS = [
    {
        header: () => <div className='font-semibold text-base pl-4'> Name </div>,
        id: 'name',
    },
    {
        header: () => <div className='font-semibold text-base pl-4'>Action</div>,
        id: 'action',
    },
    {
        header: () => <div className='font-semibold text-base pl-4'>Date</div>,
        id: 'date',
    },
    {
        header: () => <div className='font-semibold text-base pl-4'>Tier/Label</div>,
        id: 'tier', //question here since I checked the API and I do not see in the response something like label/tier maybe target?
    },
    {
        header: () => <div className='font-semibold text-base pl-4'>Made by</div>,
        id: 'author',
    },
    {
        header: () => <div className='font-semibold text-base pl-4'>Notes</div>,
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
        <div>
            {author === 'System' ? (
                <p className='pl-4'>-</p>
            ) : (
                <button
                    className='disabled:opacity-25 pl-4'
                    onClick={() => handleOnClick()}
                    disabled={!row.original.note}>
                    <AppIcon.LinedPaper size={18} />
                </button>
            )}
        </div>
    );
};

const HISTORY_COLS = [
    ({ row }: any) => <div className='pl-4'>{row.original.target}</div>,
    ({ row }: any) => <div className='pl-4'>{row.original.action}</div>,
    ({ row }: any) => <div className='pl-4'>{row.original.date}</div>,
    ({ row }: any) => <div className='pl-4'>{row.original.tagName}</div>,
    ({ row }: any) => <div className='pl-4'>{row.original.email || row.original.actor}</div>,
    ({ row }: any) => <NoteComponent row={row} />,
];

const PAGE_SIZE = 25;

/** Generates an array of column data in the success or loading states */
const getColumns = (isLoading: boolean) => {
    return BASE_COLUMNS.map((item, index) => ({
        ...item,
        cell: isLoading ? LOADING_COLS[index] : HISTORY_COLS[index],
    }));
};

const useInfiniteQueriesPage = () => {
    const getPaginatedHistory = (skip: number = 0, limit: number = PAGE_SIZE) =>
        createPaginatedFetcher(() => apiClient.getAssetGroupTagHistory(limit, skip), 'records', skip, limit);

    return useInfiniteQuery<{
        items: AssetGroupTagHistoryRecord[];
        nextPageParam?: PageParam;
    }>({
        queryKey: ['asset-group-tag-history'],
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getPaginatedHistory(pageParam.skip, pageParam.limit),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
    });
};

const HistoryContent = () => {
    const [search, setSearch] = useState('');

    const { data, isLoading } = useInfiniteQueriesPage();

    const tagsQuery = useTagsQuery();

    if (isLoading || tagsQuery.isLoading) return null;

    const tableData = data?.pages[0].items || [];

    tableData.forEach((record) => {
        const name = tagsQuery.data?.find((tag) => tag.id === record.asset_group_tag_id)?.name;
        const formattedDate = DateTime.fromISO(record.created_at).toFormat('dd/MM/yyyy');
        record.date = formattedDate;

        if (name !== undefined) {
            record.tagName = name;
        }
    });

    return (
        <div className='flex gap-8 mt-6 w-full '>
            <Card className='  min-h-[200px] '>
                <CardHeader className='flex-row ml-3 justify-between items-center'>
                    <CardTitle>History Log</CardTitle>
                    <SearchInput value={search} onInputChange={setSearch} />
                </CardHeader>

                <DataTable className='pl-2 pr-2 ' data={tableData ?? []} columns={getColumns(isLoading)} />
            </Card>
            <HistoryNotes />
        </div>
    );
};

export default HistoryContent;
