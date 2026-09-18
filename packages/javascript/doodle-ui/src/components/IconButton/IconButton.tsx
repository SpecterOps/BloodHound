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
import { cva } from 'class-variance-authority';
import * as React from 'react';
import { buttonBaseClasses, primaryClasses, secondaryClasses } from '../Button/Button.styles';
import { Icon, type IconProps } from '../Icon';
import { cn } from '../utils';

export const IconButtonVariants = cva(
    [
        ...buttonBaseClasses,
        'inline-grid h-fit aspect-square box-border',
        'shrink-0 place-items-center align-middle rounded-full border-0 p-2',
    ],
    {
        variants: {
            variant: {
                primary: primaryClasses,
                secondary: secondaryClasses,
            },
        },
        defaultVariants: {
            variant: 'primary',
        },
    }
);

export interface IconButtonProps extends Omit<BaseUIButton.Props, 'children' | 'className'> {
    variant?: 'primary' | 'secondary';
    className?: BaseUIButton.Props['className'];
    'aria-label': string;
    children: IconProps['children'];
    size?: number;
}

type IconButtonStyle = React.CSSProperties & {
    '--icon-button-icon-size': string;
};

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
    {
        variant = 'primary',
        'aria-label': ariaLabel,
        children,
        className,
        disabled = false,
        size = 16,
        ...props
    },
    ref
) {
    return (
        <BaseUIButton
            {...props}
            ref={ref}
            aria-label={ariaLabel}
            disabled={disabled}
            className={(state) =>
                cn(IconButtonVariants({ variant }), typeof className === 'function' ? className(state) : className)
            }
            style={(state) =>
                ({
                    ...(typeof props.style === 'function' ? props.style(state) : props.style),
                    '--icon-button-icon-size': `${size}px`,
                }) as IconButtonStyle
            }>
            <Icon
                aria-label={ariaLabel}
                className='size-[var(--icon-button-icon-size)] shrink-0 items-center justify-center [&>svg]:size-full'>
                {children}
            </Icon>
        </BaseUIButton>
    );
});

IconButton.displayName = 'IconButton';
