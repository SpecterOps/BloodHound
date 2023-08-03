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

import { useRegisterEvents, useSetSettings, useSigma } from '@react-sigma/core';
import { AbstractGraph, Attributes } from 'graphology-types';
import { FC, useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { setSelectedEdge } from 'src/ducks/edgeinfo/edgeSlice';
import { getEdgeDataFromKey, getEdgeSourceAndTargetDisplayData, resetCamera } from 'src/ducks/graph/utils';
import { bezier } from 'src/rendering/utils/bezier';
import { AppState, useAppDispatch } from 'src/store';

export interface GraphEventProps {
    onDoubleClickNode?: (id: string) => void;
    onClickNode?: (id: string) => void;
    onClickEdge?: (id: string, relatedFindingType?: string | null) => void;
    onClickStage?: () => void;
    edgeReducer?: (edge: string, data: Attributes, graph: AbstractGraph) => Attributes;
}

export const GraphEvents: FC<GraphEventProps> = ({
    onDoubleClickNode,
    onClickNode,
    onClickEdge,
    onClickStage,
    edgeReducer,
}) => {
    const dispatch = useAppDispatch();
    const selectedEdge = useSelector((state: AppState) => state.edgeinfo.selectedEdge);

    const sigma = useSigma();
    const registerEvents = useRegisterEvents();
    const setSettings = useSetSettings();
    const edgeEventsCanvasRef = useRef<HTMLCanvasElement | null>(document.querySelector('#sigma-container'));

    const [hoveredNode, setHoveredNode] = useState<string | null>(null);
    const [draggedNode, setDraggedNode] = useState<string | null>(null);
    const [selectedNode, setSelectedNode] = useState<string | null>(null);
    const [isDragging, setIsDragging] = useState(false);

    const isLongPress = useRef(false);
    const dragTimerRef = useRef<ReturnType<typeof setTimeout>>();
    const clickTimerRef = useRef<ReturnType<typeof setTimeout>>();
    const prevent = useRef(false);

    const clearDraggedNode = () => {
        setDraggedNode(null);
        setIsDragging(false);
        clearTimeout(dragTimerRef.current);
        isLongPress.current = false;
    };

    useEffect(() => {
        registerEvents({
            enterNode: (event) => {
                if (edgeEventsCanvasRef.current) edgeEventsCanvasRef.current.style.cursor = 'grab';
                setHoveredNode(event.node);
            },
            leaveNode: () => {
                if (edgeEventsCanvasRef.current) edgeEventsCanvasRef.current.style.cursor = 'default';
                setHoveredNode(null);
            },
            downNode: (event) => {
                setDraggedNode(event.node);

                dragTimerRef.current = setTimeout(() => {
                    isLongPress.current = true;
                }, 150);
            },
            mouseup: () => clearDraggedNode(),
            mousemovebody: (event) => {
                if (draggedNode) {
                    setIsDragging(true);
                    // Prevent sigma to move camera:
                    event.preventSigmaDefault();
                    event.original.preventDefault();
                    event.original.stopPropagation();

                    // Get new position of node
                    const position = sigma.viewportToGraph(event);
                    sigma.getGraph().setNodeAttribute(draggedNode, 'x', position.x);
                    sigma.getGraph().setNodeAttribute(draggedNode, 'y', position.y);
                }
            },
            mousemove: (event) => {
                if (draggedNode) {
                    // Prevent sigma to move camera:
                    event.preventSigmaDefault();
                    event.original.preventDefault();
                    event.original.stopPropagation();
                }
            },
            doubleClickNode: (event) => {
                if (!onDoubleClickNode) return;
                event.preventSigmaDefault();
                clearDraggedNode();
                clearTimeout(clickTimerRef.current);
                prevent.current = true;
                onDoubleClickNode(event.node);
            },
            clickNode: (event) => {
                if (onClickNode && !isLongPress.current && !isDragging) {
                    clearDraggedNode();

                    clickTimerRef.current = setTimeout(function () {
                        if (!prevent.current) {
                            onClickNode(event.node);
                            setSelectedNode(event.node);
                            dispatch(setSelectedEdge(null));
                        }
                        prevent.current = false;
                    }, 200);
                }
            },
            // We need our reducers to run again when the camera state gets updated to position edge labels correctly.
            // Despite its name, this event only triggers on camera update, not any Sigma update
            updated: () => sigma.refresh(),
            clickStage: () => onClickStage && onClickStage(),
        });
    }, [
        dispatch,
        registerEvents,
        onDoubleClickNode,
        onClickNode,
        onClickEdge,
        onClickStage,
        sigma,
        draggedNode,
        isDragging,
        selectedNode,
    ]);

    useEffect(() => {
        setSettings({
            nodeReducer: (node, data) => {
                const camera = sigma.getCamera();
                const newData: Attributes = {
                    ...data,
                    highlighted: data.highlighted || false,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(camera.ratio),
                };

                if (node === selectedNode) {
                    newData.highlighted = true;
                }

                return newData;
            },
            edgeReducer: (edge, data) => {
                const graph = sigma.getGraph();
                const camera = sigma.getCamera();
                const newData: Attributes = {
                    ...data,
                    hidden: false,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(camera.ratio),
                };

                if (edge === selectedEdge?.id) {
                    newData.selected = true;
                    newData.forceLabel = true;
                } else {
                    newData.selected = false;
                    newData.forceLabel = false;
                }

                // We calculate control points for all curved edges here and pass those along as edge attributes in both viewport and framed graph
                // coordinates. We can then use those control points in our edge, edge label, and edge arrow programs.
                if (data.type === 'curved') {
                    const edgeData = getEdgeDataFromKey(edge);
                    if (edgeData !== null) {
                        const nodeDisplayData = getEdgeSourceAndTargetDisplayData(
                            edgeData.source,
                            edgeData.target,
                            sigma
                        );
                        if (nodeDisplayData !== null) {
                            const sourceCoordinates = { x: nodeDisplayData.source.x, y: nodeDisplayData.source.y };
                            const targetCoordinates = { x: nodeDisplayData.target.x, y: nodeDisplayData.target.y };

                            const height = bezier.calculateCurveHeight(
                                data.groupSize,
                                data.groupPosition,
                                data.direction
                            );
                            const control = bezier.getControlAtMidpoint(height, sourceCoordinates, targetCoordinates);

                            newData.control = control;
                            newData.controlInViewport = sigma.framedGraphToViewport(control);
                        }
                    }
                }

                if (edgeReducer) return edgeReducer(edge, newData, graph);

                return newData;
            },
        });
    }, [hoveredNode, draggedNode, selectedNode, selectedEdge, edgeReducer, setSettings, sigma]);

    // Toggle off edge labels when dragging a node. Since these are rendered on a 2d canvas, dragging nodes with lots of edges
    // can tank performance
    useEffect(() => {
        if (draggedNode) {
            setSettings({ renderEdgeLabels: false });
        } else {
            setSettings({ renderEdgeLabels: true });
        }
    }, [draggedNode, setSettings]);

    useEffect(() => {
        resetCamera(sigma);
    }, [sigma]);

    useEffect(() => {
        if (selectedEdge) setSelectedNode(null);
    }, [selectedEdge]);

    return null;
};
