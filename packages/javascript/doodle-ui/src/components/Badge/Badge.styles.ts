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
import { dark, light } from '../../tailwind/colors';

export const badgeColors = {
    'badge-primary-surface': light.badge.primary.fill,
    'badge-primary-surface-dark': dark.badge.primary.fill,
    'badge-primary-accent': light.badge.primary.outline,
    'badge-primary-accent-dark': dark.badge.primary.outline,
    'badge-secondary-surface': light.badge.secondary.fill,
    'badge-secondary-surface-dark': dark.badge.secondary.fill,
    'badge-secondary-accent': light.badge.secondary.outline,
    'badge-secondary-accent-dark': dark.badge.secondary.outline,
    'badge-grey-surface': light.badge.grey.fill,
    'badge-grey-surface-dark': dark.badge.grey.fill,
    'badge-grey-accent': light.badge.grey.outline,
    'badge-grey-accent-dark': dark.badge.grey.outline,
    'badge-red-surface': light.badge.red.fill,
    'badge-red-surface-dark': dark.badge.red.fill,
    'badge-red-accent': light.badge.red.outline,
    'badge-red-accent-dark': dark.badge.red.outline,
    'badge-orange-surface': light.badge.orange.fill,
    'badge-orange-surface-dark': dark.badge.orange.fill,
    'badge-orange-accent': light.badge.orange.outline,
    'badge-orange-accent-dark': dark.badge.orange.outline,
    'badge-green-surface': light.badge.green.fill,
    'badge-green-surface-dark': dark.badge.green.fill,
    'badge-green-accent': light.badge.green.outline,
    'badge-green-accent-dark': dark.badge.green.outline,
    'badge-blue-surface': light.badge.blue.fill,
    'badge-blue-surface-dark': dark.badge.blue.fill,
    'badge-blue-accent': light.badge.blue.outline,
    'badge-blue-accent-dark': dark.badge.blue.outline,
};

const badgeBaseClasses = [
    'inline-flex min-w-16 items-center justify-center',
    'rounded px-2 py-1',
    'text-base text-main',
];

export const BadgeVariants = cva(badgeBaseClasses, {
    variants: {
        variant: {
            fill: '',
            outline: 'border bg-badge-grey-surface dark:bg-badge-grey-surface-dark',
        },
        color: {
            primary: '',
            secondary: '',
            grey: '',
            red: '',
            orange: '',
            green: '',
            blue: '',
            error: '',
            warning: '',
            success: '',
            disabled: '',
        },
        iconPosition: {
            left: 'flex-row',
            right: 'flex-row-reverse',
        },
        hasIcon: {
            true: 'gap-2',
            false: '',
        },
    },
    compoundVariants: [
        // Fill variants
        {
            variant: 'fill',
            color: 'primary',
            className: 'bg-badge-primary-surface dark:bg-badge-primary-surface-dark',
        },
        {
            variant: 'fill',
            color: 'secondary',
            className: 'bg-badge-secondary-surface dark:bg-badge-secondary-surface-dark',
        },
        {
            variant: 'fill',
            color: 'grey',
            className: 'bg-badge-grey-surface dark:bg-badge-grey-surface-dark',
        },
        {
            variant: 'fill',
            color: 'red',
            className: 'bg-badge-red-surface dark:bg-badge-red-surface-dark',
        },
        {
            variant: 'fill',
            color: 'orange',
            className: 'bg-badge-orange-surface dark:bg-badge-orange-surface-dark',
        },
        {
            variant: 'fill',
            color: 'green',
            className: 'bg-badge-green-surface dark:bg-badge-green-surface-dark',
        },
        {
            variant: 'fill',
            color: 'blue',
            className: 'bg-badge-blue-surface dark:bg-badge-blue-surface-dark',
        },
        {
            variant: 'fill',
            color: 'error',
            className: 'bg-status-error-main dark:bg-status-error-fill',
        },
        {
            variant: 'fill',
            color: 'warning',
            className: 'bg-status-warning-main dark:bg-status-warning-fill',
        },
        {
            variant: 'fill',
            color: 'success',
            className: 'bg-status-success-main dark:bg-status-success-fill',
        },
        { variant: 'fill', color: 'disabled', className: 'bg-neutral-500' },

        // Outline variants
        {
            variant: 'outline',
            color: 'primary',
            className: [
                'border-badge-primary-accent dark:border-badge-primary-accent-dark',
                '[&>span>svg]:text-badge-primary-accent dark:[&>span>svg]:text-badge-primary-accent-dark',
            ],
        },
        {
            variant: 'outline',
            color: 'secondary',
            className: [
                'border-badge-secondary-accent dark:border-badge-secondary-accent-dark',
                '[&>span>svg]:text-badge-secondary-accent dark:[&>span>svg]:text-badge-secondary-accent-dark',
            ],
        },
        {
            variant: 'outline',
            color: 'grey',
            className: ['border-badge-grey-accent dark:border-badge-grey-accent-dark', '[&>span>svg]:text-main'],
        },
        {
            variant: 'outline',
            color: 'red',
            className: [
                'border-badge-red-accent dark:border-badge-red-accent-dark',
                '[&>span>svg]:text-badge-red-accent dark:[&>span>svg]:text-badge-red-accent-dark',
            ],
        },
        {
            variant: 'outline',
            color: 'orange',
            className: [
                'border-badge-orange-accent dark:border-badge-orange-accent-dark',
                '[&>span>svg]:text-badge-orange-accent dark:[&>span>svg]:text-badge-orange-accent-dark',
            ],
        },
        {
            variant: 'outline',
            color: 'green',
            className: [
                'border-badge-green-accent dark:border-badge-green-accent-dark',
                '[&>span>svg]:text-badge-green-accent dark:[&>span>svg]:text-badge-green-accent-dark',
            ],
        },
        {
            variant: 'outline',
            color: 'blue',
            className: [
                'border-badge-blue-accent dark:border-badge-blue-accent-dark',
                '[&>span>svg]:text-badge-blue-accent dark:[&>span>svg]:text-badge-blue-accent-dark',
            ],
        },
        {
            variant: 'outline',
            color: 'error',
            className: 'border-status-error-main [&>span>svg]:text-status-error-main',
        },
        {
            variant: 'outline',
            color: 'warning',
            className: 'border-status-warning-main [&>span>svg]:text-status-warning-main',
        },
        {
            variant: 'outline',
            color: 'success',
            className: 'border-status-success-main [&>span>svg]:text-status-success-main',
        },
        {
            variant: 'outline',
            color: 'disabled',
            className: 'border-neutral-500 [&>span>svg]:text-neutral-500',
        },
    ],
    defaultVariants: {
        variant: 'outline',
        color: 'grey',
    },
});

export const badgeIconClasses = ['badge-icon flex shrink-0 items-center', '[&>svg]:h-3 [&>svg]:w-3'];

export const badgeLabelClasses = 'badge-label translate-y-[1px]';
