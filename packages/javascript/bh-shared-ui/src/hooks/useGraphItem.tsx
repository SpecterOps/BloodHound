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
import { apiClient, parseItemId } from '../utils';

export interface BaseItemResponse {
    id: string;
    kind: string;
    label: string;
    lastSeen: string;
    properties: any;
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

export type ItemResponse = NodeResponse | EdgeResponse;

export const isNode = (response: ItemResponse): response is NodeResponse => {
    return 'objectId' in response;
};

export const isEdge = (response: ItemResponse): response is EdgeResponse => {
    return 'source' in response;
};

export const useGraphItem = (itemId?: string) => {
    return useQuery<ItemResponse>(
        ['getGraphItem', itemId],
        () => {
            const parsedItem = parseItemId(itemId!);
            return apiClient.cypherSearch(parsedItem.cypherQuery, undefined, true).then((res) => {
                if (!itemId) {
                    return undefined;
                }
                if (parsedItem.itemType === 'edge') {
                    const edgeResponse = res.data?.data?.edges?.[0];
                    const sourceNode = { id: edgeResponse.source, ...res.data?.data?.nodes?.[edgeResponse.source] };
                    const targetNode = { id: edgeResponse.target, ...res.data?.data?.nodes?.[edgeResponse.target] };
                    return {
                        id: itemId,
                        ...edgeResponse,
                        sourceNode,
                        targetNode,
                    };
                }
                return {
                    id: itemId,
                    ...res.data?.data?.nodes?.[itemId || ''],
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
