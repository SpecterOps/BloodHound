// Copyright 2026 Specter Ops, Inc.
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
import {
    type ColumnDef,
    ColumnPinningState,
    ColumnSizingState,
    type Header,
    OnChangeFn,
    type Row,
    type TableOptions,
    createColumnHelper,
    flexRender,
    getCoreRowModel,
    useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { getCommonPinnedStyles, getConditionalPinnedStyles } from '../DataTable/pinnedStyles';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../Table';
import { cn } from '../utils';
import NoDataFallback from './NoDataFallback';

interface DataTableProps<TData, TValue> extends React.HTMLAttributes<HTMLDivElement> {
    /**
     * If using DataTable with a high order component and you need to spread ...rest props, consider casting the  columns prop `as ColumnDef<unknown, unknown>[]` to avoid a noisy TS error.
     *
     * *[Link to more info](https://github.com/microsoft/TypeScript/issues/28938#issuecomment-450636046)*
     * ***
     * Note: Use `ColumnDef<RowInterface>[]` for intellisense when building columns.
     */
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    onRowClick?: (row: TData) => void;
    selectedRow?: string;
    noResultsFallback?: string | React.ReactNode;
    growLastColumn?: boolean;
    virtualizationOptions?: Partial<Parameters<typeof useVirtualizer>[0]>;
    tableOptions?: Omit<TableOptions<TData>, 'columns' | 'data' | 'getCoreRowModel'>;
    enableResizing?: boolean;
    columnPinning?: ColumnPinningState;
    columnSizing?: ColumnSizingState;
    onColumnPinningChange?: OnChangeFn<ColumnPinningState>;
    onColumnSizingChange?: OnChangeFn<ColumnSizingState>;
    TableProps?: React.ComponentPropsWithoutRef<typeof Table> & {
        disableDefaultOverflowAuto?: boolean;
        tableContainerClassName?: string;
        heightContainerClassName?: string;
    };
    TableHeaderProps?: React.ComponentPropsWithoutRef<typeof TableHeader>;
    TableHeaderRowProps?: React.ComponentPropsWithoutRef<typeof TableRow>;
    TableHeadProps?: React.ComponentPropsWithoutRef<typeof TableHead>;
    TableBodyProps?: React.ComponentPropsWithoutRef<typeof TableBody>;
    TableBodyRowProps?: React.ComponentPropsWithoutRef<typeof TableRow>;
    TableCellProps?: React.ComponentPropsWithoutRef<typeof TableCell>;
}

interface ResizeIndicatorProps<TData, TValue> {
    header: Header<TData, TValue>;
    tableHeight: number | null;
    onSetExpanded: (header: Header<TData, TValue>) => void;
}

const ResizeIndicator = <TData, TValue>({
    header,
    tableHeight,
    onSetExpanded,
}: ResizeIndicatorProps<TData, TValue>) => {
    return (
        <div
            {...{
                onDoubleClick: () => onSetExpanded(header),
                onMouseDown: header.getResizeHandler(),
                onTouchStart: header.getResizeHandler(),
                className: cn(
                    `resizer opacity-0 absolute top-0 right-0 w-1 bg-neutral-5 dark:bg-neutral-light-5 cursor-col-resize select-none touch-none group-hover:opacity-100 group-[.isResizing~_&:hover]:opacity-0 group-[.isResizing~_&:hover]:cursor-default [th:has(~_.isResizing)_>_&]:!opacity-0 [th:has(~_.isResizing)_>_&]:cursor-default`,
                    header.column.getIsResizing() ? 'bg-primary opacity-100 dark:bg-secondary' : ''
                ),
            }}
            style={{
                height: tableHeight ? `${tableHeight}px` : '100%',
                transform: 'translateZ(0)',
            }}
        />
    );
};

const DataTable = <TData, TValue>(props: DataTableProps<TData, TValue>) => {
    const {
        columns: columnsProp,
        data,
        onRowClick,
        selectedRow,
        growLastColumn,
        tableOptions = {},
        columnPinning,
        columnSizing,
        onColumnPinningChange,
        onColumnSizingChange,
        className,
        TableProps,
        TableHeaderProps,
        TableHeaderRowProps,
        TableHeadProps,
        TableBodyProps,
        TableBodyRowProps,
        TableCellProps,
        noResultsFallback,
        virtualizationOptions: virtualizationOptionsProp = {},
        enableResizing = false,
        ...wrapperRest
    } = props;

    const [isOverflowing, setIsOverflowing] = useState(false);

    const columns = useMemo(() => {
        const columnHelper = createColumnHelper<TData>();

        if (growLastColumn && !isOverflowing) {
            return columnsProp.concat([
                columnHelper.display({
                    id: 'empty-column',
                }),
            ]);
        }
        return columnsProp;
    }, [growLastColumn, columnsProp, isOverflowing]);

    useEffect(() => {
        if (growLastColumn && parentRef.current) {
            const { scrollWidth, clientWidth } = parentRef.current;
            setIsOverflowing(scrollWidth > clientWidth);
        }
    }, [columns, growLastColumn]);

    const parentRef = useRef(null);
    const tableHeaderRef = useRef<HTMLTableSectionElement>(null);
    const tableHeadRef = useRef<HTMLTableCellElement>(null);
    const tableCellRef = useRef<HTMLTableCellElement>(null);

    const defaultInitialState = useMemo(() => {
        const defaultRowSelection = selectedRow ? { [selectedRow]: true } : {};

        return {
            rowSelection: defaultRowSelection,
        };
    }, [selectedRow]);

    const table = useReactTable<TData>({
        data,
        columns,
        getCoreRowModel: getCoreRowModel(),
        initialState: defaultInitialState,
        columnResizeMode: 'onChange',
        enableColumnResizing: enableResizing,
        state: {
            ...(columnPinning && { columnPinning }),
            ...(columnSizing && { columnSizing }),
        },
        ...(onColumnPinningChange && { onColumnPinningChange }),
        ...(onColumnSizingChange && { onColumnSizingChange }),
        ...tableOptions,
    });

    useEffect(() => {
        if (!selectedRow && table.getIsSomeRowsSelected()) {
            table.setRowSelection({});
        }

        if (selectedRow && !table.getState().rowSelection[selectedRow]) {
            table.setRowSelection({ [selectedRow]: true });
        }
    }, [selectedRow, table]);

    const virtualizationOptions = useMemo(
        () => ({
            count: data.length,
            getScrollElement: () => parentRef.current,
            estimateSize: () => 57,
            overscan: 20,
            ...virtualizationOptionsProp,
        }),
        [data.length, virtualizationOptionsProp]
    );

    const virtualizer = useVirtualizer(virtualizationOptions);

    const handleRowClick = useCallback(
        (row: Row<TData>) => {
            if (typeof onRowClick === 'function') {
                const isAlreadySelected = table.getState().rowSelection[row.id];

                if (isAlreadySelected) {
                    table.setRowSelection({});
                } else {
                    table.setRowSelection({ [row.id]: true });
                }

                onRowClick(row?.original);
            }
        },
        [onRowClick, table]
    );

    const { className: tableClassName, heightContainerClassName, ...restTableProps } = TableProps || {};
    const { className: headerClassName, ...restHeaderProps } = TableHeaderProps || {};
    const { className: headerRowClassName, ...restHeaderRowProps } = TableHeaderRowProps || {};

    const tableRows = table.getRowModel().rows;

    const haveLeftPinnedColumns = (columnPinning?.left?.length || 0) > 0;

    const tableWidth = enableResizing && !growLastColumn ? table.getCenterTotalSize() : '100%';

    //Expand column to accommodate full width of content
    const setExpanded = (header: Header<TData, TValue>) => {
        const meta = header.column.columnDef.meta;
        const headerLabel = typeof meta === 'object' && !!meta.label ? meta.label : header.id;
        table.setColumnSizing((prev) => ({
            ...prev,
            // estimate the width of the header/cell content
            [header.column.id]: getExpandedColWidth(header.column.id, headerLabel),
        }));
    };

    const getExpandedColWidth = (colId: string, headerId: string) => {
        const longestCellVal = getLongestCellValue(colId);
        const cellWidth = getTextWidth(longestCellVal, 'cell');
        const headerWidth = getTextWidth(headerId, 'header');
        //extra padding to allow for header icons
        return cellWidth > headerWidth ? cellWidth : headerWidth + 16;
    };

    const getLongestCellValue = (columnId: string) => {
        const allRows = table.getCoreRowModel().flatRows;
        const stringArr: string[] = allRows.map((row) => {
            return row.getValue(columnId);
        });

        const longestCellString = stringArr.reduce((longest, current) => {
            return current && current.length > longest.length ? current : longest;
        }, '');

        return longestCellString;
    };

    const getTextWidth = (text: string, valueType: string) => {
        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d') as CanvasRenderingContext2D;

        const theRef = valueType === 'header' ? tableHeadRef.current : tableCellRef.current;
        const computedStyle = theRef ? window.getComputedStyle(theRef) : null;
        const fallbackStyle = valueType === 'header' ? 'bold 1.2rem sans-serif' : 'bold .875rem sans-serif';

        context.font = computedStyle
            ? `${computedStyle?.fontWeight} ${computedStyle?.fontSize} ${computedStyle?.fontFamily}`
            : fallbackStyle;

        const metrics = context.measureText(text);
        return metrics.width + 24;
    };

    return (
        <div className={cn('w-full bg-neutral-light dark:bg-neutral-dark', className)} {...wrapperRest} ref={parentRef}>
            <div
                style={{
                    height: `${virtualizer.getTotalSize()}px`,
                }}
                className={cn(heightContainerClassName)}>
                <Table
                    {...restTableProps}
                    style={{
                        width: tableWidth,
                    }}
                    className={cn(
                        'after:inline-block after:h-[var(--prevent-vanishing-sticky-header)]',
                        enableResizing && haveLeftPinnedColumns && 'table-fixed',
                        tableClassName
                    )}>
                    <TableHeader
                        {...restHeaderProps}
                        className={cn(headerClassName, '[&_tr]:border-0')}
                        ref={tableHeaderRef}>
                        {table.getHeaderGroups().map((headerGroup) => {
                            return (
                                <TableRow
                                    key={headerGroup.id}
                                    {...restHeaderRowProps}
                                    className={cn(headerRowClassName, 'border-0')}>
                                    {headerGroup.headers.map((header, index, array) => {
                                        let propsClassName,
                                            tableHeadRest = {};

                                        if (TableHeadProps) {
                                            const { className, ...rest } = TableHeadProps;
                                            propsClassName = className;
                                            tableHeadRest = rest;
                                        }

                                        let width = `${header.getSize()}px`;
                                        if (growLastColumn) {
                                            const isLastColumn = index === array.length - 1;

                                            const lastColumnStyle = isOverflowing ? width : 'auto';
                                            width = isLastColumn ? lastColumnStyle : width;
                                        }
                                        return (
                                            <TableHead
                                                key={header.id}
                                                ref={tableHeadRef}
                                                className={cn(
                                                    'bg-neutral-light-2 dark:bg-neutral-dark-2 text-nowrap group relative z-10 overflow-x-clip',
                                                    `${header.column.getIsResizing() ? 'isResizing' : ''}`,
                                                    propsClassName
                                                )}
                                                {...tableHeadRest}
                                                style={{
                                                    width,
                                                    ...(header.column.getIsPinned() === 'left' &&
                                                        getCommonPinnedStyles(header.column.getStart('left'))),
                                                    ...(haveLeftPinnedColumns &&
                                                        header.column.getIsFirstColumn('center') && {
                                                            paddingLeft: '12px',
                                                        }),
                                                }}>
                                                {header.isPlaceholder
                                                    ? null
                                                    : flexRender(header.column.columnDef.header, header.getContext())}

                                                {header.column.getCanResize() && (
                                                    <ResizeIndicator
                                                        header={header as Header<TData, TValue>}
                                                        tableHeight={
                                                            tableHeaderRef.current &&
                                                            tableHeaderRef.current?.clientHeight +
                                                                virtualizer.getTotalSize()
                                                        }
                                                        onSetExpanded={setExpanded}
                                                    />
                                                )}
                                            </TableHead>
                                        );
                                    })}
                                </TableRow>
                            );
                        })}
                    </TableHeader>
                    <TableBody
                        {...TableBodyProps}
                        ref={(ref) => {
                            if (!ref) return;
                            const height = virtualizer.getTotalSize() - ref.getBoundingClientRect().height;
                            document.documentElement.style.setProperty(
                                '--prevent-vanishing-sticky-header',
                                `${height}px`
                            );
                        }}>
                        {tableRows.length ? (
                            virtualizer.getVirtualItems().map((virtualRow, index) => {
                                const row = tableRows[virtualRow.index];
                                const isLastRow = virtualRow.index === tableRows.length - 1;

                                let propsClassName,
                                    tableBodyRowRest = {};

                                if (TableBodyRowProps) {
                                    const { className, ...rest } = TableBodyRowProps;
                                    propsClassName = className;
                                    tableBodyRowRest = rest;
                                }

                                return (
                                    <TableRow
                                        key={row.id}
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            handleRowClick(row);
                                        }}
                                        data-state={row.getIsSelected() && 'selected'}
                                        className={cn(
                                            'hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4',
                                            {
                                                // Border is tricky on <tr> https://github.com/TanStack/virtual/issues/620
                                                'shadow-[inset_0px_0px_0px_2px_var(--primary)] dark:shadow-[inset_0px_0px_0px_2px_#4A42B5]':
                                                    row.getIsSelected(),
                                                // Not using CSS odd:even since those values are not tied to data in a virtualized table
                                                'bg-neutral-light-3 dark:bg-neutral-dark-3': row.index % 2 === 0,
                                                'bg-neutral-light-2 dark:bg-neutral-dark-2': row.index % 2 !== 0,
                                                'cursor-pointer': onRowClick,
                                                'cursor-default': !onRowClick,
                                            },

                                            propsClassName
                                        )}
                                        {...tableBodyRowRest}
                                        style={{
                                            height: `${virtualRow.size}px`,
                                            transform: `translateY(${virtualRow.start - index * virtualRow.size}px)`,
                                        }}>
                                        {row.getVisibleCells().map((cell, index, array) => {
                                            let propsClassName,
                                                tableCellRest = {};
                                            const isLastPinnedColumn = cell.column.getIsLastColumn('left');

                                            if (TableCellProps) {
                                                const { className, ...rest } = TableCellProps;
                                                propsClassName = className;
                                                tableCellRest = rest;
                                            }

                                            let width = `${cell.column.getSize()}px`;
                                            if (growLastColumn) {
                                                const isLastColumn = index === array.length - 1;
                                                const lastColumnStyle = isOverflowing ? width : 'auto';

                                                width = isLastColumn ? lastColumnStyle : width;
                                            }
                                            return (
                                                <TableCell
                                                    key={cell.id}
                                                    ref={tableCellRef}
                                                    className={cn('text-left overflow-x-clip', propsClassName)}
                                                    {...tableCellRest}
                                                    style={{
                                                        width,
                                                        ...(cell.column.getIsPinned() === 'left' && {
                                                            backgroundColor: 'inherit',
                                                            ...getCommonPinnedStyles(cell.column.getStart('left')),
                                                            ...getConditionalPinnedStyles(
                                                                isLastPinnedColumn,
                                                                isLastRow,
                                                                row.getIsSelected()
                                                            ),
                                                        }),
                                                        ...(haveLeftPinnedColumns &&
                                                            cell.column.getIsFirstColumn('center') && {
                                                                paddingLeft: '12px',
                                                            }),
                                                    }}>
                                                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                                </TableCell>
                                            );
                                        })}
                                    </TableRow>
                                );
                            })
                        ) : (
                            <NoDataFallback fallback={noResultsFallback} colSpan={columns.length} />
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
};

export { DataTable };
