// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { Box, Paper, SxProps } from '@mui/material';
import { usePaneStyles } from 'bh-shared-ui';
import React, { useState } from 'react';
import EdgeInfoContent from 'src/views/Explore/EdgeInfo/EdgeInfoContent';
import Header from 'src/views/Explore/EdgeInfo/EdgeInfoHeader';

const EdgeInfoPane: React.FC<{ sx?: SxProps; selectedEdge?: any }> = ({ sx, selectedEdge }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    // const { panelSelection } = useExploreParams();
    // const edgeComposition = (panelSelection as string).split('_');

    // const selectedEdgeCypherQuery = (sourceId: string | number, targetId: string | number, edgeKind: string): string =>
    //     `MATCH p= (s)-[r:${edgeKind}]->(t) WHERE ID(s) = ${sourceId} AND ID(t) = ${targetId} RETURN p`;

    // const { data: cypherResponse } = useQuery(['edge-info-get', panelSelection], ({ signal }) => {
    //     return apiClient
    //         .cypherSearch(
    //             selectedEdgeCypherQuery(edgeComposition[0], edgeComposition[2], edgeComposition[1]),
    //             { signal },
    //             true
    //         )
    //         .then((result: any) => {
    //             if (!result.data.data) return { nodes: {}, edges: [] };
    //             return result.data.data;
    //         });
    // });
    // if (!cypherResponse) return; // to do: change this

    // const selectedEdgeResponseObject = cypherResponse.edges[0];
    // const selectedEdgeNodes = cypherResponse.nodes;
    // console.log('cypherResponse', cypherResponse);
    // console.log('selectedEdgeResponseObject', selectedEdgeResponseObject);
    // console.log('selectedEdgeNodes', selectedEdgeNodes);

    // const selectedEdge = {
    //     id: panelSelection,
    //     name: selectedEdgeResponseObject.label || '',
    //     data: selectedEdgeResponseObject.properties || {},
    //     sourceNode: {
    //         id: selectedEdgeResponseObject.source,
    //         objectId: selectedEdgeNodes[selectedEdgeResponseObject.source]?.objectId,
    //         name: selectedEdgeNodes[selectedEdgeResponseObject.source]?.label,
    //         type: selectedEdgeNodes[selectedEdgeResponseObject.source]?.kind,
    //     },
    //     targetNode: {
    //         id: selectedEdgeResponseObject.target,
    //         objectId: selectedEdgeNodes[selectedEdgeResponseObject.target]?.objectId,
    //         name: selectedEdgeNodes[selectedEdgeResponseObject.target]?.label,
    //         type: selectedEdgeNodes[selectedEdgeResponseObject.target]?.kind,
    //     },
    // };

    return (
        <Box sx={sx} className={styles.container} data-testid='explore_edge-information-pane'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={selectedEdge?.name || 'None'}
                    expanded={expanded}
                    onToggleExpanded={(expanded) => {
                        setExpanded(expanded);
                    }}
                />
            </Paper>
            <Paper
                elevation={0}
                classes={{ root: styles.contentPaperRoot }}
                sx={{
                    display: expanded ? 'initial' : 'none',
                }}>
                {selectedEdge === null ? 'No information to display.' : <EdgeInfoContent selectedEdge={selectedEdge} />}
            </Paper>
        </Box>
    );
};

export default EdgeInfoPane;
