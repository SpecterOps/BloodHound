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

export const triggerStyles =
    'items-center max-w-56 text-sm text-main rounded-md bg-transparent hover:bg-secondary hover:text-common-white border shadow-outer-0 hover:border-transparent focus-visible:border-transparent border-dropdown-trigger-border group';

export const popoverContentStyles =
    'flex flex-col p-0 rounded-md border border-dropdown-popover-border bg-dropdown-popover-fill';

// TODO optionStyles is nested but used globally in RulesAccordion, ObjectsAccordion, Zone and Label Selector BED-6572
export const optionStyles =
    'has-[svg]:px-4 px-4 truncate rounded-none w-full justify-normal text-main hover:no-underline hover:bg-dropdown-option-hover-fill disabled:bg-dropdown-option-disabled-fill group';

export const tooltipStyles = 'max-w-80 border-0 dark:bg-dropdown-tooltip-fill text-main dark:text-contrast';
