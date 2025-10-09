import {
    type ColumnDef,
    type Row,
    type TableOptions,
    createColumnHelper,
    flexRender,
    getCoreRowModel,
    useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from 'components/Table';
import NoDataFallback from './NoDataFallback';
import { cn } from 'components/utils';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

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
    TableProps?: React.ComponentPropsWithoutRef<typeof Table> & { disableDefaultOverflowAuto?: boolean };
    TableHeaderProps?: React.ComponentPropsWithoutRef<typeof TableHeader>;
    TableHeaderRowProps?: React.ComponentPropsWithoutRef<typeof TableRow>;
    TableHeadProps?: React.ComponentPropsWithoutRef<typeof TableHead>;
    TableBodyProps?: React.ComponentPropsWithoutRef<typeof TableBody>;
    TableBodyRowProps?: React.ComponentPropsWithoutRef<typeof TableRow>;
    TableCellProps?: React.ComponentPropsWithoutRef<typeof TableCell>;
}

const DataTable = <TData, TValue>(props: DataTableProps<TData, TValue>) => {
    const {
        columns: columnsProp,
        data,
        onRowClick,
        selectedRow,
        growLastColumn,
        tableOptions = {},
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
        ...tableOptions,
    });

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
                table.setRowSelection({ [row.id]: true });

                onRowClick(row?.original);
            }
        },
        [onRowClick, table]
    );

    const { className: tableClassName, ...restTableProps } = TableProps || {};

    const tableRows = table.getRowModel().rows;

    return (
        <div className={cn('w-full bg-neutral-light dark:bg-neutral-dark', className)} {...wrapperRest} ref={parentRef}>
            <div style={{ height: `${virtualizer.getTotalSize()}px` }}>
                <Table
                    {...restTableProps}
                    className={cn(
                        'after:inline-block after:h-[var(--prevent-vanishing-sticky-header)]',
                        tableClassName
                    )}>
                    <TableHeader {...TableHeaderProps}>
                        {table.getHeaderGroups().map((headerGroup) => {
                            return (
                                <TableRow key={headerGroup.id} {...TableHeaderRowProps}>
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
                                                className={cn(
                                                    'bg-neutral-light-2 dark:bg-neutral-dark-2',
                                                    propsClassName
                                                )}
                                                {...tableHeadRest}
                                                style={{ width }}>
                                                {header.isPlaceholder
                                                    ? null
                                                    : flexRender(header.column.columnDef.header, header.getContext())}
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
                                        onClick={() => handleRowClick(row)}
                                        data-state={row.getIsSelected() && 'selected'}
                                        className={cn(
                                            'hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4',
                                            {
                                                // Border is tricky on <tr> https://github.com/TanStack/virtual/issues/620
                                                'shadow-[inset_0px_0px_0px_2px_var(--primary)]': row.getIsSelected(),
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
                                                    className={cn('text-left', propsClassName)}
                                                    {...tableCellRest}
                                                    style={{ width }}>
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
