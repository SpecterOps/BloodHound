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

// import { useSetSettings } from '@react-sigma/core';
import type { Attributes } from 'graphology-types';
import { forwardRef, useCallback, useEffect, useImperativeHandle, useLayoutEffect, useMemo, useState } from 'react';
import { Sigma } from 'sigma';
import type {
    CameraEvents,
    CameraState,
    Coordinates,
    MouseCaptorEvents,
    MouseCoords,
    SigmaEdgeEventPayload,
    SigmaEventPayload,
    SigmaEvents,
    SigmaNodeEventPayload,
    TouchCaptorEvents,
    TouchCoords,
} from 'sigma/types';
import {
    DRAG_THRESHOLD,
    MOUSE_BUTTON_PRIMARY,
    getDistanceBetween,
    getNodeOffset,
    resetCamera,
} from 'src/ducks/graph/utils';
import { useAppSelector } from 'src/store';
import { preventAllDefaults } from 'src/utils';
import { sequentialLayout, standardLayout } from 'src/views/Explore/utils';

interface SigmaChartRef {
    resetCamera: () => void;
    zoomTo: (id: string) => void;
    runSequentialLayout: () => void;
    runStandardLayout: () => void;
}

interface GraphEventProps {
    sigma: Sigma<Attributes, Attributes, Attributes>;
    highlightedItem: string | null;
    onClickEdge?: (id: string) => void;
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

type AllEvents = SigmaEvents & TouchCaptorEvents & MouseCaptorEvents & CameraEvents;

const sigmaEventKeys = [
    'clickNode',
    'rightClickNode',
    'downNode',
    'enterNode',
    'leaveNode',
    'doubleClickNode',
    'wheelNode',
    'clickEdge',
    'rightClickEdge',
    'downEdge',
    'enterEdge',
    'leaveEdge',
    'doubleClickEdge',
    'wheelEdge',
    'clickStage',
    'rightClickStage',
    'downStage',
    'doubleClickStage',
    'wheelStage',
    'beforeRender',
    'afterRender',
    'kill',
    'upStage',
    'upEdge',
    'upNode',
    'enterStage',
    'leaveStage',
    'resize',
    'afterClear',
    'afterProcess',
    'beforeClear',
    'beforeProcess',
    'moveBody',
] as const;

const mouseEventKeys = [
    'click',
    'rightClick',
    'doubleClick',
    'mouseup',
    'mousedown',
    'mousemove',
    'mousemovebody',
    'mouseleave',
    'mouseenter',
    'wheel',
] as const;

const touchEventKeys = ['touchup', 'touchdown', 'touchmove', 'touchmovebody', 'tap', 'doubletap'] as const;

const cameraEventKeys = ['updated'] as const;

const allEventKeys = [...sigmaEventKeys, ...mouseEventKeys, ...cameraEventKeys] as const;

export const GraphEvents = forwardRef(function GraphEvents(
    { sigma, highlightedItem, onClickNode, onClickStage, onRightClickNode, onClickEdge }: GraphEventProps,
    ref
) {
    const exploreLayout = useAppSelector((state) => state.global.view.exploreLayout);

    const graph = sigma.getGraph();
    const sigmaContainer = sigma.getContainer();

    const [neighbors, setNeighbors] = useState<Set<string>>(new Set());
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

    const events = useMemo((): Partial<Record<(typeof allEventKeys)[number], (event?: any) => void>> => {
        return {
            enterNode: () => {
                sigmaContainer.style.cursor = 'grab';
            },
            leaveNode: () => {
                sigmaContainer.style.cursor = 'default';
            },
            downNode: (event: SigmaNodeEventPayload) => {
                //@ts-ignore
                if (event.event.original.button === MOUSE_BUTTON_PRIMARY) {
                    sigma.setSettings({ renderEdgeLabels: false });
                    const node = graph.getNodeAttributes(event.node);

                    setNeighbors(new Set(graph.neighbors(event.node)));
                    setDraggedMeta({
                        cancelNextClick: false,
                        id: event.node,
                        offset: getNodeOffset(node, sigma.viewportToGraph(event.event)),
                        origin: { x: node.x, y: node.y },
                    });
                }
            },
            upNode: () => {
                if (draggedNode) {
                    sigma.setSettings({ renderEdgeLabels: true });
                    // Timeout prevents state update race conditions between this an mousemovebody.
                    setTimeout(() => setDraggedMeta(DEFAULT_DRAGGED_META), 10);
                }
            },
            mousemovebody: (event: MouseCoords) => {
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
            doubleClickNode: (event: SigmaNodeEventPayload) => {
                // Prevent zoom when node is double clicked
                preventAllDefaults(event);
            },
            clickNode: (event: SigmaNodeEventPayload) => {
                if (draggedMeta.cancelNextClick) {
                    // Click handler is skipped, canceling the click. State is unset
                    setDraggedMeta({ ...draggedMeta, cancelNextClick: false });
                } else {
                    onClickNode?.(event.node);
                }
            },
            rightClickNode: (event: SigmaNodeEventPayload) => {
                preventAllDefaults(event);
                onRightClickNode?.(event);
            },
            enterEdge: () => {
                sigmaContainer.style.cursor = 'pointer';
            },
            leaveEdge: () => {
                sigmaContainer.style.cursor = 'default';
            },
            clickEdge: (event: SigmaEdgeEventPayload) => {
                onClickEdge?.(event.edge);
            },
            // Reducers must run again on camera state update to position edge labels correctly
            // Despite the name, this event only triggers on camera update, not any Sigma update
            updated: () => sigma.refresh(),
            downStage: () => sigma.setSettings({ renderEdgeLabels: false }),
            upStage: () => sigma.setSettings({ renderEdgeLabels: true }),

            clickStage: () => {
                onClickStage?.();
                setNeighbors(new Set());
            },
        };
    }, [
        draggedMeta,
        draggedNode,
        graph,
        onClickNode,
        onClickStage,
        onRightClickNode,
        sigma,
        sigmaContainer,
        onClickEdge,
    ]);

    const registerEvents = useCallback(
        (events: Partial<AllEvents>): void => {
            Object.entries(events).forEach(([event, handler]) => {
                if (sigmaEventKeys.includes(event as (typeof sigmaEventKeys)[number])) {
                    sigma.on(event as keyof SigmaEvents, handler as (payload: SigmaEventPayload) => void);
                }

                if (mouseEventKeys.includes(event as (typeof mouseEventKeys)[number])) {
                    sigma
                        .getMouseCaptor()
                        .on(event as keyof MouseCaptorEvents, handler as (payload: MouseCoords) => void);
                }

                if (touchEventKeys.includes(event as (typeof touchEventKeys)[number])) {
                    sigma
                        .getTouchCaptor()
                        .on(event as keyof TouchCaptorEvents, handler as (coordinates: TouchCoords) => void);
                }

                if (cameraEventKeys.includes(event as (typeof cameraEventKeys)[number])) {
                    sigma.getCamera().on(event as keyof CameraEvents, handler as (state: CameraState) => void);
                }
            });
        },
        [sigma]
    );

    const removeEvents = useCallback(() => {
        allEventKeys.forEach((event) => {
            const eventHandler = events[event] as (...args: unknown[]) => void;

            if (!eventHandler) return;

            if (sigmaEventKeys.includes(event as (typeof sigmaEventKeys)[number])) {
                sigma.off(event as keyof SigmaEvents, eventHandler);
            }
            if (mouseEventKeys.includes(event as (typeof mouseEventKeys)[number])) {
                sigma.getMouseCaptor().off(event as keyof MouseCaptorEvents, eventHandler);
            }
            if (touchEventKeys.includes(event as (typeof touchEventKeys)[number])) {
                sigma.getTouchCaptor().off(event as keyof TouchCaptorEvents, eventHandler);
            }
            if (cameraEventKeys.includes(event as (typeof cameraEventKeys)[number])) {
                sigma.getCamera().off(event as keyof CameraEvents, eventHandler);
            }
        });
    }, [sigma, events]);

    useEffect(() => {
        registerEvents(events);

        return () => removeEvents();
    }, [sigma, events, registerEvents, removeEvents]);

    useEffect(() => {
        sigma.setSettings({
            nodeReducer: (node, data) => {
                const camera = sigma.getCamera();
                const highlighted = node === highlightedItem;
                const newData: Attributes = {
                    ...data,
                    highlighted,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(camera.ratio),
                };

                if (highlightedItem && !highlighted && !neighbors.has(node)) {
                    newData.label = '';
                    newData.color = '';
                }
                return newData;
            },
            edgeReducer: (edge, data) => {
                const camera = sigma.getCamera();
                const highlighted = edge === highlightedItem;

                const newData: Attributes = {
                    ...data,
                    hidden: false,
                    highlighted: highlighted,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(camera.ratio),
                };

                if (
                    highlightedItem &&
                    !highlightedItem.includes('_') &&
                    !graph
                        .extremities(edge)
                        .every((n) => n === highlightedItem || graph.areNeighbors(n, highlightedItem))
                ) {
                    newData.hidden = true;
                } else {
                    newData.hidden = false;
                }
                return newData;
            },
        });
    }, [highlightedItem, sigma]);

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

    return null;
});
