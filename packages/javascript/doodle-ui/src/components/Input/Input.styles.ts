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
import { cva } from 'class-variance-authority';
import { invalidControlClasses } from '../formControl.styles';
import { cn } from '../utils';

const inputBaseClasses = cn(
    'flex h-10 w-full',
    'text-base text-main placeholder:text-[color:var(--input-placeholder-text)]',
    'disabled:cursor-not-allowed disabled:opacity-50',
    'file:cursor-pointer file:border-0 file:bg-transparent file:pr-3 file:text-sm file:font-medium file:text-main'
);

export const InputVariants = cva(inputBaseClasses, {
    variants: {
        variant: {
            outlined: cn(
                'rounded-md border border-[color:var(--input-outlined-border-default)] bg-[color:var(--input-fill)] px-3 py-2 text-sm',
                'enabled:hover:border-[color:var(--input-outlined-border-hover)]',
                'focus:outline-none focus:focus-ring focus:border-[color:var(--input-outlined-border-default)] enabled:hover:focus:border-[color:var(--input-outlined-hover-focus-border)]',
                'disabled:border-[color:var(--input-border-disabled)] disabled:bg-[color:var(--input-disabled-fill)] disabled:text-disabled disabled:placeholder:text-disabled disabled:opacity-100',
                invalidControlClasses,
                'enabled:aria-[invalid=true]:border-status-error-main enabled:aria-[invalid=true]:hover:border-status-error-main enabled:aria-[invalid=true]:focus:border-status-error-main aria-[invalid=true]:enabled:hover:focus:border-status-error-main disabled:aria-[invalid=true]:border-[color:var(--input-disabled-invalid-border)]'
            ),
            underlined: cn(
                'rounded-sm border-b border-b-[color:var(--input-border-default)] bg-transparent px-1',
                'hover:border-b-2',
                'focus:border-b-2 focus:border-[color:var(--input-border-focus)] focus:outline-none focus:focus-ring',
                // Border images ignore border-radius, keeping the underline flat without changing the focus outline.
                '[border-image:linear-gradient(var(--input-border-default),var(--input-border-default))_1]',
                'focus:[border-image:linear-gradient(var(--input-border-focus),var(--input-border-focus))_1]'
            ),
        },
        intent: {
            time: cn(
                'appearance-none',
                '[&::-webkit-calendar-picker-indicator]:hidden',
                '[&::-webkit-calendar-picker-indicator]:appearance-none'
            ),
        },
    },
    defaultVariants: {
        variant: 'underlined',
    },
});

export const inputStyles = {
    field: 'grid gap-1.5',
    requiredIndicator: 'text-status-error-main',
    helperText: 'text-main',
    errorMessage: 'text-error',
};
