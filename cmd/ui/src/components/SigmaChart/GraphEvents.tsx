// Copyright 2025 Specter Ops, Inc.
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

import { useRegisterEvents, useSetSettings, useSigma } from '@react-sigma/core';
import type { Attributes } from 'graphology-types';
import { forwardRef, useCallback, useEffect, useImperativeHandle, useLayoutEffect, useState } from 'react';
import type { SigmaNodeEventPayload } from 'sigma/sigma';
import type { Coordinates } from 'sigma/types';
import {
    DRAG_THRESHOLD,
    MOUSE_BUTTON_PRIMARY,
    getDistanceBetween,
    getEdgeDataFromKey,
    getEdgeSourceAndTargetDisplayData,
    getNodeOffset,
    graphToFramedGraph,
    resetCamera,
} from 'src/ducks/graph/utils';
import { bezier } from 'src/rendering/utils/bezier';
import { getNodeRadius, preventAllDefaults } from 'src/rendering/utils/utils';
import { useAppSelector } from 'src/store';
import { sequentialLayout, standardLayout } from 'src/views/Explore/utils';

interface SigmaChartRef {
    resetCamera: () => void;
    zoomTo: (id: string) => void;
    runSequentialLayout: () => void;
    runStandardLayout: () => void;
}

interface GraphEventProps {
    highlightedItem: string | null;
    onClickNode?: (id: string) => void;
    onClickStage?: () => void;
    onRightClickNode?: (event: SigmaNodeEventPayload) => void;
    showNodeLabels?: boolean;
    showEdgeLabels?: boolean;
}

/** Meta info about the currently dragged node */
type DragMetadata = {
    /**
     * If node has been dragged outside of threshold, then click is cancelled
     * so node is not selected.
     */
    cancelNextClick: boolean;

    /** Id of the dragged graph node. */
    id: string | null;

    /**
     * Vector from node's center to cusor position at mouse down. Used to
     * prevent the node from jumping to the cursor at drag start.
     */
    offset: Coordinates | null;

    /** Node's original position, for determining how far it's been dragged. */
    origin: Coordinates | null;
};

const DEFAULT_DRAGGED_META = {
    id: null,
    cancelNextClick: false,
    offset: null,
    origin: null,
};

