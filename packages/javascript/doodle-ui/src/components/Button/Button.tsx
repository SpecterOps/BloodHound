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
import { buttonBaseClasses, primaryClasses, secondaryClasses } from './Button.styles';

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
    { className, children, disabled = false, variant, size, fontColor, type = 'button', ...props },
    ref
) {
    return (
        <BaseUIButton
            {...props}
            ref={ref}
            disabled={disabled}
            type={type}
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

// TODO - add type='button' in BED-6062
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
