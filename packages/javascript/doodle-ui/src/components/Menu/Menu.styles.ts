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

const menuPopupBaseStyles = [
    'z-50 min-w-[8rem] overflow-hidden rounded-md bg-[color:var(--menu-content-fill)] p-1 shadow-md',
    'data-[state=open]:animate-in data-[state=closed]:animate-out',
    'data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0',
];

export const menuContentStyles = [
    ...menuPopupBaseStyles,
    'border-[color:var(--menu-content-border)]',
    'data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2',
];

export const menuItemStyles = [
    'relative flex cursor-pointer select-none items-center rounded-lg border-none p-2 text-sm outline-none',
    'data-[highlighted]:border-[color:var(--menu-item-highlight-border)] data-[highlighted]:bg-[color:var(--menu-item-highlight-fill)] data-[highlighted]:text-[color:var(--menu-item-highlight-text)]',
    'data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)]',
    'data-[highlighted]:active:bg-[color:var(--menu-item-active-fill)]',
    'data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:bg-[color:var(--menu-item-disabled-fill)] data-[disabled]:text-[color:var(--menu-item-disabled-text)]',
];

export const menuSubTriggerStyles = [
    'relative flex cursor-pointer select-none items-center rounded-lg border border-transparent px-2 py-1.5 text-sm outline-none',
    'data-[highlighted]:border-[color:var(--menu-item-highlight-border)] data-[highlighted]:bg-[color:var(--menu-item-highlight-fill)] data-[highlighted]:text-[color:var(--menu-item-highlight-text)]',
    'data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)] data-[state=open]:shadow-[inset_3px_0_0_var(--focus-ring)]',
    'data-[state=open]:bg-[color:var(--menu-item-highlight-fill)] data-[state=open]:text-[color:var(--menu-item-highlight-text)]',
];

export const menuSubContentStyles = [
    ...menuPopupBaseStyles,
    'border-[color:var(--menu-sub-content-border)]',
    'data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2',
];

export const menuSeparatorStyles = 'my-1 h-px bg-[color:var(--menu-separator-fill)]';
