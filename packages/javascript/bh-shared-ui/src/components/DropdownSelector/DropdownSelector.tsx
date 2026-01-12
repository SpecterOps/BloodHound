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

import { Button, ButtonProps, Popover, PopoverContent, PopoverTrigger, Tooltip } from '@bloodhoundenterprise/doodleui';
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

    // triggerStyles match ZoneSelectorTrigger & LabelSelectorTrigger & EnvironmentSelectorTrigger & SimpleEnvironmentSelector
    const triggerStyles =
        'min-w-40 w-fit text-sm rounded-md bg-transparent border shadow-outer-0 hover:bg-secondary hover:border-secondary hover:text-white dark:hover:text-white group';
    // popoverContentStyles match styles in SimpleEnvironmentSelector & LabelSelector & ZoneSelector
    const popoverContentStyles = 'flex flex-col p-0 rounded-md border border-neutral-5 bg-neutral-1';
    // optionStyles match styles in ZoneSelector & LabelSelector
    const optionStyles = 'rounded-none hover:no-underline hover:bg-neutral-4 justify-normal px-4 py-1';
    // tooltipStyles match styles in ZoneSelectorOption
    const tooltipStyles = 'max-w-80 border-0 dark:bg-neutral-4 dark:text-white';

    return (
        <Popover open={open} onOpenChange={handleOpenChange}>
            <PopoverTrigger asChild>
                <Button
                    variant={variant}
                    className={cn(
                        'uppercase',
                        buttonPrimary && 'w-full text-sm',
                        !buttonPrimary && triggerStyles,
                        open && 'bg-primary text-white'
                    )}
                    size='small'
                    data-testid='dropdown_context-selector'>
                    <span className={cn('inline-flex justify-between gap-4 items-center', { 'w-full': buttonPrimary })}>
                        <p className='pt-0.5 truncate font-bold'>{selectedText}</p>
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
            <PopoverContent align={align} className={cn(popoverContentStyles, 'w-52', { 'w-64': buttonPrimary })}>
                <ul>
                    {options.map((option: DropdownOption) => {
                        return (
                            <li key={option.key}>
                                <Tooltip tooltip={option.value} contentProps={{ className: tooltipStyles }}>
                                    <Button
                                        variant={'text'}
                                        className={cn('w-full', optionStyles)}
                                        data-testid={option.value}
                                        onClick={() => {
                                            onChange(option);
                                            handleClose();
                                        }}>
                                        <span className={cn('max-w-96 truncate', { uppercase: buttonPrimary })}>
                                            {option.value}
                                        </span>
                                        {option.icon && (
                                            <FontAwesomeIcon
                                                style={{ width: '10%', alignSelf: 'center' }}
                                                icon={option.icon}
                                                data-testid={`dropdown-icon-${option.icon.iconName}`}
                                                size='sm'
                                            />
                                        )}
                                    </Button>
                                </Tooltip>
                            </li>
                        );
                    })}
                </ul>
            </PopoverContent>
        </Popover>
    );
};

export default DropdownSelector;
