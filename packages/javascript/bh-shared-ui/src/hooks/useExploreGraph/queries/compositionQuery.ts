import { apiClient } from '../../../utils/api';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    transformToFlatGraphResponse,
} from './utils';

const compositionSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { relationshipQueryItemId, searchType } = paramOptions;

    if (searchType !== 'composition' || !relationshipQueryItemId) {
        return {
            enabled: false,
        };
    }

    const [_, sourceId, edgeType, targetId] = relationshipQueryItemId.split('_'); // TODO: determined in entity panel work

    if (!sourceId || !edgeType || !targetId || isNaN(Number(sourceId)) || isNaN(Number(targetId)))
        return {
            enabled: false,
        };

    return {
        queryKey: [ExploreGraphQueryKey, searchType, relationshipQueryItemId],
        queryFn: async () => {
            const res = await apiClient.getEdgeComposition(Number(sourceId), Number(targetId), edgeType);

            const data = res.data;
            if (!data.data.nodes) {
                throw new Error('empty graph');
            }

            return transformToFlatGraphResponse(data);
        },
        refetchOnWindowFocus: false,
    };
};

const getCompositionErrorMessage = (error: any): ExploreGraphQueryError => {
    if (error?.response?.status) {
        return { message: 'Composition not found.', key: 'NodeSearchQueryFailure' };
    } else {
        return { message: 'An unknown error occurred.', key: 'NodeSearchUnknown' };
    }
};

export const compositionSearchQuery: ExploreGraphQuery = {
    getQueryConfig: compositionSearchGraphQuery,
    getErrorMessage: getCompositionErrorMessage,
};

/**
 * TODO:
 * edge panel can only open one accordion at a time? no?
 * dont refetch graph on clicking in the window
 */
