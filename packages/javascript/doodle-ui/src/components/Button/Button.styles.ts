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

export const buttonBaseClasses = [
    'inline-flex items-center justify-center whitespace-nowrap rounded-3xl transition-colors',
    'focus:outline-none focus-visible:focus-ring',
    'disabled:text-[#616161] dark:disabled:text-[#A6A6A6] disabled:pointer-events-none disabled:opacity-50',
    'has-[svg]:gap-2 [&>svg]:shrink-0',
];

export const primaryClasses = [
    'bg-primary text-common-white shadow-outer-1 dark:text-common-dark',
    'hover:bg-secondary',
    'focus-visible:bg-secondary',
    'active:bg-[#0D0A30] dark:active:bg-[#8D8BF8]',
    // Implement text-common when token experiment is ready - #0D0A30 matches light.primary.variant, #8D8BF8 matches dark.primary.variant.
    'disabled:shadow-none disabled:bg-[#E3E7EA] dark:disabled:bg-[#2E2E2E]',
    // disabled is neutral.light[200], text -> common.disabled // disabled:bg -> neutral.dark[700] or common.disabled.dark (token experiment)
];

export const secondaryClasses = [
    'bg-secondary-btn-fill text-common-dark shadow-outer-1 dark:text-common-white',
    'hover:bg-secondary hover:text-common-white dark:hover:text-common-dark',
    'focus-visible:bg-secondary focus-visible:text-common-white dark:focus-visible:text-common-dark',
    'active:bg-secondary-btn-active-fill active:text-common-dark dark:active:text-common-white',
    'disabled:bg-btn-disabled-fill disabled:shadow-none',
];
