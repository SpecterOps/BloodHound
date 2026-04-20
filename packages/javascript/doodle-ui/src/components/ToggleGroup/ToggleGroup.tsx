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
import * as ToggleGroupPrimitive from '@radix-ui/react-toggle-group';
import { type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { ToggleVariants } from '../Toggle';
import { cn } from '../utils';

const ToggleGroupContext = React.createContext<VariantProps<typeof ToggleVariants>>({
    size: 'sm',
});

const ToggleGroup = React.forwardRef<
    React.ElementRef<typeof ToggleGroupPrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof ToggleGroupPrimitive.Root> & VariantProps<typeof ToggleVariants>
>(({ className, size, children, ...props }, ref) => (
    // TODO: Replace hardcoded hex colors with design token CSS variables once the token system is ready.
    <ToggleGroupPrimitive.Root
        ref={ref}
        className={cn(
            'flex items-center justify-center gap-2 p-1 rounded-lg bg-[#F4F4F4] dark:bg-[#222222] shadow-[0_1px_2px_0_rgba(0,0,0,0.30)]',
            className
        )}
        {...props}>
        <ToggleGroupContext.Provider value={{ size }}>{children}</ToggleGroupContext.Provider>
    </ToggleGroupPrimitive.Root>
));

ToggleGroup.displayName = ToggleGroupPrimitive.Root.displayName;

const ToggleGroupItem = React.forwardRef<
    React.ElementRef<typeof ToggleGroupPrimitive.Item>,
    React.ComponentPropsWithoutRef<typeof ToggleGroupPrimitive.Item> & VariantProps<typeof ToggleVariants>
>(({ className, children, size, ...props }, ref) => {
    const context = React.useContext(ToggleGroupContext);

    return (
        <ToggleGroupPrimitive.Item
            ref={ref}
            className={cn(
                ToggleVariants({
                    size: context.size || size,
                }),
                className
            )}
            {...props}>
            {children}
        </ToggleGroupPrimitive.Item>
    );
});

ToggleGroupItem.displayName = ToggleGroupPrimitive.Item.displayName;

export { ToggleGroup, ToggleGroupItem };
