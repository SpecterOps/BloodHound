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

import {
    Announcements,
    DndContext,
    DndContextProps,
    DragCancelEvent,
    type DragEndEvent,
    DragOverEvent,
    DragStartEvent,
    KeyboardSensor,
    MouseSensor,
    TouchSensor,
    closestCenter,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import { restrictToHorizontalAxis } from '@dnd-kit/modifiers';
import { SortableContext, arrayMove, horizontalListSortingStrategy } from '@dnd-kit/sortable';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { getCommonPinnedStyles, getConditionalPinnedStyles } from '../DataTable/pinnedStyles';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../Table';
import { cn } from '../utils';
import NoDataFallback from './NoDataFallback';
import PinDialog from './components/PinDialog';
import { ResizeIndicator } from './components/ResizeIndicator';
import { getExpandedColWidth, updateColumnOrder } from './utils';

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
    enableDragAndDrop?: boolean; //table level prop to enable drag and drop
    columnOrder?: string[];
    columnPinning?: ColumnPinningState;
    columnSizing?: ColumnSizingState;
    onColumnPinningChange?: OnChangeFn<ColumnPinningState>;
    onColumnSizingChange?: OnChangeFn<ColumnSizingState>;
    setColumnPinning?: (columnPinning: ColumnPinningState) => void;
    onColumnOrderChange?: OnChangeFn<string[]>;
    TableProps?: React.ComponentPropsWithoutRef<typeof Table> & {
        disableDefaultOverflowAuto?: boolean;
        tableContainerClassName?: string;
        heightContainerClassName?: string;
    };
    TableHeaderProps?: React.ComponentPropsWithoutRef<typeof TableHeader>;
    TableHeaderRowProps?: React.ComponentPropsWithoutRef<typeof TableRow>;
    TableHeadProps?: Omit<React.ComponentPropsWithoutRef<typeof TableHead>, 'header' | 'enableDragging'>;
    TableBodyProps?: React.ComponentPropsWithoutRef<typeof TableBody>;
    TableBodyRowProps?: React.ComponentPropsWithoutRef<typeof TableRow>;
    TableCellProps?: Omit<React.ComponentPropsWithoutRef<typeof TableCell>, 'cell' | 'enableDragging'>;
}

const DndWrapper = (props: DndContextProps & { disabled?: boolean }) => {
    if (props.disabled) return <>{props.children}</>;

    return <DndContext {...props} />;
};

