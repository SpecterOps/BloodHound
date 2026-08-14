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

const primaryClasses = [
    'bg-primary text-common-white shadow-outer-1 dark:text-common-dark',
    'hover:bg-secondary',
    'focus-visible:bg-secondary',
    'active:bg-[#0D0A30] dark:active:bg-[#8D8BF8]',
    // Implement text-common when token experiment is ready - #0D0A30 matches light.primary.variant, #8D8BF8 matches dark.primary.variant.
    'disabled:shadow-none disabled:bg-[#E3E7EA] dark:disabled:bg-[#2E2E2E]',
    // disabled is neutral.light[200], text -> common.disabled // disabled:bg -> neutral.dark[700] or common.disabled.dark (token experiment)
];

const secondaryClasses = [
    'bg-secondary-btn-fill text-common-dark shadow-outer-1 dark:text-common-white',
    'hover:bg-secondary hover:text-common-white dark:hover:text-common-dark',
    'focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
    'active:bg-secondary-btn-active-fill active:text-common-dark dark:active:text-common-white',
    'disabled:bg-btn-disabled-fill disabled:shadow-none',
];

export const ButtonVariants = cva(buttonBaseClasses, {
    variants: {
        variant: {
            primary: primaryClasses,
            secondary: secondaryClasses,
            // TODO - remove in BED-7642
            // used in DropdownTriggerContents & EnvironmentSelectorTrigger
            /**
             * @deprecated Use TextButton instead.
             */
            transparent: [
                'border border-transparent-btn-border bg-transparent text-main',
                'hover:border-secondary hover:bg-secondary hover:text-common-white hover:no-underline dark:hover:text-common-dark',
                'focus-visible:border-primary focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
            ],
            // TODO - legacy, remove in BED-7635
            /**
             * @deprecated Use IconButton instead.
             */
            icon: [
                'rounded-full text-common-dark bg-icon-btn-fill shadow-outer-1 has-[svg]:p-2',
                'hover:border-2 hover:border-primary',
                'active:border-none',
            ],
        },
        // TODO - remove as the only usage of this is with variant="text"
        /**
         * @deprecated Use TextButton instead.
         */
        fontColor: {
            primary: 'text-primary',
        },
        /**
         * @deprecated .
         */
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
    'px-2 py-1 has-[svg]:px-1',
    'active:text-[#0D0A30] dark:active:text-primary',
    'hover:text-secondary',
    'focus-visible:ring-0 focus-visible:ring-transparent focus-visible:text-secondary'
);

export const TextButtonVariants = cva(TextButtonBaseClasses, {
    variants: {
        fontColor: {
            default: 'text-main',
            primary: 'text-primary',
        },
    },
    defaultVariants: {
        fontColor: 'default',
    },
});

type TextButtonBaseProps = Omit<BaseUIButton.Props, 'children' | 'render'> & {
    fontColor?: 'primary' | 'default' | null;
};

type TextButtonContent =
    | {
          children: React.ReactNode;
          render?: BaseUIButton.Props['render'];
      }
    | {
          children?: React.ReactNode;
          render: NonNullable<BaseUIButton.Props['render']>;
      };

export type TextButtonProps = TextButtonBaseProps & TextButtonContent;

export const TextButton = React.forwardRef<React.ComponentRef<typeof BaseUIButton>, TextButtonProps>(
    function TextButton({ className, disabled = false, fontColor, ...props }, ref) {
        return (
            <BaseUIButton
                {...props}
                ref={ref}
                disabled={disabled}
                className={(state) =>
                    cn(
                        TextButtonBaseClasses,
                        fontColor === 'primary' ? 'text-primary' : 'text-main',
                        typeof className === 'function' ? className(state) : className
                    )
                }
            />
        );
    }
);

TextButton.displayName = 'TextButton';

// TODO remove/refactor in BED-6062
const defaultIconButtonClasses = [
    'hover:text-primary dark:hover:text-primary',
    'active:bg-transparent active:text-secondary dark:active:text-secondary',
    'focus-visible:ring-transparent focus-visible:ring-offset-0 focus-visible:ring-offset-transparent focus-visible:text-primary dark:focus-visible:text-primary',
];

export const IconButtonVariants = cva(
    [
        ...buttonBaseClasses,
        'inline-grid h-fit min-h-8 aspect-square box-border',
        'shrink-0 place-items-center align-middle rounded-full border-0 p-2',
        '[&>svg]:h-[var(--icon-button-icon-size)]',
        '[&>svg]:w-[var(--icon-button-icon-size)]',
    ],
    {
        variants: {
            variant: {
                default: defaultIconButtonClasses,
                primary: primaryClasses,
                secondary: secondaryClasses,
            },
        },
        defaultVariants: {
            variant: 'default',
        },
    }
);
export interface IconButtonProps extends Omit<BaseUIButton.Props, 'children' | 'className'> {
    variant?: 'default' | 'primary' | 'secondary';
    color?: string;
    className?: BaseUIButton.Props['className'];
    'aria-label': string;
    children?: React.ReactNode;
    size?: number;
}

type IconButtonStyle = React.CSSProperties & {
    '--icon-button-icon-size': string;
};

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
    { variant = 'default', children, className, color, disabled = false, size = 16, ...props },
    ref
) {
    // TODO remove/refactor BED-6062
    // allow for Icon prop and chosing the icon
    // add tooltip / see Icon component
    return (
        <BaseUIButton
            {...props}
            ref={ref}
            disabled={disabled}
            className={(state) =>
                cn(IconButtonVariants({ variant }), typeof className === 'function' ? className(state) : className)
            }
            style={(state) =>
                ({
                    ...(typeof props.style === 'function' ? props.style(state) : props.style),
                    '--icon-button-icon-size': `${size}px`,
                    color,
                }) as IconButtonStyle
            }>
            {children}
        </BaseUIButton>
    );
});

IconButton.displayName = 'IconButton';
