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
import * as React from 'react';

import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { faGripVertical } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Cell, Header } from '@tanstack/react-table';
import { Tooltip } from '../Tooltip';
import { cn } from '../utils';

const Table = React.forwardRef<
    HTMLTableElement,
    React.HTMLAttributes<HTMLTableElement> & { disableDefaultOverflowAuto?: boolean; tableContainerClassName?: string }
>(({ className, disableDefaultOverflowAuto, tableContainerClassName, ...props }, ref) => (
    <div className={cn('relative w-full', { 'overflow-auto': !disableDefaultOverflowAuto }, tableContainerClassName)}>
        <table ref={ref} className={cn('w-full caption-bottom text-sm', className)} {...props} />
    </div>
));
Table.displayName = 'Table';

const TableHeader = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => (
        <thead
            ref={ref}
            className={cn('[&_tr]:border-b bg-neutral-light-2 dark:bg-neutral-dark-2', className)}
            {...props}
        />
    )
);
TableHeader.displayName = 'TableHeader';

const TableBody = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => (
        <tbody
            ref={ref}
            className={cn('[&_tr:last-child]:border-0 bg-neutral-light-2 dark:bg-neutral-dark-2', className)}
            {...props}
        />
    )
);
TableBody.displayName = 'TableBody';

const TableFooter = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
    ({ className, ...props }, ref) => (
        <tfoot
            ref={ref}
            className={cn('border-t bg-muted/50 font-medium [&>tr]:last:border-b-0', className)}
            {...props}
        />
    )
);
TableFooter.displayName = 'TableFooter';

const TableRow = React.forwardRef<HTMLTableRowElement, React.HTMLAttributes<HTMLTableRowElement>>(
    ({ className, ...props }, ref) => (
        <tr
            ref={ref}
            className={cn(
                'border-b dark:border-b-0 transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted',
                className
            )}
            {...props}
        />
    )
);
TableRow.displayName = 'TableRow';

interface TableHeadProps<TData, TValue> extends React.ThHTMLAttributes<HTMLTableCellElement> {
    header?: Header<TData, TValue>;
    // True when table level and column level drag and drop enabled.
    enableDragging?: boolean;
}

const TableHead = React.forwardRef(function TableHead<TData, TValue>(
    { className, header, enableDragging = true, style: propsStyle, ...props }: TableHeadProps<TData, TValue>,
    ref: React.ForwardedRef<HTMLTableCellElement>
) {
    const { attributes, isDragging, listeners, setNodeRef, transform } = useSortable({
        id: header?.column.id ?? '',
        disabled: !enableDragging,
    });

    const headerStyle: React.CSSProperties = {
        opacity: isDragging ? 0.8 : 1,
        position: 'relative',
        transform: CSS.Translate.toString(transform),
        transition: 'width transform 0.2s ease-in-out',
        borderRadius: isDragging ? '4px 4px 0 0' : '',
    };

    const zIndex = isDragging ? 40 : header?.column.getIsPinned() ? 30 : 1;

    return (
        <th
            ref={setNodeRef}
            className={cn(
                'h-12 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0',
                className,
                isDragging ? '!bg-neutral-light-1 dark:!bg-neutral-dark-1' : ''
            )}
            {...props}
            style={{
                ...headerStyle,
                ...propsStyle,
                zIndex: zIndex,
            }}>
            <div ref={ref} className='flex'>
                {enableDragging && header?.id !== 'empty-column' && (
                    <Tooltip tooltip='Drag to reorder'>
                        <button
                            className={cn(isDragging ? 'cursor-grabbing' : 'cursor-grab')}
                            {...attributes}
                            {...listeners}>
                            <FontAwesomeIcon icon={faGripVertical} className='text-sm mr-2' />
                        </button>
                    </Tooltip>
                )}
                {props.children}
            </div>
        </th>
    );
}) as (<TData, TValue>(
    props: TableHeadProps<TData, TValue> & React.RefAttributes<HTMLTableCellElement>
) => React.ReactElement | null) & { displayName?: string };
TableHead.displayName = 'TableHead';

interface TableCellProps<TData, TValue> extends React.TdHTMLAttributes<HTMLTableCellElement> {
    cell?: Cell<TData, TValue>;
    // True when table level and column level drag and drop enabled.
    enableDragging?: boolean;
}

const TableCell = React.forwardRef(function TableCell<TData, TValue>(
    { className, cell, enableDragging = true, style: propsStyle, ...props }: TableCellProps<TData, TValue>,
    ref: React.ForwardedRef<HTMLTableCellElement>
) {
    const { isDragging, setNodeRef, transform } = useSortable({
        id: cell?.column.id ?? '',
        disabled: !enableDragging || !cell,
    });
    const cellStyle: React.CSSProperties = {
        opacity: isDragging ? 0.8 : 1,
        position: 'relative',
        transform: CSS.Translate.toString(transform),
        transition: 'width transform 0.2s ease-in-out',
    };
    const zIndex = isDragging ? 40 : cell?.column.getIsPinned() ? 30 : 1;

    return (
        <td
            ref={setNodeRef}
            className={cn(
                'p-4 pl-0 align-middle [&:has([role=checkbox])]:pr-0',
                className,
                isDragging ? '!bg-neutral-light-1 dark:!bg-neutral-dark-1' : ''
            )}
            {...props}
            style={{ ...cellStyle, ...propsStyle, zIndex: zIndex }}>
            <div ref={ref}>{props.children}</div>
        </td>
    );
}) as (<TData, TValue>(
    props: TableCellProps<TData, TValue> & React.RefAttributes<HTMLTableCellElement>
) => React.ReactElement | null) & { displayName?: string };
TableCell.displayName = 'TableCell';

const TableCaption = React.forwardRef<HTMLTableCaptionElement, React.HTMLAttributes<HTMLTableCaptionElement>>(
    ({ className, ...props }, ref) => (
        <caption ref={ref} className={cn('mt-4 text-sm text-muted-foreground', className)} {...props} />
    )
);
TableCaption.displayName = 'TableCaption';

export { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow };
