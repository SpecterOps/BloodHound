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

import { Path, SetURLSearchParams, To, createSearchParams } from 'react-router';
import { EnvironmentQueryParams } from '../../hooks/useEnvironmentParams';
import { ExploreQueryParams } from '../../hooks/useExploreParams';

export type SearchParamKeys = keyof EnvironmentQueryParams | keyof ExploreQueryParams;
export const GloballySupportedSearchParams = ['environmentId', 'environmentAggregation'] satisfies SearchParamKeys[];

type EmptyParam = undefined | null | '';

export type AppNavigateProps = { discardQueryParams?: boolean };

export const isEmptyParam = <T>(value: T | EmptyParam): value is EmptyParam => {
    return value === undefined || value === null || value === '';
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
            if (isEmptyParam(value)) {
                searchParams.delete(key);
            } else {
                if (Array.isArray(value)) {
                    searchParams.delete(key);
                    value.forEach((item) => searchParams.append(key, item));
                } else {
                    searchParams.set(key, value);
                }
            }
        }
    };
};

/**
 * returns a function for updating all availableParams in search params.
 * @param setSearchParams react-routers setSearchParams
 * @param availableParams all params that can be controlled
 * @returns
 */
export const setParamsFactory = <T>(setSearchParams: SetURLSearchParams, availableParams: Array<keyof T>) => {
    return (updatedParams: Partial<T>) => {
        setSearchParams((params) => {
            const setParam = setSingleParamFactory(updatedParams, params);

            availableParams.forEach((param) => setParam(param));

            return params;
        });
    };
};

export const persistSearchParams = (persistentSearchParams: string[]) => {
    const prevParams = new URLSearchParams(location.search);
    const newParams = new URLSearchParams();

    persistentSearchParams.forEach((param) => {
        const prevParam = prevParams.get(param);
        if (prevParam) newParams.set(param, prevParam);
    });

    return newParams;
};

// Utility function for adding type safety/autocomplete to search param construction
export const createTypedSearchParams = <T>(params: Partial<T>) => {
    const result: any = {};
    Object.entries(params).forEach(([key, value]) => (result[key] = value));
    return createSearchParams(result).toString();
};

// The 'To' type from react router can either be a string or a 'Path' object, we should be able to handle both cases
export const applyPreservedParams = (to: To, preservedParams: URLSearchParams): To => {
    if (typeof to === 'string') {
        return applyParamsToString(to, preservedParams);
    } else {
        return applyParamsToObject(to, preservedParams);
    }
};

const applyParamsToString = (to: string, preservedParams: URLSearchParams): string => {
    const parts = to.split('?');

    // Query params already exist, we need to merge the two and prioritize those passed on the individual link
    if (parts.length === 2) {
        const baseSearchParams = new URLSearchParams(parts[1]);
        const combined = new URLSearchParams({
            ...Object.fromEntries(preservedParams),
            ...Object.fromEntries(baseSearchParams),
        });
        return parts[0] + '?' + combined.toString();
    }

    // No query params exist on link, append our preserved params
    if (parts.length === 1) {
        const params = preservedParams.toString();
        return params ? `${parts[0]}?${params}` : parts[0];
    }

    // Fallback to passing through the 'to' param as-is
    return to;
};

const applyParamsToObject = (to: Partial<Path>, preservedParams: URLSearchParams): Partial<Path> => {
    // If search field already has values, merge the two, prioritizing incoming from the link
    if (to.search) {
        const baseSearchParams = new URLSearchParams(to.search);
        const combined = new URLSearchParams({
            ...Object.fromEntries(preservedParams),
            ...Object.fromEntries(baseSearchParams),
        });
        return { ...to, search: combined.toString() };
    }

    // Otherwise just set a new search field with preserved params
    return { ...to, search: preservedParams.toString() };
};
