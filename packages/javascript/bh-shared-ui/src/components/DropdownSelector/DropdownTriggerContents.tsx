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
import { Button, ButtonProps } from 'doodle-ui';
import { forwardRef, type FC } from 'react';
import { cn } from '../../utils';
import { AppIcon } from '../AppIcon';
import { dropdownIconStateStyles, selectorIconStyles, triggerStyles } from './constants';

export type DropdownTriggerContentsProps = ButtonProps & {
    open: boolean;
    selectedText: JSX.Element | string;
    buttonProps?: ButtonProps;
    StartAdornment?: FC;
    EndAdornment?: FC;
    testId?: string;
    variant?: ButtonProps['variant'];
    readOnly?: boolean;
};

const DropdownTriggerContents = forwardRef<HTMLButtonElement, DropdownTriggerContentsProps>(
    (
        {
            open,
            selectedText,
            buttonProps,
            StartAdornment,
            EndAdornment,
            testId,
            variant,
            readOnly,
            className,
            ...props
        },
        ref
    ) => {
        const buttonPrimary = variant === 'primary';

        return (
            <Button
                ref={ref}
                {...props}
                variant={variant ?? 'transparent'}
                className={cn(
                    'uppercase group',
                    buttonPrimary && `w-full text-sm ${dropdownIconStateStyles}`,
                    {
                        [triggerStyles]: !buttonPrimary,
                        'bg-primary text-common-white dark:text-common-dark border-transparent': open,
                    },
                    className,
                    buttonProps?.className
                )}
                size='small'
                data-testid={testId ? testId : 'dropdown_context-selector'}>
                <span
                    className={cn('flex justify-center items-center max-w-full', {
                        'justify-between': StartAdornment,
                    })}>
                    <div className='flex items-center truncate'>
                        {StartAdornment && <StartAdornment />}
                        <p className='truncate font-bold mr-2'>{selectedText}</p>
                    </div>
                    {EndAdornment ? (
                        <EndAdornment />
                    ) : (
                        <span
                            className={cn({
                                'rotate-180 transition-transform': open,
                                'justify-self-end': buttonPrimary,
                                hidden: readOnly,
                            })}>
                            <AppIcon.CaretDown className={selectorIconStyles} size={12} />
                        </span>
                    )}
                </span>
            </Button>
        );
    }
);
DropdownTriggerContents.displayName = 'DropdownTriggerContents';

export default DropdownTriggerContents;
