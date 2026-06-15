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

import { DataTable, ScrollArea } from 'doodle-ui';
import fileDownload from 'js-file-download';
import { json2csv } from 'json-2-csv';
import { ChangeEvent, memo, useCallback, useMemo, useState } from 'react';
import { useAddKeyBinding, useExploreGraph, useExploreSelectedItem, useToggle } from '../../hooks';
import { cn } from '../../utils';
import TableControls from './TableControls';
import {
    DEFAULT_EXPLORE_TABLE_COLUMN_KEYS,
    ExploreTableProps,
    MungedTableRowWithId,
    createColumnStateFromKeys,
    defaultColumns,
    getExploreTableData,
} from './explore-table-utils';
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
    className: 'px-2 text-center',
};

const tableCellProps: DataTableProps['TableCellProps'] = {
    className: 'truncate group relative p-0 pl-2',
};

const tableOptions: DataTableProps['tableOptions'] = {
    getRowId: (row) => row.id,
};

const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
    estimateSize: () => 40,
};

const ExploreTable = ({
    onClose,
    onKebabMenuClick,
    onManageColumnsChange,
    onChangePinnedColumns,
    selectedColumns = defaultColumns,
    pinnedColumns,
}: ExploreTableProps) => {
    const { data: graphData } = useExploreGraph();
    const { selectedItem, setSelectedItem, clearSelectedItem } = useExploreSelectedItem();

    const [searchInput, setSearchInput] = useState('');
    const [isExpanded, toggleIsExpanded] = useToggle(false);

    //controlled state for reset size button
    const [columnSizing, setColumnSizing] = useState({});
    const handleResetColumnSize = () => {
        setColumnSizing({});
    };

    const handleSearchInputChange = useCallback(
        (e: ChangeEvent<HTMLInputElement>) => {
            setSearchInput(e.target.value);
        },
        [setSearchInput]
    );

    const exploreTableData = useMemo(() => getExploreTableData(graphData), [graphData]);

    const { columnOptionsForDropdown, sortedFilteredRows, tableColumns, resultsCount, columnOrder, setColumnOrder } =
        useExploreTableRowsAndColumns({
            onKebabMenuClick,
            searchInput,
            selectedColumns,
            exploreTableData,
        });

    const effectivePinnedColumns = pinnedColumns ?? DEFAULT_EXPLORE_TABLE_COLUMN_KEYS;

    const columnPinning = useMemo(() => {
        return {
            left: effectivePinnedColumns,
        };
    }, [effectivePinnedColumns]);

    const leftPinnedColumns = createColumnStateFromKeys(effectivePinnedColumns);

    const searchInputProps = useMemo(
        () => ({
            onChange: handleSearchInputChange,
            value: searchInput,
            placeholder: 'Search',
        }),
        [handleSearchInputChange, searchInput]
    );

    const handleRowClick = useCallback(
        (row: MungedTableRowWithId) => {
            if (row.id !== selectedItem) {
                setSelectedItem(row.id);
            } else {
                clearSelectedItem();
            }
        },
        [setSelectedItem, selectedItem, clearSelectedItem]
    );

    const handleKeydown = useCallback(
        (event: KeyboardEvent) => {
            if (event.code === 'Escape') {
                if (typeof onClose === 'function') {
                    onClose();
                }
            }
        },
        [onClose]
    );

    useAddKeyBinding(handleKeydown);

    const handleDownloadClick = useCallback(
        (columns: 'all' | 'selected') => {
            try {
                const nodes = exploreTableData?.nodes;
                if (nodes) {
                    const nodeValues = Object.values(nodes)?.map((node) => {
                        const nodeClone = Object.assign({}, node);
                        const flattenedNodeClone = Object.assign(nodeClone, node.properties);

                        delete flattenedNodeClone.properties;
                        return flattenedNodeClone;
                    });
                    const csv = json2csv(nodeValues, {
                        keys:
                            columns === 'all'
                                ? exploreTableData.node_keys
                                : exploreTableData.node_keys.filter((key) => selectedColumns[key]),
                        emptyFieldValue: '',
                        preventCsvInjection: true,
                    });

                    fileDownload(csv, 'nodes.csv');
                }
            } catch (err) {
                console.error('Failed to export CSV:', err);
            }
        },
        [exploreTableData, selectedColumns]
    );

    const handleSetColumnPinning = useCallback(
        (pinnedCols: NonNullable<DataTableProps['columnPinning']>) => {
            onChangePinnedColumns?.(pinnedCols.left ?? []);
        },
        [onChangePinnedColumns]
    );

    return (
        <div
            data-testid='explore-table-container-wrapper'
            className={cn(
                'dark:bg-neutral-dark-5 absolute z-10 bottom-4 left-4 right-4 bg-neutral-light-2 rounded-lg shadow-outer-1 border',
                {
                    'h-1/2': !isExpanded,
                    'h-[calc(100%-72px)]': isExpanded,
                    'w-[calc(100%-450px)]': selectedItem,
                }
            )}>
            <div className='explore-table-container w-full h-full overflow-hidden grid grid-rows-[72px,1fr]'>
                <TableControls
                    columns={columnOptionsForDropdown}
                    selectedColumns={selectedColumns}
                    pinnedColumns={leftPinnedColumns}
                    onDownloadClick={handleDownloadClick}
                    onExpandClick={toggleIsExpanded}
                    onManageColumnsChange={onManageColumnsChange}
                    onChangePinnedColumns={onChangePinnedColumns}
                    onCloseClick={onClose}
                    onResetColumnSize={handleResetColumnSize}
                    tableName='Results'
                    resultsCount={resultsCount}
                    SearchInputProps={searchInputProps}
                />
                <ScrollArea>
                    <MemoDataTable
                        TableHeaderProps={tableHeaderProps}
                        TableHeadProps={tableHeadProps}
                        TableProps={tableProps}
                        TableCellProps={tableCellProps}
                        columnPinning={columnPinning}
                        setColumnPinning={handleSetColumnPinning}
                        onRowClick={handleRowClick}
                        selectedRow={selectedItem || undefined}
                        data={sortedFilteredRows}
                        columns={tableColumns as DataTableProps['columns']}
                        tableOptions={tableOptions}
                        virtualizationOptions={virtualizationOptions}
                        columnSizing={columnSizing}
                        onColumnSizingChange={setColumnSizing}
                        columnOrder={columnOrder}
                        onColumnOrderChange={(newOrder) => {
                            setColumnOrder(newOrder);
                        }}
                        growLastColumn
                        enableResizing
                        enableDragAndDrop
                    />
                </ScrollArea>
            </div>
        </div>
    );
};

export default ExploreTable;