const DataTable = <TData, TValue>(props: DataTableProps<TData, TValue>) => {
    const {
        columns: columnsProp,
        data,
        onRowClick,
        selectedRow,
        growLastColumn,
        tableOptions = {},
        columnOrder,
        columnPinning,
        setColumnPinning,
        columnSizing,
        onColumnPinningChange,
        onColumnSizingChange,
        onColumnOrderChange,
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
        enableDragAndDrop = false,
        ...wrapperRest
    } = props;

    const [isOverflowing, setIsOverflowing] = useState(false);
    const [isPinDialogOpen, setIsPinDialogOpen] = useState(false);
    const [focusedCell, setFocusedCell] = useState<{
        rowIndex: number;
        colIndex: number;
        childFocused?: boolean;
    } | null>(null);

    interface PinDialogState {
        action: 'pin' | 'unpin' | null;
        activeId: string | number;
        overId: string | number;
        label?: string;
    }

    const [pinDialogState, setPinDialogState] = useState<PinDialogState>({
        action: null,
        activeId: '',
        overId: '',
        label: '',
    });

    const columns = useMemo(() => {
        const columnHelper = createColumnHelper<TData>();

        if (growLastColumn && !isOverflowing) {
            return columnsProp.concat([
                columnHelper.display({
                    id: 'empty-column',
                    header: () => <span className='sr-only'>empty placeholder column</span>,
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

    const parentRef = useRef<HTMLDivElement>(null);
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
            columnOrder,
        },
        ...(onColumnPinningChange && { onColumnPinningChange }),
        ...(onColumnSizingChange && { onColumnSizingChange }),
        ...(onColumnOrderChange && { onColumnOrderChange }),
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
                // When selectedRow is provided the parent controls selection.
                if (selectedRow === undefined) {
                    const isAlreadySelected = table.getState().rowSelection[row.id];

                    if (isAlreadySelected) {
                        table.setRowSelection({});
                    } else {
                        table.setRowSelection({ [row.id]: true });
                    }
                }

                onRowClick(row?.original);
            }
        },
        [onRowClick, selectedRow, table]
    );

    const isCrossBoundaryDrag = useCallback(
        (activeId: string | number, overId: string | number) => {
            const pinnedLeft = columnPinning?.left ?? [];
            const isActivePinned = pinnedLeft.includes(activeId as string);
            const isOverPinned = pinnedLeft.includes(overId as string);
            return isActivePinned !== isOverPinned;
        },
        [columnPinning]
    );

    const handleDragEnd = useCallback(
        (event: DragEndEvent) => {
            const { active, over } = event;

            // Guard: only act on a real positional change
            if (!active || !over || active.id === over.id) return;

            const activeId = active.id as string;
            const overId = over.id as string;
            const pinnedLeft = columnPinning?.left;

            if (pinnedLeft && isCrossBoundaryDrag(activeId, overId)) {
                // Moving between pinned ↔ unpinned zones — fire confirmation dialog
                setPinDialogState({
                    action: pinnedLeft.includes(activeId) ? 'unpin' : 'pin',
                    activeId,
                    overId,
                    label: table.getColumn(activeId)?.columnDef.meta?.label,
                });
                setIsPinDialogOpen(true);
            } else if (pinnedLeft?.includes(activeId) && pinnedLeft.includes(overId)) {
                // Reordering within the pinned zone
                setColumnPinning?.({
                    left: arrayMove(pinnedLeft, pinnedLeft.indexOf(activeId), pinnedLeft.indexOf(overId)),
                });
            } else {
                // Reordering within the unpinned zone
                onColumnOrderChange?.((order) => arrayMove(order, order.indexOf(activeId), order.indexOf(overId)));
            }
        },
        [columnPinning, isCrossBoundaryDrag, onColumnOrderChange, setColumnPinning, table]
    );

    const handlePinDialogConfirm = useCallback(
        (activeId: string | number, overId: string | number) => {
            if (!columnPinning?.left) return;

            const pinnedLeft = columnPinning.left;

            if (pinDialogState.action === 'unpin') {
                setColumnPinning?.({ left: pinnedLeft.filter((id) => id !== activeId) });
                onColumnOrderChange?.((order) => updateColumnOrder(order, activeId, overId));
            } else {
                setColumnPinning?.({ left: updateColumnOrder([...pinnedLeft, activeId as string], activeId, overId) });
            }
        },
        [columnPinning, onColumnOrderChange, pinDialogState.action, setColumnPinning]
    );

    const sensors = useSensors(useSensor(MouseSensor, {}), useSensor(TouchSensor, {}), useSensor(KeyboardSensor, {}));

    const { className: tableClassName, heightContainerClassName, ...restTableProps } = TableProps || {};
    const { className: headerClassName, ...restHeaderProps } = TableHeaderProps || {};
    const { className: headerRowClassName, ...restHeaderRowProps } = TableHeaderRowProps || {};

    const tableRows = table.getRowModel().rows;
    const haveLeftPinnedColumns = (columnPinning?.left?.length || 0) > 0;
    const tableWidth = enableResizing && !growLastColumn ? table.getCenterTotalSize() : '100%';

    const handleCellKeyDown = useCallback(
        (e: React.KeyboardEvent<HTMLTableCellElement>, rowIndex: number, colIndex: number) => {
            const totalRows = tableRows.length;
            const totalCols = tableRows[rowIndex]?.getVisibleCells().length ?? 0;

            let nextRow = rowIndex;
            let nextCol = colIndex;

            switch (e.key) {
                case 'ArrowUp':
                    e.preventDefault();
                    nextRow = Math.max(0, rowIndex - 1);
                    break;
                case 'ArrowDown':
                    e.preventDefault();
                    nextRow = Math.min(totalRows - 1, rowIndex + 1);
                    break;
                case 'ArrowLeft':
                    e.preventDefault();
                    nextCol = Math.max(0, colIndex - 1);
                    break;
                case 'ArrowRight':
                    e.preventDefault();
                    nextCol = Math.min(totalCols - 1, colIndex + 1);
                    break;
                case 'Tab':
                    // Only intercept Tab when the cell itself is focused.
                    if (e.target !== e.currentTarget) {
                        return;
                    }
                    if (e.shiftKey) {
                        if (rowIndex === 0) {
                            return;
                        }
                        e.preventDefault();
                        nextRow = Math.min(totalRows - 1, rowIndex - 1);
                        nextCol = 0;
                    } else {
                        if (rowIndex === totalRows - 1) {
                            return;
                        }
                        e.preventDefault();
                        nextRow = Math.min(totalRows - 1, rowIndex + 1);
                        nextCol = 0;
                    }
                    break;

                default:
                    return;
            }

            setFocusedCell({ rowIndex: nextRow, colIndex: nextCol });
        },
        [tableRows]
    );

    // When focusedCell changes, scroll the virtualizer to bring the target row into view,
    // then focus the corresponding <td> element via its data attributes.
    // If the cell contains a focusable interactive element (button, input, etc.), focus that instead.
    // Skip when childFocused is true
    useEffect(() => {
        if (!focusedCell || focusedCell.childFocused) return;

        virtualizer.scrollToIndex(focusedCell.rowIndex, { align: 'auto' });

        requestAnimationFrame(() => {
            const cell = parentRef.current?.querySelector(
                `td[data-row-index="${focusedCell.rowIndex}"][data-col-index="${focusedCell.colIndex}"]`
            ) as HTMLElement | null;

            if (!cell) return;

            const focusableChild = cell.querySelector<HTMLElement>(
                'button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])'
            );
            (focusableChild ?? cell).focus();
        });
    }, [focusedCell, virtualizer]);

    //Expand column to accommodate full width of content
    const setExpanded = (header: Header<TData, TValue>) => {
        const meta = header.column.columnDef.meta;
        const headerLabel = typeof meta === 'object' && !!meta.label ? meta.label : header.id;
        table.setColumnSizing((prev) => ({
            ...prev,
            // estimate the width of the header/cell content
            [header.column.id]: getExpandedColWidth(
                table,
                header.column.id,
                headerLabel,
                tableHeadRef,
                tableCellRef,
                enableDragAndDrop && header.column.columnDef.meta?.enableDragging !== false
            ),
        }));
    };

    const announcements: Announcements = {
        onDragStart({ active }: DragStartEvent) {
            const item = active.id;
            return `Picked up sortable item ${item}.`;
        },
        onDragOver({ active, over }: DragOverEvent) {
            if (over) {
                const item = active.id;
                return `Sortable item ${item} was moved over ${over.id}.`;
            }
        },
        onDragEnd({ active, over }: DragEndEvent) {
            if (over) {
                const item = active.id;
                return `Sortable item ${item} was dropped.`;
            }
        },
        onDragCancel({ active }: DragCancelEvent) {
            const item = active.id;
            return `Dragging was cancelled. Sortable item ${item} was dropped.`;
        },
    };

    return (
        <DndWrapper
            collisionDetection={closestCenter}
            modifiers={[restrictToHorizontalAxis]}
            onDragEnd={handleDragEnd}
            sensors={sensors}
            disabled={!enableDragAndDrop}
            accessibility={{ announcements, screenReaderInstructions: { draggable: 'draggable column header' } }}>
            <div
                className={cn('w-full bg-neutral-light dark:bg-neutral-dark', className)}
                {...wrapperRest}
                ref={parentRef}>
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

                                            //column level prop to enable drag and drop
                                            //set in columnDef meta: {enableDragging: boolean}
                                            const isColDraggingEnabled =
                                                header.column.columnDef.meta?.enableDragging !== false;

                                            const headerContent = (
                                                <TableHead
                                                    ref={tableHeadRef}
                                                    key={header.id}
                                                    header={header}
                                                    scope='col'
                                                    enableDragging={enableDragAndDrop && isColDraggingEnabled}
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
                                                        : flexRender(
                                                              header.column.columnDef.header,
                                                              header.getContext()
                                                          )}

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

                                            return isColDraggingEnabled ? (
                                                <SortableContext
                                                    key={header.id}
                                                    items={columnOrder ?? []}
                                                    strategy={horizontalListSortingStrategy}>
                                                    {headerContent}
                                                </SortableContext>
                                            ) : (
                                                headerContent
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

                                    const {
                                        onKeyDown: onTableBodyRowKeyDown,
                                        role: tableBodyRowRole,
                                        tabIndex: tableBodyRowTabIndex,
                                        ...restTableBodyRowAttributes
                                    } = tableBodyRowRest as React.HTMLAttributes<HTMLTableRowElement>;

                                    return (
                                        <TableRow
                                            key={row.id}
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleRowClick(row);
                                            }}
                                            onKeyDown={(event) => {
                                                onTableBodyRowKeyDown?.(event);

                                                if (
                                                    event.defaultPrevented ||
                                                    event.currentTarget !== event.target ||
                                                    !onRowClick ||
                                                    (event.key !== 'Enter' && event.key !== ' ')
                                                ) {
                                                    return;
                                                }

                                                event.preventDefault();
                                                event.stopPropagation();
                                                handleRowClick(row);
                                            }}
                                            role={tableBodyRowRole}
                                            tabIndex={tableBodyRowTabIndex ?? (onRowClick ? 0 : undefined)}
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
                                                    'focus:outline-none focus-visible:focus-ring focus-visible:bg-neutral-light-4 dark:focus-visible:bg-neutral-dark-4':
                                                        onRowClick,
                                                },

                                                propsClassName
                                            )}
                                            {...restTableBodyRowAttributes}
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

                                                //column level prop to enable drag and drop
                                                //set in columnDef meta: {enableDragging: boolean}
                                                const isColDraggingEnabled =
                                                    enableDragAndDrop &&
                                                    cell.column.columnDef.meta?.enableDragging !== false;

                                                const isCellFocused = focusedCell
                                                    ? focusedCell.rowIndex === row.index &&
                                                      focusedCell.colIndex === index
                                                    : row.index === 0 && index === 0;

                                                const cellContent = (
                                                    <TableCell
                                                        ref={tableCellRef}
                                                        key={cell.id}
                                                        cell={cell}
                                                        enableDragging={enableDragAndDrop && isColDraggingEnabled}
                                                        className={cn(
                                                            'text-left overflow-x-clip -outline-offset-2',
                                                            propsClassName
                                                        )}
                                                        {...tableCellRest}
                                                        tabIndex={isCellFocused && !focusedCell?.childFocused ? 0 : -1}
                                                        data-row-index={row.index}
                                                        data-col-index={index}
                                                        onKeyDown={(e: React.KeyboardEvent<HTMLTableCellElement>) =>
                                                            handleCellKeyDown(e, row.index, index)
                                                        }
                                                        onFocus={(e) =>
                                                            setFocusedCell({
                                                                rowIndex: row.index,
                                                                colIndex: index,
                                                                childFocused: e.target !== e.currentTarget,
                                                            })
                                                        }
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

                                                return isColDraggingEnabled ? (
                                                    <SortableContext
                                                        key={cell.id}
                                                        items={columnOrder ?? []}
                                                        strategy={horizontalListSortingStrategy}>
                                                        {cellContent}
                                                    </SortableContext>
                                                ) : (
                                                    cellContent
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
            <PinDialog
                open={isPinDialogOpen}
                onClose={() => setIsPinDialogOpen(false)}
                pinDialogState={pinDialogState}
                onConfirm={handlePinDialogConfirm}></PinDialog>
        </DndWrapper>
    );
};

export { DataTable };
