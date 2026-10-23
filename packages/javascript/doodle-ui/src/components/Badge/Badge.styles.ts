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

const badgeBaseClasses = [
    'inline-flex min-w-16 items-center justify-center',
    'rounded px-2 py-1',
    'text-base text-main',
];

export const BadgeVariants = cva(badgeBaseClasses, {
    variants: {
        variant: {
            fill: 'bg-[color:var(--badge-surface)]',
            outline: [
                'border border-[color:var(--badge-accent)] bg-[color:var(--badge-grey-surface)]',
                '[&>span>svg]:text-[color:var(--badge-accent)]',
            ],
        },
        color: {
            primary: '[--badge-surface:var(--badge-primary-surface)] [--badge-accent:var(--badge-primary-accent)]',
            secondary:
                '[--badge-surface:var(--badge-secondary-surface)] [--badge-accent:var(--badge-secondary-accent)]',
            grey: '[--badge-surface:var(--badge-grey-surface)] [--badge-accent:var(--badge-grey-accent)]',
            red: '[--badge-surface:var(--badge-red-surface)] [--badge-accent:var(--badge-red-accent)]',
            orange: '[--badge-surface:var(--badge-orange-surface)] [--badge-accent:var(--badge-orange-accent)]',
            green: '[--badge-surface:var(--badge-green-surface)] [--badge-accent:var(--badge-green-accent)]',
            blue: '[--badge-surface:var(--badge-blue-surface)] [--badge-accent:var(--badge-blue-accent)]',
            error: '[--badge-surface:var(--badge-error-surface)] [--badge-accent:var(--status-error-main)]',
            warning: '[--badge-surface:var(--badge-warning-surface)] [--badge-accent:var(--status-warning-main)]',
            success: '[--badge-surface:var(--badge-success-surface)] [--badge-accent:var(--status-success-main)]',
            disabled: '[--badge-surface:var(--neutral-500)] [--badge-accent:var(--neutral-500)]',
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
    compoundVariants: [{ variant: 'outline', color: 'grey', className: '[&>span>svg]:text-inherit' }],
    defaultVariants: {
        variant: 'outline',
        color: 'grey',
    },
});

export const badgeIconClasses = ['badge-icon flex shrink-0 items-center', '[&>svg]:h-3 [&>svg]:w-3'];

export const badgeLabelClasses = 'badge-label translate-y-[1px]';
