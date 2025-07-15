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
import { apiClient } from '../../../utils/api';
import { parseItemId } from '../../../utils/parseItemId';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    sharedGraphQueryOptions,
} from './utils';

const aclInheritanceSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { relationshipQueryItemId, searchType } = paramOptions;

    if (searchType !== 'aclinheritance' || !relationshipQueryItemId) {
        return {
            enabled: false,
        };
    }

    const { itemType, sourceId, edgeType, targetId } = parseItemId(relationshipQueryItemId);

    if (
        itemType !== 'edge' ||
        !sourceId ||
        !edgeType ||
        !targetId ||
        isNaN(Number(sourceId)) ||
        isNaN(Number(targetId))
    ) {
        return {
            enabled: false,
        };
    }

    return {
        ...sharedGraphQueryOptions,
        queryKey: [ExploreGraphQueryKey, searchType, relationshipQueryItemId],
        queryFn: ({ signal }) =>
            apiClient.getACLInheritance(Number(sourceId), Number(targetId), edgeType, { signal }).then((res) => {
                const data = res.data;
                if (!data.data.nodes) {
                    throw new Error('empty result set');
                }

                return data;
            }),
        refetchOnWindowFocus: false,
    };
};

const getACLInheritanceErrorMessage = (): ExploreGraphQueryError => {
    return { message: 'Query failed. Please try again.', key: 'edgeACLInheritanceGraphQuery' };
};

export const aclInheritanceSearchQuery: ExploreGraphQuery = {
    getQueryConfig: aclInheritanceSearchGraphQuery,
    getErrorMessage: getACLInheritanceErrorMessage,
};
