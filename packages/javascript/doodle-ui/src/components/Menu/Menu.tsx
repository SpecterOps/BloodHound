import * as DropdownMenuPrimitive from '@radix-ui/react-dropdown-menu';
import * as React from 'react';
import { LargeRightArrow } from '../../styleguide/components/AppIcons/components/LargeRightArrow';
import { cn } from '../utils';

/**
 * Root
 */
const Menu = DropdownMenuPrimitive.Root;

/**
 * Trigger
 */
const MenuTrigger = DropdownMenuPrimitive.Trigger;

/**
 * Portal
 */
const MenuPortal = DropdownMenuPrimitive.Portal;

/**
 * Content
 */
const MenuContent = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
    <DropdownMenuPrimitive.Portal>
        <DropdownMenuPrimitive.Content
            ref={ref}
            sideOffset={sideOffset}
            className={cn(
                'z-50 min-w-[8rem] overflow-hidden rounded-md border border-neutral-light-4 bg-neutral-light-2 p-1 shadow-md',
                'dark:border-neutral-dark-4 dark:bg-neutral-dark-2',
                'data-[state=open]:animate-in data-[state=closed]:animate-out',
                'data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0',
                'data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2',
                className
            )}
            {...props}
        />
    </DropdownMenuPrimitive.Portal>
));
MenuContent.displayName = DropdownMenuPrimitive.Content.displayName;

/**
 * Item
 */
interface MenuItemProps extends React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Item> {
    /** Icon rendered on the left side of the item */
    icon?: React.ReactNode;
    /** Whether to show the left icon */
    iconLeft?: boolean;
    /** Whether to show a right-arrow indicator for a secondary/sub menu */
    secondaryMenu?: boolean;
}

const MenuItem = React.forwardRef<React.ElementRef<typeof DropdownMenuPrimitive.Item>, MenuItemProps>(
    ({ className, icon, iconLeft = false, secondaryMenu = false, children, ...props }, ref) => (
        <DropdownMenuPrimitive.Item
            ref={ref}
            className={cn(
                'relative flex cursor-pointer select-none items-center rounded-lg border border-transparent px-2 py-1.5 text-sm outline-none',
                'data-[highlighted]:border-[#4A3BD7] data-[highlighted]:bg-[#4A3BD7] data-[highlighted]:text-white dark:data-[highlighted]:bg-neutral-dark-3',
                'active:bg-[#2C2677]',
                'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
                className
            )}
            {...props}>
            {iconLeft && icon && <span className='mr-2 flex items-center'>{icon}</span>}
            <span className='flex-1'>{children}</span>
            {secondaryMenu && <LargeRightArrow className='ml-2' />}
        </DropdownMenuPrimitive.Item>
    )
);
MenuItem.displayName = DropdownMenuPrimitive.Item.displayName;

/**
 * Label
 */
const MenuLabel = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.Label>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Label>
>(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Label ref={ref} className={cn('px-2 py-1.5 text-sm font-semibold', className)} {...props} />
));
MenuLabel.displayName = DropdownMenuPrimitive.Label.displayName;

/**
 * Separator
 */
const MenuSeparator = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.Separator>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Separator
        ref={ref}
        className={cn('my-1 h-px bg-neutral-light-4 dark:bg-neutral-dark-4', className)}
        {...props}
    />
));
MenuSeparator.displayName = DropdownMenuPrimitive.Separator.displayName;

export { Menu, MenuContent, MenuItem, MenuLabel, MenuPortal, MenuSeparator, MenuTrigger };
