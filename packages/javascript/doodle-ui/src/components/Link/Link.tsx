// Copyright 2025 Specter Ops, Inc.
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

import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

const linkVariants = cva(
    'inline-flex items-center rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 focus-visible:ring-ring',
    {
        variants: {
            variant: {
                styled: 'text-link underline underline-offset-1 decoration-secondary-variant-2 hover:text-link hover:decoration-link dark:no-underline dark:hover:underline',
                unstyled: 'text-inherit no-underline',
            },
        },
        defaultVariants: {
            variant: 'styled',
        },
    }
);

type LinkProps = Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'children' | 'href'> &
    VariantProps<typeof linkVariants> & {
        children: string;
        href: string;
    };

export const Link = React.forwardRef<HTMLAnchorElement, LinkProps>(
    ({ className, variant, children, href, ...props }, ref) => {
        return (
            <a ref={ref} href={href} className={cn(linkVariants({ variant }), className)} {...props}>
                {children}
            </a>
        );
    }
);

Link.displayName = 'Link';
