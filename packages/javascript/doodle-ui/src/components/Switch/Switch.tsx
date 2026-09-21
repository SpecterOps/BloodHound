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
import * as SwitchPrimitives from '@radix-ui/react-switch';
import { cva } from 'class-variance-authority';
import * as React from 'react';
import { forwardRef } from 'react';
import { cn } from '../utils';

const switchWrapperStyles = cva(
    'inline-flex items-center rounded-sm py-1 px-0.5 gap-1 focus-within:focus-ring'
);

const switchRootStyles = cva(
    'group flex h-3 w-6 items-center rounded-3xl transition-all ease-in-out bg-switch-fill data-[state=checked]:bg-primary disabled:data-[state=checked]:bg-disabled disabled:cursor-not-allowed disabled:bg-disabled focus:outline-none'
);

const switchThumbStyles = cva(
    'h-2.5 w-2.5 translate-x-px rounded-full shadow-outer-1 transition-all ease-in-out bg-switch-thumb-fill data-[state=checked]:translate-x-[13px] disabled:bg-switch-thumb-disabled-fill group-data-[disabled]:bg-switch-thumb-disabled-fill group-data-[disabled]:data-[state=checked]:bg-switch-thumb-disabled-fill'
);

const LabelVariants = cva('text-base text-main transition-all ease-in-out px-0.5', {
    variants: {
        position: {
            left: 'mr-2',
            right: 'ml-2',
        },
        disabled: {
            true: 'cursor-not-allowed',
            false: 'hover:cursor-pointer',
        },
    },
    defaultVariants: {
        position: 'right',
        disabled: false,
    },
});

type SwitchProps = Omit<React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root>, 'checked'> & {
    checked: boolean;
    label?: string;
    labelPosition?: 'left' | 'right';
    labelClassName?: string;
};

const Switch = forwardRef<React.ElementRef<typeof SwitchPrimitives.Root>, SwitchProps>(
    ({ className, disabled, id, label, labelPosition = 'right', labelClassName, ...props }, ref) => {
        const ariaLabel = label ?? 'switch';
        const generatedId = React.useId();
        const switchId = id ?? generatedId;

        return (
            <div className={cn(switchWrapperStyles(), label && 'py-0')}>
                {labelPosition === 'left' && label && (
                    <label
                        data-state={props.checked ? 'checked' : 'unchecked'}
                        className={cn(
                            LabelVariants({ position: 'left', disabled }),
                            'data-[state=checked]:font-medium',
                            labelClassName
                        )}
                        htmlFor={switchId}>
                        {label}
                    </label>
                )}

                <SwitchPrimitives.Root
                    ref={ref}
                    id={switchId}
                    aria-label={!label ? ariaLabel : undefined}
                    className={cn(switchRootStyles(), className)}
                    disabled={disabled}
                    {...props}>
                    <SwitchPrimitives.Thumb className={switchThumbStyles()} />
                </SwitchPrimitives.Root>

                {labelPosition !== 'left' && label && (
                    <label
                        data-state={props.checked ? 'checked' : 'unchecked'}
                        className={cn(
                            LabelVariants({ position: 'right', disabled }),
                            'data-[state=checked]:font-medium',
                            labelClassName
                        )}
                        htmlFor={switchId}>
                        {label}
                    </label>
                )}
            </div>
        );
    }
);

Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };
