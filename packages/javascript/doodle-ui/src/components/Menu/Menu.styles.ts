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
import { common, dark, light, palette } from '../../tailwind/colors';

export const menuBackgroundColors = {
    'menu-content-fill': common.white,
    'menu-content-fill-dark': palette.neutral.dark[400],
    'menu-item-highlight-fill': light.secondary.main,
    'menu-item-highlight-fill-dark': dark.secondary.main,
    'menu-item-active-fill': light.primary.main,
    'menu-item-active-fill-dark': dark.primary.main,
    'menu-item-disabled-fill': palette.neutral.light[200],
    'menu-item-disabled-fill-dark': palette.neutral.dark[700],
    'menu-separator-fill': palette.neutral.light[300],
    'menu-separator-fill-dark': palette.neutral.dark[600],
};

export const menuBorderColors = {
    'menu-content-border': palette.neutral.light[300],
    'menu-content-border-dark': palette.neutral.dark[600],
    'menu-item-highlight-border': light.secondary.main,
    'menu-item-highlight-border-dark': dark.secondary.main,
};

export const menuTextColors = {
    'menu-item-highlight-text': common.white,
    'menu-item-highlight-text-dark': common.black,
    'menu-item-disabled-text': palette.grey[700],
    'menu-item-disabled-text-dark': common.disabled,
};

export const menuTriggerStyles = ['focus:outline-none', 'focus-visible:focus-ring data-[state=open]:focus-ring'];

const menuPopupBaseStyles = [
    'z-50 min-w-[8rem] overflow-hidden rounded-md bg-menu-content-fill p-1 shadow-md dark:bg-menu-content-fill-dark',
    'data-[state=open]:animate-in data-[state=closed]:animate-out',
    'data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0',
];

export const menuContentStyles = [
    ...menuPopupBaseStyles,
    'border-menu-content-border dark:border-menu-content-border-dark',
    'data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2',
];

export const menuItemStyles = [
    'relative flex cursor-pointer select-none items-center rounded-lg border-none p-2 text-sm outline-none',
    'data-[highlighted]:border-menu-item-highlight-border data-[highlighted]:bg-menu-item-highlight-fill data-[highlighted]:text-menu-item-highlight-text',
    'dark:data-[highlighted]:border-menu-item-highlight-border-dark dark:data-[highlighted]:bg-menu-item-highlight-fill-dark dark:data-[highlighted]:text-menu-item-highlight-text-dark',
    'data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)]',
    'data-[highlighted]:active:bg-menu-item-active-fill dark:data-[highlighted]:active:bg-menu-item-active-fill-dark dark:data-[highlighted]:active:text-menu-item-highlight-text-dark',
    'data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:bg-menu-item-disabled-fill data-[disabled]:text-menu-item-disabled-text',
    'dark:data-[disabled]:bg-menu-item-disabled-fill-dark dark:data-[disabled]:text-menu-item-disabled-text-dark',
];

export const menuItemIconStyles = 'mr-2 flex items-center';

export const menuItemTextStyles = 'flex-1 leading-4';

export const menuSecondaryIconStyles = 'ml-2';

export const menuLabelStyles = 'px-2 py-1.5 text-sm font-semibold';

export const menuSubTriggerStyles = [
    'relative flex cursor-pointer select-none items-center rounded-lg border border-transparent px-2 py-1.5 text-sm outline-none',
    'data-[highlighted]:border-menu-item-highlight-border data-[highlighted]:bg-menu-item-highlight-fill data-[highlighted]:text-menu-item-highlight-text',
    'dark:data-[highlighted]:border-menu-item-highlight-border-dark dark:data-[highlighted]:bg-menu-item-highlight-fill-dark dark:data-[highlighted]:text-menu-item-highlight-text-dark',
    'data-[highlighted]:shadow-[inset_3px_0_0_var(--focus-ring)] data-[state=open]:shadow-[inset_3px_0_0_var(--focus-ring)]',
    'data-[state=open]:bg-menu-item-highlight-fill data-[state=open]:text-menu-item-highlight-text',
    'dark:data-[state=open]:bg-menu-item-highlight-fill-dark dark:data-[state=open]:text-menu-item-highlight-text-dark',
];

export const menuSubTriggerTextStyles = 'flex flex-1 items-center';

export const menuSubContentStyles = [
    ...menuPopupBaseStyles,
    'dark:border-menu-content-border-dark',
    'data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2',
];

export const menuSeparatorStyles = 'my-1 h-px bg-menu-separator-fill dark:bg-menu-separator-fill-dark';
