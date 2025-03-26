import { ExploreQueryParams, transformToFlatGraphResponse, useExploreGraph, useExploreParams } from 'bh-shared-ui';
import { FlatGraphResponse, GraphResponse } from 'js-client-library';
import { useMemo } from 'react';

export const normalizeGraphDataToSigma = (
    graphData: GraphResponse | FlatGraphResponse | undefined,
    searchType: ExploreQueryParams['searchType']
): FlatGraphResponse => {
    if (!graphData) return {};
    switch (searchType) {
        case 'node':
        case 'relationship': {
            return graphData as FlatGraphResponse;
        }
        case 'cypher':
        case 'composition':
        case 'pathfinding': {
            return transformToFlatGraphResponse(graphData as GraphResponse);
        }
    }
    return {};
};

export const useSigmaExploreGraph = () => {
    const { searchType } = useExploreParams();
    const graphState = useExploreGraph();
    const normalizedGraphData = useMemo(
        () => normalizeGraphDataToSigma(graphState.data, searchType),
        [graphState.data, searchType]
    );
    return { ...graphState, data: normalizedGraphData };
};
