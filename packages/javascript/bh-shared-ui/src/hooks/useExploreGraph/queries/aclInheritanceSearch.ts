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
import { RelationshipDetailsWithInfo } from 'js-client-library';
import { apiClient } from '../../../utils/api';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    sharedGraphQueryOptions,
} from './utils';

const aclInheritanceSearchGraphQuery = (
    paramOptions: Partial<ExploreQueryParams>,
    relationshipDetails: RelationshipDetailsWithInfo | undefined
): ExploreGraphQueryOptions => {
    const { searchType } = paramOptions;

    if (
        searchType !== 'aclinheritance' ||
        !relationshipDetails ||
        !relationshipDetails.source_node_id ||
        !relationshipDetails.target_node_id
    ) {
        return {
            enabled: false,
        };
    }

    return {
        ...sharedGraphQueryOptions,
        queryKey: [ExploreGraphQueryKey, searchType, relationshipDetails.relationship_id.toString()],
        queryFn: ({ signal }) =>
            apiClient
                .getACLInheritance(
                    relationshipDetails.source_node_id!,
                    relationshipDetails.target_node_id!,
                    relationshipDetails.kind.name,
                    { signal }
                )
                .then((res) => {
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

export const aclInheritanceSearchQuery = (
    paramOptions: Partial<ExploreQueryParams>,
    relationshipDetails: RelationshipDetailsWithInfo | undefined
): ExploreGraphQuery => {
    return {
        getQueryConfig: () => aclInheritanceSearchGraphQuery(paramOptions, relationshipDetails),
        getErrorMessage: getACLInheritanceErrorMessage,
    };
};
