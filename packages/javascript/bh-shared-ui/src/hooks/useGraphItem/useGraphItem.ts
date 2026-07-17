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

import {
    GraphNode,
    NodeDetails,
    NodeDetailsWithInfo,
    RelationshipDetails,
    RelationshipDetailsWithInfo,
} from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient, REL_ID_PREFIX } from '../../utils';
import { escapeCypherString } from '../../utils/cypher';

export const isRelationshipResponse = (
    response: RelationshipDetails | RelationshipDetailsWithInfo | NodeDetails | NodeDetailsWithInfo
): response is RelationshipDetails | RelationshipDetailsWithInfo => {
    return 'relationship_id' in response;
};

export const isNodeResponse = (
    response?: RelationshipDetails | RelationshipDetailsWithInfo | NodeDetails | NodeDetailsWithInfo
): response is NodeDetails | NodeDetailsWithInfo => {
    return response ? 'node_id' in response : false;
};

export const useGetRelationshipById = (id?: number) => {
    return useQuery({
        queryKey: ['getRelationshipById', id],
        queryFn: async () => {
            return apiClient
                .getRelationshipByID(id!, { params: { 'include-info': true } })
                .then((res) => res.data.data);
        },
        enabled: !!id,
        retryOnMount: false,
        retry: false,
        refetchOnWindowFocus: false,
        keepPreviousData: true,
    });
};

export const useGetNodeById = (id?: number) => {
    return useQuery({
        queryKey: ['getNodeById', id],
        queryFn: async () => {
            return apiClient.getNodeByID(id!, { params: { 'include-info': true } }).then((res) => res.data.data);
        },
        enabled: !!id,
        retryOnMount: false,
        retry: false,
        refetchOnWindowFocus: false,
        keepPreviousData: true,
    });
};

export const useGraphItem = (itemId?: string | null) => {
    const isRelationship = !!itemId?.includes(REL_ID_PREFIX);

    const relationshipId = isRelationship && itemId ? parseInt(itemId.slice(REL_ID_PREFIX.length)) : undefined;
    const relQuery = useGetRelationshipById(relationshipId);

    const nodeId = itemId ? parseInt(itemId) : undefined;
    const nodeQuery = useGetNodeById(nodeId);

    return isRelationship ? relQuery : nodeQuery;
};

export const useNodeByObjectId = (objectId?: string) => {
    return useQuery({
        queryKey: ['getGraphNodeByObjectId', objectId],
        queryFn: async () => {
            const safeObjectId = objectId ? escapeCypherString(objectId) : '';
            return apiClient
                .cypherSearch(`MATCH (n) WHERE n.objectid = "${safeObjectId}" RETURN n LIMIT 1`, undefined, true)
                .then((res) => {
                    if (!objectId) {
                        return undefined;
                    }

                    const nodes = res.data?.data?.nodes;
                    if (!nodes || Object.keys(nodes).length === 0) {
                        return undefined;
                    }

                    const firstElement: GraphNode = Object.values(nodes)[0];
                    const id = Object.keys(nodes)[0];

                    const node: NodeDetails = {
                        node_id: parseInt(id),
                        kinds: firstElement.kinds.map((kind) => {
                            return { name: kind, node_kind_id: null };
                        }),
                        properties: { objectid: objectId, ...firstElement.properties },
                    };

                    return node;
                });
        },

        enabled: !!objectId,
        retry: false,
        refetchOnWindowFocus: false,
        keepPreviousData: true,
    });
};
