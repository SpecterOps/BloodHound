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
import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import * as React from 'react';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { cn } from '../utils';

const TooltipProvider = TooltipPrimitive.Provider;

type RootProps = React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Root>;
const TooltipRoot = TooltipPrimitive.Root;

const TooltipPortal = TooltipPrimitive.Portal;

type TriggerProps = React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Trigger>;

const TooltipTrigger = React.forwardRef<React.ElementRef<typeof TooltipPrimitive.Trigger>, TriggerProps>(
    (props, ref) => {
        const { children, asChild = !!children, className, type, ...rest } = props;
        return (
            <TooltipPrimitive.Trigger
                ref={ref}
                className={cn('focus:outline-none focus-visible:focus-ring', className)}
                asChild={asChild}
                type={asChild ? type : (type ?? 'button')}
                {...rest}>
                {children ?? <AppIcon.Info size={16} aria-hidden='true' />}
            </TooltipPrimitive.Trigger>
        );
    }
);

TooltipTrigger.displayName = TooltipPrimitive.Trigger.displayName;

interface ContentProps extends React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content> {
    contentWidth?: 'default' | 'match_trigger' | 'lg' | 'md' | 'sm';
}

const TooltipContent = React.forwardRef<React.ElementRef<typeof TooltipPrimitive.Content>, ContentProps>(
    ({ className, sideOffset = 4, contentWidth = 'default', ...props }, ref) => {
        const widthOptions: Record<typeof contentWidth, string> = {
            default: '',
            match_trigger: 'w-[var(--radix-tooltip-trigger-width)]',
            lg: 'max-w-[300px]',
            md: 'max-w-[200px]',
            sm: 'max-w-[150px]',
        };
        return (
            <TooltipPrimitive.Content
                ref={ref}
                sideOffset={sideOffset}
                className={cn(
                    'TooltipContent',
                    'text-main rounded-md border dark:border-0 bg-neutral-light-2 dark:bg-neutral-dark-5 px-3 py-1.5 text-xs text-popover-foreground shadow-md',
                    'z-[1700] overflow-hidden animate-in fade-in-0 zoom-in-95',
                    'data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95',
                    'data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
                    className,
                    widthOptions[contentWidth]
                )}
                {...props}
            />
        );
    }
);
TooltipContent.displayName = TooltipPrimitive.Content.displayName;

interface TooltipProps extends React.PropsWithChildren {
    tooltip: React.ReactNode;
    renderTrigger?: (trigger: React.ReactElement<TriggerProps>) => React.ReactElement;
    open?: RootProps['open'];
    defaultOpen?: RootProps['defaultOpen'];
    onOpenChange?: RootProps['onOpenChange'];
    delayDuration?: RootProps['delayDuration'];
    rootProps?: Omit<RootProps, 'open' | 'defaultOpen' | 'onOpenChange' | 'delayDuration'>;
    triggerProps?: TriggerProps;
    contentWidth?: ContentProps['contentWidth'];
    contentProps?: Omit<ContentProps, 'contentWidth'>;
}

const Tooltip: React.FC<TooltipProps> = (props) => {
    const {
        tooltip,
        renderTrigger,
        open,
        defaultOpen,
        onOpenChange,
        delayDuration = 300,
        rootProps = {},
        triggerProps = {},
        contentWidth,
        contentProps = {},
    } = props;
    const defaultTriggerLabel = !props.children && !renderTrigger
        ? typeof tooltip === 'string'
            ? tooltip
            : 'Show more information'
        : undefined;
    const trigger = renderTrigger ? (
        <TooltipTrigger {...triggerProps} asChild={false} />
    ) : (
        <TooltipTrigger
            children={props.children}
            aria-label={triggerProps['aria-label'] ?? defaultTriggerLabel}
            {...triggerProps}
        />
    );

    return (
        <TooltipProvider>
            <TooltipRoot
                open={open}
                defaultOpen={defaultOpen}
                onOpenChange={onOpenChange}
                delayDuration={delayDuration}
                {...rootProps}>
                {renderTrigger ? renderTrigger(trigger) : trigger}
                <TooltipPortal>
                    <TooltipContent contentWidth={contentWidth} {...contentProps}>
                        {tooltip}
                    </TooltipContent>
                </TooltipPortal>
            </TooltipRoot>
        </TooltipProvider>
    );
};

export { Tooltip, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger };
