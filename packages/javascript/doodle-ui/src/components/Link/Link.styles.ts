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
import { cn } from '../utils';

export const styledLinkClasses = cn(
    'text-link underline underline-offset-1 decoration-secondary',
    'hover:text-link-hover hover:decoration-link',
    'focus-visible:underline focus-visible:decoration-2 focus-visible:underline-offset-2',
    'dark:no-underline dark:hover:underline'
);

export const linkVariants = cva('inline-flex items-center rounded-sm focus:outline-none focus-visible:focus-ring', {
    variants: {
        variant: {
            styled: styledLinkClasses,
            unstyled: 'text-inherit no-underline',
        },
    },
    defaultVariants: {
        variant: 'styled',
    },
});

export const linkExternalIconClasses = 'ml-0.5 shrink-0 text-[0.85em] align-baseline';
