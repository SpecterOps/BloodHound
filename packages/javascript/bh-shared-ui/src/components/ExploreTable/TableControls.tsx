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
import React from 'react';
import { cn } from '../../utils';

const ICON_CLASSES = 'cursor-pointer bg-slate-200 p-2 h-4 w-4 rounded-full';

export const TableControls = React.forwardRef<
    HTMLTableSectionElement,
    React.HTMLAttributes<HTMLTableSectionElement> & {
        SearchInputProps?: InputProps;
        resultsCount?: number;
        tableName?: string;
        className?: string;
        onDownloadClick?: () => void;
        onManageColumnsClick?: () => void;
        onExpandClick?: () => void;
        onCloseClick?: () => void;
    }
>(
    (
        {
            className,
            resultsCount,
            tableName = 'Results',
            SearchInputProps,
            onDownloadClick,
            onCloseClick,
            onExpandClick,
            onManageColumnsClick,
        },
        ref
    ) => (
        <div ref={ref} className={cn('flex p-3 justify-between', className)}>
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
                            onClick={onManageColumnsClick}>
                            Manage Columns
                        </Button>
                    </div>
                )}
                {onCloseClick && (
                    <div>
                        <FontAwesomeIcon onClick={onCloseClick} className={ICON_CLASSES} icon={faClose} />
                    </div>
                )}
            </div>
        </div>
    )
);
TableControls.displayName = 'TableControls';
