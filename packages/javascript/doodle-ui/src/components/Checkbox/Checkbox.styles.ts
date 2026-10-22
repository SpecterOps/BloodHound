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
import { common, dark, light } from '../../tailwind/colors';

export const checkboxBorderColors = {
    'checkbox-border': 'var(--border-control)',
    'checkbox-hover': light.secondary.main,
    'checkbox-hover-dark': dark.badge.secondary.outline,
};

export const checkboxTextColors = {
    'checkbox-border': 'var(--border-control)',
    'checkbox-check': common.white,
    'checkbox-check-dark': common.dark,
};

export const checkboxBackgroundColors = {
    'checkbox-hover': light.secondary.main,
    'checkbox-hover-dark': dark.badge.secondary.outline,
    'checkbox-unchecked-fill': 'var(--surface)',
    'checkbox-fill': common.dark,
    'checkbox-fill-dark': common.white,
};

export const checkboxRingOffsetColors = {
    'checkbox-unchecked-fill': 'var(--surface)',
};

export const CheckboxVariants = cva(
    [
        'peer shrink-0 rounded-sm border border-checkbox-border ring-offset-background',
        'enabled:hover:border-checkbox-hover dark:enabled:hover:border-checkbox-hover-dark enabled:active:border-primary-variant',
        'dark:border-checkbox-border dark:hover:border-checkbox-hover-dark dark:active:border-primary-variant',
        'focus:outline-none',
        'data-[state=checked]:bg-checkbox-fill dark:data-[state=checked]:bg-checkbox-fill-dark data-[state=checked]:text-checkbox-check dark:data-[state=checked]:text-checkbox-check-dark dark:data-[state=checked]:bg-primary',
        'enabled:hover:data-[state=checked]:border-checkbox-hover dark:enabled:hover:data-[state=checked]:border-checkbox-hover-dark enabled:hover:data-[state=checked]:bg-checkbox-hover dark:enabled:hover:data-[state=checked]:bg-checkbox-hover-dark enabled:active:data-[state=checked]:bg-primary-variant',
        'dark:data-[state=checked]:hover:bg-checkbox-hover-dark dark:data-[state=checked]:active:bg-primary-variant',
        'enabled:hover:data-[state=indeterminate]:border-checkbox-hover dark:enabled:hover:data-[state=indeterminate]:border-checkbox-hover-dark enabled:hover:data-[state=indeterminate]:bg-secondary',
        'enabled:data-[state=checked]:border-primary enabled:data-[state=checked]:bg-primary enabled:data-[state=checked]:text-checkbox-check dark:enabled:data-[state=checked]:text-checkbox-check-dark',
        'enabled:data-[state=indeterminate]:border-primary enabled:data-[state=indeterminate]:bg-primary enabled:data-[state=indeterminate]:text-checkbox-check dark:enabled:data-[state=indeterminate]:text-checkbox-check-dark',
        'disabled:cursor-not-allowed disabled:border-input-border-disabled dark:disabled:border-input-border-disabled-dark disabled:bg-input-fill-disabled disabled:text-icon-disabled',
        'dark:disabled:bg-input-fill-disabled-dark dark:disabled:border-input-border-disabled-dark dark:hover:disabled:border-input-border-disabled-dark',
        'enabled:aria-[invalid=true]:data-[state=unchecked]:border-status-error-main',
    ],
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
export const checkboxCheckedIconClasses = 'hidden h-full w-full group-data-[state=checked]:block';
export const checkboxIndeterminateIconClasses = 'hidden h-full w-full group-data-[state=indeterminate]:block';

export const checkboxLabelClasses = [
    'inline-flex items-center gap-2 rounded-sm',
    '[&:has(:focus-visible)]:ring-2 [&:has(:focus-visible)]:ring-secondary',
    'dark:[&:has(:focus-visible)]:ring-secondary-variant-2',
    '[&:has(:focus-visible)]:ring-offset-2',
    '[&:has(:focus-visible)]:ring-offset-checkbox-unchecked-fill',
];
