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
import { AccessibleIcon } from '../AccessibleIcon';
import { buttonBaseClasses } from '../Button/Button.styles';
import { Tooltip } from '../Tooltip';
import { cn, cnWithState } from '../utils';

const defaultIconButtonClasses = [
    'hover:text-primary dark:hover:text-primary',
    'active:bg-transparent active:text-secondary dark:active:text-secondary',
    'focus-visible:ring-transparent focus-visible:ring-offset-0 focus-visible:ring-offset-transparent focus-visible:text-primary dark:focus-visible:text-primary',
];

const iconPrimaryClasses = [
    'bg-primary text-common-white shadow-outer-1 dark:text-common-dark',
    'hover:bg-secondary',
    'focus-visible:bg-secondary',
    'active:bg-[#0D0A30] dark:active:bg-[#8D8BF8]',
    // Implement text-common when token experiment is ready - #0D0A30 matches light.primary.variant, #8D8BF8 matches dark.primary.variant.
    'disabled:shadow-none disabled:bg-[#E3E7EA] dark:disabled:bg-[#2E2E2E]',
    // disabled is neutral.light[200], text -> common.disabled // disabled:bg -> neutral.dark[700] or common.disabled.dark (token experiment)
];

const iconSecondaryClasses = [
    'bg-secondary-btn-fill text-common-dark shadow-outer-1 dark:text-common-white',
    'hover:bg-secondary hover:text-common-white dark:hover:text-common-dark',
    'focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
    'active:bg-secondary-btn-active-fill active:text-common-dark dark:active:text-common-white',
    'disabled:bg-btn-disabled-fill disabled:shadow-none',
];

export const IconButtonVariants = cva(
    [
        ...buttonBaseClasses,
        'inline-grid h-fit aspect-square box-border',
        'shrink-0 place-items-center align-middle rounded-full border-0 p-2',
        '[&>svg]:h-[var(--icon-button-icon-size)]',
        '[&>svg]:w-[var(--icon-button-icon-size)]',
    ],
    {
        variants: {
            variant: {
                default: defaultIconButtonClasses,
                primary: iconPrimaryClasses,
                secondary: iconSecondaryClasses,
            },
        },
        defaultVariants: {
            variant: 'default',
        },
    }
);
export interface IconButtonProps extends Omit<BaseUIButton.Props, 'children' | 'className'> {
    'aria-label': string;
    children: React.ReactElement;
    className?: BaseUIButton.Props['className'];
    color?: string;
    hideTooltip?: boolean;
    size?: number;
    tooltip?: React.ReactNode;
    /** Classes for positioning the tooltip trigger in its parent layout. */
    triggerClassName?: string;
    variant?: 'default' | 'primary' | 'secondary';
}

type IconButtonStyle = React.CSSProperties & {
    '--icon-button-icon-size': string;
};

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
    {
        'aria-label': ariaLabel,
        children,
        className,
        color,
        disabled = false,
        hideTooltip = false,
        size = 16,
        tooltip = ariaLabel,
        triggerClassName,
        variant = 'default',
        ...props
    },
    ref
) {
    const buttonIcon = React.cloneElement(children, {
        'aria-hidden': true,
        ...(children.type === AccessibleIcon ? { hideTooltip: true } : {}),
    } as React.HTMLAttributes<HTMLElement> & { hideTooltip?: boolean });
    const renderButton = (render?: BaseUIButton.Props['render'], triggerClasses?: string) => (
        <BaseUIButton
            {...props}
            render={render}
            ref={ref}
            aria-label={ariaLabel}
            disabled={disabled}
            className={cnWithState(IconButtonVariants({ variant }), className, triggerClasses)}
            style={(state) =>
                ({
                    ...(typeof props.style === 'function' ? props.style(state) : props.style),
                    '--icon-button-icon-size': `${size}px`,
                    color,
                }) as IconButtonStyle
            }>
            <span className='inline-flex size-[var(--icon-button-icon-size)] shrink-0 items-center justify-center [&>*]:size-full [&_svg]:size-full'>
                {buttonIcon}
            </span>
        </BaseUIButton>
    );

    if (hideTooltip) return renderButton(undefined, triggerClassName);

    return (
        <Tooltip
            tooltip={tooltip}
            contentProps={{ side: 'bottom', align: 'start' }}
            renderTrigger={disabled ? undefined : (trigger) => renderButton(trigger, triggerClassName)}>
            {disabled ? <span className={cn('inline-flex', triggerClassName)}>{renderButton()}</span> : undefined}
        </Tooltip>
    );
});

IconButton.displayName = 'IconButton';
