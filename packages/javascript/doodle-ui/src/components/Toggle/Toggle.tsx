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
import * as TogglePrimitive from '@radix-ui/react-toggle';
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

const ToggleVariants = cva(
    'inline-flex items-center justify-center rounded text-sm transition-colors gap-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 border border-neutral-light-5 bg-neutral-1 text-neutral-dark-1 hover:border-secondary hover:bg-secondary hover:text-neutral-1 hover:underline focus:outline-none focus-visible:focus-ring focus-visible:border-secondary focus-visible:bg-secondary focus-visible:text-neutral-1 active:border-primary-variant active:bg-primary-variant active:text-neutral-1 data-[state=on]:border-primary data-[state=on]:bg-primary data-[state=on]:text-neutral-1 disabled:cursor-not-allowed disabled:border-neutral-light-3 disabled:bg-neutral-light-3 disabled:text-text-disabled dark:border-neutral-dark-1 dark:bg-neutral-dark-1 dark:text-neutral-light-1 dark:hover:border-secondary dark:hover:bg-secondary dark:hover:text-neutral-dark-1 dark:focus-visible:border-secondary dark:focus-visible:bg-secondary dark:focus-visible:text-neutral-dark-1 dark:active:border-primary-variant dark:active:bg-primary-variant dark:active:text-neutral-dark-1 dark:data-[state=on]:border-primary dark:data-[state=on]:bg-primary dark:data-[state=on]:text-neutral-dark-1 dark:disabled:border-neutral-dark-5 dark:disabled:bg-neutral-dark-5 dark:disabled:text-text-disabled',
    {
        variants: {
            size: {
                sm: 'h-6 px-2 py-1',
                lg: 'h-10 px-2',
            },
        },
        defaultVariants: {
            size: 'sm',
        },
    }
);

const Toggle = React.forwardRef<
    React.ElementRef<typeof TogglePrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof TogglePrimitive.Root> & VariantProps<typeof ToggleVariants>
>(({ className, size, ...props }, ref) => (
    <TogglePrimitive.Root ref={ref} className={cn(ToggleVariants({ size }), className)} {...props} />
));

Toggle.displayName = TogglePrimitive.Root.displayName;

export { Toggle, ToggleVariants };
