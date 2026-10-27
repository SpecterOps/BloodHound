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

import { TextButtonVariants } from '../Button/Button.styles';
import { cn } from '../utils';

export const calendarDropdownTriggerClasses =
    'py-2 px-1 rounded-sm border-0 h-7 w-fit font-medium text-base text-[color:var(--calendar-caption-text)] [.is-between_&]:hidden [.is-end_&]:hidden [.is-start.is-end_&]:flex';

export const calendarSelectItemClasses = 'text-sm min-w-[var(--radix-popper-anchor-width)]';

export const calendarClassNames = {
    months: 'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0',
    month: 'space-y-4',
    caption_start: 'is-start',
    caption_between: 'is-between',
    caption_end: 'is-end',
    caption: 'flex justify-center pt-1 relative items-center gap-2',
    caption_label:
        'flex h-7 font-medium text-[color:var(--calendar-caption-text)] justify-center items-center grow [.is-multiple_&]:flex',
    caption_dropdowns: 'flex justify-center gap-1 grow dropdowns pl-8 pr-9',
    multiple_months: 'is-multiple',
    vhidden: 'hidden [.is-between_&]:flex [.is-end_&]:flex [.is-start.is-end_&]:hidden',
    nav: "flex items-center [&:has([name='previous-month'])]:order-first [&:has([name='next-month'])]:order-last gap-1",
    nav_button: cn(TextButtonVariants(), 'h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100'),
    nav_button_previous: 'absolute left-1',
    nav_button_next: 'absolute right-1',
    table: 'w-full border-collapse space-y-1',
    head_row: 'flex',
    head_cell: 'rounded-md w-9 font-normal text-[0.8rem] text-[color:var(--calendar-heading-text)] opacity-70',
    row: 'flex w-full mt-2',
    cell: cn(
        'h-9 w-9 text-center text-sm p-0 relative',
        '[&:has([aria-selected]:not(.day-range))]:rounded-md',
        '[&:has([aria-selected].day-range-end)]:rounded-r-md',
        '[&:has([aria-selected].day-range-start)]:rounded-l-md',
        '[&:has([aria-selected].day-outside)]:bg-[color:var(--calendar-outside-selected-fill)]',
        '[&:has([aria-selected])]:bg-[color:var(--calendar-cell-selected-fill)]',
        'first:[&:has([aria-selected])]:rounded-l-md',
        'last:[&:has([aria-selected])]:rounded-r-md',
        'focus-within:relative focus-within:z-20'
    ),
    day: cn(
        TextButtonVariants(),
        'h-9 w-9 p-0 font-normal',
        'aria-selected:opacity-100',
        'hover:no-underline hover:bg-[color:var(--calendar-day-hover-fill)]',
        'dark:aria-selected:text-[color:var(--calendar-selected-day-text)] aria-selected:hover:bg-secondary aria-selected:hover:text-[color:var(--calendar-selected-hover-text)]',
        'rounded-md'
    ),
    day_range_end: 'day-range-end day-range',
    day_range_start: 'day-range-start day-range',
    day_selected: 'bg-[color:var(--calendar-selected-fill)] [&&]:text-[color:var(--calendar-selected-text)]',
    day_today: '[&&]:underline underline-offset-4 decoration-2',
    day_outside: 'day-outside opacity-50 aria-selected:bg-neutral-dark-1/50 aria-selected:opacity-30',
    day_disabled: 'text-neutral-dark-5 opacity-50',
    day_range_middle:
        'aria-selected:bg-[color:var(--calendar-range-fill)] aria-selected:text-[color:var(--calendar-range-text)] [&&]:dark:aria-selected:text-[color:var(--calendar-range-text)] day-range',
    day_hidden: 'invisible',
};
