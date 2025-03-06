import { useQuery } from 'react-query';
import { apiClient, parseItemId } from '../utils';

export interface BaseItemResponse {
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
    sourceNodeId: string;
    sourceNode: NodeResponse;
    targetNodeId: string;
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
                    const sourceNode = res.data?.data?.nodes?.[edgeResponse.source];
                    const targetNode = res.data?.data?.nodes?.[edgeResponse.target];
                    return {
                        ...edgeResponse,
                        sourceNodeId: edgeResponse.source,
                        targetNodeId: edgeResponse.target,
                        sourceNode,
                        targetNode,
                    };
                }
                return res.data?.data?.nodes?.[itemId || ''];
            });
        },
        {
            enabled: !!itemId,
            retry: false,
            refetchOnWindowFocus: false,
        }
    );
};
