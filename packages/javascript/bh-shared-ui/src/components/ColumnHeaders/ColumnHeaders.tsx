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

import {
    Button,
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from '@bloodhoundenterprise/doodleui';
import { faInfoCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { SortOrder } from '../../types';
import { adaptClickHandlerToKeyDown, cn } from '../../utils';
import { AppIcon } from '../AppIcon';

interface BaseColumnHeader extends React.HTMLAttributes<HTMLDivElement> {
    textAlign: 'left' | 'center' | 'right';
    title: string;
}

export const BaseColumnHeader: React.FC<BaseColumnHeader> = (props) => {
    const { textAlign, title, className } = props;

    const textAlignment = {
        'text-left': textAlign === 'left',
        'text-center': textAlign === 'center',
        'text-right': textAlign === 'right',
    };

    return <div className={cn('font-semibold text-base -mb-1', textAlignment, className)}>{title}</div>;
};

interface SortableHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
    title: string;
    tooltipText?: string;
    sortOrder?: SortOrder;
    disable?: boolean;
    classes?: {
        container?: React.HTMLAttributes<HTMLDivElement>['className'];
        button?: React.HTMLAttributes<HTMLButtonElement>['className'];
    };
    onSort: () => void;
}

export const SortableHeader: React.FC<SortableHeaderProps> = (props) => {
    const { title, tooltipText, sortOrder, disable, classes, onSort, ...rest } = props;

    const containerClass = classes && classes.container ? classes.container : '';
    const buttonClass = classes && classes.button ? classes.button : '';

    let IconComponent = AppIcon.SortEmpty;
    if (sortOrder === 'asc') IconComponent = AppIcon.SortAsc;
    if (sortOrder === 'desc') IconComponent = AppIcon.SortDesc;

    return (
        <TooltipProvider>
            <TooltipRoot>
                <TooltipTrigger asChild>
                    <div {...rest} data-testid='column-header_sort-button' className={containerClass}>
                        <Button
                            aria-label={`Sort by ${title}`}
                            className={cn('p-0 font-semibold text-base hover:no-underline relative', buttonClass)}
                            onClick={onSort}
                            onKeyDown={adaptClickHandlerToKeyDown(onSort)}
                            tabIndex={0}
                            variant={'text'}>
                            {title}
                            {/* SortIcon stays inside the button when no tooltip to avoid affecting other header style layouts - example on the ObjectsAccordion used on Attack Paths and PZ Zone builder pages */}
                            {!tooltipText && <IconComponent size={12} className={cn('absolute -right-5 m-1')} />}
                            {tooltipText && (
                                <>
                                    <span
                                        className='flex items-center'
                                        role='img'
                                        aria-label='More information in tooltip'
                                        data-testid='column-header_tooltip-trigger-icon'>
                                        <FontAwesomeIcon className={cn('m-1')} size={'sm'} icon={faInfoCircle} />
                                    </span>
                                    <span className='flex items-center'>
                                        <IconComponent size={12} />
                                    </span>
                                    <TooltipPortal>
                                        <TooltipContent
                                            className='max-w-80 dark:bg-neutral-dark-5 border-0'
                                            data-testid='column-header_tooltip-content-text'>
                                            {tooltipText}
                                        </TooltipContent>
                                    </TooltipPortal>
                                </>
                            )}
                        </Button>
                    </div>
                </TooltipTrigger>
            </TooltipRoot>
        </TooltipProvider>
    );
};
