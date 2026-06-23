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
export const focusedControlStateStyles =
    'focus:bg-primary focus:text-white dark:focus:text-neutral-dark-1 focus-visible:bg-primary focus-visible:text-white dark:focus-visible:text-neutral-dark-1';

export const dropdownIconStateStyles =
    'hover:[&_svg]:text-white dark:hover:[&_svg]:text-neutral-dark-1 hover:[&_svg]:fill-current hover:[&_svg_*]:text-white dark:hover:[&_svg_*]:text-neutral-dark-1 hover:[&_svg_*]:fill-current focus:[&_svg]:text-white dark:focus:[&_svg]:text-neutral-dark-1 focus:[&_svg]:fill-current focus:[&_svg_*]:text-white dark:focus:[&_svg_*]:text-neutral-dark-1 focus:[&_svg_*]:fill-current focus-visible:[&_svg]:text-white dark:focus-visible:[&_svg]:text-neutral-dark-1 focus-visible:[&_svg]:fill-current focus-visible:[&_svg_*]:text-white dark:focus-visible:[&_svg_*]:text-neutral-dark-1 focus-visible:[&_svg_*]:fill-current';

export const triggerStyles =
    `max-w-56 text-sm text-contrast rounded-md bg-transparent hover:bg-primary hover:text-white dark:hover:text-neutral-dark-1 ${focusedControlStateStyles} border shadow-outer-0 hover:border-transparent focus:border-transparent focus-visible:border-transparent border-neutral-light-5 group ${dropdownIconStateStyles}`;

export const popoverContentStyles = 'flex flex-col p-0 rounded-md border border-neutral-5 bg-neutral-1';

export const optionStyles =
    `px-4 py-1 rounded-none w-full justify-normal text-contrast dark:text-neutral-light-1 hover:no-underline hover:bg-neutral-4 dark:hover:text-neutral-dark-1 ${focusedControlStateStyles} ${dropdownIconStateStyles} disabled:bg-neutral-4 group`;

export const selectorIconStyles =
    'group-hover:text-white dark:group-hover:text-neutral-dark-1 group-focus:text-white dark:group-focus:text-neutral-dark-1 group-focus-visible:text-white dark:group-focus-visible:text-neutral-dark-1';

export const optionIconStyles =
    selectorIconStyles;

export const tooltipStyles = 'max-w-80 border-0 dark:bg-neutral-4 dark:text-white';
