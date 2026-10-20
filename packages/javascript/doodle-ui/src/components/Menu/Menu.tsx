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
import { faCaretRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import * as DropdownMenuPrimitive from '@radix-ui/react-dropdown-menu';
import * as React from 'react';
import { cn } from '../utils';
import {
    menuContentStyles,
    menuItemIconStyles,
    menuItemStyles,
    menuItemTextStyles,
    menuLabelStyles,
    menuSecondaryIconStyles,
    menuSeparatorStyles,
    menuSubContentStyles,
    menuSubTriggerStyles,
    menuSubTriggerTextStyles,
    menuTriggerStyles,
} from './Menu.styles';

/**
 * Root
 */
const Menu = DropdownMenuPrimitive.Root;

/**
 * Trigger
 */
const MenuTrigger = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Trigger>
>(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Trigger ref={ref} className={cn(menuTriggerStyles, className)} {...props} />
));
MenuTrigger.displayName = DropdownMenuPrimitive.Trigger.displayName;

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
            className={cn(menuContentStyles, className)}
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
        <DropdownMenuPrimitive.Item ref={ref} className={cn(menuItemStyles, className)} {...props}>
            {iconLeft && icon && <span className={menuItemIconStyles}>{icon}</span>}
            <span className={menuItemTextStyles}>{children}</span>
            {secondaryMenu && <FontAwesomeIcon icon={faCaretRight} className={menuSecondaryIconStyles} />}
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
    <DropdownMenuPrimitive.Label ref={ref} className={cn(menuLabelStyles, className)} {...props} />
));
MenuLabel.displayName = DropdownMenuPrimitive.Label.displayName;

/**
 * Sub
 */
const MenuSub = DropdownMenuPrimitive.Sub;

/**
 * SubTrigger
 */
const MenuSubTrigger = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.SubTrigger>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.SubTrigger>
>(({ className, children, ...props }, ref) => (
    <DropdownMenuPrimitive.SubTrigger ref={ref} className={cn(menuSubTriggerStyles, className)} {...props}>
        <span className={menuSubTriggerTextStyles}>{children}</span>
        <FontAwesomeIcon icon={faCaretRight} className={menuSecondaryIconStyles} />
    </DropdownMenuPrimitive.SubTrigger>
));
MenuSubTrigger.displayName = DropdownMenuPrimitive.SubTrigger.displayName;

/**
 * SubContent
 */
const MenuSubContent = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.SubContent>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.SubContent>
>(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Portal>
        <DropdownMenuPrimitive.SubContent ref={ref} className={cn(menuSubContentStyles, className)} {...props} />
    </DropdownMenuPrimitive.Portal>
));
MenuSubContent.displayName = DropdownMenuPrimitive.SubContent.displayName;

/**
 * Separator
 */
const MenuSeparator = React.forwardRef<
    React.ElementRef<typeof DropdownMenuPrimitive.Separator>,
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Separator ref={ref} className={cn(menuSeparatorStyles, className)} {...props} />
));
MenuSeparator.displayName = DropdownMenuPrimitive.Separator.displayName;

export {
    Menu,
    MenuContent,
    MenuItem,
    MenuLabel,
    MenuPortal,
    MenuSeparator,
    MenuSub,
    MenuSubContent,
    MenuSubTrigger,
    MenuTrigger,
};
