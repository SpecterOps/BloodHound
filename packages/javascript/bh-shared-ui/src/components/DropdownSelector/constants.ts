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
const focusVisibleControlStateStyles =
    'focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark';

export const focusedControlStateStyles = `focus:bg-primary focus:text-common-white dark:focus:text-common-dark ${focusVisibleControlStateStyles}`;

const focusVisibleDropdownIconStateStyles =
    'focus-visible:[&_svg]:text-common-white dark:focus-visible:[&_svg]:text-common-dark focus-visible:[&_svg]:fill-current focus-visible:[&_svg_*]:text-common-white dark:focus-visible:[&_svg_*]:text-common-dark focus-visible:[&_svg_*]:fill-current';

const focusedDropdownIconStateStyles = `focus:[&_svg]:text-common-white dark:focus:[&_svg]:text-common-dark focus:[&_svg]:fill-current focus:[&_svg_*]:text-common-white dark:focus:[&_svg_*]:text-common-dark focus:[&_svg_*]:fill-current ${focusVisibleDropdownIconStateStyles}`;

export const dropdownIconStateStyles = `hover:[&_svg]:text-common-white dark:hover:[&_svg]:text-common-dark hover:[&_svg]:fill-current hover:[&_svg_*]:text-common-white dark:hover:[&_svg_*]:text-common-dark hover:[&_svg_*]:fill-current ${focusedDropdownIconStateStyles}`;

export const triggerStyles = `max-w-56 text-sm text-main rounded-md bg-transparent hover:bg-primary hover:text-common-white ${focusedControlStateStyles} border shadow-outer-0 hover:border-transparent focus:border-transparent focus-visible:border-transparent border-dropdown-trigger-border group ${dropdownIconStateStyles}`;

export const popoverContentStyles =
    'flex flex-col p-0 rounded-md border border-dropdown-popover-border bg-dropdown-popover-fill';

export const optionStyles = `px-4 py-1 rounded-none w-full justify-normal text-main hover:text-main hover:no-underline hover:bg-dropdown-option-hover-fill ${focusVisibleControlStateStyles} ${focusVisibleDropdownIconStateStyles} disabled:bg-dropdown-option-disabled-fill group`;

export const focusedOptionIconStyles =
    'group-focus-visible:text-common-white dark:group-focus-visible:text-common-dark';

export const selectorIconStyles = `group-hover:text-common-white dark:group-hover:text-common-dark group-focus:text-common-white dark:group-focus:text-common-dark ${focusedOptionIconStyles}`;

export const optionIconStyles = `group-hover:text-main ${focusedOptionIconStyles}`;

export const tooltipStyles = 'max-w-80 border-0 dark:bg-dropdown-tooltip-fill text-main';
