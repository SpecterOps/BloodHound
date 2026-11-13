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
import { Label } from '../Label';
import { cn } from '../utils';
import { radioItemControlClasses, radioItemIndicatorClasses, radioItemLabelClasses } from './RadioGroup.styles';

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
        <RadioGroupPrimitive.Root
            ref={ref}
            className={cn('rounded-md focus-within:focus-ring', row ? 'flex' : '', className)}
            {...props}>
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
        <Label className={radioItemLabelClasses}>
            <RadioGroupPrimitive.Item value={value} ref={ref} className={radioItemControlClasses} {...props}>
                <RadioGroupPrimitive.Indicator className={radioItemIndicatorClasses} />
            </RadioGroupPrimitive.Item>
            <span className={className}>{label}</span>
        </Label>
    )
);
RadioItem.displayName = RadioGroupPrimitive.Item.displayName;

export { RadioGroup, RadioItem };
