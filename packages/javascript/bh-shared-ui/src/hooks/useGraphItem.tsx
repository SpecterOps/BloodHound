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

import { NodeDetails, NodeDetailsWithInfo, RelationshipDetails, RelationshipDetailsWithInfo } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient, REL_ID_PREFIX } from '../utils';

export interface BaseItemResponse {
    id: string;
    kind: string;
    label: string;
    lastSeen: string;
    properties?: any;
}

export interface NodeResponse extends BaseItemResponse {
    objectId: string;
    isTierZero: boolean;
    isOwnedObject: boolean;
}

export interface EdgeResponse extends BaseItemResponse {
    source: string;
    sourceNode: NodeResponse;
    target: string;
    targetNode: NodeResponse;
}

export type ItemResponse = NodeResponse | EdgeResponse | undefined;

export const isNode = (response: ItemResponse): response is NodeResponse => {
    if (!response) return false;
    return 'objectId' in response;
};

export const isEdge = (response: ItemResponse): response is EdgeResponse => {
    if (!response) return false;
    return 'source' in response;
};

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
            return apiClient.getRelationshipByID(id!).then((res) => res.data.data);
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
            return apiClient.getNodeByID(id!).then((res) => res.data.data);
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

    const relationshipId = itemId ? parseInt(itemId.slice(REL_ID_PREFIX.length)) : undefined;
    const relQuery = useGetRelationshipById(relationshipId);

    const nodeId = itemId ? parseInt(itemId) : undefined;
    const nodeQuery = useGetNodeById(nodeId);

    return isRelationship ? relQuery : nodeQuery;
};

export const useNodeByObjectId = (itemId?: string) => {
    return useQuery<NodeResponse>(
        ['getGraphNodeByObjectId', itemId],
        () => {
            return apiClient
                .cypherSearch(`MATCH (n) WHERE n.objectid = "${itemId}" RETURN n LIMIT 1`, undefined, true)
                .then((res) => {
                    if (!itemId) {
                        return undefined;
                    }

                    const firstElement: any = Object.values(res.data?.data?.nodes)[0];

                    const id = Object.keys(res.data?.data?.nodes)[0];

                    return {
                        id,
                        ...firstElement,
                    };
                });
        },
        {
            enabled: !!itemId,
            retry: false,
            refetchOnWindowFocus: false,
            keepPreviousData: true,
        }
    );
};
