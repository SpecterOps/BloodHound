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

export const buttonBaseClasses = cn(
    'inline-flex items-center justify-center whitespace-nowrap rounded-3xl transition-colors',
    'focus:outline-none focus-visible:focus-ring',
    'disabled:text-[color:var(--button-disabled-text)] disabled:pointer-events-none disabled:opacity-50',
    'has-[svg]:gap-2 [&>svg]:shrink-0'
);

export const primaryClasses = cn(
    'bg-primary text-[color:var(--button-primary-text)] shadow-outer-1',
    'hover:bg-secondary',
    'focus-visible:bg-secondary',
    'active:bg-[color:var(--button-primary-active-fill)]',
    'disabled:shadow-none disabled:bg-[color:var(--button-disabled-fill)]'
);

export const secondaryClasses = cn(
    'bg-[color:var(--button-secondary-fill)] text-[color:var(--button-secondary-text)] shadow-outer-1',
    'hover:bg-secondary hover:text-[color:var(--button-highlight-text)]',
    'focus-visible:bg-secondary focus-visible:text-[color:var(--button-highlight-text)]',
    'active:bg-[color:var(--button-secondary-active-fill)] active:text-[color:var(--button-secondary-text)]',
    'disabled:bg-[color:var(--button-disabled-fill)] disabled:shadow-none'
);

export const ButtonVariants = cva(buttonBaseClasses, {
    variants: {
        variant: {
            primary: primaryClasses,
            secondary: secondaryClasses,
            // TODO - remove in BED-7642
            // used in DropdownTriggerContents & EnvironmentSelectorTrigger
            /**
             * @deprecated Use TextButton instead.
             */
            transparent: cn(
                'border border-[color:var(--button-transparent-border)] bg-transparent text-main',
                'hover:border-secondary hover:bg-secondary hover:text-[color:var(--button-highlight-text)] hover:no-underline',
                'focus-visible:border-primary focus-visible:bg-secondary focus-visible:text-[color:var(--button-highlight-text)]'
            ),
            // TODO - legacy, remove in BED-7635
            /**
             * @deprecated Use IconButton instead.
             */
            icon: cn(
                'rounded-full bg-[color:var(--button-icon-fill)] text-common-dark shadow-outer-1 has-[svg]:p-2',
                'hover:border-2 hover:border-primary',
                'active:border-none'
            ),
        },
        // TODO - remove as the only usage of this is with variant="text"
        /**
         * @deprecated Use TextButton instead.
         */
        fontColor: {
            primary: 'text-primary',
        },
        /**
         * @deprecated .
         */
        size: {
            // TODO remove small variant in BED-7635
            small: 'h-9 px-4 py-1 text-xs',
            medium: 'h-10 px-6 py-2 text-sm/5',
            // TODO remove large variant in BED-7635
            large: 'h-11 px-8 py-3 text-base',
        },
    },

    defaultVariants: {
        variant: 'primary',
        size: 'medium',
    },
});

export const TextButtonBaseClasses = cn(
    buttonBaseClasses,
    'px-2 py-1 has-[svg]:px-1',
    'active:text-[color:var(--button-text-active)]',
    'hover:text-secondary',
    'focus-visible:ring-0 focus-visible:ring-transparent focus-visible:text-secondary'
);

export const TextButtonVariants = cva(TextButtonBaseClasses, {
    variants: {
        fontColor: {
            default: 'text-main',
            primary: 'text-primary',
        },
    },
    defaultVariants: {
        fontColor: 'default',
    },
});
