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
import { cva } from 'class-variance-authority';
import { ElementType, forwardRef } from 'react';
import { cn } from '../utils';
import { DEFAULT_VARIANT, Variant, variantMapping } from './utils';

// leading = line-height
// tracking = letter-spacing

export const TypographyVariants = cva('break-words', {
    variants: {
        variant: {
            h1: 'font-heading text-2xl font-bold leading-7 tracking-normal text-text-main',
            h2: 'font-heading text-[1.375rem] font-bold leading-6 tracking-normal text-text-main',
            h3: 'font-heading text-xl font-bold leading-[1.375rem] tracking-normal text-text-main',
            h4: 'font-heading text-xl font-semibold leading-[1.375rem] tracking-normal text-text-main',
            h5: 'font-heading text-lg font-bold leading-5 tracking-[.25px] text-text-main',
            h6: 'font-heading text-base font-semibold leading-[1.125rem] tracking-[.25px] text-text-main',
            body1: 'font-sans text-base font-normal leading-6 tracking-normal text-text-muted dark:text-text-main',
            body2: 'font-sans text-sm font-normal leading-[1.375rem] tracking-normal text-text-muted dark:text-text-main',
            subtitle1: 'font-sans text-[.9375rem] font-medium leading-6 tracking-[.25px] text-text-main',
            subtitle2: 'font-sans text-[.8125rem] font-medium leading-[1.375rem] tracking-[.25px] text-text-main',
            caption: 'font-sans text-xs font-normal leading-5 tracking-[.25px] text-text-muted dark:text-text-main',
        },
    },
    defaultVariants: {
        variant: DEFAULT_VARIANT,
    },
});

interface TypographyProps extends React.HTMLAttributes<HTMLElement> {
    variant?: Variant;
    component?: ElementType;
}

const Typography = forwardRef<HTMLElement, TypographyProps>(
    ({ variant, component, children, className, ...rest }, ref) => {
        const Tag = (component || variantMapping[variant ?? DEFAULT_VARIANT]) as ElementType;

        return (
            <Tag
                ref={ref}
                className={cn(TypographyVariants({ variant }), `typography-${variant}`, className)}
                {...rest}>
                {children}
            </Tag>
        );
    }
);

Typography.displayName = 'Typography';

export { Typography };
