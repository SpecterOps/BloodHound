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

const compositionSearchGraphQuery = (
    paramOptions: Partial<ExploreQueryParams>,
    relationshipDetails: RelationshipDetailsWithInfo | undefined
): ExploreGraphQueryOptions => {
    const { searchType } = paramOptions;

    if (
        searchType !== 'composition' ||
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
                .getEdgeComposition(
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

const getCompositionErrorMessage = (): ExploreGraphQueryError => {
    return { message: 'Query failed. Please try again.', key: 'edgeCompositionGraphQuery' };
};

export const compositionSearchQuery = (
    paramOptions: Partial<ExploreQueryParams>,
    relationshipDetails: RelationshipDetailsWithInfo | undefined
): ExploreGraphQuery => {
    return {
        getQueryConfig: () => compositionSearchGraphQuery(paramOptions, relationshipDetails),
        getErrorMessage: getCompositionErrorMessage,
    };
};
