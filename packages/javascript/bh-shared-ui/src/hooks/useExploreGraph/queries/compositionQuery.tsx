import { apiClient } from '../../../utils/api';
import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQueryKey, ExploreGraphQueryOptions } from './utils';

// const selectedEdgeCypherQuery = (sourceId: string, targetId: string, edgeKind: string): string =>
// `MATCH (s)-[r:${edgeKind}]->(t) WHERE ID(s) = ${sourceId} AND ID(t) = ${targetId} RETURN r LIMIT 1`;

export const compositionSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { relationshipQueryItemId, searchType } = paramOptions;

    if (searchType !== 'composition' || !relationshipQueryItemId) {
        return {
            enabled: false,
        };
    }

    const [_, sourceId, edgeType, targetId] = relationshipQueryItemId.split('_'); // TODO: determined in entity panel effort
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

            return data;
        },
        refetchOnWindowFocus: false,
    };
};

/**
 * TODO:
 * edge panel can only open one accordion at a time? no?
 * dont refetch graph on clicking in the window
 */
