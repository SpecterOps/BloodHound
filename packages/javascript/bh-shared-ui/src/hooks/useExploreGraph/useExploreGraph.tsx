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
    compositionSearchQuery,
    cypherSearchQuery,
    fallbackQuery,
    nodeSearchQuery,
    pathfindingSearchQuery,
    relationshipSearchQuery,
} from './queries';

export function exploreGraphQueryFactory(paramOptions: Partial<ExploreQueryParams>): ExploreGraphQuery {
    switch (paramOptions.searchType) {
        case 'node':
            return nodeSearchQuery;
        case 'pathfinding':
            return pathfindingSearchQuery;
        case 'relationship':
            return relationshipSearchQuery;
        case 'composition':
            return compositionSearchQuery;
        default:
            return fallbackQuery;
    }
}

// Hook for maintaining the top level graph query powering the explore page
// TODO: Proposal to make this an explicit optional param in the hook? To overrwrite for table view?
export const useExploreGraph = (paramOptions?: Partial<ExploreQueryParams>, includeProperties?: boolean) => {
    let params = useExploreParams() as Partial<ExploreQueryParams>;

    if (paramOptions) {
        params = paramOptions as Partial<ExploreQueryParams>;
    }
    const { addNotification } = useNotifications();

    const query = params?.searchType === 'cypher' ? cypherSearchQuery : exploreGraphQueryFactory(params);

    const queryConfig =
        paramOptions?.searchType === 'cypher'
            ? (query as CypherExploreGraphQuery).getQueryConfig(params, includeProperties)
            : query.getQueryConfig(params);

    return useQuery({
        ...queryConfig,
        onError: (error: any) => {
            const { message, key } = query.getErrorMessage(error);
            addNotification(message, key, {
                autoHideDuration: SNACKBAR_DURATION_LONG,
            });
        },
    });
};
