import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQueryKey, ExploreGraphQueryOptions } from './utils';

const fakeEndpointMap: any = {};

export const relationshipSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { relationshipQueryType, relationshipQueryItemId, searchType } = paramOptions;

    const isEdgeId = relationshipQueryItemId?.includes('_'); // TODO: tobe determined from entity panel work

    if (searchType !== 'relationship' || isEdgeId || !relationshipQueryItemId || !relationshipQueryType) {
        return {
            enabled: false,
        };
    }

    const endpoint = fakeEndpointMap[relationshipQueryType];

    return {
        queryKey: [ExploreGraphQueryKey, searchType, relationshipQueryItemId, relationshipQueryType],
        queryFn: async () => endpoint({ id: relationshipQueryItemId, type: 'graph' }),
        refetchOnWindowFocus: false,
    };
};

/**
 * TODO:
 * how can we test
 *   make sure the right endpoint it ran with the correct params
 */
