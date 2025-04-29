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

import { UseQueryOptions, UseQueryResult } from 'react-query';

export type GenericQueryOptions<T> = Omit<UseQueryOptions<T, unknown, T, string[]>, 'queryFn'>;

export const getQueryKey = (baseKey: string[], customKeys?: string[]) => {
    return customKeys?.length ? [...baseKey, ...customKeys] : baseKey;
};

/**
 * Use this when you want to treat the loading or error state of multiple queries as one.
 * For example, display a component when 2 separate APIs are done loading.
 * @param queries Variadic function that will check if any query isLoading or isError
 * @returns isLoading = true if any query is loading, and isError = true if any query has an error
 */
export const queriesAreLoadingOrErrored = (...queries: UseQueryResult<unknown, unknown>[]) => {
    let isLoading = false;
    let isError = false;

    queries.forEach((query) => {
        if (query.isLoading) isLoading = true;
        if (query.isError) isError = true;
    });

    return { isLoading, isError };
};
