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

export const SelectTriggerVariants = cva(
    [
        'flex h-10 w-full items-center justify-between rounded-lg p-2',
        'bg-select-trigger-fill',
        'placeholder:text-select-trigger-placeholder-text dark:data-[placeholder]:text-select-trigger-placeholder-text',
        'focus:outline-none focus-visible:focus-ring data-[state=open]:focus-ring',
        'disabled:cursor-not-allowed disabled:opacity-50',
        '[&>span]:line-clamp-1 [&[data-state=open]>svg]:rotate-180',
    ],
    {
        variants: {
            variant: {
                outlined: [
                    'rounded-md border border-input-outlined-border-default bg-input-fill px-3 py-2 text-sm dark:bg-input-fill-dark',
                    'enabled:hover:border-input-outlined-border-hover enabled:hover:bg-input-outlined-border-hover enabled:hover:text-common-white',
                    'dark:enabled:hover:text-common-dark dark:data-[placeholder]:enabled:hover:text-common-dark',
                    'focus:outline-none focus:focus-ring dark:focus:border-input-outlined-border-default',
                    'disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-disabled disabled:placeholder:text-disabled disabled:opacity-100 dark:disabled:bg-input-fill-disabled-dark',
                    'aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main',
                    'dark:aria-[invalid=true]:focus:border-status-error-main',
                    'aria-[invalid=true]:data-[placeholder]:text-input-placeholder-text aria-[invalid=true]:data-[placeholder]:enabled:hover:text-common-white dark:aria-[invalid=true]:data-[placeholder]:enabled:hover:text-common-dark',
                    '[&[aria-invalid=true]:not([data-placeholder])>span]:text-status-error-main [&[aria-invalid=true]:not([data-placeholder]):enabled:hover>span]:text-common-white dark:[&[aria-invalid=true]:not([data-placeholder]):enabled:hover>span]:text-common-dark',
                    'data-[state=open]:border-primary data-[state=open]:bg-primary data-[state=open]:text-common-white',
                    'data-[state=open]:enabled:hover:border-primary data-[state=open]:enabled:hover:bg-primary',
                    '[&[data-state=open]>span]:text-common-white [&[aria-invalid=true][data-state=open]:not([data-placeholder])>span]:text-common-white',
                    'dark:data-[state=open]:text-common-dark dark:data-[placeholder]:data-[state=open]:text-common-dark dark:[&[data-state=open]>span]:text-common-dark dark:[&[aria-invalid=true][data-state=open]:not([data-placeholder])>span]:text-common-dark',
                    'bg-select-trigger-outlined-fill',
                ],
                underlined: [
                    'rounded-sm border-b border-b-select-border-default bg-transparent ring-none',
                    'hover:border-b-2',
                    'focus-visible:border-b-2 focus-visible:border-select-border-focus',
                ],
            },
        },
        defaultVariants: {
            variant: 'outlined',
        },
    }
);

export const selectIconClasses = 'h-3 w-3';

export const selectLoadingIconClasses = [selectIconClasses, 'animate-spin motion-reduce:animate-none'];

export const selectScrollButtonClasses = 'flex cursor-default items-center justify-center py-1';

export const selectContentClasses = [
    'relative z-[1500] max-h-96 overflow-hidden rounded-lg border',
    'border-select-content-border bg-select-content-fill text-main shadow-outer-2',
    'data-[state=open]:animate-in data-[state=closed]:animate-out',
    'data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0',
    'data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95',
    'data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2',
    'data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2',
];

export const selectPopperContentClasses = [
    'data-[side=bottom]:-translate-y-1 data-[side=top]:-translate-y-1',
    'data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1',
];

export const selectViewportClasses = 'p-1';

export const selectPopperViewportClasses = [
    'h-[var(--radix-select-trigger-height)]',
    'w-full min-w-[var(--radix-select-trigger-width)]',
];

export const selectLabelClasses = 'py-1.5 pl-8 pr-2 font-semibold';

export const selectItemClasses = [
    'relative flex w-full cursor-default select-none items-center rounded-lg p-2 outline-none',
    'data-[highlighted]:bg-secondary data-[highlighted]:text-common-white',
    'data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)]',
    'dark:data-[highlighted]:text-common-dark',
    'data-[disabled]:pointer-events-none data-[disabled]:bg-dropdown-option-disabled-fill',
    'data-[disabled]:text-disabled data-[disabled]:opacity-100',
    'data-[state=checked]:font-bold data-[state=checked]:text-select-item-checked-text',
    'data-[highlighted]:data-[state=checked]:text-common-white',
    'dark:data-[highlighted]:data-[state=checked]:text-common-dark',
];

export const selectSeparatorClasses = '-mx-1 my-1 h-px bg-select-separator-fill';
