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

import * as RadioGroupPrimitive from '@radix-ui/react-radio-group';
import React from 'react';
import { Label } from '../Label/Label';
import { cn } from '../utils';

interface RadioGroupProps extends React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Root> {
    row?: boolean;
}

/**
 * Implementation:
 *
 * `<RadioGroup  value={}  onValueChange={(value) =>  doStuff(value)}>...`
 *
 * Props:
 * - row?: boolean -- optional prop to display radio items horizontally
 */

const RadioGroup = React.forwardRef<React.ElementRef<typeof RadioGroupPrimitive.Root>, RadioGroupProps>(
    ({ className, children, row, ...props }, ref) => (
        <RadioGroupPrimitive.Root ref={ref} className={cn(row ? 'flex' : '', className)} {...props}>
            {children}
        </RadioGroupPrimitive.Root>
    )
);
RadioGroup.displayName = RadioGroupPrimitive.Root.displayName;

interface RadioGroupItemProps extends React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Item> {
    label: string;
    value: string;
}

const RadioItem = React.forwardRef<React.ElementRef<typeof RadioGroupPrimitive.Item>, RadioGroupItemProps>(
    ({ className, label, value, ...props }, ref) => (
        <div className='flex items-center mb-1 px-1 mr-4 rounded-md [&:has(:focus-visible)]:ring-2 [&:has(:focus-visible)]:ring-secondary '>
            <Label className='font-normal cursor-pointer flex items-center'>
                <RadioGroupPrimitive.Item
                    value={value}
                    ref={ref}
                    className={cn(
                        'w-4 h-4 rounded rounded-full border border-neutral-5 dark:border-neutral-light-5 hover:border-secondary dark:hover:border-secondary-variant-2 active:border-primary-variant mr-2 relative data-[disabled]:border-neutral-4 data-[disabled]:bg-neutral-2 peer',
                        className
                    )}
                    {...props}>
                    <RadioGroupPrimitive.Indicator
                        className={cn(
                            'bg-primary dark:bg-secondary-variant-2 w-2 h-2 rounded rounded-full absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2'
                        )}
                    />
                </RadioGroupPrimitive.Item>
                <span>{label}</span>
            </Label>
        </div>
    )
);
RadioItem.displayName = RadioGroupPrimitive.Item.displayName;

export { RadioGroup, RadioItem };
