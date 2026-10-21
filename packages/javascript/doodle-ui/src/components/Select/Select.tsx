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
import { faChevronDown, faChevronUp, faSpinner } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import * as SelectPrimitive from '@radix-ui/react-select';
import { VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';
import {
    selectContentClasses,
    selectIconClasses,
    selectItemClasses,
    selectLabelClasses,
    selectLoadingIconClasses,
    selectPopperContentClasses,
    selectPopperViewportClasses,
    selectScrollButtonClasses,
    selectSeparatorClasses,
    SelectTriggerVariants,
    selectViewportClasses,
} from './Select.styles';

export { SelectTriggerVariants } from './Select.styles';

const Select = SelectPrimitive.Root;

const SelectPortal = SelectPrimitive.Portal;

const SelectGroup = SelectPrimitive.Group;

interface SelectValueProps extends React.ComponentPropsWithoutRef<typeof SelectPrimitive.Value> {
    required?: boolean;
}

const SelectValue = React.forwardRef<React.ElementRef<typeof SelectPrimitive.Value>, SelectValueProps>(
    ({ placeholder, required = false, ...props }, ref) => (
        <SelectPrimitive.Value
            ref={ref}
            placeholder={required && placeholder != null ? <>{placeholder} *</> : placeholder}
            {...props}
        />
    )
);
SelectValue.displayName = SelectPrimitive.Value.displayName;

const SelectTrigger = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger> &
        VariantProps<typeof SelectTriggerVariants> & { isLoading?: boolean }
>(({ className, variant, children, isLoading = false, disabled, ...props }, ref) => (
    <SelectPrimitive.Trigger
        ref={ref}
        className={cn(SelectTriggerVariants({ variant, className }))}
        disabled={disabled || isLoading}
        aria-busy={isLoading}
        {...props}>
        {children}
        <SelectPrimitive.Icon asChild>
            {isLoading ? (
                <FontAwesomeIcon icon={faSpinner} className={cn(selectLoadingIconClasses)} />
            ) : (
                <FontAwesomeIcon icon={faChevronDown} className={selectIconClasses} />
            )}
        </SelectPrimitive.Icon>
    </SelectPrimitive.Trigger>
));
SelectTrigger.displayName = SelectPrimitive.Trigger.displayName;

const SelectScrollUpButton = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.ScrollUpButton>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollUpButton>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.ScrollUpButton ref={ref} className={cn(selectScrollButtonClasses, className)} {...props}>
        <FontAwesomeIcon icon={faChevronUp} className={selectIconClasses} />
    </SelectPrimitive.ScrollUpButton>
));
SelectScrollUpButton.displayName = SelectPrimitive.ScrollUpButton.displayName;

const SelectScrollDownButton = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.ScrollDownButton>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollDownButton>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.ScrollDownButton ref={ref} className={cn(selectScrollButtonClasses, className)} {...props}>
        <FontAwesomeIcon icon={faChevronDown} className={selectIconClasses} />
    </SelectPrimitive.ScrollDownButton>
));
SelectScrollDownButton.displayName = SelectPrimitive.ScrollDownButton.displayName;

const SelectContent = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = 'popper', ...props }, ref) => (
    <SelectPrimitive.Content
        ref={ref}
        className={cn(selectContentClasses, position === 'popper' && selectPopperContentClasses, className)}
        position={position}
        {...props}>
        <SelectScrollUpButton />
        <SelectPrimitive.Viewport
            className={cn(selectViewportClasses, position === 'popper' && selectPopperViewportClasses)}>
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
    <SelectPrimitive.Label ref={ref} className={cn(selectLabelClasses, className)} {...props} />
));
SelectLabel.displayName = SelectPrimitive.Label.displayName;

const SelectItem = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Item>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
    <SelectPrimitive.Item ref={ref} className={cn(selectItemClasses, className)} {...props}>
        <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
    </SelectPrimitive.Item>
));
SelectItem.displayName = SelectPrimitive.Item.displayName;

const SelectSeparator = React.forwardRef<
    React.ElementRef<typeof SelectPrimitive.Separator>,
    React.ComponentPropsWithoutRef<typeof SelectPrimitive.Separator>
>(({ className, ...props }, ref) => (
    <SelectPrimitive.Separator ref={ref} className={cn(selectSeparatorClasses, className)} {...props} />
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
