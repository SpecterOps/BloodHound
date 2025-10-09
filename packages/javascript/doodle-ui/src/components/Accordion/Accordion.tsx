import * as React from 'react';
import * as AccordionPrimitive from '@radix-ui/react-accordion';

import { cn } from 'components/utils';

const Accordion = AccordionPrimitive.Root;

const AccordionItem = React.forwardRef<
    React.ElementRef<typeof AccordionPrimitive.Item>,
    React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Item>
>(({ ...props }, ref) => (
    <AccordionPrimitive.Item ref={ref} className='bg-neutral-light-2 rounded-lg mb-4' {...props} />
));
AccordionItem.displayName = 'AccordionItem';

const AccordionHeader = React.forwardRef<
    React.ElementRef<typeof AccordionPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
    <AccordionPrimitive.Header className='flex'>
        <AccordionPrimitive.Trigger
            ref={ref}
            className={cn(
                'flex flex-1 gap-1 items-center font-bold transition-all hover:underline [&[data-state=open]>svg]:rotate-180 p-4',
                className
            )}
            {...props}>
            {children}
        </AccordionPrimitive.Trigger>
    </AccordionPrimitive.Header>
));
AccordionHeader.displayName = AccordionPrimitive.Header.displayName;

const AccordionContent = React.forwardRef<
    React.ElementRef<typeof AccordionPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Content>
>(({ className, children, ...props }, ref) => (
    <AccordionPrimitive.Content
        ref={ref}
        className='overflow-hidden text-sm transition-all data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down'
        {...props}>
        <div className={cn('p-4', className)}>{children}</div>
    </AccordionPrimitive.Content>
));

AccordionContent.displayName = AccordionPrimitive.Content.displayName;

export { Accordion, AccordionItem, AccordionHeader, AccordionContent };
