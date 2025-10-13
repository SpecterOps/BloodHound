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

import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { forwardRef, useEffect, useMemo, useRef, useState } from 'react';
import { Sigma } from 'sigma';
import { Settings } from 'sigma/settings';
import { SigmaNodeEventPayload } from 'sigma/types';
import { MAX_CAMERA_RATIO, MIN_CAMERA_RATIO } from 'src/ducks/graph/utils';
// import EdgeArrowProgram from 'src/rendering/programs/edge.arrow';
// import CurvedEdgeArrowProgram from 'src/rendering/programs/edge.curvedArrow';
// import SelfEdgeArrowProgram from 'src/rendering/programs/edge.selfArrow';
import { EdgeCurvedArrowProgram } from '@sigma/edge-curve';
import { createNodeImageProgram } from '@sigma/node-image';
import { EdgeArrowProgram } from 'sigma/rendering';
import drawEdgeLabel from 'src/rendering/programs/edge-label';
import drawHover from 'src/rendering/programs/node-hover';
import drawLabel from 'src/rendering/programs/node-label';
import { GraphEvents } from './GraphEvents';

interface SigmaChartProps {
    graph: MultiDirectedGraph<Attributes, Attributes, Attributes>;
    highlightedItem: string | null;
    onClickNode: (id: string) => void;
    onClickEdge: (id: string) => void;
    onClickStage: () => void;
    handleContextMenu: (event: SigmaNodeEventPayload) => void;
    showNodeLabels?: boolean;
    showEdgeLabels?: boolean;
}

interface SigmaChartRef {
    resetCamera: () => void;
    zoomTo: (id: string) => void;
    runSequentialLayout: () => void;
    runStandardLayout: () => void;
}

const defaultSettings = {
    nodeProgramClasses: {
        image: createNodeImageProgram(),
        // combined: getNodeCombinedProgram(),
        // glyphs: getNodeGlyphsProgram(),
    },
    defaultNodeType: 'image',
    edgeProgramClasses: {
        // straightNoArrow: EdgeRectangleProgram,
        // curvedNoArrow: EdgeCurveProgram,
        // straightArrow: EdgeArrowProgram,
        // curvedArrow: EdgeCurvedArrowProgram,
        // straightDoubleArrow: EdgeDoubleArrowProgram,
        // curvedDoubleArrow: EdgeCurvedDoubleArrowProgram,
        arrow: EdgeArrowProgram,
        curved: EdgeCurvedArrowProgram,
        self: EdgeCurvedArrowProgram,
    },
    renderEdgeLabels: true,
    renderLabels: true,
    // hoverRenderer: drawHover,
    defaultDrawNodeHover: drawHover,
    defaultDrawEdgeLabel: drawEdgeLabel,
    edgeLabelRenderer: drawEdgeLabel,
    edgeLabelSize: 12,
    enableEdgeEvents: true,
    labelSize: 12,
    labelFont: 'Roboto',
    labelColor: { color: '#000' },
    labelRenderer: drawLabel,
    maxCameraRatio: MAX_CAMERA_RATIO,
    minCameraRatio: MIN_CAMERA_RATIO,
    allowInvalidContainer: true,
};

const SigmaChart = forwardRef<SigmaChartRef, SigmaChartProps>(function SigmaChart(
    {
        graph,
        highlightedItem,
        onClickNode,
        onClickEdge,
        onClickStage,
        handleContextMenu,
        showNodeLabels = true,
        showEdgeLabels = true,
    }: SigmaChartProps,
    ref
) {
    const containerRef = useRef<HTMLDivElement>(null);
    const labelColor = containerRef.current
        ? getComputedStyle(containerRef.current).getPropertyValue('--contrast')
        : '#000';

    const settings: Partial<Settings> = useMemo(() => {
        return {
            ...defaultSettings,
            renderEdgeLabels: showEdgeLabels,
            renderLabels: showNodeLabels,
            labelColor: { color: labelColor },
        };
    }, [showEdgeLabels, showNodeLabels, labelColor]);

    const [sigma, setSigma] = useState<Sigma<Attributes, Attributes, Attributes> | undefined>();

    useEffect(() => {
        if (containerRef.current) setSigma(new Sigma(graph, containerRef.current, settings));
        return () => {
            sigma?.kill();
        };
    }, [settings, graph]);

    if (!sigma)
        return (
            <div data-testid='sigma-container-wrapper'>
                <div
                    id='sigma-container'
                    className='absolute top-0 left-0 h-full w-full bg-neutral-1'
                    ref={containerRef}
                />
            </div>
        );

    return (
        <div data-testid='sigma-container-wrapper'>
            <div id='sigma-container' className='absolute top-0 left-0 h-full w-full bg-neutral-1' ref={containerRef}>
                <GraphEvents
                    sigma={sigma}
                    highlightedItem={highlightedItem ?? null}
                    onClickEdge={onClickEdge}
                    onClickNode={onClickNode}
                    onClickStage={onClickStage}
                    onRightClickNode={handleContextMenu}
                    showNodeLabels={showNodeLabels}
                    showEdgeLabels={showEdgeLabels}
                    ref={ref}
                />
            </div>
        </div>
    );
});

export default SigmaChart;
