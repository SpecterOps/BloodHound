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

export const ToggleVariants = cva(
    cn(
        'inline-flex items-center justify-center gap-2 rounded border border-[color:var(--toggle-btn-border)] bg-[color:var(--toggle-btn-fill)] text-sm text-main transition-colors',
        '[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0',
        'hover:border-[color:var(--toggle-hover-border)] hover:bg-secondary hover:text-[color:var(--toggle-highlight-text)] hover:underline',
        'focus:outline-none',
        'focus-visible:border-[color:var(--toggle-hover-border)] focus-visible:bg-secondary focus-visible:text-[color:var(--toggle-highlight-text)] focus-visible:focus-ring',
        'active:border-[color:var(--toggle-active-border)] active:bg-primary-variant active:text-[color:var(--toggle-highlight-text)]',
        'data-[state=on]:border-[color:var(--toggle-selected-border)] data-[state=on]:bg-primary data-[state=on]:text-[color:var(--toggle-highlight-text)]',
        'disabled:hover:data-[state=on]:text-[color:var(--toggle-highlight-text)]',
        'disabled:hover:text-[color:var(--toggle-disabled-hover-text)]',
        'disabled:cursor-not-allowed disabled:border-[color:var(--toggle-disabled-border)] disabled:bg-[color:var(--button-disabled-fill)] disabled:text-disabled'
    ),
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

export const toggleGroupClasses =
    'flex items-center justify-center gap-2 p-1 rounded-lg bg-[color:var(--toggle-group-fill)] shadow-outer-1';
