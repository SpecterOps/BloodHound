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
        'flex h-10 w-full items-center justify-between rounded-lg bg-select-trigger-fill p-2',
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
                    'disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-text-disabled disabled:placeholder:text-text-disabled disabled:opacity-100 dark:disabled:bg-input-fill-disabled-dark',
                    'aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main dark:aria-[invalid=true]:focus:border-status-error-main',
                    'aria-[invalid=true]:data-[placeholder]:text-input-placeholder-text aria-[invalid=true]:data-[placeholder]:enabled:hover:text-common-white dark:aria-[invalid=true]:data-[placeholder]:enabled:hover:text-common-dark',
                    '[&[aria-invalid=true]:not([data-placeholder])>span]:text-status-error-main [&[aria-invalid=true]:not([data-placeholder]):enabled:hover>span]:text-common-white dark:[&[aria-invalid=true]:not([data-placeholder]):enabled:hover>span]:text-common-dark',
                    'data-[state=open]:border-primary data-[state=open]:bg-primary data-[state=open]:text-common-white data-[state=open]:enabled:hover:border-primary data-[state=open]:enabled:hover:bg-primary',
                    '[&[data-state=open]>span]:text-common-white [&[aria-invalid=true][data-state=open]:not([data-placeholder])>span]:text-common-white',
                    'dark:data-[state=open]:text-common-dark dark:data-[placeholder]:data-[state=open]:text-common-dark dark:[&[data-state=open]>span]:text-common-dark dark:[&[aria-invalid=true][data-state=open]:not([data-placeholder])>span]:text-common-dark',
                    'bg-select-trigger-outlined-fill',
                ],
                underlined:
                    'rounded-sm ring-none bg-transparent border-b-select-border-default border-b hover:border-b-2 focus-visible:border-select-border-focus focus-visible:border-b-2',
            },
        },
        defaultVariants: {
            variant: 'outlined',
        },
    }
);
