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

import { entityRelationshipEndpoints } from '../../../utils/content';
import { parseItemId } from '../../../utils/parseItemId';
import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQuery, ExploreGraphQueryError, ExploreGraphQueryKey, ExploreGraphQueryOptions } from './utils';

const relationshipSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { relationshipQueryType, relationshipQueryItemId, searchType } = paramOptions;

    if (searchType !== 'relationship' || !relationshipQueryItemId || !relationshipQueryType) {
        return {
            enabled: false,
        };
    }

    const parsedQueryItemId = parseItemId(relationshipQueryItemId);

    if (parsedQueryItemId.itemType !== 'node') {
        return {
            enabled: false,
        };
    }

    const endpoint = entityRelationshipEndpoints[relationshipQueryType];

    return {
        queryKey: [ExploreGraphQueryKey, searchType, relationshipQueryItemId, relationshipQueryType],
        queryFn: async () => endpoint({ id: relationshipQueryItemId, type: 'graph' }),
        refetchOnWindowFocus: false,
    };
};

const getRelationshipErrorMessage = (error: any): ExploreGraphQueryError => {
    if (error?.response?.status) {
        return { message: 'Relationship not found.', key: 'NodeSearchQueryFailure' };
    } else {
        return { message: 'An unknown error occurred.', key: 'NodeSearchUnknown' };
    }
};

export const relationshipSearchQuery: ExploreGraphQuery = {
    getQueryConfig: relationshipSearchGraphQuery,
    getErrorMessage: getRelationshipErrorMessage,
};

/**
 * TODO:
 * use parseId on selectedItem
 * what should these errors be
 * better tests
 */
