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
//import { useAppSelector } from 'src/store';
import { SNACKBAR_DURATION_LONG } from '../../constants';
import { useNotifications } from '../../providers';
import { ExploreQueryParams, useExploreParams } from '../useExploreParams';

import { useDisableQueryLimitContext } from '../../views/Explore/providers/DisableQueryLimitProvider/DisableQueryLimitContext';
import {
    CypherExploreGraphQuery,
    ExploreGraphQuery,
    ExploreGraphQueryOptions,
    aclInheritanceSearchQuery,
    compositionSearchQuery,
    cypherSearchQuery,
    fallbackQuery,
    nodeSearchQuery,
    pathfindingSearchQuery,
    relationshipSearchQuery,
} from './queries';

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

// Hook for maintaining the top level graph query powering the explore page
export const useExploreGraph = (options: ExploreGraphQueryOptions = {}) => {
    const params = useExploreParams();
    const { onError, ...rest } = options;

    const { addNotification } = useNotifications();

    //const { isDisableQueryLimit } = useDisableQueryLimitContext();

    //console.log(isDisableQueryLimit);

    const query = exploreGraphQueryFactory(params);

    const queryConfig = query.getQueryConfig(params);

    const userSettings = useUserSettings();

    //console.log(userSettings);

    return useQuery({
        ...queryConfig,
        onError: (error: any) => {
            const { message, key } = query.getErrorMessage(error);
            if (onError) {
                onError(message);
            }

            addNotification(message, key, {
                autoHideDuration: SNACKBAR_DURATION_LONG,
            });
        },
        ...rest,
        ...userSettings,
    });
};

export const useUserSettings = () => {
    const { isDisableQueryLimit } = useDisableQueryLimitContext();
    //console.log(isDisableQueryLimit);

    //let settings = { headers: '' };

    let waitTime;

    let settings = {
        headers: {
            Prefer: `${waitTime}`,
        },
    };

    if (isDisableQueryLimit) {
        settings.headers = { Prefer: 'wait=-1' };
    }

    return settings;
};
