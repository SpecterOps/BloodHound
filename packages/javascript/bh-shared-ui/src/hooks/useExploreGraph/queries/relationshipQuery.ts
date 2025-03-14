import { entityRelationshipEndpoints } from '../../../utils/content';
import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQuery, ExploreGraphQueryError, ExploreGraphQueryKey, ExploreGraphQueryOptions } from './utils';

const relationshipSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { relationshipQueryType, relationshipQueryItemId, searchType } = paramOptions;

    const isEdgeId = relationshipQueryItemId?.includes('_'); // TODO: tobe determined from entity panel work

    if (searchType !== 'relationship' || isEdgeId || !relationshipQueryItemId || !relationshipQueryType) {
        return {
            enabled: false,
        };
    }

    const endpoint = entityRelationshipEndpoints[relationshipQueryType];

    return {
        queryKey: [ExploreGraphQueryKey, searchType, relationshipQueryItemId, relationshipQueryType],
        queryFn: async () => endpoint({ id: relationshipQueryItemId, type: 'graph' }),
        refetchOnWindowFocus: false,
    };
};

const getRelationshipErrorMessage = (error: any): ExploreGraphQueryError => {
    if (error?.response?.status) {
        return { message: 'Relationship not found.', key: 'NodeSearchQueryFailure' };
    } else {
        return { message: 'An unknown error occurred.', key: 'NodeSearchUnknown' };
    }
};

export const relationshipSearchQuery: ExploreGraphQuery = {
    getQueryConfig: relationshipSearchGraphQuery,
    getErrorMessage: getRelationshipErrorMessage,
};

/**
 * TODO:
 * how can we test
 */
