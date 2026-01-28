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

import { Button, ButtonProps, Popover, PopoverContent, Tooltip } from '@bloodhoundenterprise/doodleui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { PopperContentProps } from '@radix-ui/react-popper';
import { FC, useState } from 'react';
import { cn } from '../../utils';
import DropdownTrigger from './DropdownTrigger';
import { optionStyles, popoverContentStyles, tooltipStyles } from './constants';
import { DropdownOption } from './types';

const DropdownSelector: FC<{
    options: DropdownOption[];
    selectedText: JSX.Element | string;
    onChange: (selection: DropdownOption) => void;
    StartAdornment?: React.FC;
    align?: PopperContentProps['align'];
    variant?: ButtonProps['variant'];
}> = ({ variant, options, selectedText, StartAdornment, align = 'start', onChange }) => {
    const [open, setOpen] = useState<boolean>(false);

    const handleClose = () => setOpen(false);

    const buttonPrimary = variant === 'primary';

    const handleOpenChange: (open: boolean) => void = (open) => setOpen(open);

    return (
        <Popover open={open} onOpenChange={handleOpenChange}>
            <DropdownTrigger
                open={open}
                selectedText={selectedText}
                variant={variant}
                StartAdornment={StartAdornment}
            />
            <PopoverContent align={align} className={cn(popoverContentStyles, 'w-48', { 'w-64': buttonPrimary })}>
                <ul>
                    {options.map((option: DropdownOption) => {
                        return (
                            <li key={option.key}>
                                <Tooltip tooltip={option.value} contentProps={{ className: tooltipStyles }}>
                                    <Button
                                        variant={'text'}
                                        className={optionStyles}
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
