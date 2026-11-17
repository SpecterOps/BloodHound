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

export const sliderRootStyles =
    'relative flex w-full touch-none select-none items-start gap-4 data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50';

export const sliderControlStyles = 'group/slider relative flex w-full grow items-center py-3';

export const sliderTrackStyles =
    'relative h-1 w-full grow overflow-hidden rounded-full bg-neutral-200 data-[disabled]:bg-disabled';

export const sliderIndicatorStyles = cva(
    'h-full rounded-full group-has-[:focus-visible]/slider:bg-secondary data-[disabled]:bg-disabled',
    {
        variants: {
            active: {
                true: 'bg-primary',
                false: 'bg-neutral-200',
            },
        },
        defaultVariants: {
            active: false,
        },
    }
);

export const sliderThumbStyles = cva(
    'group/thumb relative size-4 rounded-full border border-[color:var(--slider-thumb-border)] shadow-outer-1 has-[:focus-visible]:bg-[color:var(--slider-thumb-focus-fill)] data-[disabled]:bg-disabled',
    {
        variants: {
            active: {
                true: 'bg-[color:var(--slider-thumb-active-fill)]',
                false: 'bg-[color:var(--slider-thumb-fill)]',
            },
        },
        defaultVariants: {
            active: false,
        },
    }
);

export const sliderThumbDotStyles = cva(
    [
        'pointer-events-none absolute left-1/2 top-1/2 size-1.5 -translate-x-1/2 -translate-y-1/2 rounded-full',
        'group-has-[:focus-visible]/thumb:block group-has-[:focus-visible]/thumb:bg-secondary',
        'group-data-[disabled]/thumb:hidden',
    ],
    {
        variants: {
            active: {
                true: 'block bg-primary',
                false: 'hidden',
            },
        },
        defaultVariants: {
            active: false,
        },
    }
);

export const sliderStyles = {
    invalidValueSpacing: 'mb-5',
    controlWrapper: 'min-w-0 flex-1',
    ticks: 'relative h-7 text-[10px] leading-4',
    disabledText: 'text-disabled',
    text: 'text-main',
    tick: 'absolute flex -translate-x-1/2 rtl:translate-x-1/2 flex-col items-center gap-1',
    tickDot: 'size-1 rounded-full',
    disabledTick: 'bg-disabled',
    activeTick: 'bg-[color:var(--slider-current-color,var(--primary))]',
    inactiveTick: 'bg-[color:var(--slider-tick-fill)]',
    value: 'relative shrink-0',
    valueInput: 'h-7 w-16 px-2 py-1 text-xs',
    valueInputWithSuffix: 'pr-5',
    valueSuffix: 'pointer-events-none absolute end-2 top-1.5 text-xs',
    valueError: 'absolute start-0 top-full text-xs text-status-error-main',
};
