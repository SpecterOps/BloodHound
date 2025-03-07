import { FC, ReactNode } from 'react';
import { useQuery } from 'react-query';
import { useExploreParams } from '../../hooks';
import { SelectedEdge } from '../../store';
import { apiClient } from '../../utils';

export const EdgeDataFetcher: FC<{ children: (selectedEdge: SelectedEdge) => ReactNode }> = ({ children }) => {
    const { panelSelection } = useExploreParams();
    const panelSelectionItems = (panelSelection as string).split('_');
    const selectedEdgeCypherQuery = (): string =>
        `MATCH p= (s)-[r:${panelSelectionItems[1]}]->(t) WHERE ID(s) = ${panelSelectionItems[0]} AND ID(t) = ${panelSelectionItems[2]} RETURN p`;

    const { data: cypherResponse } = useQuery(['selected-edge', panelSelection], ({ signal }) => {
        return apiClient
            .cypherSearch(selectedEdgeCypherQuery(), { signal }, true)
            .then((result: any) => {
                if (!result.data.data) return { nodes: {}, edges: [] };
                return result.data.data;
            })
            .catch((err) => {
                if (err.response.status === 404) {
                    // To do: show error message here
                }
            });
    });

    const selectedEdgeResponseObject = cypherResponse?.edges[0];
    const selectedEdgeNodes = cypherResponse?.nodes;

    const selectedEdge: SelectedEdge =
        selectedEdgeResponseObject && selectedEdgeNodes
            ? {
                  id: panelSelection as string,
                  name: selectedEdgeResponseObject.label || '',
                  data: selectedEdgeResponseObject.properties || {},
                  sourceNode: {
                      id: selectedEdgeResponseObject.source,
                      objectId: selectedEdgeNodes[selectedEdgeResponseObject.source]?.objectId,
                      name: selectedEdgeNodes[selectedEdgeResponseObject.source]?.label,
                      type: selectedEdgeNodes[selectedEdgeResponseObject.source]?.kind,
                  },
                  targetNode: {
                      id: selectedEdgeResponseObject.target,
                      objectId: selectedEdgeNodes[selectedEdgeResponseObject.target]?.objectId,
                      name: selectedEdgeNodes[selectedEdgeResponseObject.target]?.label,
                      type: selectedEdgeNodes[selectedEdgeResponseObject.target]?.kind,
                  },
              }
            : null;
    return children(selectedEdge);
};
