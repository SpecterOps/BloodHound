import * as React from 'react';
import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import { cn } from 'components/utils';

const TooltipProvider = TooltipPrimitive.Provider;

type RootProps = React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Root>;
const TooltipRoot = TooltipPrimitive.Root;

const TooltipPortal = TooltipPrimitive.Portal;

type TriggerProps = React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Trigger>;

const TooltipTrigger = React.forwardRef<React.ElementRef<typeof TooltipPrimitive.Trigger>, TriggerProps>(
    (props, ref) => {
        const { className, ...rest } = props;
        const asChild = !!props.children;
        return (
            <TooltipPrimitive.Trigger ref={ref} className={cn(className)} asChild={asChild} {...rest}>
                {asChild ? (
                    props.children
                ) : (
                    <span
                        className='border rounded-full border-neutral-dark-1 text-neutral-dark-1 dark:border-neutral-light-1 dark:text-neutral-light-1 size-3 grid grid-rows-7 grid-cols-7'
                        role='img'>
                        <span
                            className='bg-neutral-dark-1 dark:bg-neutral-light-1 col-start-4 row-start-2'
                            role='img'
                        />
                        <span
                            className='bg-neutral-dark-1 dark:bg-neutral-light-1 col-start-4 row-start-4 row-end-7'
                            role='img'
                        />
                    </span>
                )}
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
                    'z-50 overflow-hidden rounded-md border bg-neutral-light-2 px-3 py-1.5 text-xs text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
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
    tooltip: string | React.ReactNode;
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
        open,
        defaultOpen,
        onOpenChange,
        delayDuration = 300,
        rootProps = {},
        triggerProps = {},
        contentWidth,
        contentProps = {},
    } = props;

    return (
        <TooltipProvider>
            <TooltipRoot
                open={open}
                defaultOpen={defaultOpen}
                onOpenChange={onOpenChange}
                delayDuration={delayDuration}
                {...rootProps}>
                <TooltipTrigger children={props.children} {...triggerProps} />
                <TooltipPortal>
                    <TooltipContent contentWidth={contentWidth} {...contentProps}>
                        {tooltip}
                    </TooltipContent>
                </TooltipPortal>
            </TooltipRoot>
        </TooltipProvider>
    );
};

export { Tooltip, TooltipProvider, TooltipRoot, TooltipTrigger, TooltipPortal, TooltipContent };
