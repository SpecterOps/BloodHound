import * as React from 'react';
import * as TabsPrimitive from '@radix-ui/react-tabs';
import { cn } from 'components/utils';

const Tabs = TabsPrimitive.Root;

const TabsList = React.forwardRef<
    React.ElementRef<typeof TabsPrimitive.List>,
    React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
    <TabsPrimitive.List
        ref={ref}
        className={cn('inline-flex text-base dark:text-white items-center justify-center gap-4', className)}
        {...props}
    />
));
TabsList.displayName = TabsPrimitive.List.displayName;

const TabsTrigger = React.forwardRef<
    React.ElementRef<typeof TabsPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => {
    return (
        <TabsPrimitive.Trigger
            ref={ref}
            className={cn(
                `grid grid-cols-1 text-center whitespace-nowrap p-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:font-bold data-[state=active]:border-b-primary data-[state=active]:border-b-[1px] data-[state=active]:text-primary dark:data-[state=active]:text-secondary-variant-2 dark:data-[state=active]:border-b-secondary-variant-2`,
                className
            )}
            {...props}></TabsPrimitive.Trigger>
    );
});
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName;

const TabsContent = React.forwardRef<
    React.ElementRef<typeof TabsPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
    <TabsPrimitive.Content
        ref={ref}
        className={cn('mt-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2', className)}
        {...props}
    />
));
TabsContent.displayName = TabsPrimitive.Content.displayName;

export { Tabs, TabsList, TabsTrigger, TabsContent };
