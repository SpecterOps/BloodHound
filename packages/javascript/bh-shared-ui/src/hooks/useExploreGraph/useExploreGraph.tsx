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
import { useNotifications } from '../../providers';
import { ExploreQueryParams, useExploreParams } from '../useExploreParams';
import { ExploreGraphQueryContext, nodeSearchQueryContext, pathfindingSearchQueryContext } from './queries';

export function getExploreGraphQueryContext(paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryContext {
    switch (paramOptions.searchType) {
        case 'node':
            return nodeSearchQueryContext;
        case 'pathfinding':
            return pathfindingSearchQueryContext;
        // case 'cypher':
        //     return {};
        // case 'relationship':
        //     return {};
        // case 'composition':
        //     return {};
        default:
            return { getQueryConfig: () => ({ enabled: false }) };
    }
}

// Hook for maintaining the top level graph query powering the explore page
export const useExploreGraph = () => {
    const params = useExploreParams();
    const { addNotification } = useNotifications();

    const queryContext = getExploreGraphQueryContext(params);
    const query = useQuery(queryContext.getQueryConfig(params));

    if (query.error && queryContext.getGraphError) {
        const error = queryContext.getGraphError(query.error);
        addNotification(error.message, error.key);
    }

    return query;
};
