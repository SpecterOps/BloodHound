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
import { palette } from '../../tailwind/colors';

export const toggleBackgroundColors = {
    'toggle-btn-fill': 'var(--surface)',
    'toggle-group-fill': 'var(--surface-subtle)',
};

export const toggleBorderColors = {
    'toggle-btn-border': palette.neutral.light[500],
    'toggle-btn-border-dark': palette.neutral.dark[600],
};

export const ToggleVariants = cva(
    [
        'inline-flex items-center justify-center gap-2 rounded border border-toggle-btn-border dark:border-toggle-btn-border-dark bg-toggle-btn-fill text-sm text-main transition-colors',
        '[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0',
        'hover:border-secondary hover:bg-secondary hover:text-common-white hover:underline dark:hover:text-common-dark',
        'focus:outline-none focus-visible:border-secondary focus-visible:bg-secondary focus-visible:text-common-white focus-visible:focus-ring dark:focus-visible:text-common-dark',
        'active:border-primary-variant active:bg-primary-variant active:text-common-white dark:active:text-common-dark',
        'data-[state=on]:border-primary data-[state=on]:bg-primary data-[state=on]:text-common-white dark:data-[state=on]:text-common-dark',
        'disabled:cursor-not-allowed disabled:border-btn-disabled-fill disabled:bg-btn-disabled-fill disabled:text-disabled',
    ],
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
    'flex items-center justify-center gap-2 p-1 rounded-lg bg-toggle-group-fill shadow-outer-1';
