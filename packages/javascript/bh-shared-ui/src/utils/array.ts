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

    const sortedA = sortFn ? [...a].sort(sortFn) : [...a].sort();
    const sortedB = sortFn ? [...b].sort(sortFn) : [...b].sort();

    return sortedA.every((item, index) => (compareFn ? compareFn(item, sortedB[index]) : item === sortedB[index]));
};

/**
 * Split an array into consecutive chunks of a given size.
 *
 * @param arr - The input array to split
 * @param size - Desired chunk size; must be a positive integer
 * @returns An array of subarrays where each subarray has at most `size` elements (the final chunk may be smaller)
 * @throws RangeError if `size` is not a positive integer
 */
export function chunk<T>(arr: T[], size: number): T[][] {
    if (!Number.isInteger(size) || size <= 0) {
        throw new RangeError(`chunk size must be a positive integer, got: ${size}`);
    }
    const result: T[][] = [];
    for (let i = 0; i < arr.length; i += size) {
        result.push(arr.slice(i, i + size));
    }
    return result;
}

/**
 * Flattens a one-level array of values and arrays into a single array of values.
 *
 * @param arr - Input array containing elements of type `T` or arrays of `T` (one nesting level)
 * @returns A new array with all `T` elements from `arr` in their original order
 */
export function flatten<T>(arr: (T | T[])[]): T[] {
    return arr.reduce((acc: T[], val) => {
        if (Array.isArray(val)) {
            for (const item of val) acc.push(item);
        } else {
            acc.push(val);
        }
        return acc;
    }, []);
}