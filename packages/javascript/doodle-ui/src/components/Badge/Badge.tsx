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

const BadgeVariants = cva(
    'inline-flex items-center justify-center rounded border bg-neutral-light-3 dark:bg-neutral-dark-3 text-neutral-dark-1 dark:text-white border-neutral-light-5 dark:border-neutral-dark-5',
    {
        variants: {
            size: {
                small: 'h-6 px-1 font-medium min-w-14 w-14',
                medium: 'h-8 px-2 text-base min-w-16',
            },
            hasIcon: {
                true: 'pr-3',
            },
        },
        defaultVariants: {
            size: 'medium',
        },
    }
);

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof BadgeVariants> {
    label: string;
    labelClassName?: string;
    icon?: React.ReactNode;
    iconClassName?: string;
    color?: string;
    backgroundColor?: string;
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
    ({ label, labelClassName, icon, iconClassName, color, backgroundColor, size, className, ...rest }, ref) => {
        return (
            <div
                ref={ref}
                {...rest}
                className={cn(BadgeVariants({ size, hasIcon: !!icon }), className)}
                style={{
                    borderColor: color,
                    backgroundColor: backgroundColor,
                }}>
                {icon && (
                    <span style={{ color }} className={iconClassName}>
                        {icon}
                    </span>
                )}
                <span className={labelClassName}>{label}</span>
            </div>
        );
    }
);

export { Badge };
