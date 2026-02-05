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

import { Button, Input, InputProps } from '@bloodhoundenterprise/doodleui';
import { faClose, faDownload, faExpand, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ColumnDef } from '@tanstack/react-table';
import { useMemo } from 'react';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { ManageColumnsComboBox, ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';

const ICON_CLASSES = 'cursor-pointer bg-slate-200 p-2 h-4 w-4 rounded-full dark:text-black';

type TableControlsProps<TData, TValue> = {
    SearchInputProps?: InputProps;
    columns: ColumnDef<TData, TValue>[];
    selectedColumns: Record<string, boolean>;
    pinnedColumns?: Record<string, boolean>;
    resultsCount?: number;
    tableName?: string;
    className?: string;
    onDownloadClick?: () => void;
    onExpandClick?: () => void;
    onCloseClick?: () => void;
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    onResetColumnSize?: () => void;
};

const TableControls = <TData, TValue>({
    className,
    resultsCount,
    columns,
    pinnedColumns = {},
    tableName = 'Results',
    selectedColumns,
    SearchInputProps,
    onDownloadClick,
    onCloseClick,
    onExpandClick,
    onManageColumnsChange,
    onResetColumnSize,
}: TableControlsProps<TData, TValue>) => {
    const parsedColumns: ManageColumnsComboBoxOption[] = useMemo(
        () =>
            columns?.map((columnDef: ColumnDef<TData, TValue>) => ({
                id: columnDef?.id || '',
                value: formatPotentiallyUnknownLabel(columnDef?.id || ''),
                isPinned: pinnedColumns[columnDef?.id || ''] || false,
            })),
        [columns, pinnedColumns]
    );

    const DISABLED_CLASSNAME = 'pointer-events-none *:dark:text-neutral-500 *:text-neutral-400';
    const noResults = !resultsCount;
    return (
        <div className={cn('flex p-3 justify-between relative', className)}>
            <div>
                <div className='font-bold text-lg'>{tableName}</div>
                {typeof resultsCount === 'number' && <div className='text-sm'>{resultsCount} results</div>}
            </div>
            <div className='flex justify-end items-center w-1/2 gap-3'>
                {SearchInputProps && (
                    <div className='flex justify-center items-center relative'>
                        <Input
                            data-testid='explore-table-search'
                            disabled={noResults}
                            className={cn('border-0 w-48 rounded-none border-b-2 border-black bg-inherit', {
                                [DISABLED_CLASSNAME]: noResults,
                                'border-neutral-400': noResults,
                            })}
                            {...SearchInputProps}
                        />
                        <FontAwesomeIcon
                            className={cn('absolute right-2', { [DISABLED_CLASSNAME]: noResults })}
                            icon={faSearch}
                        />
                    </div>
                )}
                {onDownloadClick && (
                    <button
                        aria-disabled={noResults}
                        onClick={onDownloadClick}
                        data-testid='download-button'
                        aria-label='Download CSV'
                        className={cn({ [DISABLED_CLASSNAME]: noResults })}>
                        <FontAwesomeIcon className={ICON_CLASSES} icon={faDownload} />
                    </button>
                )}
                {onExpandClick && (
                    <div
                        role='button'
                        tabIndex={0}
                        onClick={onExpandClick}
                        onKeyDown={adaptClickHandlerToKeyDown(onExpandClick)}
                        data-testid='expand-button'
                        aria-label='Expand table view'>
                        <FontAwesomeIcon className={ICON_CLASSES} icon={faExpand} />
                    </div>
                )}
                {onManageColumnsChange && (
                    <ManageColumnsComboBox
                        disabled={noResults}
                        allColumns={parsedColumns}
                        selectedColumns={selectedColumns}
                        onChange={onManageColumnsChange}
                        onResetColumnSize={onResetColumnSize}
                    />
                )}
                {onCloseClick && (
                    <Button
                        variant='text'
                        onClick={onCloseClick}
                        onKeyDown={adaptClickHandlerToKeyDown(onCloseClick)}
                        data-testid='close-button'
                        aria-label='Close table view'>
                        <FontAwesomeIcon className={ICON_CLASSES} icon={faClose} />
                    </Button>
                )}
            </div>
        </div>
    );
};

TableControls.displayName = 'TableControls';

export default TableControls;