export const GraphEvents = forwardRef(function GraphEvents(
    {
        highlightedItem,
        onClickNode,
        onClickStage,
        onRightClickNode,
        showNodeLabels = true,
        showEdgeLabels = true,
    }: GraphEventProps,
    ref
) {
    const exploreLayout = useAppSelector((state) => state.global.view.exploreLayout);

    const sigma = useSigma();
    const graph = sigma.getGraph();

    const registerEvents = useRegisterEvents();
    const setSettings = useSetSettings();

    const [draggedMeta, setDraggedMeta] = useState<DragMetadata>(DEFAULT_DRAGGED_META);
    const draggedNode = draggedMeta.id && graph.getNodeAttributes(draggedMeta.id);

    const sigmaChartRef = ref as React.MutableRefObject<SigmaChartRef | null>;

    useImperativeHandle(
        sigmaChartRef,
        () => {
            return {
                zoomTo: (id: string) => {
                    const node = sigma.getNodeDisplayData(id);

                    if (node) {
                        sigma.getCamera().animate(
                            {
                                x: node?.x,
                                y: node?.y,
                                ratio: 1,
                            },
                            {
                                easing: 'quadraticOut',
                            },
                            () => {
                                sigma.scheduleRefresh();
                            }
                        );
                    }
                },
                resetCamera: () => {
                    resetCamera(sigma);
                },

                runSequentialLayout: () => {
                    sequentialLayout(graph);
                    resetCamera(sigma);
                },
                runStandardLayout: () => {
                    standardLayout(graph);
                    resetCamera(sigma);
                },
            };
        },
        [sigma, graph]
    );

    const sigmaContainer = document.getElementById('sigma-container');
    const { getControlAtMidpoint, getLineLength, calculateCurveHeight } = bezier;

    const curvedEdgeReducer = useCallback(
        (edge: string, data: Attributes, newData: Attributes) => {
            // We calculate control points for all curved edges here and pass those along as edge attributes in both viewport and framed graph
            // coordinates. We can then use those control points in our edge, edge label, and edge arrow programs.
            const edgeData = getEdgeDataFromKey(edge);
            if (edgeData !== null) {
                const nodeDisplayData = getEdgeSourceAndTargetDisplayData(edgeData.source, edgeData.target, sigma);
                if (nodeDisplayData !== null) {
                    const sourceCoordinates = { x: nodeDisplayData.source.x, y: nodeDisplayData.source.y };
                    const targetCoordinates = { x: nodeDisplayData.target.x, y: nodeDisplayData.target.y };

                    const height = calculateCurveHeight(data.groupSize, data.groupPosition, data.direction);
                    const control = getControlAtMidpoint(height, sourceCoordinates, targetCoordinates);

                    newData.control = control;
                    newData.controlInViewport = sigma.framedGraphToViewport(control);
                }
            }
        },
        [sigma, calculateCurveHeight, getControlAtMidpoint]
    );

    const selfEdgeReducer = useCallback(
        (edge: string, newData: Attributes) => {
            const edgeData = getEdgeDataFromKey(edge);
            if (edgeData !== null) {
                const nodeDisplayData = getEdgeSourceAndTargetDisplayData(edgeData.source, edgeData.target, sigma);
                if (nodeDisplayData !== null) {
                    const nodeRadius = getNodeRadius(false, newData.inverseSqrtZoomRatio, nodeDisplayData.source.size);

                    const framedGraphNodeRadius = getLineLength(
                        graphToFramedGraph(sigma, { x: 0, y: 0 }),
                        graphToFramedGraph(sigma, { x: nodeRadius, y: nodeRadius })
                    );

                    newData.framedGraphNodeRadius = framedGraphNodeRadius;
                }
            }
        },
        [sigma, getLineLength]
    );

    useEffect(() => {
        registerEvents({
            enterNode: () => {
                if (sigmaContainer) sigmaContainer.style.cursor = 'grab';
            },
            leaveNode: () => {
                if (sigmaContainer) sigmaContainer.style.cursor = 'default';
            },
            downNode: (event) => {
                if (event.event.original.button === MOUSE_BUTTON_PRIMARY) {
                    const node = graph.getNodeAttributes(event.node);
                    setDraggedMeta({
                        cancelNextClick: false,
                        id: event.node,
                        offset: getNodeOffset(node, sigma.viewportToGraph(event.event)),
                        origin: { x: node.x, y: node.y },
                    });
                }
            },
            mouseup: () => {
                if (draggedNode) {
                    // Timeout prevents state update race conditions between this an mousemovebody.
                    setTimeout(() => setDraggedMeta(DEFAULT_DRAGGED_META), 10);
                }
            },
            mousemovebody: (event) => {
                if (draggedNode && draggedMeta.offset) {
                    // Get new position of node
                    const position = sigma.viewportToGraph(event);

                    // Prevent Sigma from moving camera
                    preventAllDefaults(event);

                    // DRAG_THRESHOLD
                    const newPosition = {
                        x: position.x - draggedMeta.offset.x,
                        y: position.y - draggedMeta.offset.y,
                    };

                    // Determine if node has been dragged past click-cancel threshold
                    if (!draggedMeta.cancelNextClick && draggedMeta.origin) {
                        const dragDistance = getDistanceBetween(draggedMeta.origin, newPosition);
                        if (dragDistance > DRAG_THRESHOLD) {
                            setDraggedMeta({ ...draggedMeta, cancelNextClick: true });
                        }
                    }

                    graph.setNodeAttribute(draggedMeta.id, 'x', newPosition.x);
                    graph.setNodeAttribute(draggedMeta.id, 'y', newPosition.y);
                }
            },
            doubleClickNode: (event) => {
                // Prevent zoom when node is double clicked
                preventAllDefaults(event);
            },
            clickNode: (event) => {
                if (draggedMeta.cancelNextClick) {
                    // Click handler is skipped, canceling the click. State is unset
                    setDraggedMeta({ ...draggedMeta, cancelNextClick: false });
                } else {
                    onClickNode?.(event.node);
                }
            },
            rightClickNode: (event) => {
                preventAllDefaults(event);
                onRightClickNode?.(event);
            },
            // Reducers must run again on camera state update to position edge labels correctly
            // Despite the name, this event only triggers on camera update, not any Sigma update
            updated: () => sigma.refresh(),
            clickStage: () => onClickStage?.(),
        });
    }, [
        draggedMeta,
        draggedMeta.id,
        draggedMeta.offset,
        draggedNode,
        graph,
        onClickNode,
        onClickStage,
        onRightClickNode,
        registerEvents,
        sigma,
        sigmaContainer,
    ]);

    useEffect(() => {
        setSettings({
            nodeReducer: (node, data) => {
                const camera = sigma.getCamera();
                return {
                    ...data,
                    highlighted: node === highlightedItem,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(camera.ratio),
                };
            },
            edgeReducer: (edge, data) => {
                const camera = sigma.getCamera();
                const newData: Attributes = {
                    ...data,
                    hidden: false,
                    highlighted: edge === highlightedItem,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(camera.ratio),
                };

                if (data.type === 'curved') {
                    curvedEdgeReducer(edge, data, newData);
                } else if (data.type === 'self') {
                    selfEdgeReducer(edge, newData);
                }

                return newData;
            },
        });
    }, [curvedEdgeReducer, highlightedItem, selfEdgeReducer, setSettings, sigma]);

    // Toggle off edge labels when dragging a node to avoid performance hit
    useLayoutEffect(() => {
        setSettings({ renderEdgeLabels: draggedNode ? false : showEdgeLabels });
    }, [draggedNode, setSettings, showEdgeLabels]);

    useLayoutEffect(() => {
        if (sigmaChartRef?.current) {
            if (exploreLayout === 'sequential') {
                sigmaChartRef?.current?.runSequentialLayout();
            } else if (exploreLayout === 'standard') {
                sigmaChartRef?.current?.runStandardLayout();
            } else {
                resetCamera(sigma);
            }
        } else {
            resetCamera(sigma);
        }
    }, [sigma, exploreLayout, sigmaChartRef]);

    useLayoutEffect(() => {
        setSettings({
            renderLabels: showNodeLabels,
            renderEdgeLabels: showEdgeLabels,
        });
    }, [setSettings, showNodeLabels, showEdgeLabels]);

    return null;
});
