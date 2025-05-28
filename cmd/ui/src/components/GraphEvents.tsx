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
import { useExploreSelectedItem } from 'bh-shared-ui';
import { Attributes } from 'graphology-types';
import { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import {
    getEdgeDataFromKey,
    getEdgeSourceAndTargetDisplayData,
    graphToFramedGraph,
    resetCamera,
} from 'src/ducks/graph/utils';
import { bezier } from 'src/rendering/utils/bezier';
import { DRAG_DELAY, MOUSE_BUTTON_PRIMARY, getNodeRadius, preventAllDefaults } from 'src/rendering/utils/utils';
import { sequentialLayout, standardLayout } from 'src/views/Explore/utils';

export interface GraphEventProps {
    onClickNode?: (id: string) => void;
    onClickEdge?: (id: string, relatedFindingType?: string | null) => void;
    onClickStage?: () => void;
    onRightClickNode?: (event: SigmaNodeEventPayload) => void;
    showNodeLabels?: boolean;
    showEdgeLabels?: boolean;
}

export const GraphEvents = forwardRef(function GraphEvents(
    {
        onClickNode,
        onClickEdge,
        onClickStage,
        onRightClickNode,
        showNodeLabels = true,
        showEdgeLabels = true,
    }: GraphEventProps,
    ref
) {
    const { cancelHighlight, highlightedItem, selectedItem } = useExploreSelectedItem();

    const sigma = useSigma();
    const registerEvents = useRegisterEvents();
    const setSettings = useSetSettings();

    const [draggedNode, setDraggedNode] = useState<string | null>(null);
    const isDragging = !!draggedNode;

    const dragTimerRef = useRef<ReturnType<typeof setTimeout>>();

    const graph = sigma.getGraph();

    useImperativeHandle(
        ref,
        () => {
            return {
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

    const clearDraggedNode = () => {
        setDraggedNode(null);
        clearTimeout(dragTimerRef.current);
    };

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
                    // Drag after a timeout, to prevent accidental dragging
                    dragTimerRef.current = setTimeout(() => setDraggedNode(event.node), DRAG_DELAY);
                }
            },
            mouseup: () => clearDraggedNode(),
            mousemovebody: (event) => {
                if (draggedNode) {
                    // Prevent Sigma from moving camera
                    preventAllDefaults(event);

                    // Get new position of node
                    const position = sigma.viewportToGraph(event);
                    sigma.getGraph().setNodeAttribute(draggedNode, 'x', position.x);
                    sigma.getGraph().setNodeAttribute(draggedNode, 'y', position.y);
                }
            },
            mousemove: (event) => {
                if (draggedNode) {
                    // Prevent Sigma from moving camera
                    preventAllDefaults(event);
                }
            },
            doubleClickNode: (event) => {
                // Prevent zoom when node is double clicked
                preventAllDefaults(event);
            },
            clickNode: (event) => {
                // Dragged node check ensures that node isn't selected when drag ends
                // This would also sometimes cause the drag to not be cleared
                if (!draggedNode) {
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
            clickStage: () => {
                cancelHighlight();
                onClickStage?.();
            },
        });
    }, [
        cancelHighlight,
        draggedNode,
        isDragging,
        onClickEdge,
        onClickNode,
        onClickStage,
        onRightClickNode,
        registerEvents,
        selectedItem,
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
    }, [curvedEdgeReducer, draggedNode, highlightedItem, selectedItem, selfEdgeReducer, setSettings, sigma]);

    // Toggle off edge labels when dragging a node to avoid performance hit
    useEffect(() => {
        setSettings({ renderEdgeLabels: !draggedNode && showEdgeLabels });
    }, [draggedNode, setSettings, showEdgeLabels]);

    useEffect(() => {
        resetCamera(sigma);
    }, [sigma]);

    useEffect(() => {
        setSettings({
            renderLabels: showNodeLabels,
            renderEdgeLabels: showEdgeLabels,
        });
    }, [setSettings, showNodeLabels, showEdgeLabels]);

    return null;
});
