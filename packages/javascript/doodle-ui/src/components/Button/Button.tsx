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

export const ButtonVariants = cva(
    'inline-flex items-center justify-center whitespace-nowrap h-10 px-6 py-2 rounded-3xl text-sm transition-colors hover:underline focus:outline-none focus-visible:focus-ring disabled:pointer-events-none disabled:opacity-50 active:no-underline',
    {
        variants: {
            variant: {
                primary:
                    'bg-primary text-common-white dark:text-common-dark shadow-outer-1 hover:bg-secondary focus-visible:bg-secondary',
                secondary:
                    'bg-secondary-btn-fill text-common-dark shadow-outer-1 dark:text-common-white hover:bg-secondary hover:text-common-white dark:hover:text-common-dark active:bg-secondary-btn-active-fill active:text-common-dark dark:active:text-common-white focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark disabled:bg-btn-disabled-fill disabled:text-text-disabled disabled:shadow-none disabled:!opacity-100',
                transparent:
                    'bg-transparent border border-transparent-btn-border text-main hover:bg-primary hover:text-common-white dark:hover:text-common-dark hover:border-primary hover:no-underline focus-visible:bg-primary focus-visible:text-common-white dark:focus-visible:text-common-dark focus-visible:border-primary',
                text: 'text-primary px-2 py-0 hover:text-secondary active:text-primary-variant focus-visible:text-secondary disabled:text-text-disabled disabled:!opacity-100',
                // icon: 'text-common-dark bg-icon-btn-fill shadow-outer-1 hover:bg-primary hover:border-2 hover:border-primary focus-visible:border-2 focus-visible:border-secondary active:border-none',
            },
            fontColor: {
                primary: 'text-primary',
            },
            size: {
                small: 'h-9 px-4 py-1 text-xs',
                medium: 'h-10 px-6 py-2',
                large: 'h-11 px-8 py-3 text-base',
                icon: 'size-10 p-0 rounded-full hover:no-underline',
            },
        },
        defaultVariants: {
            variant: 'primary',
        },
    }
);

export interface ButtonProps extends BaseUIButton.Props, VariantProps<typeof ButtonVariants> {}

export const Button = React.forwardRef<React.ComponentRef<typeof BaseUIButton>, ButtonProps>(function Button(
    { className, variant, size, fontColor, ...props },
    ref
) {
    return (
        <BaseUIButton
            {...props}
            ref={ref}
            className={(state) =>
                cn(
                    ButtonVariants({ variant, size, fontColor }),
                    typeof className === 'function' ? className(state) : className
                )
            }
        />
    );
});

Button.displayName = 'Button';

export interface IconButtonProps extends Omit<ButtonProps, 'size' | 'children'> {
    variant?: 'primary' | 'secondary';
    'aria-label': string;
    children: React.ReactElement;
}

export const IconButton = React.forwardRef<React.ComponentRef<typeof BaseUIButton>, IconButtonProps>(
    function IconButton({ variant = 'secondary', children, ...props }, ref) {
        return (
            <Button {...props} ref={ref} variant={variant} size='icon'>
                {children}
            </Button>
        );
    }
);

IconButton.displayName = 'IconButton';
