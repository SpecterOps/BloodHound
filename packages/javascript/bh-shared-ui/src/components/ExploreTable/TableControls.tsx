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

import { Input, InputProps } from '@bloodhoundenterprise/doodleui';
import { faClose, faDownload, faExpand, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ColumnDef } from '@tanstack/react-table';
import { useMemo } from 'react';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
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
                            className='border-0 w-48 rounded-none border-b-2 border-black bg-inherit'
                            {...SearchInputProps}
                        />
                        <FontAwesomeIcon icon={faSearch} className='absolute right-2' />
                    </div>
                )}
                {onDownloadClick && (
                    <div onClick={onDownloadClick} data-testid='download-button'>
                        <FontAwesomeIcon className={ICON_CLASSES} icon={faDownload} />
                    </div>
                )}
                {onExpandClick && (
                    <div onClick={onExpandClick} data-testid='expand-button'>
                        <FontAwesomeIcon className={ICON_CLASSES} icon={faExpand} />
                    </div>
                )}
                {onManageColumnsChange && (
                    <ManageColumnsComboBox
                        allColumns={parsedColumns}
                        selectedColumns={selectedColumns}
                        onChange={onManageColumnsChange}
                    />
                )}
                {onCloseClick && (
                    <div onClick={onCloseClick} data-testid='close-button'>
                        <FontAwesomeIcon className={ICON_CLASSES} icon={faClose} />
                    </div>
                )}
            </div>
        </div>
    );
};

TableControls.displayName = 'TableControls';

export default TableControls;
