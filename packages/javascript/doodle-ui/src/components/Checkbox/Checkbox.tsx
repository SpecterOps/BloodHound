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
import * as CheckboxPrimitive from '@radix-ui/react-checkbox';
import { cva, type VariantProps } from 'class-variance-authority';
import { Check, Minus } from 'lucide-react';
import * as React from 'react';
import { Label } from '../Label';
import { cn } from '../utils';

const CheckboxVariants = cva(
    cn(
        'peer shrink-0 rounded-sm border-2',
        'border-input-border-default dark:border-input-border-default',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary dark:focus-visible:ring-secondary',
        'focus-visible:ring-offset-2 focus-visible:ring-offset-neutral-light-1 dark:focus-visible:ring-offset-neutral-dark-1',

        // enabled states
        'enabled:hover:border-secondary',
        'enabled:hover:data-[state=checked]:border-secondary enabled:hover:data-[state=checked]:bg-secondary',
        'enabled:hover:data-[state=indeterminate]:border-secondary enabled:hover:data-[state=indeterminate]:bg-secondary',
        'enabled:active:border-primary-variant',
        'enabled:data-[state=checked]:border-primary enabled:data-[state=checked]:bg-primary enabled:data-[state=checked]:text-neutral-light-1',
        'enabled:data-[state=indeterminate]:border-primary enabled:data-[state=indeterminate]:bg-primary enabled:data-[state=indeterminate]:text-neutral-light-1',
        'enabled:dark:data-[state=checked]:text-neutral-dark-1',
        'enabled:dark:data-[state=indeterminate]:text-neutral-dark-1',

        // disabled state
        'disabled:cursor-not-allowed disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-icon-disabled',

        // error state
        'enabled:aria-[invalid=true]:border-status-error-main enabled:aria-[invalid=true]:text-status-error-main',
        'enabled:aria-[invalid=true]:data-[state=checked]:border-status-error-main enabled:aria-[invalid=true]:data-[state=checked]:bg-status-error-main enabled:aria-[invalid=true]:data-[state=checked]:text-neutral-light-1',
        'enabled:aria-[invalid=true]:data-[state=indeterminate]:border-status-error-main enabled:aria-[invalid=true]:data-[state=indeterminate]:bg-status-error-main enabled:aria-[invalid=true]:data-[state=indeterminate]:text-neutral-light-1',
        'enabled:dark:aria-[invalid=true]:data-[state=checked]:text-neutral-dark-1',
        'enabled:dark:aria-[invalid=true]:data-[state=indeterminate]:text-neutral-dark-1'
    ),
    {
        variants: {
            size: {
                lg: 'size-[24px]',
                md: 'size-[18px]',
                sm: 'size-[12px]',
            },
        },
    }
);

type CheckboxProps = React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root> & {
    size?: VariantProps<typeof CheckboxVariants>['size'];
    icon?: React.ReactNode;
};

const Checkbox = React.forwardRef<React.ElementRef<typeof CheckboxPrimitive.Root>, CheckboxProps>(
    ({ size = 'md', icon, className, ...props }, ref) => {
        return (
            <CheckboxPrimitive.Root ref={ref} className={cn(CheckboxVariants({ size, className }))} {...props}>
                <CheckboxPrimitive.Indicator className={cn('group flex items-center justify-center text-current')}>
                    {icon ? (
                        icon
                    ) : (
                        <>
                            <Check
                                className='hidden h-full w-full group-data-[state=checked]:block'
                                absoluteStrokeWidth={true}
                                strokeWidth={3}
                            />
                            <Minus
                                className='hidden h-full w-full group-data-[state=indeterminate]:block'
                                absoluteStrokeWidth={true}
                                strokeWidth={3}
                            />
                        </>
                    )}
                </CheckboxPrimitive.Indicator>
            </CheckboxPrimitive.Root>
        );
    }
);
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

type CheckboxWithLabelProps = React.ComponentPropsWithoutRef<typeof Checkbox> & {
    label: React.ReactNode;
    error?: boolean;
    labelClassName?: string;
    fieldClassName?: string;
};

const CheckboxWithLabel = React.forwardRef<React.ElementRef<typeof Checkbox>, CheckboxWithLabelProps>(
    ({ id, label, error = false, disabled, className, labelClassName, fieldClassName, ...props }, ref) => {
        const generatedId = React.useId();
        const checkboxId = id ?? generatedId;

        return (
            <div
                className={cn(
                    'inline-flex items-center gap-2 rounded-sm',
                    '[&:has(:focus-visible)]:ring-2 [&:has(:focus-visible)]:ring-secondary',
                    'dark:[&:has(:focus-visible)]:ring-secondary-variant-2',
                    '[&:has(:focus-visible)]:ring-offset-2 [&:has(:focus-visible)]:ring-offset-neutral-light-1',
                    'dark:[&:has(:focus-visible)]:ring-offset-neutral-dark-1',
                    fieldClassName
                )}>
                <Checkbox
                    ref={ref}
                    id={checkboxId}
                    disabled={disabled}
                    aria-invalid={error || props['aria-invalid']}
                    className={cn('focus-visible:ring-0 focus-visible:ring-offset-0', className)}
                    {...props}
                />

                <Label
                    htmlFor={checkboxId}
                    className={cn(
                        'cursor-pointer text-base font-normal leading-[18px]',
                        error && 'text-error',
                        disabled && 'cursor-not-allowed',
                        labelClassName
                    )}>
                    {label}
                </Label>
            </div>
        );
    }
);

export { Checkbox, CheckboxWithLabel };
