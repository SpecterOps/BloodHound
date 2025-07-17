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
import { cn } from '../../utils';
import TableControls from './TableControls';
import { ExploreTableProps, MungedTableRowWithId, requiredColumns } from './explore-table-utils';
import useExploreTableRowsAndColumns from './useExploreTableRowsAndColumns';

const MemoDataTable = memo(DataTable<MungedTableRowWithId, any>);

type DataTableProps = React.ComponentProps<typeof MemoDataTable>;

const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
    className: 'sticky top-0 z-10',
};

const tableHeadProps: DataTableProps['TableHeadProps'] = {
    className: 'pr-2',
};

const ExploreTable = ({
    data,
    selectedNode,
    onClose,
    onRowClick,
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

    const { columnOptionsForDropdown, sortedFilteredRows, tableColumns, resultsCount } = useExploreTableRowsAndColumns({
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

    if (!data) return null;

    return (
        <div
            data-testid='explore-table-container-wrapper'
            className={cn(
                'dark:bg-neutral-dark-5 border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 bg-neutral-light-2',
                {
                    'h-1/2': !isExpanded,
                    'h-[calc(100%-72px)]': isExpanded,
                    'w-[calc(100%-450px)]': selectedNode,
                }
            )}>
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
                    resultsCount={resultsCount}
                    SearchInputProps={searchInputProps}
                />
                <MemoDataTable
                    className='h-full *:h-[calc(100%-72px)]'
                    TableHeaderProps={tableHeaderProps}
                    TableHeadProps={tableHeadProps}
                    onRowClick={onRowClick}
                    selectedRow={selectedNode || undefined}
                    data={sortedFilteredRows}
                    columns={tableColumns as DataTableProps['columns']}
                />
            </div>
        </div>
    );
};

export default ExploreTable;
