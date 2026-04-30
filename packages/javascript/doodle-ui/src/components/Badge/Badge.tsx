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
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

type BadgeColor = NonNullable<VariantProps<typeof BadgeVariants>['color']>;

const BadgeVariants = cva('inline-flex items-center justify-center rounded text-main text-sm p-2', {
    variants: {
        color: {
            indeterminate: 'border-badge-indeterminate-border bg-badge-indeterminate-fill border',
            primary: 'bg-badge-primary',
            secondary: 'bg-badge-secondary',
            grey: 'bg-badge-grey',
            red: 'bg-badge-red',
            orange: 'bg-badge-orange',
            green: 'bg-badge-green',
            blue: 'bg-badge-blue',
            purple: 'bg-badge-purple',
        },
        iconPosition: {
            left: 'flex-row',
            right: 'flex-row-reverse',
        },
        hasIcon: {
            true: 'gap-1',
            false: null,
        },
    },
    defaultVariants: {
        color: 'indeterminate',
    },
});

type BadgeProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'color'> &
    Omit<VariantProps<typeof BadgeVariants>, 'hasIcon' | 'color'> & {
        color?: BadgeColor;
        label: string;
        labelClassName?: string;
        icon?: React.ReactNode;
        iconClassName?: string;
    };

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
    ({ label, labelClassName, icon, iconClassName, iconPosition, color, className, ...rest }, ref) => {
        return (
            <div
                ref={ref}
                {...rest}
                className={cn(
                    BadgeVariants({ color, iconPosition: icon ? iconPosition ?? 'left' : undefined, hasIcon: !!icon }),
                    className
                )}>
                {icon && (
                    <span className={cn('shrink-0 [&>svg]:block [&>svg]:h-4 [&>svg]:w-4', iconClassName)}>{icon}</span>
                )}
                <span className={cn('translate-y-[0.5px]', labelClassName)}>{label}</span>
            </div>
        );
    }
);

export { Badge };
