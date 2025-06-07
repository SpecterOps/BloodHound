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
import { decodeCypherQuery } from '../utils';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    sharedGraphQueryOptions,
} from './utils';

export const cypherSearchGraphQuery = (
    paramOptions: Partial<ExploreQueryParams>,
    includeProperties = false
): ExploreGraphQueryOptions => {
    const { searchType, cypherSearch } = paramOptions;

    if (!cypherSearch || !searchType) {
        return { enabled: false };
    }

    const decoded = decodeCypherQuery(cypherSearch);

    return {
        ...sharedGraphQueryOptions,
        queryKey: [ExploreGraphQueryKey, searchType, cypherSearch],
        queryFn: ({ signal }) => apiClient.cypherSearch(decoded, { signal }, includeProperties).then((res) => res.data),
        retry: false,
        enabled: !!(searchType && cypherSearch),
    };
};

const getCypherErrorMessage = (error: any): ExploreGraphQueryError => {
    const status = error?.response?.status;
    const message = error?.response?.data?.errors?.[0]?.message;

    if (status === 404) {
        return { message: 'No results match your criteria', key: 'CypherSearchEmptyResponse' };
    } else if (message) {
        return { message, key: 'CypherSearchBadRequest' };
    } else {
        return { message: 'An unknown error occurred.', key: 'CypherSearchUnknown' };
    }
};

export type CypherExploreGraphQuery = ExploreGraphQuery & {
    // TODO: make generic?
    getQueryConfig: (
        paramOptions: Partial<ExploreQueryParams>,
        includeProperties?: boolean
    ) => ExploreGraphQueryOptions;
};

export const cypherSearchQuery: CypherExploreGraphQuery = {
    getQueryConfig: cypherSearchGraphQuery,
    getErrorMessage: getCypherErrorMessage,
};
