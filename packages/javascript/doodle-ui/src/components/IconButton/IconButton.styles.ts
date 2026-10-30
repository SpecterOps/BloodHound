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
import { buttonBaseClasses } from '../Button/Button.styles';

const defaultIconButtonClasses = [
    'hover:text-primary dark:hover:text-primary',
    'active:bg-transparent active:text-secondary dark:active:text-secondary',
    'focus-visible:ring-transparent focus-visible:ring-offset-0 focus-visible:ring-offset-transparent focus-visible:text-primary dark:focus-visible:text-primary',
];

export const iconPrimaryClasses = [
    'bg-primary text-[color:var(--button-primary-text)] shadow-outer-1',
    'hover:bg-secondary',
    'focus-visible:bg-secondary',
    'active:bg-[color:var(--button-primary-active-fill)]',
    'disabled:shadow-none disabled:bg-[color:var(--button-disabled-fill)]',
];

export const iconSecondaryClasses = [
    'bg-[color:var(--button-secondary-fill)] text-[color:var(--button-secondary-text)] shadow-outer-1',
    'hover:bg-secondary hover:text-[color:var(--button-highlight-text)]',
    'focus-visible:bg-secondary focus-visible:text-[color:var(--button-highlight-text)]',
    'active:bg-[color:var(--button-secondary-active-fill)] active:text-[color:var(--button-secondary-text)]',
    'disabled:bg-[color:var(--button-disabled-fill)] disabled:shadow-none',
];

export const IconButtonVariants = cva(
    [
        ...buttonBaseClasses,
        'inline-grid h-fit aspect-square box-border',
        'shrink-0 place-items-center align-middle rounded-full border-0 p-2',
    ],
    {
        variants: {
            variant: {
                default: defaultIconButtonClasses,
                primary: iconPrimaryClasses,
                secondary: iconSecondaryClasses,
            },
        },
        defaultVariants: {
            variant: 'default',
        },
    }
);
