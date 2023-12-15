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

import { SigmaContainer } from '@react-sigma/core';
import '@react-sigma/core/lib/react-sigma.min.css';
import Graph, { MultiDirectedGraph } from 'graphology';
import { AbstractGraph, Attributes } from 'graphology-types';
import { FC } from 'react';
import GraphButtons from 'src/components/GraphButtons';
import { GraphButtonOptions } from 'src/components/GraphButtons/GraphButtons';
import { GraphEvents } from 'src/components/GraphEvents';
import { MAX_CAMERA_RATIO, MIN_CAMERA_RATIO } from 'src/ducks/graph/utils';
import { RankDirection } from 'src/hooks/useLayoutDagre/useLayoutDagre';
import drawEdgeLabel from 'src/rendering/programs/edge-label';
import EdgeArrowProgram from 'src/rendering/programs/edge.arrow';
import CurvedEdgeArrowProgram from 'src/rendering/programs/edge.curvedArrow';
import drawHover from 'src/rendering/programs/node-hover';
import drawLabel from 'src/rendering/programs/node-label';
import getNodeCombinedProgram from 'src/rendering/programs/node.combined';
import getNodeGlyphsProgram from 'src/rendering/programs/node.glyphs';
import GraphEdgeEvents from './GraphEdgeEvents';
import { Box } from '@mui/material';
import { GraphNodes } from 'js-client-library';
import { GraphButtonProps, SearchCurrentNodes } from 'bh-shared-ui';
import { SigmaNodeEventPayload } from 'sigma/sigma';

interface SigmaChartProps {
    rankDirection: RankDirection;
    options: GraphButtonOptions;
    nonLayoutButtons: GraphButtonProps[];
    isCurrentSearchOpen: boolean;
    graph: Graph<Attributes, Attributes, Attributes>;
    currentNodes: GraphNodes;
    toggleCurrentSearch: () => void;
    onDoubleClickNode: (id: string) => void;
    onClickNode: (id: string) => void;
    onClickEdge: (id: string, relatedFindingType?: string | null) => void;
    onClickStage: () => void;
    edgeReducer: (edge: string, data: Attributes, graph: AbstractGraph) => Attributes;
    handleContextMenu: (event: SigmaNodeEventPayload) => void;
}

const SigmaChart: FC<Partial<SigmaChartProps>> = ({
    rankDirection,
    options,
    nonLayoutButtons,
    isCurrentSearchOpen,
    graph,
    currentNodes,
    toggleCurrentSearch,
    onDoubleClickNode,
    onClickNode,
    onClickEdge,
    onClickStage,
    edgeReducer,
    handleContextMenu,
}) => {
    return (
        <div
            // prevent browser's default right-click behavior
            onContextMenu={(e) => e.preventDefault()}>
            <SigmaContainer
                id='sigma-container'
                style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    height: '100%',
                    width: '100%',
                    background: 'linear-gradient(rgb(228, 233, 235) 0%, rgb(228, 233, 235) 100%)',
                }}
                graph={graph || MultiDirectedGraph}
                settings={{
                    nodeProgramClasses: {
                        combined: getNodeCombinedProgram(),
                        glyphs: getNodeGlyphsProgram(),
                    },
                    edgeProgramClasses: { curved: CurvedEdgeArrowProgram, arrow: EdgeArrowProgram },
                    renderEdgeLabels: true,
                    hoverRenderer: drawHover,
                    edgeLabelRenderer: drawEdgeLabel,
                    edgeLabelSize: 12,
                    labelSize: 12,
                    labelFont: 'Roboto',
                    labelRenderer: drawLabel,
                    maxCameraRatio: MAX_CAMERA_RATIO,
                    minCameraRatio: MIN_CAMERA_RATIO,
                }}>
                <GraphEdgeEvents />
                <GraphEvents
                    onDoubleClickNode={onDoubleClickNode}
                    onClickNode={onClickNode}
                    onClickEdge={onClickEdge}
                    onClickStage={onClickStage}
                    edgeReducer={edgeReducer}
                    onRightClickNode={handleContextMenu}
                />
                <Box position={'absolute'} bottom={16}>
                    {isCurrentSearchOpen && (
                        <SearchCurrentNodes
                            sx={{ marginLeft: 2, padding: 1 }}
                            currentNodes={currentNodes || {}}
                            onSelect={(node) => {
                                onClickNode?.(node.id);
                                toggleCurrentSearch?.();
                            }}
                            onClose={toggleCurrentSearch}
                        />
                    )}
                    <GraphButtons rankDirection={rankDirection} options={options} nonLayoutButtons={nonLayoutButtons} />
                </Box>
            </SigmaContainer>
        </div>
    );
};

export default SigmaChart;
