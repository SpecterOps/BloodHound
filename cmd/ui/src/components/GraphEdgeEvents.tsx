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
import { setEdgeInfoOpen, setSelectedEdge } from 'bh-shared-ui';
import { FC, useCallback } from 'react';
import { setEntityInfoOpen } from 'src/ducks/entityinfo/actions';
import {
    calculateEdgeDistanceForLabel,
    getEdgeDataFromKey,
    getEdgeLabelTextLength,
    getEdgeSourceAndTargetDisplayData,
} from 'src/ducks/graph/utils';
import { bezier } from 'src/rendering/utils/bezier';
import { useAppDispatch, useAppSelector } from 'src/store';

const GraphEdgeEvents: FC = () => {
    const dispatch = useAppDispatch();
    const graphState = useAppSelector((state) => state.explore);

    const sigma = useSigma();
    const camera = sigma.getCamera();
    const ratio = camera.getState().ratio;
    const canvases = sigma.getCanvases();
    const sigmaContainer = document.getElementById('sigma-container');
    const mouseCanvas = canvases.mouse;
    const edgeLabelsCanvas = canvases.edgeLabels;
    const { height, width } = mouseCanvas.style;

    const onClickEdge = useCallback(
        (id: string) => {
            const exploreGraphId = sigma.getGraph().getEdgeAttribute(id, 'exploreGraphId');
            const selectedItem = graphState.chartProps.items?.[id] || graphState.chartProps.items?.[exploreGraphId];
            if (!selectedItem) return;

            dispatch(setEntityInfoOpen(false));
            dispatch(setEdgeInfoOpen(true));
            dispatch(
                setSelectedEdge({
                    id: id,
                    name: selectedItem.label?.text || '',
                    data: selectedItem.data || {},
                    sourceNode: {
                        name: graphState.chartProps.items?.[selectedItem.id1].data.name,
                        id: selectedItem.id1,
                        objectId: graphState.chartProps.items?.[selectedItem.id1].data.objectid,
                        type: graphState.chartProps.items?.[selectedItem.id1].data.nodetype,
                    },
                    targetNode: {
                        name: graphState.chartProps.items?.[selectedItem.id2].data.name,
                        id: selectedItem.id2,
                        objectId: graphState.chartProps.items?.[selectedItem.id2].data.objectid,
                        type: graphState.chartProps.items?.[selectedItem.id2].data.nodetype,
                    },
                })
            );
        },
        [graphState.chartProps.items, dispatch, sigma]
    );

    const handleEdgeEvents = useCallback(
        (event: any) => {
            if (event.type === 'click' || event.type === 'mousemove') {
                const inverseSqrtZoomRatio = 1 / Math.sqrt(ratio);

                const graph = sigma.getGraph();
                const size = sigma.getSettings().edgeLabelSize * inverseSqrtZoomRatio;
                const context = edgeLabelsCanvas.getContext('2d');
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
                    if (!edgeDistance.distance) continue;

                    const textLength = getEdgeLabelTextLength(context!, attributes.label, edgeDistance.distance);
                    if (!textLength) continue;

                    let point = { x: edgeDistance.cx, y: edgeDistance.cy };
                    if (attributes.type === 'curved') {
                        const curveHeight = bezier.calculateCurveHeight(
                            attributes.groupSize,
                            attributes.groupPosition,
                            attributes.direction
                        );
                        const control = bezier.getControlAtMidpoint(curveHeight, sourceCoordinates, targetCoordinates);

                        point = bezier.getCoordinatesAlongCurve(
                            sourceCoordinatesViewport,
                            targetCoordinatesViewport,
                            sigma.framedGraphToViewport(control),
                            0.5
                        );
                    }

                    //Determine label dimensions and position
                    const xPadding = 4 * inverseSqrtZoomRatio;
                    const labelWidth = textLength * 1.4 * inverseSqrtZoomRatio + 2 * xPadding;
                    const labelHeight = size * 1.4;
                    const x1 = point.x + (-(textLength / 2) * inverseSqrtZoomRatio - xPadding);
                    const y1 = point.y + ((attributes.size / 2) * inverseSqrtZoomRatio - size);
                    const x2 = x1 + labelWidth;
                    const y2 = y1 + labelHeight;

                    const offsetY = edgeLabelsCanvas.getBoundingClientRect().y;
                    const { x: viewportX, y: viewportY } = {
                        x: event.clientX,
                        y: event.clientY - offsetY,
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
        [sigma, mouseCanvas, edgeLabelsCanvas, onClickEdge, ratio, sigmaContainer]
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

export default GraphEdgeEvents;
