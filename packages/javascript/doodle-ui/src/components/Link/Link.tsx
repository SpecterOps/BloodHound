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

import { faExternalLink } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';
import { linkExternalIconClasses, linkVariants } from './Link.styles';

export type LinkProps = Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'children' | 'href' | 'target' | 'rel'> &
    VariantProps<typeof linkVariants> & {
        children: string;
        href: string;
        allowReferrer?: boolean;
    };

export const Link = React.forwardRef<HTMLAnchorElement, LinkProps>(
    ({ className, variant = 'styled', children, href, allowReferrer = false, ...props }, ref) => {
        return (
            <a
                ref={ref}
                href={href}
                target='_blank'
                rel={`noopener ${allowReferrer ? '' : 'noreferrer'}`}
                className={cn(linkVariants({ variant }), className)}
                {...props}>
                {children}{' '}
                <FontAwesomeIcon icon={faExternalLink} aria-hidden='true' className={linkExternalIconClasses} />
                <span className='sr-only'> (opens in a new tab)</span>
            </a>
        );
    }
);

Link.displayName = 'Link';
