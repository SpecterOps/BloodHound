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

import { ElementType, forwardRef } from 'react';
import { cn } from '../utils';

import { TypographyVariants } from './Typography.styles';
import { DEFAULT_VARIANT, Variant, variantMapping } from './utils';
export { TypographyVariants } from './Typography.styles';

// leading = line-height
// tracking = letter-spacing

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
