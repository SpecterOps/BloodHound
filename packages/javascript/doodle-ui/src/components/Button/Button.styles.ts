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
import { cn } from '../utils';

export const buttonBackgroundColors = {
    'secondary-btn-fill': palette.neutral.light[300],
    'secondary-btn-fill-dark': palette.neutral.dark[700],
    'secondary-btn-active-fill': palette.neutral.light[400],
    'secondary-btn-active-fill-dark': palette.neutral.dark[700],
    'icon-btn-fill': palette.neutral.light[400],
    'btn-disabled-fill': palette.neutral.light[200],
    'btn-disabled-fill-dark': palette.neutral.dark[700],
};

export const buttonBorderColors = {
    'transparent-btn-border': palette.neutral.light[400],
    'transparent-btn-border-dark': palette.neutral.dark[700],
    'btn-disabled-fill': palette.neutral.light[200],
    'btn-disabled-fill-dark': palette.neutral.dark[700],
};

export const buttonBaseClasses = [
    'inline-flex items-center justify-center whitespace-nowrap rounded-3xl transition-colors',
    'focus:outline-none focus-visible:focus-ring',
    'disabled:text-[#616161] dark:disabled:text-[#A6A6A6] disabled:pointer-events-none disabled:opacity-50',
    'has-[svg]:gap-2 [&>svg]:shrink-0',
];

export const primaryClasses = [
    'bg-primary text-common-white shadow-outer-1 dark:text-common-dark',
    'hover:bg-secondary',
    'focus-visible:bg-secondary',
    'active:bg-[#0D0A30] dark:active:bg-[#8D8BF8]',
    // Implement text-common when token experiment is ready - #0D0A30 matches light.primary.variant, #8D8BF8 matches dark.primary.variant.
    'disabled:shadow-none disabled:bg-[#E3E7EA] dark:disabled:bg-[#2E2E2E]',
    // disabled is neutral.light[200], text -> common.disabled // disabled:bg -> neutral.dark[700] or common.disabled.dark (token experiment)
];

export const secondaryClasses = [
    'bg-secondary-btn-fill text-common-dark shadow-outer-1 dark:bg-secondary-btn-fill-dark dark:text-common-white',
    'hover:bg-secondary hover:text-common-white dark:hover:text-common-dark',
    'focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
    'active:bg-secondary-btn-active-fill active:text-common-dark dark:active:bg-secondary-btn-active-fill-dark dark:active:text-common-white',
    'disabled:bg-btn-disabled-fill disabled:shadow-none dark:disabled:bg-btn-disabled-fill-dark',
];

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
            transparent: [
                'border border-transparent-btn-border bg-transparent text-main dark:border-transparent-btn-border-dark',
                'hover:border-secondary hover:bg-secondary hover:text-common-white hover:no-underline dark:hover:text-common-dark',
                'focus-visible:border-primary focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
            ],
            // TODO - legacy, remove in BED-7635
            /**
             * @deprecated Use IconButton instead.
             */
            icon: [
                'rounded-full bg-icon-btn-fill text-common-dark shadow-outer-1 has-[svg]:p-2',
                'hover:border-2 hover:border-primary',
                'active:border-none',
            ],
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
    ...buttonBaseClasses,
    'px-2 py-1 has-[svg]:px-1',
    'active:text-[#0D0A30] dark:active:text-primary',
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
