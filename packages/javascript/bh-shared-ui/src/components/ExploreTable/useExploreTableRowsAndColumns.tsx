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
import { createColumnHelper, DataTable } from '@bloodhoundenterprise/doodleui';
import { faCancel, faCheck, faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button } from '@mui/material';
import { useCallback, useMemo, useState } from 'react';
import { EntityField, format } from '../../utils';
import NodeIcon from '../NodeIcon';
import { requiredColumns, type ExploreTableProps, type MungedTableRowWithId } from './explore-table-utils';
import ExploreTableHeaderCell from './ExploreTableHeaderCell';

const columnHelper = createColumnHelper<MungedTableRowWithId>();

type DataTableProps = React.ComponentProps<typeof DataTable>;

const filterKeys: (keyof MungedTableRowWithId)[] = ['displayname', 'objectid'];

type UseExploreTableRowsAndColumnsProps = Pick<
    ExploreTableProps,
    'onKebabMenuClick' | 'allColumnKeys' | 'selectedColumns' | 'data'
> & { searchInput: string };

const useExploreTableRowsAndColumns = ({
    onKebabMenuClick,
    searchInput,
    allColumnKeys,
    selectedColumns,
    data,
}: UseExploreTableRowsAndColumnsProps) => {
    const [sortBy, setSortBy] = useState<keyof MungedTableRowWithId>();
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>();

    const handleSort = useCallback(
        (sortByColumn: keyof MungedTableRowWithId) => {
            if (sortByColumn) {
                if (sortBy === sortByColumn) {
                    switch (sortOrder) {
                        case 'desc':
                            setSortOrder('asc');
                            break;
                        case 'asc':
                            setSortOrder('desc');
                            break;
                        default:
                            setSortOrder('desc');
                            break;
                    }
                } else {
                    setSortBy(sortByColumn);
                    setSortOrder('desc');
                }
            }
        },
        [sortBy, sortOrder]
    );

    const handleKebabMenuClick = useCallback(
        (e: React.MouseEvent, id: string) => {
            if (onKebabMenuClick) onKebabMenuClick({ x: e.clientX, y: e.clientY, id });
        },
        [onKebabMenuClick]
    );

    const makeColumnDef = useCallback(
        (key: keyof MungedTableRowWithId) =>
            columnHelper.accessor(String(key), {
                header: () => (
                    <ExploreTableHeaderCell
                        sortBy={sortBy}
                        sortOrder={sortOrder}
                        onClick={() => handleSort(key)}
                        headerKey={key}
                    />
                ),
                cell: (info) => {
                    const value = info.getValue() as EntityField['value'];

                    if (typeof value === 'boolean') {
                        return (
                            <div className='h-full w-full flex justify-center items-center text-center'>
                                <FontAwesomeIcon
                                    icon={value ? faCheck : faCancel}
                                    color={value ? 'green' : 'lightgray'}
                                    className='scale-125'
                                />
                            </div>
                        );
                    }

                    return format({ keyprop: key?.toString(), value, label: String(key) }) || '--';
                },
                id: key?.toString(),
            }),
        [handleSort, sortOrder, sortBy]
    );

    const requiredColumnDefinitions = useMemo(
        () => [
            columnHelper.accessor('', {
                id: 'action-menu',
                cell: ({ row }) => (
                    <Button
                        data-testid='kebab-menu'
                        onClick={(e) => handleKebabMenuClick(e, row?.original?.id)}
                        className='pl-4 pr-2 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0'>
                        <FontAwesomeIcon icon={faEllipsis} className='rotate-90 dark:text-neutral-light-1 text-black' />
                    </Button>
                ),
            }),
            columnHelper.accessor('nodetype', {
                id: 'nodetype',
                header: () => {
                    return <span className='dark:text-neutral-light-1'>Type</span>;
                },
                cell: (info) => {
                    return (
                        <div className='flex justify-center items-center relative'>
                            <NodeIcon nodeType={(info.getValue() as string) || ''} />
                        </div>
                    );
                },
            }),
            ...['objectid', 'displayname'].map(makeColumnDef),
        ],
        [handleKebabMenuClick, makeColumnDef]
    );

    const rows = useMemo(
        () =>
            ((data &&
                Object.entries(data).map(([key, value]) => ({
                    ...value.data,
                    id: key,
                    displayname: value?.label?.text,
                }))) ||
                []) as MungedTableRowWithId[],
        [data]
    );

    const filteredRows: MungedTableRowWithId[] = useMemo(() => {
        const lowercaseSearchInput = searchInput?.toLowerCase();

        return rows.filter((item) => {
            const filterTargets = filterKeys.map((filterKey) => {
                const stringyValue = String(item?.[filterKey]);

                return stringyValue?.toLowerCase();
            });

            return filterTargets.some((filterTarget) => filterTarget?.includes(lowercaseSearchInput));
        });
    }, [rows, searchInput]);

    const sortedFilteredRows = useMemo(() => {
        const dataToSort = filteredRows.slice();
        if (sortBy) {
            if (sortOrder === 'asc') {
                dataToSort.sort((a, b) => {
                    return a[sortBy]?.toString().localeCompare(b[sortBy]?.toString());
                });
            } else {
                dataToSort.sort((a, b) => {
                    return b[sortBy]?.toString().localeCompare(a[sortBy]?.toString());
                });
            }
        }

        return dataToSort;
    }, [filteredRows, sortBy, sortOrder]);

    const nonRequiredColumnDefinitions = useMemo(
        () => allColumnKeys?.filter((key) => !requiredColumns[key]).map(makeColumnDef) || [],
        [allColumnKeys, makeColumnDef]
    );

    const selectedColumnDefinitions = useMemo(
        () => nonRequiredColumnDefinitions.filter((columnDef) => selectedColumns?.[columnDef?.id || '']),
        [nonRequiredColumnDefinitions, selectedColumns]
    );

    const tableColumns = useMemo(
        () => [...requiredColumnDefinitions, ...selectedColumnDefinitions],
        [requiredColumnDefinitions, selectedColumnDefinitions]
    ) as DataTableProps['columns'];

    const columnOptionsForDropdown = useMemo(
        () => [...requiredColumnDefinitions, ...nonRequiredColumnDefinitions],
        [requiredColumnDefinitions, nonRequiredColumnDefinitions]
    );

    return {
        columnOptionsForDropdown,
        tableColumns,
        sortedFilteredRows,
    };
};

export default useExploreTableRowsAndColumns;
