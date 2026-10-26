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

export const switchWrapperStyles = cn('inline-flex items-center rounded-sm py-1 px-0.5 gap-1 focus-within:focus-ring');

export const switchRootStyles = cn(
    'group flex h-3 w-6 items-center rounded-3xl transition-all ease-in-out bg-[color:var(--switch-fill)] data-[state=checked]:bg-primary disabled:data-[state=checked]:bg-[color:var(--switch-disabled-fill)] disabled:cursor-not-allowed disabled:bg-[color:var(--switch-disabled-fill)] focus:outline-none'
);

export const switchThumbStyles = cn(
    'h-2.5 w-2.5 translate-x-px rounded-full shadow-outer-1 transition-all ease-in-out bg-[color:var(--switch-thumb-fill)] data-[state=checked]:translate-x-[13px]'
);

export const LabelVariants = cva('text-base text-main transition-all ease-in-out px-0.5', {
    variants: {
        position: {
            left: 'mr-2',
            right: 'ml-2',
        },
        disabled: {
            true: 'cursor-not-allowed',
            false: 'hover:cursor-pointer',
        },
    },
    defaultVariants: {
        position: 'right',
        disabled: false,
    },
});
