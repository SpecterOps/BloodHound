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

/** Returns Object.entries with results retaining their types */
export const typedEntries = <T extends object>(obj: T): [keyof T, T[keyof T]][] => Object.entries(obj) as any;

/** Returns true if both objects are deeply equal, by value */
export const isDeepEqual = (a: unknown, b: unknown): boolean => {
    if (a === b) {
        return true;
    }

    if (a && b && typeof a === 'object' && typeof b === 'object') {
        // Handle Array case
        if (Array.isArray(a) && Array.isArray(b)) {
            if (a.length !== b.length) return false;
            return a.every((val, i) => isDeepEqual(val, b[i]));
        }

        // If one is array and the other isn't
        if (Array.isArray(a) !== Array.isArray(b)) {
            return false;
        }

        // Handle plain objects
        const keysA = Object.keys(a as object);
        const keysB = Object.keys(b as object);

        if (keysA.length !== keysB.length) {
            return false;
        }

        return keysA.every((key) => {
            if (!(key in (b as object))) {
                return false;
            }
            return isDeepEqual((a as Record<string, unknown>)[key], (b as Record<string, unknown>)[key]);
        });
    }

    // For primitives (number, string, boolean, null, undefined, symbol)
    return false;
};
