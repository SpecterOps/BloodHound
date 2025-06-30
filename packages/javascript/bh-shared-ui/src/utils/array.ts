// Copyright 2025 Specter Ops, Inc.
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

/**
 * Returns true if array `a` and `b` contain the same elements in any order
 *
 * Uses referential comparison; provide a sort and compare function for non-primitive array elements.
 */
export const areArraysSimilar = <T>(
    a: T[],
    b: T[],
    sortFn?: (i: T, j: T) => number,
    compareFn?: (i: T, j: T) => boolean
) => {
    if (a.length !== b.length) {
        return false;
    }

    if (sortFn && !compareFn) {
        throw new Error('');
    }

    const sortedA = sortFn ? [...a].sort(sortFn) : [...a].sort();
    const sortedB = sortFn ? [...b].sort(sortFn) : [...b].sort();

    return sortedA.every((item, index) => (compareFn ? compareFn(item, sortedB[index]) : item === sortedB[index]));
};
