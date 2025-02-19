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

import { SetURLSearchParams } from 'react-router-dom';
import { nil } from '../types';

export const isNotNullish = <T>(value: T | nil): value is T => {
    return value !== undefined && value !== null && value !== 0 && value !== '';
};

/**
 * returns a function for updating a single search param found in updatedParams
 * @param updatedParams all keys in the updatedParams will either update or delete the matching urlParam
 * @param searchParams current url params
 * @returns
 */
export const setSingleParamFactory = <T>(updatedParams: T, searchParams: URLSearchParams) => {
    return (param: keyof T) => {
        const key = param as string;
        const value = (updatedParams as Record<string, string>)[key];

        // only set keys that have been passed via updatedParams
        if (key in (updatedParams as Record<string, string>)) {
            if (isNotNullish(value)) {
                if (Array.isArray(value)) {
                    searchParams.delete(key);
                    value.forEach((item) => searchParams.append(key, item));
                } else {
                    searchParams.set(key, value);
                }
            } else {
                searchParams.delete(key);
            }
        }
    };
};

/**
 * returns a function for updating all availableParams in search params.
 * @param setSearchParams react-router-doms setSearchParams
 * @param availableParams all params that can be controlled
 * @param deleteNil if a key in availableParams is passed and it has a falsy value, this will remove it from searchParams
 * @returns
 */
export const setParamsFactory = <T>(setSearchParams: SetURLSearchParams, availableParams: Array<keyof T>) => {
    return (updatedParams: T) => {
        setSearchParams((params) => {
            const setParam = setSingleParamFactory(updatedParams, params);

            availableParams.forEach((param) => setParam(param));

            return params;
        });
    };
};
