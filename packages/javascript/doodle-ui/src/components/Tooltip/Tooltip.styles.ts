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

import { cn } from '../utils';

export const tooltipDefaultTriggerClasses = 'inline-flex items-center justify-center rounded-full';

export const tooltipTriggerClasses = 'focus:outline-none focus-visible:focus-ring';

export const tooltipContentClasses = cn(
    'TooltipContent',
    'text-main text-xs text-balance break-words rounded-lg border bg-neutral-light-2 border-[#939597] dark:bg-neutral-dark-5 px-3 py-1.5 text-popover-foreground shadow-outer-1',
    'z-[1700] overflow-hidden animate-in fade-in-0 zoom-in-95',
    'data-[closed]:animate-out data-[closed]:fade-out-0 data-[closed]:zoom-out-95',
    'data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2'
);

export const tooltipContentWidthClasses = {
    match_trigger: 'w-[var(--anchor-width)]',
    lg: 'w-fit max-w-[300px]',
    md: 'w-fit max-w-[225px]',
    sm: 'w-fit max-w-[150px]',
};
