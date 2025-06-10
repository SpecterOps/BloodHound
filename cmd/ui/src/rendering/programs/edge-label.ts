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

/**
 * Sigma.js Canvas Renderer Edge Label Component
 * =============================================
 *
 * Function used by the canvas renderer to display a single edge's label.
 * @module
 */
import { Attributes } from 'graphology-types';
import { Settings } from 'sigma/settings';
import { Coordinates, NodeDisplayData, PartialButFor } from 'sigma/types';
import { EdgeDistanceProperties, calculateEdgeDistanceForLabel } from 'src/ducks/graph/utils';
import { bezier } from 'src/rendering/utils/bezier';
import { GraphItemData } from '../utils/utils';
import { getControlPointsFromGroupSize } from './edge.self';
import drawLabel from './node-label';

const PADDING_SCALAR = 5;

const getXPadding = (inverseSqrtZoomRatio: number) => {
    return PADDING_SCALAR * inverseSqrtZoomRatio;
};

export const getBackgroundBoundInfo = (
    inverseSqrtZoomRatio: number,
    textLength: number,
    edgeSize: number,
    edgeLabelSize: number
) => {
    const xPadding = getXPadding(inverseSqrtZoomRatio);
    const deltaX = -textLength / 2 - xPadding;
    const deltaY = (edgeSize / 2) * inverseSqrtZoomRatio - edgeLabelSize * inverseSqrtZoomRatio;
    const width = textLength + 2 * xPadding;
    const height = edgeLabelSize * inverseSqrtZoomRatio * 1.4;

    return { deltaX: deltaX, deltaY: deltaY, width: width, height: height };
};

const getCurvedEdgeStartingPoint = (
    sourceCoords: Coordinates,
    targetCoords: Coordinates,
    control: Coordinates
): Coordinates => {
    return bezier.getCoordinatesAlongQuadraticBezier(sourceCoords, targetCoords, control, 0.5);
};

export const getSelfEdgeStartingPoint = (
    c1: Coordinates,
    c2: Coordinates,
    c3: Coordinates,
    c4: Coordinates
): Coordinates => {
    return bezier.getCoordinatesAlongCubicBezier(c1, c2, c3, c4, 0.5);
};

const getStartingPoint = (
    edgeData: Attributes,
    sourceData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    targetData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    edgeDistance: EdgeDistanceProperties
): Coordinates => {
    if (edgeData.controlInViewport) {
        return getCurvedEdgeStartingPoint(sourceData, targetData, edgeData.controlInViewport);
    } else if (edgeData.type === 'self') {
        const inverseSqrtZoomRatio = edgeData.inverseSqrtZoomRatio || 1;
        const radius = bezier.getLineLength(
            { x: 0, y: 0 },
            {
                x: sourceData.size * Math.pow(inverseSqrtZoomRatio, 2),
                y: sourceData.size * Math.pow(inverseSqrtZoomRatio, 2),
            }
        );

        const { control2, control3 } = getControlPointsFromGroupSize(
            edgeData.groupPosition,
            radius * 3,
            sourceData,
            false,
            true
        );

        return getSelfEdgeStartingPoint(sourceData, control2, control3, sourceData);
    } else {
        return { x: edgeDistance.cx, y: edgeDistance.cy };
    }
};

export default function draw(
    context: CanvasRenderingContext2D,
    edgeData: Attributes,
    sourceData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    targetData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    settings: Settings
): void {
    const label = edgeData.label;

    if (!label) return;

    const edgeDistance = calculateEdgeDistanceForLabel(sourceData, targetData);
    const startingPoint = getStartingPoint(edgeData, sourceData, targetData, edgeDistance);

    // Overwrite edge positions with custom-calculated position since non-default edge renderer is used
    const newData = { ...edgeData, ...startingPoint } as GraphItemData;

    drawLabel(context, newData, settings);
}
