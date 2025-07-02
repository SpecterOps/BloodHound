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
import { faCancel, faCheck, faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { GraphNode } from 'js-client-library';
import { ChangeEvent, memo, useCallback, useMemo, useState } from 'react';
import { useToggle } from '../../hooks';
import { WrappedExploreTableItem } from '../../types';
import { EntityField, format, formatPotentiallyUnknownLabel } from '../../utils';
import NodeIcon from '../NodeIcon';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';
import { TableControls } from './TableControls';

const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = ['nodetype', 'objectid', 'displayname'];

const requiredColumns = REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.reduce(
    (acc, curr) => ({ ...acc, [curr]: true }),
    {}
) as Record<string, boolean>;

type MungedTableRowWithId = WrappedExploreTableItem['data'] & { id: string };

const columnhelper = createColumnHelper();

interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    data?: Record<string, WrappedExploreTableItem>;
    visibleColumns?: Record<string, boolean>;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
}

const MemoDataTable = memo(DataTable);

const makeColumnDef = (key: string) =>
    columnhelper.accessor(key, {
        header: formatPotentiallyUnknownLabel(key),
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
    });

const ExploreTable = ({
    data,
    open,
    onClose,
    onManageColumnsChange,
    allColumnKeys,
    visibleColumns,
}: ExploreTableProps) => {
    const [searchInput, setSearchInput] = useState('');
    const [isExpanded, toggleIsExpanded] = useToggle(false);

    const mungedData = useMemo(
        () =>
            // TODO: remove id and just use objectid for onRowClick/getRowId?
            ((data && Object.entries(data).map(([key, value]) => ({ ...value.data, id: key }))) ||
                []) as MungedTableRowWithId[],
        [data]
    );

    const filteredData = useMemo(
        () =>
            mungedData?.filter((item) => {
                const filterKeys: (keyof GraphNode)[] = ['displayname', 'objectid'];
                const filterTagets = filterKeys.map((filterKey) => {
                    const stringyValue = String(item?.[filterKey]);

                    return stringyValue?.toLowerCase();
                });

                return filterTagets.some((filterTarget) => filterTarget?.includes(searchInput?.toLowerCase()));
            }),
        [searchInput, mungedData]
    );

    const nonRequiredColumnDefinitions = useMemo(
        () => allColumnKeys?.filter((key) => !requiredColumns[key]).map(makeColumnDef) || [],
        [allColumnKeys]
    );

    const visibleColumnDefinitions = useMemo(
        () => nonRequiredColumnDefinitions.filter((columnDef) => visibleColumns?.[columnDef?.id || '']),
        [nonRequiredColumnDefinitions, visibleColumns]
    );

    const requiredColumnDefinitions = useMemo(
        () => [
            {
                accessorKey: '',
                id: 'action-menu',
                cell: () => (
                    <Button className='pl-4 pr-2 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0'>
                        <FontAwesomeIcon icon={faEllipsis} className='rotate-90 dark:text-neutral-light-1 text-black' />
                    </Button>
                ),
            },
            {
                accessorKey: 'nodetype',
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
            },
            ...['objectid', 'displayname'].map(makeColumnDef),
        ],
        []
    );

    const tableColumns = useMemo(
        () => [...requiredColumnDefinitions, ...visibleColumnDefinitions],
        [requiredColumnDefinitions, visibleColumnDefinitions]
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
            className={`border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 bg-neutral-light-2 ${isExpanded ? `h-[calc(100%-72px)]` : 'h-1/2'}`}>
            <div className='explore-table-container w-full h-full'>
                <TableControls
                    className={`h-[72px]`}
                    columns={columnOptionsForDropdown}
                    visibleColumns={visibleColumns || requiredColumns}
                    pinnedColumns={requiredColumns}
                    onDownloadClick={() => console.log('download icon clicked')}
                    onExpandClick={toggleIsExpanded}
                    onManageColumnsChange={onManageColumnsChange}
                    onCloseClick={onClose}
                    tableName='Results'
                    resultsCount={filteredData?.length}
                    SearchInputProps={searchInputProps}
                />
                <MemoDataTable
                    className={`h-full *:h-[calc(100%-72px)]`}
                    TableHeaderProps={tableHeaderProps}
                    TableHeadProps={tableHeadProps}
                    data={filteredData}
                    columns={tableColumns}
                />
            </div>
        </div>
    );
};

export default ExploreTable;
