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

interface RadioGroupProps extends React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Root> {}

/**
 * Description for RadioGroup
 */
const RadioGroup = React.forwardRef<React.ElementRef<typeof RadioGroupPrimitive.Root>, RadioGroupProps>(
    ({ className, children, ...props }, ref) => (
        <RadioGroupPrimitive.Root ref={ref} className={cn(className)} {...props}>
            {children}
        </RadioGroupPrimitive.Root>
    )
);
RadioGroup.displayName = RadioGroupPrimitive.Root.displayName;

interface RadioGroupItemProps extends React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Item> {
    label: string;
    value: string;
}

// const indicatorStyle = 'w-4 h-4 rounded rounded-full border';
const RadioItem = React.forwardRef<React.ElementRef<typeof RadioGroupPrimitive.Item>, RadioGroupItemProps>(
    ({ className, label, value, ...props }, ref) => (
        <div className='flex items-center mb-1'>
            <RadioGroupPrimitive.Item
                value={value}
                ref={ref}
                className={cn(
                    'w-4 h-4 rounded rounded-full border border-neutral-5 hover:border-secondary active:border-primary-variant mr-3 relative data-[disabled]:border-neutral-4 data-[disabled]:bg-neutral-2 peer',
                    className
                )}
                {...props}>
                <RadioGroupPrimitive.Indicator
                    className={cn(
                        'bg-primary w-2 h-2 rounded rounded-full absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2'
                    )}
                />
            </RadioGroupPrimitive.Item>
            <Label className='font-normal'>{label}</Label>
        </div>
    )
);
RadioItem.displayName = RadioGroupPrimitive.Item.displayName;

export { RadioGroup, RadioItem };
