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
import { ExploreQueryParams, useExploreParams } from '../useExploreParams/useExploreParams';
import { ExploreGraphQueryOptions, nodeSearchGraphQuery } from './queries';
import { compositionSearchGraphQuery } from './queries/compositionQuery';
import { relationshipSearchGraphQuery } from './queries/relationshipQuery';

export function getExploreGraphQuery(paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions {
    switch (paramOptions.searchType) {
        case 'node':
            return nodeSearchGraphQuery(paramOptions);
        case 'pathfinding':
            return {};
        case 'cypher':
            return {};
        case 'relationship':
            return relationshipSearchGraphQuery(paramOptions);
        case 'composition':
            return compositionSearchGraphQuery(paramOptions);
        default:
            return { enabled: false };
    }
    // else some unidentified type, display error, set to node-search
}

// Consumer of query params example
export const useExploreGraph = <T,>() => {
    const params = useExploreParams();

    const queryConfig = getExploreGraphQuery(params);

    const { data, isLoading, isError } = useQuery(queryConfig);

    return { data: data as T, isLoading, isError };
};
