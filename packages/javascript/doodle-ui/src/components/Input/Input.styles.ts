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
import { common, dark, palette, text } from '../../tailwind/colors';

export const inputBorderColors = {
    'input-border-default': palette.grey[700],
    'input-border-default-dark': palette.neutral.light[400],
    'input-border-focus': 'var(--control-accent)',
    'input-border-disabled': palette.neutral.light[900],
    'input-border-disabled-dark': palette.neutral.dark[900],
    'input-outlined-border-default': 'var(--border-control)',
    'input-outlined-border-hover': 'var(--secondary)',
};

export const inputRingColors = {
    'input-border-default': palette.grey[700],
    'input-border-default-dark': palette.neutral.light[400],
};

export const inputTextColors = {
    'input-placeholder-text': text.placeholder,
    'input-placeholder-text-dark': dark.input.placeholder,
};

export const inputBackgroundColors = {
    'input-fill': common.white,
    'input-fill-dark': palette.neutral.dark[50],
    'input-fill-disabled': palette.neutral.light[100],
    'input-fill-disabled-dark': palette.neutral.dark[400],
    'input-outlined-border-hover': 'var(--secondary)',
};

const inputBaseClasses = [
    'flex h-10 w-full',
    'text-base text-main placeholder:text-[color:var(--input-placeholder-text)]',
    'disabled:cursor-not-allowed disabled:opacity-50',
    'file:cursor-pointer file:border-0 file:bg-transparent file:pr-3 file:text-sm file:font-medium file:text-main',
];

export const InputVariants = cva(inputBaseClasses, {
    variants: {
        variant: {
            outlined: [
                'rounded-md border border-[color:var(--input-outlined-border-default)] bg-[color:var(--input-fill)] px-3 py-2 text-sm',
                'enabled:hover:border-[color:var(--input-outlined-border-hover)]',
                'focus:outline-none focus:focus-ring focus:border-[color:var(--input-outlined-border-default)] enabled:hover:focus:border-[color:var(--input-outlined-hover-focus-border)]',
                'disabled:border-[color:var(--input-border-disabled)] disabled:bg-[color:var(--input-disabled-fill)] disabled:text-disabled disabled:placeholder:text-disabled disabled:opacity-100',
                'aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main aria-[invalid=true]:enabled:hover:focus:border-status-error-main',
            ],
            underlined: [
                'rounded-sm border-b border-b-[color:var(--input-border-default)] bg-transparent',
                'hover:border-b-2',
                'focus:border-b-2 focus:border-[color:var(--input-border-focus)] focus:outline-none focus:focus-ring',
            ],
        },
        intent: {
            time: [
                'appearance-none',
                '[&::-webkit-calendar-picker-indicator]:hidden',
                '[&::-webkit-calendar-picker-indicator]:appearance-none',
            ],
        },
    },
    defaultVariants: {
        variant: 'underlined',
    },
});

// '--input-label': 'var(--text-main)',
// '--input-border-hover': 'var(--secondary)',
