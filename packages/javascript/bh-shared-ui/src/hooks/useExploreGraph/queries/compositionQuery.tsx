import { apiClient } from '../../../utils/api';
import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQueryKey, ExploreGraphQueryOptions } from './utils';

// const selectedEdgeCypherQuery = (sourceId: string, targetId: string, edgeKind: string): string =>
// `MATCH (s)-[r:${edgeKind}]->(t) WHERE ID(s) = ${sourceId} AND ID(t) = ${targetId} RETURN r LIMIT 1`;

export const compositionSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { expandedRelationships, panelSelection, searchType } = paramOptions;

    const compositionExpanded = expandedRelationships?.includes('Composition');

    if (searchType !== 'composition' || !expandedRelationships?.length || !compositionExpanded || !panelSelection) {
        return {
            enabled: false,
        };
    }

    const [_rel, sourceId, edgeType, targetId] = panelSelection.split('_'); // TODO: what is this actually going to look like?
    if (!sourceId || !edgeType || !targetId || isNaN(Number(sourceId)) || isNaN(Number(targetId)))
        return {
            enabled: false,
        };

    return {
        queryKey: [ExploreGraphQueryKey, ...expandedRelationships, panelSelection, searchType],
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
 * selection data type to be determined in a separate ticket
 * edge panel can only open one accordion at a time? no?
 * dont refetch graph on clicking in the window
 */
