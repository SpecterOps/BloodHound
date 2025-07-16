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
import { FlatGraphResponse, GraphNodeSpreadWithProperties } from 'js-client-library';
import { ChangeEvent, memo, useCallback, useMemo, useState } from 'react';
import { useToggle } from '../../hooks';
import { EntityField, format, formatPotentiallyUnknownLabel } from '../../utils';
import NodeIcon from '../NodeIcon';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';
import TableControls from './TableControls';

const KEYS_TO_FILTER_BY: (keyof GraphNodeSpreadWithProperties)[] = ['objectId', 'displayname'];

const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = ['nodetype', 'objectId', 'displayname'];

const requiredColumns = Object.fromEntries(REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.map((key) => [key, true]));

type MungedTableRowWithId = GraphNodeSpreadWithProperties & { id: string };

const columnHelper = createColumnHelper();

const makeColumnDef = (key: string) =>
    columnHelper.accessor(key, {
        header: formatPotentiallyUnknownLabel(key),
        cell: (info) => {
            const value = info.getValue() as EntityField['value'];

            if (typeof value === 'boolean') {
                return (
                    <div className='h-full w-full flex justify-center items-center text-center'>
                        <FontAwesomeIcon
                            icon={value ? faCheck : faCancel}
                            color={value ? 'green' : 'lightgray'}
                            className='scale-125'
                        />{' '}
                    </div>
                );
            }

            return format({ keyprop: key, value, label: key }) || '--';
        },
        id: key,
    });

const requiredColumnDefinitions = [
    columnHelper.accessor('', {
        id: 'action-menu',
        cell: () => (
            <Button className='pl-4 pr-2 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0'>
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
    ...['objectId', 'displayname'].map(makeColumnDef),
];

type DataTableProps = React.ComponentProps<typeof DataTable>;

const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
    className: 'sticky top-0 z-10',
};

const tableHeadProps: DataTableProps['TableHeadProps'] = {
    className: 'pr-4',
};

interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    data?: FlatGraphResponse;
    selectedColumns?: Record<string, boolean>;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
}

const MemoDataTable = memo(DataTable);

const ExploreTable = ({
    data,
    open,
    onClose,
    onManageColumnsChange,
    allColumnKeys,
    selectedColumns,
}: ExploreTableProps) => {
    const [searchInput, setSearchInput] = useState('');
    const [isExpanded, toggleIsExpanded] = useToggle(false);

    const mungedData = useMemo(
        () =>
            (data &&
                Object.entries(data).map(([key, value]) => {
                    return Object.assign({}, value?.data || {}, { id: key }) as MungedTableRowWithId;
                })) ||
            [],
        [data]
    );

    const filteredData = useMemo(
        () =>
            mungedData?.filter((potentialRow) => {
                const filterTargets = KEYS_TO_FILTER_BY.map((filterKey) =>
                    String(potentialRow?.[filterKey]).toLowerCase()
                );

                return filterTargets.some((filterTarget) => filterTarget?.includes(searchInput?.toLowerCase()));
            }),
        [searchInput, mungedData]
    );

    const nonRequiredColumnDefinitions = useMemo(
        () => allColumnKeys?.filter((key) => !requiredColumns[key]).map(makeColumnDef) || [],
        [allColumnKeys]
    );

    const selectedColumnDefinitions = useMemo(
        () => nonRequiredColumnDefinitions.filter((columnDef) => selectedColumns?.[columnDef?.id || '']),
        [nonRequiredColumnDefinitions, selectedColumns]
    );

    const tableColumns = useMemo(
        () => [...requiredColumnDefinitions, ...selectedColumnDefinitions],
        [selectedColumnDefinitions]
    ) as DataTableProps['columns'];

    const columnOptionsForDropdown = useMemo(
        () => [...requiredColumnDefinitions, ...nonRequiredColumnDefinitions],
        [nonRequiredColumnDefinitions]
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

    if (!open || !data) return null;

    return (
        <div
            className={`border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 bg-neutral-light-2 ${isExpanded ? `h-[calc(100%-72px)]` : 'h-1/2'}`}>
            <div className='explore-table-container w-full h-full'>
                <TableControls
                    className='h-[72px]'
                    columns={columnOptionsForDropdown}
                    selectedColumns={selectedColumns || requiredColumns}
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
                    className='h-full *:h-[calc(100%-72px)]'
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
