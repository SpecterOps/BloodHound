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

export const radioItemLabelClasses = cn(
    'flex items-center font-normal cursor-pointer',
    'has-[:disabled]:cursor-default has-[:disabled]:text-disabled',
    'mb-1 mr-4 rounded-md px-1',
    '[&:has(:focus-visible)]:bg-radio-label-focus-fill'
);

export const radioItemControlClasses = cn(
    'w-4 h-4 rounded-full border border-radio-border-default',
    'hover:border-radio-border-hover',
    'active:border-primary-variant',
    'mr-2 relative',
    'focus:outline-none',
    'focus-visible:focus-ring',
    'data-[disabled]:border-radio-disabled-border data-[disabled]:bg-radio-disabled-fill',
    'peer'
);

export const radioItemIndicatorClasses =
    'bg-radio-indicator-fill w-2 h-2 rounded-full absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2';
