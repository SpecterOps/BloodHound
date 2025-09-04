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
} from '@bloodhoundenterprise/doodleui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FC, useState } from 'react';
import { cn } from '../../utils/theme';
import { AppIcon } from '../AppIcon';
import { DropdownOption } from './types';

const DropdownSelector: FC<{
    // options: DropdownOption[];
    // selectedText: JSX.Element | string;
    // onChange: (selection: DropdownOption) => void;
    // buttonClasses?: string;
    // buttonStyle?: CSSProperties;
    // menuItemSize?: number;
    // toUpper?: boolean;
    // variant: ButtonProps['variant'];

    // }> = ({ options, selectedText, onChange, buttonClasses, buttonStyle, variant = 'primary' }) => {
    // const [anchorEl, setAnchorEl] = useState(null);
    // const open = Boolean(anchorEl);

    // const handleClick = (e: any) => {
    //     setAnchorEl(e.currentTarget);
    // };

    // const handleClose = () => {
    //     setAnchorEl(null);
    // };

    // const defaultClassName = 'w-full truncate uppercase';

    //     return (
    //         <Box p={1}>
    //             <Button
    //                 variant={variant}
    //                 style={buttonStyle}
    //                 className={buttonClasses ?? defaultClassName}
    //                 onClick={handleClick}
    //                 data-testid='dropdown_context-selector'>
    //                 <span className='inline-flex justify-between gap-4 items-center w-full'>
    //                     <span>{selectedText}</span>
    //                     <span className={cn({ 'rotate-180 transition-transform': open })}>
    //                         <AppIcon.CaretDown size={12} />
    //                     </span>
    //                 </span>
    //             </Button>
    //             <Popover
    //                 open={open}
    //                 anchorEl={anchorEl}
    //                 onClose={handleClose}
    //                 anchorOrigin={{
    //                     vertical: 'bottom',
    //                     horizontal: 'left',
    //                 }}
    //                 transformOrigin={{
    //                     vertical: -10,
    //                     horizontal: 0,
    //                 }}
    //                 data-testid='dropdown_context-selector-popover'>
    //                 {options.map((option) => {
    //                     return (
    //                         <MenuItem
    //                             style={{
    //                                 display: 'flex',
    //                                 justifyContent: 'space-between',
    //                                 width: 200,
    //                                 maxWidth: 200,
    //                             }}
    //                             key={option.key}
    //                             onClick={() => {
    //                                 onChange(option);
    //                                 handleClose();
    //                             }}>
    //                             <Tooltip title={option.value}>
    //                                 <Typography
    //                                     style={{
    //                                         overflow: 'hidden',
    //                                         textTransform: 'uppercase',
    //                                         display: 'inline-block',
    //                                         textOverflow: 'ellipsis',
    //                                         maxWidth: 350,
    //                                     }}>
    //                                     {option.value}
    //                                 </Typography>
    //                             </Tooltip>
    //                             {option.icon && (
    //                                 <FontAwesomeIcon
    //                                     style={{ width: '10%', alignSelf: 'center' }}
    //                                     icon={option.icon}
    //                                     size='sm'
    //                                 />
    //                             )}
    //                         </MenuItem>
    //                     );
    //                 })}
    //             </Popover>
    //         </Box>
    //     );
    // };

    variant?: ButtonProps['variant'];
    options: DropdownOption[];
    selectedText: JSX.Element | string;
    onChange: (selection: DropdownOption) => void;
    align?: 'center' | 'start' | 'end';
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
                        'bg-transparent rounded-md border shadow-outer-0 hover:bg-neutral-3 text-black dark:text-white truncate':
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
                data-testid='dropdown_context-selector-popover'
                align={align}
                className='flex flex-col gap-2 p-4 border border-neutral-light-5 w-80'>
                <ul>
                    {options.map((option: DropdownOption, index: number) => {
                        return (
                            <li key={option.key}>
                                <Button
                                    variant={'text'}
                                    className='flex justify-between items-center gap-2 w-full'
                                    onClick={() => {
                                        onChange(option);
                                        handleClose();
                                    }}>
                                    <TooltipProvider>
                                        <TooltipRoot>
                                            <TooltipTrigger>
                                                <span className='uppercase max-w-96 truncate'>{option.value}</span>
                                            </TooltipTrigger>
                                            <TooltipPortal>
                                                <TooltipContent side='left' className='dark:bg-neutral-dark-5 border-0'>
                                                    <span className='uppercase'>{option.value}</span>
                                                </TooltipContent>
                                            </TooltipPortal>
                                        </TooltipRoot>
                                    </TooltipProvider>
                                    {option.icon && (
                                        <FontAwesomeIcon
                                            style={{ width: '10%', alignSelf: 'center' }}
                                            icon={option.icon}
                                            size='sm'
                                        />
                                    )}
                                </Button>
                            </li>
                        );
                    })}
                </ul>
            </PopoverContent>
        </Popover>
    );
};

export default DropdownSelector;
