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

const BadgeVariants = cva('inline-flex items-center justify-center rounded text-main text-sm p-2 min-w-16', {
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
            true: 'gap-2',
            false: '',
        },
    },
    defaultVariants: {
        color: 'indeterminate',
    },
});

type NamedColor = NonNullable<VariantProps<typeof BadgeVariants>['color']>;

/** @deprecated Use a named color instead. Hex support is transitional and will be removed in a future release. */
type HexColor = `#${string}`;

type BadgeProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'color'> &
    Omit<VariantProps<typeof BadgeVariants>, 'hasIcon' | 'color'> & {
        /** @deprecated Passing a hex color is transitional and will be removed in a future release. Use a named color instead. */
        color?: NamedColor | HexColor;
        label: string;
        icon?: React.ReactNode;
        iconClassName?: string;
        labelClassName?: string;
    };

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
    ({ className, color, icon, iconClassName, iconPosition, label, labelClassName, style, ...rest }, ref) => {
        const isHex = typeof color === 'string' && color.startsWith('#');

        return (
            <div
                ref={ref}
                {...rest}
                style={isHex ? { backgroundColor: color, ...style } : style}
                className={cn(
                    BadgeVariants({
                        color: isHex ? undefined : (color as NamedColor | undefined),
                        hasIcon: !!icon,
                        iconPosition: icon ? iconPosition ?? 'left' : undefined,
                    }),
                    className
                )}>
                {icon && (
                    <span className={cn('flex shrink-0 items-center [&>svg]:h-4 [&>svg]:w-4', iconClassName)}>
                        {icon}
                    </span>
                )}
                <span className={cn('translate-y-[1px]', labelClassName)}>{label}</span>
            </div>
        );
    }
);

Badge.displayName = 'Badge';

export { Badge };
export type { HexColor, NamedColor };
