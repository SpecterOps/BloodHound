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
import { useExploreGraph, useExploreSelectedItem, useToggle } from '../../hooks';
import { cn, exportToJson } from '../../utils';
import TableControls from './TableControls';
import { ExploreTableProps, MungedTableRowWithId, getExploreTableData, requiredColumns } from './explore-table-utils';
import useExploreTableRowsAndColumns from './useExploreTableRowsAndColumns';

const MemoDataTable = memo(DataTable<MungedTableRowWithId, any>);

type DataTableProps = React.ComponentProps<typeof MemoDataTable>;

const tableProps: DataTableProps['TableProps'] = {
    className: 'w-[calc(100% + 250px)] table-fixed',
    disableDefaultOverflowAuto: true,
};

const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
    className: 'sticky top-0 z-10 shadow-sm',
};

const tableHeadProps: DataTableProps['TableHeadProps'] = {
    className: 'pr-2 text-center',
};

const tableCellProps: DataTableProps['TableCellProps'] = {
    className: 'truncate group relative',
};

const tableOptions: DataTableProps['tableOptions'] = {
    getRowId: (row) => row.id,
};

const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
    estimateSize: () => 79,
};

const ExploreTable = ({
    onClose,
    onKebabMenuClick,
    onManageColumnsChange,
    selectedColumns = requiredColumns,
}: ExploreTableProps) => {
    const { data: graphData } = useExploreGraph();
    const { selectedItem, setSelectedItem } = useExploreSelectedItem();

    const [searchInput, setSearchInput] = useState('');
    const [isExpanded, toggleIsExpanded] = useToggle(false);

    const handleSearchInputChange = useCallback(
        (e: ChangeEvent<HTMLInputElement>) => {
            setSearchInput(e.target.value);
        },
        [setSearchInput]
    );

    const exploreTableData = useMemo(() => getExploreTableData(graphData), [graphData]);
    const { columnOptionsForDropdown, sortedFilteredRows, tableColumns, resultsCount } = useExploreTableRowsAndColumns({
        onKebabMenuClick,
        searchInput,
        selectedColumns,
        exploreTableData,
    });

    const searchInputProps = useMemo(
        () => ({
            onChange: handleSearchInputChange,
            value: searchInput,
            placeholder: 'Search',
        }),
        [handleSearchInputChange, searchInput]
    );

    const handleDownloadClick = useCallback(() => {
        if (exploreTableData?.nodes) {
            exportToJson({ nodes: exploreTableData?.nodes });
        }
    }, [exploreTableData?.nodes]);

    const handleRowClick = useCallback(
        (row: MungedTableRowWithId) => {
            if (row.id !== selectedItem) {
                setSelectedItem(row.id);
            }
        },
        [setSelectedItem, selectedItem]
    );

    return (
        <div
            data-testid='explore-table-container-wrapper'
            className={cn('dark:bg-neutral-dark-5 border-2 absolute z-10 bottom-4 left-4 right-4 bg-neutral-light-2', {
                'h-1/2': !isExpanded,
                'h-[calc(100%-72px)]': isExpanded,
                'w-[calc(100%-450px)]': selectedItem,
            })}>
            <div className='explore-table-container w-full h-full overflow-hidden grid grid-rows-[72px,1fr]'>
                <TableControls
                    columns={columnOptionsForDropdown}
                    selectedColumns={selectedColumns}
                    pinnedColumns={requiredColumns}
                    onDownloadClick={handleDownloadClick}
                    onExpandClick={toggleIsExpanded}
                    onManageColumnsChange={onManageColumnsChange}
                    onCloseClick={onClose}
                    tableName='Results'
                    resultsCount={resultsCount}
                    SearchInputProps={searchInputProps}
                />
                <MemoDataTable
                    className='overflow-auto'
                    TableHeaderProps={tableHeaderProps}
                    TableHeadProps={tableHeadProps}
                    TableProps={tableProps}
                    TableCellProps={tableCellProps}
                    onRowClick={handleRowClick}
                    selectedRow={selectedItem || undefined}
                    data={sortedFilteredRows}
                    columns={tableColumns as DataTableProps['columns']}
                    tableOptions={tableOptions}
                    virtualizationOptions={virtualizationOptions}
                    growLastColumn
                />
            </div>
        </div>
    );
};

export default ExploreTable;
