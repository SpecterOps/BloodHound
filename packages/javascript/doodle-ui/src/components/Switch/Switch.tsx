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
    'inline-flex items-center rounded-sm focus-within:ring-2 focus-within:ring-secondary focus-within:ring-offset-2'
);

const switchRootStyles = cva(
    'group flex h-3 w-6 items-center rounded-3xl bg-neutral-dark-5 transition-all ease-in-out data-[state=checked]:bg-primary disabled:cursor-not-allowed disabled:bg-[#616161] disabled:data-[state=checked]:bg-[#616161] dark:bg-neutral-light-5 dark:data-[state=checked]:bg-primary outline-none focus:outline-none focus-visible:outline-none focus:ring-0 focus-visible:ring-0'
);

const switchThumbStyles = cva(
    'h-2.5 w-2.5 translate-x-px rounded-full bg-neutral-light-2 shadow-outer-1 transition-all ease-in-out data-[state=checked]:translate-x-[13px] disabled:bg-neutral-5 group-data-[disabled]:bg-neutral-5 group-data-[disabled]:data-[state=checked]:bg-neutral-5 focus-visible:outline-none focus-visible:ring-0'
);

const LabelVariants = cva(
    // 'text-base text-neutral-dark-5 transition-all ease-in-out hover:cursor-pointer dark:text-white',
    'text-base text-neutral-dark-5 transition-all ease-in-out dark:text-white',
    {
        variants: {
            position: {
                left: 'mr-3',
                right: 'ml-3',
            },
            disabled: {
                true: 'cursor-not-allowed text-neutral-5 dark:text-neutral-5',
                false: 'hover:cursor-pointer',
            },
        },
        defaultVariants: {
            position: 'right',
            disabled: false,
        },
    }
);

type SwitchProps = React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root> & {
    label?: string;
    labelPosition?: 'left' | 'right';
};

const Switch = forwardRef<React.ElementRef<typeof SwitchPrimitives.Root>, SwitchProps>(
    ({ className, disabled, id, label, labelPosition = 'right', ...props }, ref) => {
        const ariaLabel = label || 'switch';
        const switchId = React.useId();

        return (
            <div className={switchWrapperStyles()}>
                {labelPosition === 'left' && label && (
                    <label className={LabelVariants({ position: 'left', disabled })} htmlFor={switchId}>
                        {label}
                    </label>
                )}

                <SwitchPrimitives.Root
                    ref={ref}
                    id={id ?? switchId}
                    aria-label={!label ? ariaLabel : undefined}
                    className={cn(switchRootStyles(), className)}
                    disabled={disabled}
                    {...props}>
                    <SwitchPrimitives.Thumb className={switchThumbStyles()} />
                </SwitchPrimitives.Root>

                {labelPosition !== 'left' && label && (
                    <label className={LabelVariants({ position: 'right', disabled })} htmlFor={switchId}>
                        {label}
                    </label>
                )}
            </div>
        );
    }
);

Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };
