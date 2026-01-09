// Copyright 2023 Specter Ops, Inc.
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
    ButtonProps,
    Popover,
    PopoverContent,
    PopoverTrigger,
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from 'doodle-ui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { PopperContentProps } from '@radix-ui/react-popper';
import { FC, useState } from 'react';
import { cn } from '../../utils';
import { AppIcon } from '../AppIcon';
import { DropdownOption } from './types';

const DropdownSelector: FC<{
    variant?: ButtonProps['variant'];
    options: DropdownOption[];
    selectedText: JSX.Element | string;
    onChange: (selection: DropdownOption) => void;
    align?: PopperContentProps['align'];
}> = ({ variant = 'primary', options, selectedText, align = 'start', onChange }) => {
    const [open, setOpen] = useState<boolean>(false);

    const handleClose = () => setOpen(false);

    const buttonPrimary = variant === 'primary';

    const handleOpenChange: (open: boolean) => void = (open) => setOpen(open);

    return (
        <Popover open={open} onOpenChange={handleOpenChange}>
            <PopoverTrigger asChild>
                <Button
                    variant={variant}
                    className={cn({
                        'rounded-md border shadow-outer-0 hover:bg-neutral-3 hover:text-primary text-black dark:text-white truncate':
                            !buttonPrimary,
                        'w-full uppercase': buttonPrimary,
                    })}
                    data-testid='dropdown_context-selector'>
                    <span className={cn('inline-flex justify-between gap-4 items-center', { 'w-full': buttonPrimary })}>
                        <span>{selectedText}</span>
                        <span
                            className={cn({
                                'rotate-180 transition-transform': open,
                                'justify-self-end': buttonPrimary,
                            })}>
                            <AppIcon.CaretDown size={12} />
                        </span>
                    </span>
                </Button>
            </PopoverTrigger>
            <PopoverContent
                align={align}
                className={cn(
                    'flex flex-col gap-2 p-1 border border-neutral-light-5',
                    { 'w-80': buttonPrimary },
                    { 'w-full': !buttonPrimary }
                )}>
                <TooltipProvider>
                    <ul>
                        {options.map((option: DropdownOption) => {
                            return (
                                <li key={option.key}>
                                    <Button
                                        variant={'text'}
                                        className='flex justify-between items-center gap-2 w-full'
                                        onClick={() => {
                                            onChange(option);
                                            handleClose();
                                        }}>
                                        <TooltipRoot>
                                            <TooltipTrigger>
                                                <span className={cn('max-w-96 truncate', { uppercase: buttonPrimary })}>
                                                    {option.value}
                                                </span>
                                            </TooltipTrigger>
                                            <TooltipPortal>
                                                <TooltipContent side='left' className='dark:bg-neutral-dark-5 border-0'>
                                                    <span className='uppercase'>{option.value}</span>
                                                </TooltipContent>
                                            </TooltipPortal>
                                        </TooltipRoot>
                                        {option.icon && (
                                            <FontAwesomeIcon
                                                style={{ width: '10%', alignSelf: 'center' }}
                                                icon={option.icon}
                                                data-testid={`dropdown-icon-${option.icon.iconName}`}
                                                size='sm'
                                            />
                                        )}
                                    </Button>
                                </li>
                            );
                        })}
                    </ul>
                </TooltipProvider>
            </PopoverContent>
        </Popover>
    );
};

export default DropdownSelector;
