import isEmpty from 'lodash/isEmpty';
import { useEffect } from 'react';
import { DEV_TABLE_VIEW } from '../../constants';
import { isGraphResponse, useExploreGraph } from '../useExploreGraph';
import { useExploreParams } from '../useExploreParams';

export const useExploreTableAutoDisplay = (setSelectedLayout: (layout: 'table') => void) => {
    const { data: graphData } = useExploreGraph();
    const { searchType } = useExploreParams();

    const isCypherSearch = searchType === 'cypher';
    const autoDisplayTableQueryCandidate = isCypherSearch && graphData && isGraphResponse(graphData);

    useEffect(() => {
        if (DEV_TABLE_VIEW && autoDisplayTableQueryCandidate) {
            const emptyEdges = isEmpty(graphData.data.edges);
            const containsNodes = !isEmpty(graphData.data.nodes);

            if (emptyEdges && containsNodes) {
                setSelectedLayout('table');
            }
        }
    }, [autoDisplayTableQueryCandidate, graphData, setSelectedLayout]);
};
