import { useQuery } from 'react-query';
import { apiClient, parseItemId } from '../utils';

export const useGraphItem = (itemId: string | null | undefined) => {
    return useQuery(
        ['getGraphItem', itemId],
        () => {
            const parsedItem = parseItemId(itemId!);
            return apiClient.cypherSearch(parsedItem.cypherQuery, undefined, true).then((res) => {
                if (!itemId) {
                    return undefined;
                }
                if (parsedItem.itemType === 'edge') {
                    return res.data?.data?.edges?.[0];
                }
                return res.data?.data?.nodes?.[itemId || ''];
            });
        },
        {
            enabled: !!itemId,
            retry: false,
        }
    );
};
