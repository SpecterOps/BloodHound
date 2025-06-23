import { Button, Input, InputProps } from '@bloodhoundenterprise/doodleui';
import { faClose, faDownload, faExpand, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ColumnDef } from '@tanstack/react-table';
import React, { ForwardedRef, useMemo, useState } from 'react';
import { cn } from '../../utils';
import { ManageColumnsComboBox, ManageColumnsComboBoxOption } from './ManageColumnsComboBox';

const ICON_CLASSES = 'cursor-pointer bg-slate-200 p-2 h-4 w-4 rounded-full';

type TableControlsProps<TData, TValue> = {
    SearchInputProps?: InputProps;
    columns: ColumnDef<TData, TValue>[];
    visibleColumns?: Record<string, boolean>;
    pinnedColumns?: Record<string, boolean>;
    resultsCount?: number;
    tableName?: string;
    className?: string;
    onDownloadClick?: () => void;
    onManageColumnsClick?: () => void;
    onExpandClick?: () => void;
    onCloseClick?: () => void;
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
};

const TableControlsInner = <TData, TValue>(
    {
        className,
        resultsCount,
        columns,
        pinnedColumns = {},
        tableName = 'Results',
        visibleColumns,
        SearchInputProps,
        onDownloadClick,
        onCloseClick,
        onExpandClick,
        onManageColumnsChange,
        onManageColumnsClick,
    }: TableControlsProps<TData, TValue>,
    ref: ForwardedRef<HTMLDivElement>
) => {
    const [showManageColumns, setShowManageColumns] = useState(false);

    const handleManageColumnsClick = () => {
        if (onManageColumnsClick) onManageColumnsClick();
        setShowManageColumns(true);
    };

    const camelCaseToWords = (s: string) => {
        const result = s.replace(/([A-Z])/g, ' $1');
        return result.charAt(0).toUpperCase() + result?.slice(1);
    };

    const parsedColumns = useMemo(
        () =>
            columns?.slice(1).map((columnDef: ColumnDef<TData, TValue>) => ({
                id: columnDef?.id || '',
                value: camelCaseToWords(columnDef?.id || ''),
                isPinned: pinnedColumns[columnDef?.id || ''] || false,
            })),
        []
    );

    return (
        <div ref={ref} className={cn('flex p-3 justify-between relative', className)}>
            <div>
                <div className='font-bold text-lg'>{tableName}</div>
                {typeof resultsCount === 'number' && <div className='text-sm'>{resultsCount} results</div>}
            </div>
            <div className='flex justify-end items-center w-1/2 gap-3'>
                {SearchInputProps && (
                    <div className='flex justify-center items-center relative'>
                        <Input
                            className='border-0 w-48 rounded-none border-b-2 border-black bg-inherit'
                            {...SearchInputProps}
                        />
                        <FontAwesomeIcon icon={faSearch} className='absolute right-2' />
                    </div>
                )}
                {onDownloadClick && (
                    <div>
                        <FontAwesomeIcon onClick={onDownloadClick} className={ICON_CLASSES} icon={faDownload} />
                    </div>
                )}
                {onExpandClick && (
                    <div>
                        <FontAwesomeIcon onClick={onExpandClick} className={ICON_CLASSES} icon={faExpand} />
                    </div>
                )}
                {onManageColumnsClick && (
                    <div className='mb-1'>
                        <Button
                            className='hover:bg-gray-300 cursor-pointer bg-slate-200 h-8 text-black rounded-full text-sm text-center'
                            onClick={handleManageColumnsClick}>
                            Manage Columns
                        </Button>
                    </div>
                )}
                {onCloseClick && (
                    <div>
                        <FontAwesomeIcon onClick={onCloseClick} className={ICON_CLASSES} icon={faClose} />
                    </div>
                )}
                {showManageColumns && onManageColumnsChange && (
                    <div className='absolute z-20 top-3'>
                        <ManageColumnsComboBox
                            allItems={parsedColumns}
                            visibleColumns={visibleColumns}
                            onChange={onManageColumnsChange}
                            onClose={() => setShowManageColumns(false)}
                        />
                    </div>
                )}
            </div>
        </div>
    );
};

export const TableControls = React.forwardRef(TableControlsInner) as <TData, TValue>(
    props: React.HTMLAttributes<HTMLTableSectionElement> & TableControlsProps<TData, TValue>
) => ReturnType<typeof TableControlsInner>;

TableControlsInner.displayName = 'TableControls';
