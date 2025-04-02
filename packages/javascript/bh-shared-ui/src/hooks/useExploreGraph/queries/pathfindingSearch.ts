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

import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    INITIAL_FILTER_TYPES,
    sharedGraphQueryOptions,
} from './utils';

// Only need to create our default filters once
const createPathFilterString = (types: string[]) => `in:${types.join(',')}`;
const DEFAULT_FILTERS = createPathFilterString(INITIAL_FILTER_TYPES);

export const pathfindingSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { searchType, primarySearch, secondarySearch, pathFilters } = paramOptions;

    // Query should occur whether or not pathFilters exist
    if (!primarySearch || !searchType || !secondarySearch) {
        return { enabled: false };
    }

    const filter = pathFilters?.length ? createPathFilterString(pathFilters) : DEFAULT_FILTERS;

    return {
        ...sharedGraphQueryOptions,
        queryKey: [ExploreGraphQueryKey, searchType, primarySearch, secondarySearch, filter],
        queryFn: ({ signal }) => {
            return apiClient
                .getShortestPathV2(primarySearch, secondarySearch, filter, { signal })
                .then((res) => res.data);
        },
        enabled: !!(searchType && primarySearch && secondarySearch),
    };
};

const getPathfindingErrorMessage = (error: any): ExploreGraphQueryError => {
    const statusCode = error?.response?.status;
    if (statusCode === 404) {
        return { message: 'Path not found.', key: 'shortestPathNotFound' };
    } else if (statusCode === 503) {
        return {
            message:
                'Calculating the requested Attack Path exceeded memory limitations due to the complexity of paths involved.',
            key: 'ShortestPathOutOfMemory',
        };
    } else if (statusCode === 504) {
        return {
            message: 'The results took too long to compute, possibly due to the complexity of paths involved.',
            key: 'ShortestPathTimeout',
        };
    } else {
        return { message: 'An unknown error occurred. Please try again.', key: 'ShortestPathUnknown' };
    }
};

export const pathfindingSearchQuery: ExploreGraphQuery = {
    getQueryConfig: pathfindingSearchGraphQuery,
    getErrorMessage: getPathfindingErrorMessage,
};
