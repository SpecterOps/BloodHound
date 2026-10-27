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

export const accordionItemClasses = 'bg-[color:var(--accordion-fill)] rounded-lg mb-4';

export const accordionHeaderClasses = cn(
    'flex flex-1 gap-1 items-center rounded-md font-bold transition-all',
    'hover:underline',
    'focus:outline-none',
    'focus-visible:focus-ring focus-visible:underline',
    '[&[data-state=open]>svg]:rotate-180',
    'p-4'
);

export const accordionContentClasses =
    'overflow-hidden text-sm transition-all data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down';
