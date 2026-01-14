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

import { Button, ButtonProps, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC, HTMLProps } from 'react';
import { cn } from '../../utils';
import { AppIcon } from '../AppIcon';

const DropdownTrigger: FC<{
    open: boolean;
    selectedText: JSX.Element | string;
    buttonProps?: HTMLProps<HTMLButtonElement>;
    StartAdornment?: React.FC;
    EndAdornment?: React.FC;
    variant?: ButtonProps['variant'];
}> = ({ open, selectedText, buttonProps, StartAdornment, EndAdornment, variant }) => {
    const buttonPrimary = variant === 'primary';

    // triggerStyles match EnvironmentSelectorTrigger
    const triggerStyles =
        'max-w-56 text-sm text-contrast rounded-md bg-transparent hover:bg-primary hover:text-white border shadow-outer-0 hover:border-transparent border-neutral-light-5 group';

    return (
        <PopoverTrigger asChild>
            <Button
                variant={variant}
                className={cn(
                    'uppercase',
                    {
                        'w-full text-sm': buttonPrimary,
                        [triggerStyles]: !buttonPrimary,
                        'bg-primary text-white border-transparent': open,
                    },
                    buttonProps?.className
                )}
                size='small'
                data-testid='dropdown_context-selector'>
                <span
                    className={cn('flex justify-center items-center max-w-full', {
                        'justify-between': StartAdornment,
                    })}>
                    <div className='flex items-center truncate'>
                        {StartAdornment && <StartAdornment />}
                        <p className='pt-0.5 truncate font-bold mr-2'>{selectedText}</p>
                    </div>
                    {EndAdornment ? (
                        <EndAdornment />
                    ) : (
                        <span
                            className={cn({
                                'rotate-180 transition-transform': open,
                                'justify-self-end': buttonPrimary,
                            })}>
                            <AppIcon.CaretDown size={12} />
                        </span>
                    )}
                </span>
            </Button>
        </PopoverTrigger>
    );
};

export default DropdownTrigger;
