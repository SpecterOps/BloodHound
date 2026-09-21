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
import { Tooltip } from '../Tooltip';
import { cn } from '../utils';

const defaultIconButtonClasses = [
    'hover:text-primary dark:hover:text-primary',
    'active:bg-transparent active:text-secondary dark:active:text-secondary',
    'focus-visible:ring-transparent focus-visible:ring-offset-0 focus-visible:ring-offset-transparent focus-visible:text-primary dark:focus-visible:text-primary',
];

export const IconButtonVariants = cva(
    [
        ...buttonBaseClasses,
        'inline-grid h-fit aspect-square box-border',
        'shrink-0 place-items-center align-middle rounded-full border-0 p-2',
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

export interface IconButtonProps extends Omit<BaseUIButton.Props, 'children' | 'className' | 'render'> {
    variant?: 'default' | 'primary' | 'secondary';
    className?: BaseUIButton.Props['className'];
    'aria-label': string;
    children: React.ReactElement;
    size?: number;
    tooltip?: React.ReactNode;
}

type IconButtonStyle = React.CSSProperties & {
    '--icon-button-icon-size': string;
};

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
    {
        variant = 'default',
        'aria-label': ariaLabel,
        children,
        className,
        disabled = false,
        size = 16,
        tooltip = ariaLabel,
        ...props
    },
    ref
) {
    const decorativeIcon = React.cloneElement(children, {
        'aria-hidden': true,
    } as React.HTMLAttributes<HTMLElement>);
    const renderButton = (render?: BaseUIButton.Props['render']) => (
        <BaseUIButton
            {...props}
            render={render}
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
            <span className='inline-flex size-[var(--icon-button-icon-size)] shrink-0 items-center justify-center [&>svg]:size-full'>
                {decorativeIcon}
            </span>
        </BaseUIButton>
    );

    return (
        <Tooltip
            tooltip={tooltip}
            contentProps={{ side: 'bottom', align: 'start' }}
            renderTrigger={disabled ? undefined : renderButton}>
            {disabled ? <span className='inline-flex'>{renderButton()}</span> : undefined}
        </Tooltip>
    );
});

IconButton.displayName = 'IconButton';
