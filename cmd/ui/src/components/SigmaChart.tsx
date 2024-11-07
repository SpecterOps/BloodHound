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
import { forwardRef } from 'react';
import { GraphEvents } from 'src/components/GraphEvents';
import { MAX_CAMERA_RATIO, MIN_CAMERA_RATIO } from 'src/ducks/graph/utils';
import drawEdgeLabel from 'src/rendering/programs/edge-label';
import EdgeArrowProgram from 'src/rendering/programs/edge.arrow';
import CurvedEdgeArrowProgram from 'src/rendering/programs/edge.curvedArrow';
import SelfEdgeArrowProgram from 'src/rendering/programs/edge.selfArrow';
import drawHover from 'src/rendering/programs/node-hover';
import drawLabel from 'src/rendering/programs/node-label';
import getNodeCombinedProgram from 'src/rendering/programs/node.combined';
import getNodeGlyphsProgram from 'src/rendering/programs/node.glyphs';
import GraphEdgeEvents from './GraphEdgeEvents';
import { useTheme } from '@mui/material';
import { SigmaNodeEventPayload } from 'sigma/sigma';

interface SigmaChartProps {
    graph: Graph<Attributes, Attributes, Attributes>;
    onDoubleClickNode: (id: string) => void;
    onClickNode: (id: string) => void;
    onClickEdge: (id: string, relatedFindingType?: string | null) => void;
    onClickStage: () => void;
    edgeReducer: (edge: string, data: Attributes, graph: AbstractGraph) => Attributes;
    handleContextMenu: (event: SigmaNodeEventPayload) => void;
}

const SigmaChart = forwardRef(function SigmaChart(
    {
        graph,
        onDoubleClickNode,
        onClickNode,
        onClickEdge,
        onClickStage,
        edgeReducer,
        handleContextMenu,
    }: Partial<SigmaChartProps>,
    ref
) {
    const theme = useTheme();

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
                    background: theme.palette.neutral.primary,
                }}
                graph={graph || MultiDirectedGraph}
                settings={{
                    nodeProgramClasses: {
                        combined: getNodeCombinedProgram(),
                        glyphs: getNodeGlyphsProgram(),
                    },
                    edgeProgramClasses: {
                        curved: CurvedEdgeArrowProgram,
                        self: SelfEdgeArrowProgram,
                        arrow: EdgeArrowProgram,
                    },
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
                    ref={ref}
                />
            </SigmaContainer>
        </div>
    );
});

export default SigmaChart;
