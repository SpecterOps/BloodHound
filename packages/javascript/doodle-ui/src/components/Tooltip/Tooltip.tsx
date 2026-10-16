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
import { Tooltip as BaseUITooltip } from '@base-ui/react/tooltip';
import * as React from 'react';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { cn } from '../utils';

type ProviderProps = React.ComponentPropsWithoutRef<typeof BaseUITooltip.Provider> & {
    delayDuration?: number;
    skipDelayDuration?: number;
};

const TooltipProvider: React.FC<ProviderProps> = ({ delay, delayDuration, skipDelayDuration, timeout, ...props }) => (
    <BaseUITooltip.Provider
        delay={delay ?? delayDuration ?? 700}
        timeout={timeout ?? skipDelayDuration ?? 300}
        {...props}
    />
);

type RootProps = React.ComponentPropsWithoutRef<typeof BaseUITooltip.Root>;
const TooltipDescriptionContext = React.createContext<string | undefined>(undefined);

const TooltipRoot: React.FC<RootProps> = ({ children, ...props }) => {
    const descriptionId = React.useId();

    return (
        <TooltipDescriptionContext.Provider value={descriptionId}>
            <BaseUITooltip.Root {...props}>{children}</BaseUITooltip.Root>
        </TooltipDescriptionContext.Provider>
    );
};

const TooltipPortal = BaseUITooltip.Portal;

type BaseTriggerProps = React.ComponentPropsWithoutRef<typeof BaseUITooltip.Trigger>;
type TriggerProps = Omit<BaseTriggerProps, 'render'> & {
    asChild?: boolean;
};

const TooltipTrigger = React.forwardRef<HTMLButtonElement, TriggerProps>((props, ref) => {
    const {
        asChild = props.children == null || React.isValidElement(props.children),
        'aria-describedby': ariaDescribedBy,
        children,
        className,
        onFocusCapture,
        type,
        ...rest
    } = props;
    const descriptionId = React.useContext(TooltipDescriptionContext);
    const describedBy = [ariaDescribedBy, descriptionId].filter(Boolean).join(' ') || undefined;
    const trigger = children ?? (
        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
        <span role='img' tabIndex={0} className='inline-flex items-center justify-center rounded-full'>
            <AppIcon.Info size={16} aria-hidden='true' />
        </span>
    );

    return (
        <BaseUITooltip.Trigger
            ref={ref}
            aria-describedby={describedBy}
            className={cn('focus:outline-none focus-visible:focus-ring', className)}
            onFocusCapture={(event) => {
                onFocusCapture?.(event);
                if (!event.currentTarget.matches(':focus-visible')) event.preventDefault();
            }}
            render={asChild && React.isValidElement(trigger) ? trigger : undefined}
            type={asChild ? type : type ?? 'button'}
            {...rest}>
            {asChild ? undefined : trigger}
        </BaseUITooltip.Trigger>
    );
});

TooltipTrigger.displayName = 'TooltipTrigger';

type PositionerProps = React.ComponentPropsWithoutRef<typeof BaseUITooltip.Positioner>;
type PopupProps = React.ComponentPropsWithoutRef<typeof BaseUITooltip.Popup>;

interface ContentProps extends Omit<PopupProps, 'className'> {
    align?: PositionerProps['align'];
    alignOffset?: PositionerProps['alignOffset'];
    avoidCollisions?: boolean;
    className?: string;
    collisionBoundary?: PositionerProps['collisionBoundary'];
    collisionPadding?: PositionerProps['collisionPadding'];
    contentWidth?: 'match_trigger' | 'lg' | 'md' | 'sm';
    side?: PositionerProps['side'];
    sideOffset?: PositionerProps['sideOffset'];
}

type TooltipPosition = Extract<NonNullable<ContentProps['side']>, 'top' | 'bottom' | 'left' | 'right'>;

