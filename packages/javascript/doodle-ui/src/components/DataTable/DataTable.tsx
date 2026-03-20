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
    DndContext,
    type DragEndEvent,
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
import { ResizeIndicator } from './components/ResizeIndicator';
import { getExpandedColWidth } from './utils';

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

    // const [columnOrder, setColumnOrder] = useState<string[]>(() =>
    //     columns.map((c) => {
    //         const col = c as { id?: string; accessorKey?: string };
    //         return col.id ?? col.accessorKey ?? '';
    //     })
    // );

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
            columnOrder,
        },
        // onColumnOrderChange: onColumnOrderChange,
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

    function handleDragEnd(event: DragEndEvent) {
        const { active, over } = event;
        if (active && over && active.id !== over.id) {
            if (
                columnPinning &&
                columnPinning.left &&
                columnPinning?.left?.includes(active.id as string) &&
                columnPinning?.left?.includes(over.id as string)
            ) {
                const pinnedArr = columnPinning.left;
                const oldIndex = pinnedArr?.indexOf(active.id as string);
                const newIndex = pinnedArr?.indexOf(over.id as string);

                setColumnPinning?.({ left: arrayMove(pinnedArr, oldIndex, newIndex) });
            } else if (!isCrossBoundaryDrag(active.id, over.id)) {
                onColumnOrderChange &&
                    onColumnOrderChange((columnOrder) => {
                        const oldIndex = columnOrder.indexOf(active.id as string);
                        const newIndex = columnOrder.indexOf(over.id as string);
                        return arrayMove(columnOrder, oldIndex, newIndex);
                    });
            }
        }
    }

    const isCrossBoundaryDrag = useCallback(
        (activeId: string | number, overId: string | number) => {
            const pinnedLeft = columnPinning?.left ?? [];
            const isActivePinned = pinnedLeft.includes(activeId as string);
            const isOverPinned = pinnedLeft.includes(overId as string);
            return isActivePinned !== isOverPinned;
        },
        [columnPinning]
    );

    const cancelDrop = useCallback(
        ({ active, over }: { active: { id: string | number }; over: { id: string | number } | null }) => {
            if (!over) return false;
            return isCrossBoundaryDrag(active.id, over.id);
        },
        [isCrossBoundaryDrag]
    );

    const sensors = useSensors(useSensor(MouseSensor, {}), useSensor(TouchSensor, {}), useSensor(KeyboardSensor, {}));

    const { className: tableClassName, heightContainerClassName, ...restTableProps } = TableProps || {};
    const { className: headerClassName, ...restHeaderProps } = TableHeaderProps || {};
    const { className: headerRowClassName, ...restHeaderRowProps } = TableHeaderRowProps || {};

    const tableRows = table.getRowModel().rows;

    const haveLeftPinnedColumns = (columnPinning?.left?.length || 0) > 0;

    // Column IDs that are NOT pinned — used as SortableContext items for unpinned columns
    // so that dragging a pinned column over unpinned columns (or vice-versa) does not cause
    // the opposing group to visually shift.
    const unpinnedColumnIds = useMemo(
        () => columnOrder?.filter((id) => !(columnPinning?.left ?? []).includes(id)),
        [columnOrder, columnPinning]
    );

    const tableWidth = enableResizing && !growLastColumn ? table.getCenterTotalSize() : '100%';

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

    return (
        <DndContext
            collisionDetection={closestCenter}
            modifiers={[restrictToHorizontalAxis]}
            onDragEnd={handleDragEnd}
            cancelDrop={cancelDrop}
            sensors={sensors}>
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
                                                    items={
                                                        header.column.getIsPinned()
                                                            ? columnPinning?.left ?? []
                                                            : unpinnedColumnIds ?? []
                                                    }
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

                                                //column level prop to enable drag and drop
                                                //set in columnDef meta: {enableDragging: boolean}
                                                const isColDraggingEnabled =
                                                    cell.column.columnDef.meta?.enableDragging !== false;

                                                const cellContent = (
                                                    <TableCell
                                                        ref={tableCellRef}
                                                        key={cell.id}
                                                        cell={cell}
                                                        enableDragging={enableDragAndDrop && isColDraggingEnabled}
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

                                                return isColDraggingEnabled ? (
                                                    <SortableContext
                                                        key={cell.id}
                                                        items={
                                                            cell.column.getIsPinned()
                                                                ? columnPinning?.left ?? []
                                                                : unpinnedColumnIds ?? []
                                                        }
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
        </DndContext>
    );
};

export { DataTable };
