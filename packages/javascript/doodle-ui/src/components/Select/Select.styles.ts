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
        'bg-[color:var(--select-trigger-fill)]',
        'placeholder:text-[color:var(--select-trigger-placeholder-text)] data-[placeholder]:text-[color:var(--select-placeholder-text)]',
        'focus:outline-none focus-visible:focus-ring data-[state=open]:focus-ring',
        'disabled:cursor-not-allowed disabled:opacity-50',
        '[&>span]:line-clamp-1 [&[data-state=open]>svg]:rotate-180',
    ],
    {
        variants: {
            variant: {
                outlined: [
                    'rounded-md border border-[color:var(--input-outlined-border-default)] bg-[color:var(--input-fill)] px-3 py-2 text-sm',
                    'enabled:hover:border-[color:var(--input-outlined-border-hover)] enabled:hover:bg-[color:var(--input-outlined-border-hover)] enabled:hover:text-[color:var(--select-highlight-text)] data-[placeholder]:enabled:hover:text-[color:var(--select-highlight-text)]',
                    'focus:outline-none focus:focus-ring focus:border-[color:var(--select-outlined-focus-border)] enabled:hover:focus:border-[color:var(--select-outlined-hover-focus-border)]',
                    'disabled:border-[color:var(--input-border-disabled)] disabled:bg-[color:var(--select-disabled-fill)] disabled:text-disabled disabled:placeholder:text-disabled disabled:opacity-100',
                    'aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main aria-[invalid=true]:enabled:hover:focus:border-status-error-main',
                    'aria-[invalid=true]:data-[placeholder]:text-[color:var(--input-placeholder-text)] aria-[invalid=true]:data-[placeholder]:enabled:hover:text-[color:var(--select-highlight-text)]',
                    '[&[aria-invalid=true]:not([data-placeholder])>span]:text-status-error-main [&[aria-invalid=true]:not([data-placeholder]):enabled:hover>span]:text-[color:var(--select-highlight-text)]',
                    'data-[state=open]:border-primary data-[state=open]:bg-[color:var(--select-open-fill)] data-[state=open]:text-[color:var(--select-highlight-text)] data-[placeholder]:data-[state=open]:text-[color:var(--select-highlight-text)]',
                    'data-[state=open]:enabled:hover:border-primary data-[state=open]:enabled:hover:bg-primary',
                    '[&[data-state=open]>span]:text-[color:var(--select-highlight-text)] [&[aria-invalid=true][data-state=open]:not([data-placeholder])>span]:text-[color:var(--select-highlight-text)]',
                    'bg-[color:var(--select-trigger-outlined-fill)]',
                    'disabled:data-[state=open]:bg-[color:var(--select-open-disabled-fill)] disabled:data-[placeholder]:text-[color:var(--select-disabled-placeholder-text)] disabled:data-[placeholder]:data-[state=open]:text-[color:var(--select-highlight-text)]',
                    'aria-[invalid=true]:data-[placeholder]:data-[state=open]:text-[color:var(--select-invalid-placeholder-open-text)]',
                    'data-[state=open]:focus:border-[color:var(--select-open-focus-border)] aria-[invalid=true]:data-[state=open]:focus:border-status-error-main',
                    'aria-[invalid=true]:data-[state=open]:enabled:hover:focus:border-[color:var(--select-open-invalid-hover-focus-border)]',
                ],
                underlined: [
                    'rounded-sm border-b border-b-[color:var(--select-border-default)] bg-transparent ring-none',
                    'hover:border-b-2',
                    'focus-visible:border-b-2 focus-visible:border-[color:var(--select-border-focus)]',
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
    'border-[color:var(--select-content-border)] bg-[color:var(--select-content-fill)] text-main shadow-outer-2',
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
    'data-[highlighted]:bg-secondary data-[highlighted]:text-[color:var(--select-highlight-text)]',
    'data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)]',
    'data-[disabled]:pointer-events-none data-[disabled]:bg-dropdown-option-disabled-fill',
    'data-[disabled]:text-disabled data-[disabled]:opacity-100',
    'data-[state=checked]:font-bold data-[state=checked]:text-[color:var(--select-item-checked-text)]',
    'data-[highlighted]:data-[state=checked]:text-[color:var(--select-highlight-text)]',
];

export const selectSeparatorClasses = '-mx-1 my-1 h-px bg-[color:var(--select-separator-fill)]';