const contentWidthOptions: Record<NonNullable<ContentProps['contentWidth']>, string> = {
    match_trigger: 'w-[var(--anchor-width)]',
    lg: 'max-w-[300px]',
    md: 'max-w-[200px]',
    sm: 'max-w-[150px]',
};

const TooltipContent = React.forwardRef<HTMLDivElement, ContentProps>(
    (
        {
            align,
            alignOffset,
            avoidCollisions = true,
            children,
            className,
            collisionBoundary,
            collisionPadding,
            contentWidth = 'md',
            id,
            side,
            sideOffset = 4,
            ...props
        },
        ref
    ) => {
        const descriptionId = React.useContext(TooltipDescriptionContext);

        return (
            <BaseUITooltip.Positioner
                align={align}
                alignOffset={alignOffset}
                collisionAvoidance={avoidCollisions ? undefined : { align: 'none', side: 'none' }}
                collisionBoundary={collisionBoundary}
                collisionPadding={collisionPadding}
                side={side}
                sideOffset={sideOffset}
                className='z-[1700]'>
                <BaseUITooltip.Popup
                    ref={ref}
                    className={cn(
                        'TooltipContent',
                        'text-main break-words rounded-lg border bg-neutral-light-2 border-[#939597] dark:bg-neutral-dark-5 px-3 py-1.5 text-xs text-popover-foreground shadow-outer-1',
                        'z-[1700] overflow-hidden animate-in fade-in-0 zoom-in-95',
                        'data-[closed]:animate-out data-[closed]:fade-out-0 data-[closed]:zoom-out-95',
                        'data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
                        contentWidthOptions[contentWidth],
                        className
                    )}
                    {...props}>
                    <span id={id ?? descriptionId} role='tooltip'>
                        {children}
                    </span>
                </BaseUITooltip.Popup>
            </BaseUITooltip.Positioner>
        );
    }
);
TooltipContent.displayName = 'TooltipContent';

interface TooltipProps extends React.PropsWithChildren {
    contentProps?: Omit<ContentProps, 'contentWidth' | 'side'>;
    contentWidth?: ContentProps['contentWidth'];
    defaultOpen?: RootProps['defaultOpen'];
    delayDuration?: number;
    onOpenChange?: RootProps['onOpenChange'];
    open?: RootProps['open'];
    renderTrigger?: (trigger: React.ReactElement<TriggerProps>) => React.ReactElement;
    rootProps?: Omit<RootProps, 'open' | 'defaultOpen' | 'onOpenChange' | 'delayDuration'>;
    tooltip: React.ReactNode;
    tooltipPosition?: TooltipPosition;
    triggerProps?: TriggerProps;
}

const Tooltip = React.forwardRef<HTMLButtonElement, TooltipProps>((props, ref) => {
    const {
        children,
        contentProps = {},
        contentWidth,
        defaultOpen,
        delayDuration = 300,
        onOpenChange,
        open,
        renderTrigger,
        rootProps = {},
        tooltip,
        tooltipPosition = 'top',
        triggerProps = {},
    } = props;
    const trigger = renderTrigger ? (
        <TooltipTrigger ref={ref} {...triggerProps} asChild={false} />
    ) : (
        <TooltipTrigger ref={ref} children={children} {...triggerProps} />
    );

    return (
        <TooltipProvider delayDuration={delayDuration}>
            <TooltipRoot open={open} defaultOpen={defaultOpen} onOpenChange={onOpenChange} {...rootProps}>
                {renderTrigger ? renderTrigger(trigger) : trigger}
                <TooltipPortal>
                    <TooltipContent contentWidth={contentWidth} {...contentProps} side={tooltipPosition}>
                        {tooltip}
                    </TooltipContent>
                </TooltipPortal>
            </TooltipRoot>
        </TooltipProvider>
    );
});

Tooltip.displayName = 'Tooltip';

export { Tooltip, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger };
export type { TooltipPosition };
