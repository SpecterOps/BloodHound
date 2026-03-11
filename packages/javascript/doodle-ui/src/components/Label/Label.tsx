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
import * as LabelPrimitive from '@radix-ui/react-label';
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

const LabelVariants = cva('leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70', {
    variants: {
        size: {
            small: 'text-sm font-medium',
            medium: 'text-base font-bold',
            large: 'text-lg font-bold',
        },
    },
    defaultVariants: {
        size: 'medium',
    },
});

const Label = React.forwardRef<
    React.ElementRef<typeof LabelPrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root> & VariantProps<typeof LabelVariants>
>(({ className, size, ...props }, ref) => (
    <LabelPrimitive.Root ref={ref} className={cn(LabelVariants({ size, className }))} {...props} />
));

Label.displayName = LabelPrimitive.Root.displayName;

export { Label };
