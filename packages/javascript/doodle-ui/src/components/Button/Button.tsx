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
import type { VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cnWithState } from '../utils';
import { ButtonVariants, TextButtonBaseClasses } from './Button.styles';

export { ButtonVariants, TextButtonBaseClasses, TextButtonVariants } from './Button.styles';

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
            className={cnWithState(ButtonVariants({ variant, size, fontColor }), className)}>
            {children}
        </BaseUIButton>
    );
});

Button.displayName = 'Button';

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
    function TextButton({ className, disabled = false, fontColor, type = 'button', ...props }, ref) {
        return (
            <BaseUIButton
                {...props}
                ref={ref}
                disabled={disabled}
                type={type}
                className={cnWithState(
                    TextButtonBaseClasses,
                    fontColor === 'primary' ? 'text-primary' : 'text-main',
                    className
                )}
            />
        );
    }
);

TextButton.displayName = 'TextButton';
