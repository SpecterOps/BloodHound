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

import { useSigma } from '@react-sigma/core';
import { collapseAllSections, useExploreSelectedItem, useFeatureFlag } from 'bh-shared-ui';
import { FC, useCallback } from 'react';
import {
    calculateEdgeDistanceForLabel,
    getEdgeDataFromKey,
    getEdgeLabelTextLength,
    getEdgeSourceAndTargetDisplayData,
} from 'src/ducks/graph/utils';
import { getBackgroundBoundInfo, getSelfEdgeStartingPoint } from 'src/rendering/programs/edge-label';
import { getControlPointsFromGroupSize } from 'src/rendering/programs/edge.self';
import { bezier } from 'src/rendering/utils/bezier';
import { useAppDispatch, useAppSelector } from 'src/store';

const GraphEdgeEventsV2: FC = () => {
    const graphState = useAppSelector((state) => state.explore);
    const { setSelectedItem: setExploreSelectedItem } = useExploreSelectedItem();
    const backButtonFlagQuery = useFeatureFlag('back_button_support');
    const dispatch = useAppDispatch();

    const sigma = useSigma();
    const canvases = sigma.getCanvases();
    const sigmaContainer = document.getElementById('sigma-container');
    const mouseCanvas = canvases.mouse;
    const edgeLabelsCanvas = canvases.edgeLabels;
    const { height, width } = mouseCanvas.style;

    const onClickEdge = useCallback(
        (id: string) => {
            if (backButtonFlagQuery.data?.enabled) {
                setExploreSelectedItem(id);
            }
            const exploreGraphId = sigma.getGraph().getEdgeAttribute(id, 'exploreGraphId');
            const selectedItem = graphState.chartProps.items?.[id] || graphState.chartProps.items?.[exploreGraphId];
            dispatch(collapseAllSections());
            if (!selectedItem) return;
        },
        [graphState.chartProps.items, sigma, setExploreSelectedItem, backButtonFlagQuery.data?.enabled, dispatch]
    );

    const handleEdgeEvents = useCallback(
        (event: any) => {
            const context = edgeLabelsCanvas.getContext('2d');
            if (!context) return;
            if (event.type === 'click' || event.type === 'mousemove') {
                const camera = sigma.getCamera();
                const ratio = camera.getState().ratio;
                const inverseSqrtZoomRatio = 1 / Math.sqrt(ratio);
                const size = sigma.getSettings().edgeLabelSize;
                const graph = sigma.getGraph();
                const edges = graph.edges();

                for (let i = 0; i < edges.length; i++) {
                    const edge: string = edges[i];
                    const attributes = graph.getEdgeAttributes(edge);

                    const edgeData = getEdgeDataFromKey(edge);
                    if (edgeData === null) continue;
                    const nodeDisplayData = getEdgeSourceAndTargetDisplayData(edgeData.source, edgeData.target, sigma);
                    if (nodeDisplayData === null) continue;

                    const { source, target } = nodeDisplayData;
                    const sourceCoordinates = { x: source.x, y: source.y };
                    const targetCoordinates = { x: target.x, y: target.y };
                    const sourceCoordinatesViewport = sigma.framedGraphToViewport(sourceCoordinates);
                    const targetCoordinatesViewport = sigma.framedGraphToViewport(targetCoordinates);

                    const edgeDistance = calculateEdgeDistanceForLabel(
                        { ...sourceCoordinatesViewport, size: source.size },
                        { ...targetCoordinatesViewport, size: target.size }
                    );
                    const textLength = getEdgeLabelTextLength(context, attributes.label, edgeDistance.distance);
                    if (!textLength) continue;

                    let point = { x: edgeDistance.cx, y: edgeDistance.cy };
                    if (attributes.type === 'curved') {
                        const curveHeight = bezier.calculateCurveHeight(
                            attributes.groupSize,
                            attributes.groupPosition,
                            attributes.direction
                        );
                        const control = sigma.framedGraphToViewport(
                            bezier.getControlAtMidpoint(curveHeight, sourceCoordinates, targetCoordinates)
                        );

                        point = bezier.getCoordinatesAlongQuadraticBezier(
                            sourceCoordinatesViewport,
                            targetCoordinatesViewport,
                            control,
                            0.5
                        );
                    } else if (attributes.type === 'self') {
                        const radius = bezier.getLineLength(
                            { x: 0, y: 0 },
                            {
                                x: source.size * Math.pow(inverseSqrtZoomRatio, 3),
                                y: target.size * Math.pow(inverseSqrtZoomRatio, 3),
                            }
                        );

                        const { control2, control3 } = getControlPointsFromGroupSize(
                            attributes.groupPosition,
                            radius * 3,
                            sourceCoordinatesViewport,
                            false,
                            true
                        );

                        point = getSelfEdgeStartingPoint(
                            sourceCoordinatesViewport,
                            control2,
                            control3,
                            sourceCoordinatesViewport
                        );
                    }

                    const { deltaX, deltaY, width, height } = getBackgroundBoundInfo(
                        inverseSqrtZoomRatio,
                        textLength,
                        attributes.size * inverseSqrtZoomRatio,
                        size
                    );

                    const x1 = point.x + deltaX;
                    const y1 = point.y + deltaY;
                    const x2 = x1 + width;
                    const y2 = y1 + height;

                    const offsetX = edgeLabelsCanvas.getBoundingClientRect().x;
                    const { x: viewportX, y: viewportY } = {
                        x: event.clientX - offsetX,
                        y: event.clientY,
                    };

                    //Check if the click happened within the bounds of the label
                    if (viewportX > x1 && viewportX < x2 && viewportY > y1 && viewportY < y2) {
                        if (event.type === 'click') {
                            onClickEdge(edge);
                        } else {
                            //Hover the edge label
                            if (sigmaContainer) sigmaContainer.style.cursor = 'pointer';
                        }
                        //Return early so as not to propagate the event to sigma handlers
                        return;
                    }
                }
            }

            if (sigmaContainer && sigmaContainer.style.cursor === 'pointer') sigmaContainer.style.cursor = 'default';

            const customEvent = new Event(event.type, { bubbles: true });
            Object.assign(customEvent, {
                clientX: event.clientX,
                clientY: event.clientY,
                deltaY: event.deltaY, // Needed for wheel events
                button: event.button, // Needed for mousedown/dragging events
            });
            mouseCanvas.dispatchEvent(customEvent);
            sigma.scheduleRefresh();
        },
        [sigma, mouseCanvas, edgeLabelsCanvas, onClickEdge, sigmaContainer]
    );

    return (
        <canvas
            id='edge-events'
            width={width.slice(0, width.length - 2)}
            height={height.slice(0, height.length - 2)}
            style={{ position: 'absolute', height: height, width: width, inset: 0 }}
            onClick={handleEdgeEvents}
            onContextMenu={handleEdgeEvents}
            onMouseDown={handleEdgeEvents}
            onWheel={handleEdgeEvents}
            onMouseOut={handleEdgeEvents}
            onMouseMove={handleEdgeEvents}
            onMouseUp={handleEdgeEvents}></canvas>
    );
};

export default GraphEdgeEventsV2;
