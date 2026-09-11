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
import { cva, VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { Label } from '../Label';
import { cn } from '../utils';
import { Typography } from '../Typography';

export const InputVariants = cva(
    'flex h-10 w-full text-base text-main placeholder:text-input-placeholder-text disabled:cursor-not-allowed disabled:opacity-50 file:border-0 file:bg-transparent file:pr-3 file:text-sm file:font-medium file:text-main file:cursor-pointer',
    {
        variants: {
            variant: {
                outlined:
                    'rounded-md border border-input-outlined-border-default bg-input-fill px-3 py-2 text-sm enabled:hover:border-input-outlined-border-hover focus:outline-none focus:focus-ring dark:focus:border-input-outlined-border-default disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-text-disabled disabled:placeholder:text-text-disabled disabled:opacity-100 aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main dark:aria-[invalid=true]:focus:border-status-error-main',
                underlined:
                    'rounded-sm bg-transparent border-b-input-border-default border-b hover:border-b-2 focus:outline-none focus:focus-ring focus:border-input-border-focus focus:border-b-2',
            },
            intent: {
                time: 'appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none',
            },
        },
        defaultVariants: {
            variant: 'underlined',
        },
    }
);

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement>, VariantProps<typeof InputVariants> {
    label?: React.ReactNode;
    error?: boolean;
    helperText?: React.ReactNode;
    errorMessage?: React.ReactNode;
    fieldClassName?: string;
    labelClassName?: string;
    helperTextClassName?: string;
    errorMessageClassName?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
    (
        {
            id,
            label,
            required = false,
            error = false,
            helperText,
            errorMessage,
            variant,
            intent,
            placeholder,
            className,
            fieldClassName,
            labelClassName,
            helperTextClassName,
            errorMessageClassName,
            'aria-describedby': ariaDescribedBy,
            'aria-invalid': ariaInvalid,
            'aria-errormessage': ariaErrorMessage,
            ...props
        },
        ref
    ) => {
        // need ID for fields with labels so screen readers have correct accessible name
        const generatedId = React.useId();
        const isComposed = !!label;
        const inputId = isComposed ? id ?? generatedId : id;
        // generate IDs for supporting text and combine them with any caller-provided IDs so assistive technology can describe the input.
        const helperTextId = `${inputId}-helper-text`;
        const errorMessageId = `${inputId}-error-message`;
        const describedBy = [
            ariaDescribedBy,
            helperText ? helperTextId : undefined,
            error && errorMessage ? errorMessageId : undefined,
        ]
            .filter(Boolean)
            .join(' ');

        const input = (
            <input
                {...props}
                ref={ref}
                id={inputId}
                required={required}
                placeholder={required && !isComposed && placeholder ? `${placeholder} *` : placeholder}
                className={cn(InputVariants({ variant, intent }), className)}
                aria-describedby={isComposed ? describedBy || undefined : ariaDescribedBy}
                aria-invalid={!!(isComposed && error) || ariaInvalid}
                aria-errormessage={isComposed && error && errorMessage ? errorMessageId : ariaErrorMessage}
            />
        );

        if (!isComposed) {
            return input;
        }

        return (
            <div className={cn('grid gap-1.5', fieldClassName)}>
                <Label htmlFor={inputId} className={labelClassName}>
                    {label}{' '}
                    {required && (
                        <span aria-hidden='true' className='text-status-error-main'>
                            *
                        </span>
                    )}
                </Label>
                {input}
                {helperText && (
                    <Typography variant='body2' id={helperTextId} className={cn('text-main', helperTextClassName)}>
                        {helperText}
                    </Typography>
                )}
                {error && errorMessage && (
                    <Typography
                        variant='body2'
                        id={errorMessageId}
                        role='alert'
                        className={cn('text-error', errorMessageClassName)}>
                        {errorMessage}
                    </Typography>
                )}
            </div>
        );
    }
);
Input.displayName = 'Input';

export { Input };
