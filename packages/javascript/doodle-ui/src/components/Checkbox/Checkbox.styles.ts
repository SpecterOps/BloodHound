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

export const CheckboxVariants = cva(
    cn(
        'peer shrink-0 rounded-sm border border-[color:var(--checkbox-border)] ring-offset-background',
        'enabled:hover:border-[color:var(--checkbox-hover)] enabled:active:border-primary-variant',
        'focus:outline-none',
        'data-[state=checked]:bg-[color:var(--checkbox-checked-fill)] data-[state=checked]:text-[color:var(--checkbox-check)]',
        'enabled:hover:data-[state=checked]:border-[color:var(--checkbox-hover)] enabled:hover:data-[state=checked]:bg-[color:var(--checkbox-hover)] enabled:active:data-[state=checked]:bg-primary-variant',
        'enabled:hover:data-[state=indeterminate]:border-[color:var(--checkbox-hover)] enabled:hover:data-[state=indeterminate]:bg-secondary',
        'enabled:data-[state=checked]:border-primary enabled:data-[state=checked]:bg-primary enabled:data-[state=checked]:text-[color:var(--checkbox-check)]',
        'enabled:data-[state=indeterminate]:border-primary enabled:data-[state=indeterminate]:bg-primary enabled:data-[state=indeterminate]:text-[color:var(--checkbox-check)]',
        'disabled:cursor-not-allowed disabled:border-[color:var(--input-border-disabled)] disabled:bg-[color:var(--checkbox-disabled-fill)] disabled:text-icon-disabled',
        'enabled:active:data-[state=checked]:border-[color:var(--checkbox-checked-active-border)] enabled:active:data-[state=indeterminate]:border-[color:var(--checkbox-checked-active-border)]',
        'enabled:hover:active:data-[state=checked]:border-[color:var(--checkbox-hover)] enabled:hover:active:data-[state=indeterminate]:border-[color:var(--checkbox-hover)]',
        'disabled:hover:data-[state=checked]:bg-[color:var(--checkbox-disabled-checked-hover-fill)]',
        'enabled:aria-[invalid=true]:data-[state=unchecked]:border-status-error-main'
    ),
    {
        variants: {
            size: {
                lg: 'size-[24px]',
                md: 'size-[18px]',
                sm: 'size-[12px]',
            },
            focusRing: {
                true: 'focus-visible:focus-ring',
                false: '',
            },
        },
        defaultVariants: {
            size: 'md',
            focusRing: true,
        },
    }
);

export const checkboxIndicatorClasses = 'group flex items-center justify-center text-current';

export const checkboxCheckIconClasses = 'hidden h-full w-full group-data-[state=checked]:block';

export const checkboxIndeterminateIconClasses = 'hidden h-full w-full group-data-[state=indeterminate]:block';

export const checkboxLabelClasses = cn(
    'inline-flex items-center gap-2 rounded-sm',
    '[&:has(:focus-visible)]:ring-2 [&:has(:focus-visible)]:ring-[color:var(--checkbox-focus-ring)]',
    '[&:has(:focus-visible)]:ring-offset-2',
    '[&:has(:focus-visible)]:ring-offset-[color:var(--checkbox-unchecked-fill)]'
);
