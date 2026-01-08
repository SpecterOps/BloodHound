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

export const isDeepEqual = <T>(obj1: T, obj2: T): boolean => {
    if (obj1 === obj2) {
        return true;
    }

    const isObject1 = typeof obj1 === 'object' && obj1 !== null;
    const isObject2 = typeof obj2 === 'object' && obj2 !== null;

    if (!isObject1 || !isObject2) {
        return false;
    }

    const isArray1 = Array.isArray(obj1);
    const isArray2 = Array.isArray(obj2);

    if (isArray1 !== isArray2) {
        return false;
    }

    if (isArray1 && isArray2) {
        const arr1 = obj1 as unknown as any[];
        const arr2 = obj2 as unknown as any[];
        if (arr1.length !== arr2.length) return false;

        for (let i = 0; i < arr1.length; i++) {
            if (!isDeepEqual(arr1[i], arr2[i])) {
                return false;
            }
        }
        return true;
    } else {
        const keys1 = Object.keys(obj1) as (keyof T)[];
        const keys2 = Object.keys(obj2) as (keyof T)[];

        if (keys1.length !== keys2.length) return false;

        for (const key of keys1) {
            if (!Object.prototype.hasOwnProperty.call(obj2, key) || !isDeepEqual(obj1[key], obj2[key])) {
                return false;
            }
        }
        return true;
    }
};
