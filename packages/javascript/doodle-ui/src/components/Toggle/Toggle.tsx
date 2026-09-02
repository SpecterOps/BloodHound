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
    'inline-flex items-center justify-center rounded text-sm transition-colors gap-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 border border-toggle-btn-border bg-toggle-btn-fill text-main hover:border-secondary hover:bg-secondary hover:text-common-white dark:hover:text-common-dark hover:underline focus:outline-none focus-visible:focus-ring focus-visible:border-secondary focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark active:border-primary-variant active:bg-primary-variant active:text-common-white dark:active:text-common-dark data-[state=on]:border-primary data-[state=on]:bg-primary data-[state=on]:text-common-white dark:data-[state=on]:text-common-dark disabled:cursor-not-allowed disabled:border-btn-disabled-fill disabled:bg-btn-disabled-fill disabled:text-text-disabled',
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
