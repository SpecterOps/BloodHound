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
        variant: {
            fill: '',
            outline: 'border bg-badge-outline-bg',
        },
        color: {
            primary: '',
            secondary: '',
            grey: '',
            red: '',
            orange: '',
            green: '',
            blue: '',
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
        { variant: 'fill', color: 'primary', className: 'bg-badge-primary-fill' },
        { variant: 'fill', color: 'secondary', className: 'bg-badge-secondary-fill' },
        { variant: 'fill', color: 'grey', className: 'bg-badge-grey-fill' },
        { variant: 'fill', color: 'red', className: 'bg-badge-red-fill' },
        { variant: 'fill', color: 'orange', className: 'bg-badge-orange-fill' },
        { variant: 'fill', color: 'green', className: 'bg-badge-green-fill' },
        { variant: 'fill', color: 'blue', className: 'bg-badge-blue-fill' },

        // Outline variants
        {
            variant: 'outline',
            color: 'primary',
            className: 'border-badge-primary-outline [&>span>svg]:text-badge-primary-outline',
        },
        {
            variant: 'outline',
            color: 'secondary',
            className: 'border-badge-secondary-outline [&>span>svg]:text-badge-secondary-outline',
        },
        {
            variant: 'outline',
            color: 'grey',
            className: 'border-badge-grey-outline [&>span>svg]:text-badge-grey-outline',
        },
        {
            variant: 'outline',
            color: 'red',
            className: 'border-badge-red-outline [&>span>svg]:text-badge-red-outline',
        },
        {
            variant: 'outline',
            color: 'orange',
            className: 'border-badge-orange-outline [&>span>svg]:text-badge-orange-outline',
        },
        {
            variant: 'outline',
            color: 'green',
            className: 'border-badge-green-outline [&>span>svg]:text-badge-green-outline',
        },
        {
            variant: 'outline',
            color: 'blue',
            className: 'border-badge-blue-outline [&>span>svg]:text-badge-blue-outline',
        },
    ],
    defaultVariants: {
        variant: 'outline',
        color: 'grey',
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
    ({ className, color, icon, iconClassName, iconPosition, label, labelClassName, style, variant, ...rest }, ref) => {
        const isHex = typeof color === 'string' && color.startsWith('#');

        const hexStyle = isHex
            ? variant === 'outline'
                ? { borderColor: color, ...style }
                : { backgroundColor: color, ...style }
            : style;

        return (
            <div
                ref={ref}
                {...rest}
                style={hexStyle}
                className={cn(
                    BadgeVariants({
                        variant,
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
