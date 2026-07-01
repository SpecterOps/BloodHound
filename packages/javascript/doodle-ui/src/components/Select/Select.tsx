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
import * as SelectPrimitive from '@radix-ui/react-select';
import { cva, VariantProps } from 'class-variance-authority';
import { ChevronDown, ChevronUp } from 'lucide-react';
import * as React from 'react';
import { cn } from '../utils';

const Select = SelectPrimitive.Root;

const SelectPortal = SelectPrimitive.Portal;

const SelectGroup = SelectPrimitive.Group;

const SelectValue = SelectPrimitive.Value;

export const SelectTriggerVariants = cva(
    'flex h-10 w-full items-center justify-between bg-neutral-2 rounded-lg p-2 placeholder:text-neutral-5 focus:outline-none focus-visible:focus-ring data-[state=open]:focus-ring disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1 [&[data-state=open]>svg]:rotate-180',
    {
        variants: {
            variant: {
                outlined:
                    'rounded-md ring-1 ring-neutral-dark-5 dark:ring-neutral-light-5 px-3 py-2 text-sm hover:ring-2 bg-neutral-1',
                underlined:
                    'rounded-sm ring-none bg-transparent border-b-neutral-dark-5 dark:border-b-neutral-light-5 border-b hover:border-b-2 focus-visible:border-secondary focus-visible:border-b-2 dark:focus-visible:border-secondary-variant-2',
            },
        },
        defaultVariants: {
            variant: 'outlined',
        },
    }
);

const SelectTrigger = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger> & VariantProps<typeof SelectTriggerVariants>
>(({ className, variant, children, ...props }, ref) => (
    <SelectPrimitive.Trigger ref={ref} className={cn(SelectTriggerVariants({ variant, className }))} {...props}>
        {children}
        <SelectPrimitive.Icon asChild>
            <ChevronDown className='h-4 w-4' />
        </SelectPrimitive.Icon>
    </SelectPrimitive.Trigger>
));
SelectTrigger.displayName = SelectPrimitive.Trigger.displayName;

const SelectScrollUpButton = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.ScrollUpButton>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollUpButton>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.ScrollUpButton
        ref={ref}
        className={cn('flex cursor-default items-center justify-center py-1', className)}
        {...props}>
        <ChevronUp className='h-4 w-4' />
    </SelectPrimitive.ScrollUpButton>
));
SelectScrollUpButton.displayName = SelectPrimitive.ScrollUpButton.displayName;

const SelectScrollDownButton = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.ScrollDownButton>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollDownButton>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.ScrollDownButton
        ref={ref}
        className={cn('flex cursor-default items-center justify-center py-1', className)}
        {...props}>
        <ChevronDown className='h-4 w-4' />
    </SelectPrimitive.ScrollDownButton>
));
SelectScrollDownButton.displayName = SelectPrimitive.ScrollDownButton.displayName;

const SelectContent = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = 'popper', ...props }, ref) => (
    <SelectPrimitive.Content
        ref={ref}
        className={cn(
            'relative z-[1500] max-h-96 overflow-hidden border rounded dark:border-neutral-light-5 bg-neutral-light-1 dark:bg-neutral-dark-2 text-black dark:text-white dark:shadow-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
            position === 'popper' &&
                'data-[side=bottom]:-translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1',
            className
        )}
        position={position}
        {...props}>
        <SelectScrollUpButton />
        <SelectPrimitive.Viewport
            className={cn(
                position === 'popper' &&
                    'h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)]'
            )}>
            {children}
        </SelectPrimitive.Viewport>
        <SelectScrollDownButton />
    </SelectPrimitive.Content>
));
SelectContent.displayName = SelectPrimitive.Content.displayName;

const SelectLabel = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Label>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Label>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.Label ref={ref} className={cn('py-1.5 pl-8 pr-2 font-semibold', className)} {...props} />
));
SelectLabel.displayName = SelectPrimitive.Label.displayName;

const SelectItem = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Item>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
    <SelectPrimitive.Item
        ref={ref}
        className={cn(
            'relative flex w-full cursor-default select-none items-center p-2 outline-none data-[highlighted]:bg-secondary data-[highlighted]:text-white data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)] dark:data-[highlighted]:bg-secondary dark:data-[highlighted]:text-neutral-dark-1 data-[disabled]:pointer-events-none data-[disabled]:opacity-50 data-[state=checked]:text-primary dark:data-[state=checked]:text-secondary-variant-2 data-[state=checked]:font-bold',
            className
        )}
        {...props}>
        <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
    </SelectPrimitive.Item>
));
SelectItem.displayName = SelectPrimitive.Item.displayName;

const SelectSeparator = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Separator>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Separator>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.Separator ref={ref} className={cn('-mx-1 my-1 h-px bg-neutral-light-3', className)} {...props} />
));
SelectSeparator.displayName = SelectPrimitive.Separator.displayName;

export {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectPortal,
    SelectScrollDownButton,
    SelectScrollUpButton,
    SelectSeparator,
    SelectTrigger,
    SelectValue,
};
