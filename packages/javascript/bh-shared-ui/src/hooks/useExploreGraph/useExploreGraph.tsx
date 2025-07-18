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
import { useQuery } from 'react-query';
import { SNACKBAR_DURATION_LONG } from '../../constants';
import { useNotifications } from '../../providers';
import { ExploreQueryParams, useExploreParams } from '../useExploreParams';
import {
    CypherExploreGraphQuery,
    ExploreGraphQuery,
    aclInheritanceSearchQuery,
    compositionSearchQuery,
    cypherSearchQuery,
    fallbackQuery,
    nodeSearchQuery,
    pathfindingSearchQuery,
    relationshipSearchQuery,
} from './queries';

type UseExploreGraphParams = {
    includeProperties?: boolean;
    enabled?: boolean;
};

export function exploreGraphQueryFactory(
    paramOptions: Partial<ExploreQueryParams>
): ExploreGraphQuery | CypherExploreGraphQuery {
    switch (paramOptions.searchType) {
        case 'node':
            return nodeSearchQuery;
        case 'pathfinding':
            return pathfindingSearchQuery;
        case 'relationship':
            return relationshipSearchQuery;
        case 'composition':
            return compositionSearchQuery;
        case 'cypher':
            return cypherSearchQuery;
        case 'aclinheritance':
            return aclInheritanceSearchQuery;
        default:
            return fallbackQuery;
    }
}

const DEFAULT_USE_EXPLORE_GRAPH_PARAMS = { includeProperties: false, enabled: true };

// Hook for maintaining the top level graph query powering the explore page
export const useExploreGraph = ({
    includeProperties = DEFAULT_USE_EXPLORE_GRAPH_PARAMS.includeProperties,
    enabled = DEFAULT_USE_EXPLORE_GRAPH_PARAMS.enabled,
}: UseExploreGraphParams = DEFAULT_USE_EXPLORE_GRAPH_PARAMS) => {
    const params = useExploreParams();

    const { addNotification } = useNotifications();

    const query = exploreGraphQueryFactory(params);

    const queryConfig =
        params?.searchType === 'cypher'
            ? query.getQueryConfig(params, includeProperties)
            : query.getQueryConfig(params);

    const shouldFetch = Boolean(enabled && queryConfig?.queryFn);
    return useQuery({
        ...queryConfig,
        onError: (error: any) => {
            const { message, key } = query.getErrorMessage(error);
            addNotification(message, key, {
                autoHideDuration: SNACKBAR_DURATION_LONG,
            });
        },
        enabled: shouldFetch,
    });
};
