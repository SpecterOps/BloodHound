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

export const LoadingSpinnerVariants = cva('LoadingSpinner inline-flex text-primary', {
    variants: {
        variant: {
            inline: 'size-10',
            overlay: 'absolute inset-0 z-[1000] items-center justify-center bg-black/20',
        },
    },
    defaultVariants: {
        variant: 'inline',
    },
});

export const LoadingSpinnerIconVariants = cva(
    [
        'block animate-loading-spinner motion-reduce:animate-none',
        // Cut a thin ring out of the Font Awesome disc.
        '[--spinner-ring:radial-gradient(farthest-side,transparent_calc(84%_-_0.5px),#000_84%,#000_calc(100%_-_0.5px),transparent_100%)]',
        // Reveal an animated arc of that ring.
        '[--spinner-arc:conic-gradient(from_calc(90deg_+_var(--doodle-spinner-start,0deg)),#000_var(--doodle-spinner-sweep,225deg),transparent_0)]',
        '[mask-image:var(--spinner-ring),var(--spinner-arc)] [mask-composite:intersect]',
    ],
    {
        variants: {
            variant: {
                inline: 'size-full',
                overlay: 'size-10',
            },
        },
        defaultVariants: {
            variant: 'inline',
        },
    }
);
