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
import * as CheckboxPrimitive from '@radix-ui/react-checkbox';
import { cva, type VariantProps } from 'class-variance-authority';
import { Check, Minus } from 'lucide-react';
import * as React from 'react';
import { cn } from '../utils';

const CheckboxVariants = cva(
    'peer shrink-0 rounded-sm border-2 border-neutral-dark-1 dark:border-neutral-light-1 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary dark:focus-visible:ring-secondary-variant-2 focus-visible:ring-offset-2 focus-visible:ring-offset-neutral-light-1 dark:focus-visible:ring-offset-neutral-dark-1 aria-[invalid=true]:border-error dark:aria-[invalid=true]:border-error aria-[invalid=true]:text-error aria-[invalid=true]:data-[state=checked]:bg-error dark:aria-[invalid=true]:data-[state=checked]:bg-error aria-[invalid=true]:data-[state=checked]:border-error dark:aria-[invalid=true]:data-[state=checked]:border-error aria-[invalid=true]:data-[state=checked]:text-neutral-light-1 dark:aria-[invalid=true]:data-[state=checked]:text-neutral-dark-1 aria-[invalid=true]:data-[state=indeterminate]:bg-error dark:aria-[invalid=true]:data-[state=indeterminate]:bg-error aria-[invalid=true]:data-[state=indeterminate]:border-error dark:aria-[invalid=true]:data-[state=indeterminate]:border-error aria-[invalid=true]:data-[state=indeterminate]:text-neutral-light-1 dark:aria-[invalid=true]:data-[state=indeterminate]:text-neutral-dark-1',
    {
        variants: {
            size: {
                lg: 'h-[24px] w-[24px]',
                md: 'h-[18px] w-[18px]',
                sm: 'h-[12px] w-[12px]',
            },
            isDisabled: {
                false: 'data-[state=checked]:bg-neutral-dark-1 data-[state=checked]:text-neutral-light-1 data-[state=indeterminate]:bg-neutral-dark-1 data-[state=indeterminate]:text-neutral-light-1 dark:data-[state=checked]:bg-neutral-light-1 dark:data-[state=checked]:text-neutral-dark-1 dark:data-[state=indeterminate]:bg-neutral-light-1 dark:data-[state=indeterminate]:text-neutral-dark-1',
                true: 'cursor-not-allowed border-input-border-disabled dark:border-input-border-disabled bg-input-fill-disabled text-icon-disabled',
            },
        },
        defaultVariants: {
            isDisabled: false,
        },
    }
);

interface CheckboxProps
    extends React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root>,
        VariantProps<typeof CheckboxVariants> {
    icon?: React.ReactNode;
}

const Checkbox = React.forwardRef<React.ElementRef<typeof CheckboxPrimitive.Root>, CheckboxProps>(
    ({ size = 'md', icon, className, checked, disabled = false, ...props }, ref) => {
        const CheckedIcon = checked === 'indeterminate' ? Minus : Check;

        return (
            <CheckboxPrimitive.Root
                ref={ref}
                checked={checked}
                disabled={disabled}
                className={cn(CheckboxVariants({ size, className, isDisabled: disabled }))}
                {...props}>
                <CheckboxPrimitive.Indicator className={cn('flex items-center justify-center text-current')}>
                    {icon ? icon : <CheckedIcon className='h-full w-full' absoluteStrokeWidth={true} strokeWidth={3} />}
                </CheckboxPrimitive.Indicator>
            </CheckboxPrimitive.Root>
        );
    }
);
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox };
