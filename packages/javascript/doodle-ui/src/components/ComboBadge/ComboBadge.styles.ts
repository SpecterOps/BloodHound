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

export const ComboBadgeVariants = cva('relative leading-[1] z-10 h-8 text-nowrap', {
    variants: {
        type: {
            slideLeft: 'clip-right-rounded flex flex-row-reverse',
            slideRight: 'clip-left-rounded',
            inlineSlideLeft: 'clip-right-rounded flex items-center flex-row-reverse',
            inlineSlideRight: 'clip-left-rounded flex items-center flex-row',
        },
    },
});

export const AdornmentVariants = cva(
    'bg-primary text-[color:var(--combo-badge-adornment-text)] rounded px-4 py-2 z-10 transition-[transform] duration-300',
    {
        variants: {
            type: {
                slideLeft: 'absolute left-2 bottom-0 top-0 peer-hover:-translate-x-full',
                slideRight: 'absolute right-2 bottom-0 top-0 peer-hover:translate-x-full',
                inlineSlideLeft: 'inline-block peer-hover:translate-x-0 translate-x-full -mr-2',
                inlineSlideRight: 'inline-block peer-hover:translate-x-0 -translate-x-full -ml-2',
            },
            displayAdornment: {
                true: 'translate-x-0 peer-hover:translate-x-0',
            },
        },
        compoundVariants: [
            {
                type: ['slideLeft'],
                displayAdornment: true,
                className: 'inline-block relative left-0 -mr-2',
            },
            {
                type: ['slideRight'],
                displayAdornment: true,
                className: 'inline-block relative right-0 -ml-2',
            },
        ],
    }
);

export const comboBadgeLabelClasses =
    'bg-[color:var(--combo-badge-fill)] text-[color:var(--combo-badge-text)] rounded inline-block px-4 py-2 relative z-20';
