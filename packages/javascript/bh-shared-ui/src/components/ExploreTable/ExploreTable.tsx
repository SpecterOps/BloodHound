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

import { DataTable } from '@bloodhoundenterprise/doodleui';
import { ChangeEvent, memo, useCallback, useMemo, useState } from 'react';
import { useToggle } from '../../hooks';
import { WrappedExploreTableItem } from '../../types';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';
import TableControls from './TableControls';
import useExploreTableRowsAndColumns from './useExploreTableRowsAndColumns';

const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = ['nodetype', 'objectid', 'displayname'];

export const requiredColumns = REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.reduce(
    (acc, curr) => ({ ...acc, [curr]: true }),
    {}
) as Record<string, boolean>;

export type MungedTableRowWithId = WrappedExploreTableItem['data'] & { id: string };

export type ExploreTableProps = {
    open?: boolean;
    onClose?: () => void;
    data?: Record<string, WrappedExploreTableItem>;
    selectedNode: string;
    selectedColumns?: Record<string, boolean>;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    onDownloadClick: () => void;
    onKebabMenuClick: (clickInfo: NodeClickInfo) => void;
};

export type DataTableProps = React.ComponentProps<typeof DataTable>;
export type NodeClickInfo = { id: string; x: number; y: number };

const MemoDataTable = memo(DataTable);

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

    const handleSearchInputChange = useCallback(
        (e: ChangeEvent<HTMLInputElement>) => setSearchInput(e.target.value),
        []
    );

    const { columnOptionsForDropdown, sortedFilteredRows, tableColumns } = useExploreTableRowsAndColumns({
        onKebabMenuClick,
        searchInput,
        allColumnKeys,
        selectedColumns,
        data,
    });

    const searchInputProps = useMemo(
        () => ({
            onChange: handleSearchInputChange,
            value: searchInput,
            placeholder: 'Search',
        }),
        [handleSearchInputChange, searchInput]
    );

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
            className={`dark:bg-neutral-dark-5 border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 bg-neutral-light-2 ${selectedNode ? 'w-[calc(100%-450px)]' : ''} ${isExpanded ? `h-[calc(100%-72px)]` : 'h-1/2'}`}>
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
