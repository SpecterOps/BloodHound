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
import type { VariantProps } from 'class-variance-authority';
import { Check, Minus } from 'lucide-react';
import * as React from 'react';
import { Label } from '../Label';
import { Typography } from '../Typography';
import { cn } from '../utils';
import { CheckboxVariants } from './Checkbox.styles';

interface CheckboxProps
    extends React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root>,
        VariantProps<typeof CheckboxVariants> {
    icon?: React.ReactNode;
}

const Checkbox = React.forwardRef<React.ElementRef<typeof CheckboxPrimitive.Root>, CheckboxProps>(
    ({ size, focusRing, icon, className, ...props }, ref) => {
        return (
            <CheckboxPrimitive.Root
                ref={ref}
                className={cn(CheckboxVariants({ size, focusRing, className }))}
                {...props}>
                <CheckboxPrimitive.Indicator className='group flex items-center justify-center text-current'>
                    {/*
    Both icons are rendered with `display: none` by default. When the indicator's `data-state` is
    set to `checked` or `indeterminate`, the appropriate icon changes to `display: block`. With
    this pattern, the checkbox supports uncontrolled usage.
*/}
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

interface CheckboxWithLabelProps extends React.ComponentPropsWithoutRef<typeof Checkbox> {
    label: React.ReactNode;
    error?: boolean;
    labelClassName?: string;
    fieldClassName?: string;
}

const CheckboxWithLabel = React.forwardRef<React.ElementRef<typeof Checkbox>, CheckboxWithLabelProps>(
    (
        {
            id,
            label,
            error = false,
            disabled,
            className,
            labelClassName,
            fieldClassName,
            'aria-invalid': ariaInvalid,
            ...props
        },
        ref
    ) => {
        const generatedId = React.useId();
        const checkboxId = id ?? generatedId;

        return (
            <Label
                htmlFor={checkboxId}
                className={cn(
                    'inline-flex items-center gap-2 rounded-sm',
                    '[&:has(:focus-visible)]:ring-2 [&:has(:focus-visible)]:ring-secondary',
                    'dark:[&:has(:focus-visible)]:ring-secondary-variant-2',
                    '[&:has(:focus-visible)]:ring-offset-2',
                    '[&:has(:focus-visible)]:ring-offset-checkbox-unchecked-fill',
                    fieldClassName
                )}>
                <Checkbox
                    {...props}
                    ref={ref}
                    id={checkboxId}
                    disabled={disabled}
                    aria-invalid={error ? true : ariaInvalid}
                    focusRing={false}
                    className={className}
                />
                <Typography
                    variant='caption'
                    className={cn(
                        'cursor-pointer font-normal leading-[18px]',
                        error && 'peer-data-[state=unchecked]:text-status-error-main',
                        disabled && 'cursor-not-allowed',
                        labelClassName
                    )}>
                    {label}
                </Typography>
            </Label>
        );
    }
);
CheckboxWithLabel.displayName = 'CheckboxWithLabel';

export { Checkbox, CheckboxWithLabel };
export type { CheckboxProps };
