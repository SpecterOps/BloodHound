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

export const RiskBadgePropVariants = cva('flex justify-center items-center rounded-full text-black', {
    variants: {
        type: {
            labeled: 'size-[32px] shadow-none w-auto leading-[1]',
            'sm-labeled': 'size-[24px] p-1 shadow-none w-auto leading-[1]',
            'sm-circle': 'size-4 p-1 drop-shadow bg-neutral-300',
            'md-circle': 'size-8 p-2 drop-shadow bg-neutral-300',
        },
    },
});

export const RiskBadgeContentVariants = cva('rounded-full', {
    variants: {
        outlined: {
            true: 'border-solid border-2 shadow-none bg-transparent text-main',
        },
        type: {
            labeled: 'size-max w-32 px-6 py-2.5 border-none text-center tracking-[.25px]',
            'sm-labeled': 'size-max px-1 py-1 border-none text-center font-bold min-w-14 w-14 text-sm',
            'sm-circle': 'size-full shadow-inner1xl',
            'md-circle': 'size-full shadow-inner1xl',
        },
    },
    compoundVariants: [
        {
            outlined: true,
            type: 'md-circle',
            className: 'border-2',
        },
        {
            outlined: true,
            type: 'sm-circle',
            className: 'border',
        },
    ],
});
