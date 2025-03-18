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
import { ExploreGraphQuery, fallbackQuery, nodeSearchQuery, pathfindingSearchQuery } from './queries';

export function exploreGraphQueryFactory(paramOptions: Partial<ExploreQueryParams>): ExploreGraphQuery {
    switch (paramOptions.searchType) {
        case 'node':
            return nodeSearchQuery;
        case 'pathfinding':
            return pathfindingSearchQuery;
        // case 'cypher':
        //     return {};
        // case 'relationship':
        //     return {};
        // case 'composition':
        //     return {};
        default:
            return fallbackQuery;
    }
}

// Hook for maintaining the top level graph query powering the explore page
export const useExploreGraph = () => {
    const params = useExploreParams();
    const { addNotification } = useNotifications();

    const query = exploreGraphQueryFactory(params);

    return useQuery({
        ...query.getQueryConfig(params),
        onError: (error: any) => {
            const { message, key } = query.getErrorMessage(error);
            addNotification(message, key);
        },
    });
};
