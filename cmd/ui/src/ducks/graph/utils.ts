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

import AbstractGraph, { Attributes } from 'graphology-types';
import Sigma from 'sigma';
import { CameraState, Coordinates, Extent, NodeDisplayData } from 'sigma/types';
import { graphExtent } from 'sigma/utils';

export const isLink = (item: any): boolean => {
    return item.id1 !== undefined;
};

export const isNode = (item: any): boolean => {
    return !isLink(item);
};

export const GROUP_SPREAD = 0.06;
export const MIN_CAMERA_RATIO = 0.5;
export const MAX_CAMERA_RATIO = 15;

export const getEdgeDataFromKey = (edgeKey: string): null | { source: string; target: string; label: string } => {
    const keyChunks = edgeKey.split('_');
    const source = keyChunks.shift();
    const target = keyChunks.pop();
    const label = keyChunks.join('_');

    if (!source) {
        console.warn('Invalid edge key. No source found');
        return null;
    }
    if (!target) {
        console.warn('Invalid edge key. No target found');
        return null;
    }
    if (!label) {
        console.warn('Invalid edge key. Unable to evaluate label');
        return null;
    }

    return { source: source, target: target, label: label };
};

export const getEdgeSourceAndTargetDisplayData = (
    source: string,
    target: string,
    sigma: Sigma
): { source: NodeDisplayData; target: NodeDisplayData } | null => {
    const sourceDisplayData = sigma.getNodeDisplayData(source);
    const targetDisplayData = sigma.getNodeDisplayData(target);

    if (!sourceDisplayData || !targetDisplayData) return null;
    return { source: sourceDisplayData, target: targetDisplayData };
};

export type LinkedNode = {
    size: number;
    x: number;
    y: number;
};

export const calculateEdgeDistanceForLabel = (
    source: LinkedNode,
    target: LinkedNode
): { distance: number; cx: number; cy: number } => {
    // Computing positions without considering nodes sizes:
    const sSize = source.size;
    const tSize = target.size;
    let sx = source.x;
    let sy = source.y;
    let tx = target.x;
    let ty = target.y;
    let cx = (sx + tx) / 2;
    let cy = (sy + ty) / 2;
    let dx = tx - sx;
    let dy = ty - sy;
    let d = Math.sqrt(dx * dx + dy * dy);

    //Don't render label since nodes sizes engulf edge
    if (d < sSize + tSize) return { distance: 0, cx: cx, cy: cy };

    // Adding nodes sizes:
    sx += (dx * sSize) / d;
    sy += (dy * sSize) / d;
    tx -= (dx * tSize) / d;
    ty -= (dy * tSize) / d;
    cx = (sx + tx) / 2;
    cy = (sy + ty) / 2;
    dx = tx - sx;
    dy = ty - sy;
    d = Math.sqrt(dx * dx + dy * dy);
    return { distance: d, cx: cx, cy: cy };
};

export const getEdgeLabelTextLength = (
    context: CanvasRenderingContext2D,
    edgeLabel: string,
    edgeDistance: number
): number => {
    let label = edgeLabel;
    let textLength = context.measureText(label).width;

    if (textLength > edgeDistance) {
        const ellipsis = '…';
        label = label + ellipsis;
        textLength = context.measureText(label).width;

        while (textLength > edgeDistance && label.length > 1) {
            label = label.slice(0, -2) + ellipsis;
            textLength = context.measureText(label).width;
        }

        if (label.length < 4) return 0;
    }
    return textLength;
};

const graphToFramedGraph = (
    sigma: Sigma<AbstractGraph<Attributes, Attributes, Attributes>>,
    coordinates: Coordinates
): Coordinates => {
    return sigma.viewportToFramedGraph(sigma.graphToViewport(coordinates));
};

const graphGetMidpoint = (graphExtent: { x: Extent; y: Extent }): Coordinates => {
    const { x: xExtent, y: yExtent } = graphExtent;
    const [xMin, xMax] = xExtent;
    const [yMin, yMax] = yExtent;
    const midX = (xMax + xMin) / 2;
    const midY = (yMax + yMin) / 2;

    return { x: midX, y: midY };
};

const centerCameraOnGraph = (sigma: Sigma<AbstractGraph<Attributes, Attributes, Attributes>>): void => {
    const graph = sigma.getGraph();
    const extent = graphExtent(graph);
    const camera = sigma.getCamera();
    const currentRatio = camera.ratio;
    const newCameraState: CameraState = camera.getState();

    // determine if we need to zoom the camera out by comparing the left edge of the graph and the left edge of the current frame
    const { x: xGraphMin, y: yGraphMin } = graphToFramedGraph(sigma, { x: extent.x[0], y: extent.y[0] });
    const { x: xGraphMax, y: yGraphMax } = graphToFramedGraph(sigma, { x: extent.x[1], y: extent.y[1] });
    const { x1: xFrameMin, x2: xFrameMax, y1: yFrameMin, y2: yFrameMax, height: frameHeight } = sigma.viewRectangle();
    const graphHeight = yGraphMax - yGraphMin;
    const graphWidth = xGraphMax - xGraphMin;
    const frameWidth = xFrameMax - xFrameMin;

    // the graph lies beyond the edge of the current frame so zoom out by some factor
    if (xGraphMin < xFrameMin || xGraphMax > xFrameMax || yGraphMin < yFrameMin || yGraphMax > yFrameMax) {
        const frameHypotenuse = Math.hypot(frameWidth, frameHeight);
        const graphHypotenuse = Math.hypot(graphWidth, graphHeight);

        let ratio =
            currentRatio *
            Math.max(graphHypotenuse / frameHypotenuse, graphHeight / frameHeight, graphWidth / frameWidth) *
            1.1;

        if (ratio < MIN_CAMERA_RATIO) ratio = MIN_CAMERA_RATIO;
        if (ratio > MAX_CAMERA_RATIO) ratio = MAX_CAMERA_RATIO;

        newCameraState.ratio = ratio;
        newCameraState.x = 0.5;
        newCameraState.y = 0.5;
    }

    camera.animate(
        newCameraState,
        {
            easing: 'quadraticOut',
            duration: 250,
        },
        () => {
            sigma.scheduleRefresh();
        }
    );
};

const calculateNewBBox = (sigma: Sigma): { x: Extent; y: Extent } => {
    const originalExtent = graphExtent(sigma.getGraph());
    const midpoint = graphGetMidpoint(originalExtent);
    const { width, height } = sigma.getDimensions();

    return {
        x: [midpoint.x - width / 2, midpoint.x + width / 2],
        y: [midpoint.y - height / 2, midpoint.y + height / 2],
    };
};

export const resetCamera = (sigma: Sigma<AbstractGraph<Attributes, Attributes, Attributes>>): void => {
    sigma.setCustomBBox(calculateNewBBox(sigma));
    setTimeout(() => {
        centerCameraOnGraph(sigma);
    }, 250);
};
