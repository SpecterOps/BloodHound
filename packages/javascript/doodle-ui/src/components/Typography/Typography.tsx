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
import { ElementType } from 'react';
import { cn } from '../utils';
import { DEFAULT_VARIANT, Variant, variantMapping } from './utils';

// leading = line-height
// tracking = letter-spacing

export const TypographyVariants = cva('', {
    variants: {
        variant: {
            h1: 'text-3xl font-bold leading-[2rem] tracking-normal',
            h2: 'text-2xl font-medium leading-[1.75rem] tracking-normal',
            h3: 'text-[1.375rem] font-medium leading-[1.5rem] tracking-normal',
            h4: 'text-xl font-medium leading-[1.5rem] tracking-normal',
            h5: 'text-lg font-bold leading-[1.5rem] tracking-[.25px]',
            h6: 'text-base font-bold leading-[1.5rem] tracking-[.25px]',
            body1: 'text-base font-normal leading-[1.5rem] tracking-normal',
            body2: 'text-sm font-normal leading-[1.375rem] tracking-normal',
            subtitle1: 'text-[.938rem] font-normal leading-[1.5rem] tracking-[.25px]',
            subtitle2: 'text-[.8125rem] font-medium leading-[1.375rem] tracking-[.25px]',
            caption: 'text-xs font-normal leading-[1.25rem] tracking-[.25px]',
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

const Typography = ({ variant, component, children, className, ...rest }: TypographyProps) => {
    const Tag = (component || variantMapping[variant ?? DEFAULT_VARIANT]) as ElementType;

    return (
        <Tag className={cn(TypographyVariants({ variant }), `typography-${variant}`, className)} {...rest}>
            {children}
        </Tag>
    );
};

Typography.displayName = 'Typography';

export { Typography };
