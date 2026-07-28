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
import { Button as BaseUIButton } from '@base-ui/react/button';
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

const buttonBaseClasses = [
    'inline-flex items-center justify-center whitespace-nowrap rounded-3xl transition-colors',
    'hover:underline',
    'focus:outline-none focus-visible:focus-ring',
    'active:no-underline',
    'disabled:text-[#616161] dark:disabled:text-[#A6A6A6] disabled:pointer-events-none disabled:opacity-50',
    'has-[svg]:gap-2 [&>svg]:shrink-0',
];

export const ButtonVariants = cva(buttonBaseClasses, {
    variants: {
        variant: {
            primary: [
                'bg-primary text-common-white shadow-outer-1 dark:text-common-dark',
                'hover:bg-secondary',
                'focus-visible:bg-secondary',
                'active:bg-[#0D0A30]',
                'dark:active:bg-[#8D8BF8]',
                // Implement text-common when token experiment is ready - #0D0A30 matches light.primary.variant, #8D8BF8 matches dark.primary.variant.
                'disabled:shadow-none disabled:bg-[#E3E7EA] dark:disabled:bg-[#2E2E2E]',
                // disabled is neutral.light[200], text -> common.disabled // disabled:bg -> neutral.dark[700] or common.disabled.dark (token experiment)
            ],

            secondary: [
                'bg-secondary-btn-fill text-common-dark shadow-outer-1 dark:text-common-white',
                'hover:bg-secondary hover:text-common-white dark:hover:text-common-dark',
                'focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
                'active:bg-secondary-btn-active-fill active:text-common-dark dark:active:text-common-white',
                'disabled:bg-btn-disabled-fill disabled:shadow-none',
            ],
            // TODO - remove in BED-7635
            transparent: [
                'border border-transparent-btn-border bg-transparent text-main',
                'hover:border-primary hover:bg-primary hover:text-common-white hover:no-underline dark:hover:text-common-dark',
                'focus-visible:border-primary focus-visible:bg-primary focus-visible:text-common-white dark:focus-visible:text-common-dark',
            ],
            // TODO - legacy, remove in BED-7635
            text: [
                'px-2 py-0 text-primary',
                'hover:text-secondary',
                'focus-visible:text-secondary',
                'active:text-[#0D0A30]',
            ],
            // TODO - legacy, remove in BED-7635
            icon: 'text-common-dark bg-icon-btn-fill shadow-outer-1 hover:bg-primary hover:border-2 hover:border-primary focus-visible:border-2 focus-visible:border-secondary active:border-none',
            default: null,
        },
        // TODO - remove as the only usage of this is with variant="text"
        fontColor: {
            primary: 'text-primary',
        },

        size: {
            // TODO remove small variant in BED-7635
            small: 'h-9 px-4 py-1 text-xs',
            medium: 'h-10 px-6 py-2 text-sm/5',
            // TODO remove large variant in BED-7635
            large: 'h-11 px-8 py-3 text-base',
        },
    },

    defaultVariants: {
        variant: 'primary',
        size: 'medium',
    },
});

export interface ButtonProps extends BaseUIButton.Props, VariantProps<typeof ButtonVariants> {}

export const Button = React.forwardRef<React.ComponentRef<typeof BaseUIButton>, ButtonProps>(function Button(
    { className, children, disabled = false, variant, size, fontColor, ...props },
    ref
) {
    return (
        <BaseUIButton
            {...props}
            ref={ref}
            disabled={disabled}
            className={(state) =>
                cn(
                    ButtonVariants({ variant, size, fontColor }),
                    typeof className === 'function' ? className(state) : className
                )
            }>
            {children}
        </BaseUIButton>
    );
});

Button.displayName = 'Button';

export const TextButtonBaseClasses = cn(
    ...buttonBaseClasses,
    'px-1 py-2 has-[svg]:px-1',
    'active:text-[#0D0A30] dark:active:text-primary',
    'hover:text-secondary',
    'focus-visible:ring-0 focus-visible:ring-transparent focus-visible:text-secondary'
);

export const TextButtonVariants = cva(TextButtonBaseClasses, {
    variants: {
        fontColor: {
            default: 'text-main hover:no-underline',
            primary: 'text-primary',
        },
    },
    defaultVariants: {
        fontColor: 'default',
    },
});

export type TextButtonProps = Omit<BaseUIButton.Props, 'children' | 'className'> & {
    children: React.ReactNode;
    className?: string;
    fontColor?: 'primary' | 'default';
};

export const TextButton = React.forwardRef<React.ComponentRef<typeof BaseUIButton>, TextButtonProps>(
    function TextButton({ className, children, disabled = false, fontColor, ...props }, ref) {
        return (
            <BaseUIButton
                {...props}
                ref={ref}
                disabled={disabled}
                className={cn(TextButtonVariants({ fontColor }), className)}>
                {children}
            </BaseUIButton>
        );
    }
);

TextButton.displayName = 'TextButton';

// TODO remove/refactor in BED-6062
export const IconButtonClasses = cn(
    'hover:text-primary',
    'active:bg-transparent active:text-secondary',
    'focus-visible:ring-transparent'
);

export interface IconButtonProps
    extends Omit<ButtonProps, 'size' | 'children' | 'className' | 'variant' | 'fontColor'> {
    variant?: 'default' | 'primary' | 'secondary';
    className?: string;
    'aria-label': string;
    children: React.ReactElement;
}

export const IconButton = React.forwardRef<React.ComponentRef<typeof BaseUIButton>, IconButtonProps>(
    function IconButton({ variant = 'default', children, className, disabled = false, ...props }, ref) {
        // TODO remove BED-6062
        return (
            <Button
                {...props}
                ref={ref}
                disabled={disabled}
                variant={variant}
                size={null}
                className={cn(
                    'size-fit box-border shrink-0 rounded-full border-0 p-2 has-[svg]:p-2 aspect-square',
                    variant === 'default' && IconButtonClasses,
                    className
                )}>
                {children}
            </Button>
        );
    }
);

IconButton.displayName = 'IconButton';
