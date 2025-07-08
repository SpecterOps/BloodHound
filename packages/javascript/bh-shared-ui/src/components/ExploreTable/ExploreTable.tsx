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

import { Button, DataTable, createColumnHelper } from '@bloodhoundenterprise/doodleui';
import { faCancel, faCaretDown, faCaretUp, faCheck, faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { GraphNode } from 'js-client-library';
import { ChangeEvent, MouseEvent, memo, useCallback, useMemo, useState } from 'react';
import { useToggle } from '../../hooks';
import { WrappedExploreTableItem } from '../../types';
import { EntityField, cn, format, formatPotentiallyUnknownLabel } from '../../utils';
import NodeIcon from '../NodeIcon';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';
import TableControls from './TableControls';

const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = ['nodetype', 'objectid', 'displayname'];

const requiredColumns = REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.reduce(
    (acc, curr) => ({ ...acc, [curr]: true }),
    {}
) as Record<string, boolean>;

type MungedTableRowWithId = WrappedExploreTableItem['data'] & { id: string };

const columnHelper = createColumnHelper<MungedTableRowWithId>();

interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    data?: Record<string, WrappedExploreTableItem>;
    selectedNode: string;
    selectedColumns?: Record<string, boolean>;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    onDownloadClick: () => void;
    onKebabMenuClick: (clickInfo: NodeClickInfo) => void;
}

export type NodeClickInfo = { id: string; x: number; y: number };

const MemoDataTable = memo(DataTable);

