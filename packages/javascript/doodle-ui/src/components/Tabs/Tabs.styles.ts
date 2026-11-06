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

export const tabsListClasses = 'inline-flex text-base text-[color:var(--tabs-text)] items-center justify-center gap-4';

export const tabsTriggerClasses = cn(
    'grid grid-cols-1 text-center whitespace-nowrap p-2',
    'focus:outline-none',
    'focus-visible:focus-ring',
    'disabled:pointer-events-none disabled:opacity-50',
    'data-[state=active]:font-bold data-[state=active]:border-b-[color:var(--tabs-active-accent)] data-[state=active]:border-b-[1px] data-[state=active]:text-[color:var(--tabs-active-accent)]'
);

export const tabsContentClasses = 'mt-2 focus:outline-none focus-visible:focus-ring';
