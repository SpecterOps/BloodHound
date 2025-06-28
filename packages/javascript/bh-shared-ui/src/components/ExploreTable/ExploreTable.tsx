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

import { Button, ColumnDef, DataTable } from '@bloodhoundenterprise/doodleui';
import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Checkbox } from '@mui/material';
import { ChangeEvent, memo, useCallback, useMemo, useState } from 'react';
import { useToggle } from '../../hooks';
import { format, formatPotentiallUnknownLabel } from '../../utils';
import NodeIcon from '../NodeIcon';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox';
import { TableControls } from './TableControls';

type HasData = { data?: object };

interface ExploreTableProps<TData extends HasData> {
    open?: boolean;
    onClose?: () => void;
    data?: Record<string, TData>;
    visibleColumns?: Record<string, boolean>;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
}

const TABLE_CONTROLS_HEIGHT = '72px';

const requiredColumns = {
    nodetype: true,
    displayname: true,
    objectid: true,
    isTierZero: true,
    enabled: true,
    pwdlastset: true,
    lastlogontimestamp: true,
} as Record<string, boolean>;

const MemoDataTable = memo(DataTable);

const makeColumnDef = (key: any) =>
    ({
        accessorKey: key,
        header: formatPotentiallUnknownLabel(key),
        cell: (info: any) => format({ keyprop: key, value: info.getValue(), label: key }) || '--',
        id: key,
    }) as ColumnDef<any, any>;

const ExploreTable = <TData extends HasData>({
    data,
    open,
    onClose,
    onManageColumnsChange,
    allColumnKeys,
    visibleColumns,
}: ExploreTableProps<TData>) => {
    const [searchInput, setSearchInput] = useState('');
    const [isExpanded, toggleIsExpanded] = useToggle(false);

    const unfilteredData = useMemo(
        () => data && Object.entries(data).map(([key, value]) => ({ ...value.data, id: key })),
        [data]
    );

    const mungedData = useMemo(
        () =>
            unfilteredData?.filter((item) => item?.displayname?.toLowerCase?.()?.includes(searchInput?.toLowerCase())),
        [searchInput, unfilteredData]
    );

    const nonRequiredColumnDefinitions: ColumnDef<any, any>[] = useMemo(
        () => allColumnKeys?.filter((key) => !requiredColumns[key]).map(makeColumnDef) || [],
        [allColumnKeys]
    );

    const visibleColumnDefinitions = useMemo(
        // TODO: import AccessorColumnDef type from doodleui for complete typing
        () => nonRequiredColumnDefinitions.filter((columnDef) => visibleColumns?.[columnDef?.accessorKey]),
        [nonRequiredColumnDefinitions, visibleColumns]
    );

    const requiredColumnDefinitions: ColumnDef<any, any>[] = useMemo(
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
                cell: ({ row }) => {
                    return (
                        <div className='flex justify-center items-center relative'>
                            <NodeIcon nodeType={row?.original?.nodetype} />
                        </div>
                    );
                },
            },
            {
                accessorKey: 'isTierZero',
                id: 'isTierZero',
                header: () => {
                    return <span className='dark:text-neutral-light-1'>Is Tier Zero</span>;
                },
                cell: (cell) => <Checkbox checked={cell.getValue()} />,
            },
            ...['objectid', 'displayname', 'enabled', 'pwdlastset', 'lastlogontimestamp'].map(makeColumnDef),
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

    const tableOptions: DataTableProps['tableOptions'] = useMemo(
        () => ({
            getRowId: (row) => row?.id,
        }),
        []
    );

    if (!open || !data) return null;

    return (
        <div
            className={`border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 h-[475px] bg-neutral-light-2 ${isExpanded ? `h-[calc(100%-${TABLE_CONTROLS_HEIGHT})]` : 'h-1/2'}`}>
            <div className='explore-table-container w-full h-full'>
                <TableControls
                    className={`h-[${TABLE_CONTROLS_HEIGHT}]`}
                    columns={columnOptionsForDropdown}
                    visibleColumns={visibleColumns || requiredColumns}
                    pinnedColumns={requiredColumns}
                    onDownloadClick={() => console.log('download icon clicked')}
                    onExpandClick={toggleIsExpanded}
                    onManageColumnsChange={onManageColumnsChange}
                    onCloseClick={onClose}
                    tableName='Results'
                    resultsCount={mungedData?.length}
                    SearchInputProps={searchInputProps}
                />
                <MemoDataTable
                    className={`h-full *:h-[calc(100%-${TABLE_CONTROLS_HEIGHT})]`}
                    TableHeaderProps={tableHeaderProps}
                    TableHeadProps={tableHeadProps}
                    tableOptions={tableOptions}
                    data={mungedData as unknown[]}
                    columns={tableColumns}
                />
            </div>
        </div>
    );
};

export default ExploreTable;
