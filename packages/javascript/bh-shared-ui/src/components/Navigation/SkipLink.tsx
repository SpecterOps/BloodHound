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

import { AnchorHTMLAttributes, FC } from 'react';
import { cn } from '../../utils';

export const SkipLink: FC<AnchorHTMLAttributes<HTMLAnchorElement>> = ({ children, className, ...props }) => (
    <a
        className={cn(
            'sr-only focus:not-sr-only focus:fixed focus:left-3 focus:top-3 focus:z-[10000]',
            'focus:rounded focus:bg-neutral-1 focus:px-4 focus:py-2 focus:text-primary focus:shadow-lg focus:focus-ring',
            className
        )}
        {...props}>
        {children}
    </a>
);
