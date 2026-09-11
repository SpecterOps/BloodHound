// Copyright 2024 Specter Ops, Inc.
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

import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * This function sets the name of our current theme as a class on the html document root. This will ensure the correct styles are applied to components attached elsewhere in the DOM, such as modals and popover menus.
 *
 * @param value - can be 'dark' or 'light'
 *
 * @returns the name of the currently set class as a string
 */
export const setRootClass = (value: 'dark' | 'light') => {
    const root = window.document.documentElement;
    root.classList.remove('dark', 'light');
    root.classList.add(value);
    return value;
};

/**
 * Utility function for conditionally constructing className strings and merging the result.
 *
 * @param inputs - any number of valid clsx statements. For reference: [clsx docs](https://github.com/lukeed/clsx#readme)
 *
 * @returns a merged class list as a string. For more information about how merging tailwind classes works: [twMerge docs](https://github.com/dcastil/tailwind-merge/blob/v2.5.4/docs/what-is-it-for.md)
 */
export const cn = (...inputs: ClassValue[]) => {
    return twMerge(clsx(inputs));
};