const HeaderCell = ({
    headerKey,
    sortBy,
    sortOrder,
    onClick,
}: {
    headerKey: string;
    sortBy?: string;
    sortOrder?: string;
    onClick: () => void;
}) => {
    return (
        <div className='flex items-center p-1 m-0 cursor-pointer h-full hover:bg-neutral-100' onClick={onClick}>
            <div>{formatPotentiallyUnknownLabel(headerKey)}</div>
            <div className={cn('pl-2', sortBy !== headerKey ? 'opacity-0' : '')}>
                {!sortOrder && <FontAwesomeIcon icon={faCaretDown} />}
                {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} />}
                {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} />}
            </div>
        </div>
    );
};
const ExploreTable = ({
    data,
    selectedNode,
    open,
    onClose,
    onDownloadClick,
    onKebabMenuClick,
    onManageColumnsChange,
    allColumnKeys,
    selectedColumns,
}: ExploreTableProps) => {
    const [searchInput, setSearchInput] = useState('');
    const [isExpanded, toggleIsExpanded] = useToggle(false);

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
                        case null:
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

    const makeColumnDef = useCallback(
        (key: keyof MungedTableRowWithId) =>
            columnHelper.accessor(key, {
                header: () => (
                    <HeaderCell sortBy={sortBy} sortOrder={sortOrder} onClick={() => handleSort(key)} headerKey={key} />
                ),
                cell: (info) => {
                    const value = info.getValue() as EntityField['value'];

                    if (typeof value === 'boolean') {
                        return value ? (
                            <div className='h-full w-full flex justify-center items-center text-center'>
                                <FontAwesomeIcon icon={faCheck} color='green' className='scale-125' />{' '}
                            </div>
                        ) : (
                            <div className='h-full w-full flex justify-center items-center text-center'>
                                <FontAwesomeIcon icon={faCancel} color='lightgray' className='scale-125' />{' '}
                            </div>
                        );
                    }

                    return format({ keyprop: key, value, label: key }) || '--';
                },
                id: key,
            }),
        [handleSort, sortOrder]
    );

    const handleKebabMenuClick = (e: MouseEvent, id: string) => {
        onKebabMenuClick({ x: e.clientX, y: e.clientY, id });
    };

    const rows = useMemo(
        () =>
            ((data && Object.entries(data).map(([key, value]) => ({ ...value.data, id: key }))) ||
                []) as MungedTableRowWithId[],
        [data]
    );

    const filteredRows = useMemo(
        () =>
            rows?.filter((item) => {
                const filterKeys: (keyof GraphNode)[] = ['displayname', 'objectid'];
                const filterTargets = filterKeys.map((filterKey) => {
                    const stringyValue = String(item?.[filterKey]);

                    return stringyValue?.toLowerCase();
                });

                return filterTargets.some((filterTarget) => filterTarget?.includes(searchInput?.toLowerCase()));
            }),
        [searchInput, rows]
    );

    const sortedFilteredRows = useMemo(() => {
        const dataToSort = filteredRows.slice();
        if (sortBy) {
            if (sortOrder === 'asc') {
                dataToSort.sort((a, b) => {
                    return a[sortBy] < b[sortBy] ? 1 : -1;
                });
            } else {
                dataToSort.sort((a, b) => {
                    return a[sortBy] < b[sortBy] ? -1 : 1;
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

    const requiredColumnDefinitions = useMemo(
        () => [
            {
                accessorKey: '',
                id: 'action-menu',
                cell: ({ row }) => (
                    <Button
                        onClick={(e) => handleKebabMenuClick(e, row?.original?.id)}
                        className='pl-4 pr-2 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0'>
                        <FontAwesomeIcon icon={faEllipsis} className='rotate-90 dark:text-neutral-light-1 text-black' />
                    </Button>
                ),
            },
            {
                accessorKey: 'nodetype',
                id: 'nodetype',
                header: () => <span className='dark:text-neutral-light-1'>Type</span>,
                cell: (info) => {
                    return (
                        <div className='flex justify-center items-center relative'>
                            <NodeIcon nodeType={(info.getValue() as string) || ''} />
                        </div>
                    );
                },
            },
            ...['objectid', 'displayname'].map(makeColumnDef),
        ],
        [makeColumnDef, handleKebabMenuClick]
    );

    const tableColumns = useMemo(
        () => [...requiredColumnDefinitions, ...selectedColumnDefinitions],
        [requiredColumnDefinitions, selectedColumnDefinitions]
    ) as DataTableProps['columns'];

    const columnOptionsForDropdown = useMemo(
        () => [...requiredColumnDefinitions, ...nonRequiredColumnDefinitions],
        [requiredColumnDefinitions, nonRequiredColumnDefinitions]
    );

    const handleSearchInputChange = useCallback(
        (e: ChangeEvent<HTMLInputElement>) => setSearchInput(e.target.value),
        []
    );

    const searchInputProps = useMemo(
        () => ({
            onChange: handleSearchInputChange,
            value: searchInput,
            placeholder: 'Search',
        }),
        [handleSearchInputChange, searchInput]
    );
    type DataTableProps = React.ComponentProps<typeof DataTable>;

    const tableHeaderProps: DataTableProps['TableHeaderProps'] = useMemo(
        () => ({
            className: 'sticky top-0 z-10',
        }),
        []
    );

    const tableHeadProps: DataTableProps['TableHeadProps'] = useMemo(
        () => ({
            className: 'pr-4',
        }),
        []
    );

    if (!open || !data) return null;

    return (
        <div
            className={`border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 bg-neutral-light-2 ${selectedNode ? 'w-[calc(100%-450px)]' : ''} ${isExpanded ? `h-[calc(100%-72px)]` : 'h-1/2'}`}>
            <div className='explore-table-container w-full h-full'>
                <TableControls
                    className='h-[72px]'
                    columns={columnOptionsForDropdown}
                    selectedColumns={selectedColumns || requiredColumns}
                    pinnedColumns={requiredColumns}
                    onDownloadClick={onDownloadClick}
                    onExpandClick={toggleIsExpanded}
                    onManageColumnsChange={onManageColumnsChange}
                    onCloseClick={onClose}
                    tableName='Results'
                    resultsCount={sortedFilteredRows?.length}
                    SearchInputProps={searchInputProps}
                />
                <MemoDataTable
                    className='h-full *:h-[calc(100%-72px)]'
                    TableHeaderProps={tableHeaderProps}
                    TableHeadProps={tableHeadProps}
                    data={sortedFilteredRows}
                    columns={tableColumns}
                />
            </div>
        </div>
    );
};

export default ExploreTable;
