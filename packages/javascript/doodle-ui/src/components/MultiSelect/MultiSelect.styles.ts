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
import { cn } from '../utils';

export const MultiSelectTriggerVariants = cva(
    cn(
        'flex h-10 w-full items-center justify-between rounded px-[14px] py-2 text-base font-normal leading-6 tracking-[0.15px]',
        'focus:outline-none focus-visible:focus-ring',
        'enabled:hover:bg-secondary enabled:hover:text-inverse',
        'data-[state=open]:bg-primary data-[state=open]:text-inverse',
        'disabled:cursor-not-allowed disabled:border disabled:border-input-border-disabled disabled:bg-input-fill-disabled dark:disabled:bg-input-fill-disabled-dark disabled:text-disabled',
        'aria-[invalid=true]:[&>svg]:text-main aria-[invalid=true]:border aria-[invalid=true]:border-status-error-main aria-[invalid=true]:bg-select-trigger-outlined-fill aria-[invalid=true]:text-input-placeholder-text aria-[invalid=true]:enabled:hover:border-status-error-main aria-[invalid=true]:enabled:hover:bg-select-trigger-outlined-fill aria-[invalid=true]:data-[state=open]:bg-select-trigger-outlined-fill aria-[invalid=true]:enabled:hover:text-input-placeholder-text aria-[invalid=true]:data-[state=open]:text-input-placeholder-text aria-[invalid=true]:[&:enabled:hover>svg]:text-main aria-[invalid=true]:[&[data-state=open]>svg]:text-main'
    ),
    {
        variants: {
            variant: {
                outlined: cn(
                    'ring-1 ring-input-border-default bg-select-trigger-outlined-fill text-input-placeholder-text',
                    '[&>svg]:text-main',
                    '[&:enabled:hover>svg]:text-inverse',
                    '[&[data-state=open]>svg]:text-inverse'
                ),
                filled: 'bg-primary text-inverse',
            },
        },
        defaultVariants: {
            variant: 'filled',
        },
    }
);

export const multiSelectRowStyles = 'flex w-full items-center gap-2 rounded-lg p-2';

export const multiSelectInteractiveRowStyles = 'group cursor-pointer hover:bg-secondary hover:text-inverse';

export const multiSelectCheckboxStyles = cn(
    'enabled:data-[state=unchecked]:!border-current',
    'group-hover:enabled:!border-current group-hover:enabled:data-[state=checked]:!bg-transparent group-hover:enabled:data-[state=checked]:!text-inverse group-hover:enabled:data-[state=indeterminate]:!bg-transparent group-hover:enabled:data-[state=indeterminate]:!text-inverse'
);

export const multiSelectIndeterminateCheckboxStyles =
    'enabled:data-[state=indeterminate]:!border-checkbox-border enabled:data-[state=indeterminate]:!bg-checkbox-unchecked-fill enabled:data-[state=indeterminate]:!text-checkbox-border';

export const multiSelectCaretClasses = 'ml-2 shrink-0 transition-transform duration-200';

export const multiSelectLoadingRowsClasses = 'flex items-center gap-2 mx-1 p-2 py-2';

export const multiSelectStateRowClasses = 'px-3 py-2 text-center text-main';

export const multiSelectContentClasses =
    'w-[var(--radix-popover-trigger-width)] rounded-lg bg-select-content-fill p-0 text-main';

export const multiSelectSearchRowClasses = 'flex items-center gap-2 p-2';

export const multiSelectSearchIconClasses = 'size-4 shrink-0 text-main';

export const multiSelectSearchInputClasses = cn(
    'h-6 border-none bg-transparent px-2 text-main leading-4',
    'placeholder:text-input-placeholder-text',
    'focus-visible:outline-none focus-visible:ring-0 focus-visible:ring-offset-0'
);

export const multiSelectScrollAreaClasses =
    'max-h-60 [--multiselect-scrollbar-thumb:var(--neutral-400)] dark:[--multiselect-scrollbar-thumb:var(--input-placeholder-text)] [&_[data-radix-scroll-area-viewport]]:max-h-60';
