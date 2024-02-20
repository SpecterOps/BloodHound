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
import { Settings } from 'sigma/settings';
import { Coordinates, NodeDisplayData, PartialButFor } from 'sigma/types';
import { bezier } from 'src/rendering/utils/bezier';
import {
    HIGHLIGHTED_LABEL_BACKGROUND_COLOR,
    HIGHLIGHTED_LABEL_FONT_COLOR,
    calculateLabelOpacity,
} from 'src/rendering/utils/utils';
import { getEdgeLabelTextLength, calculateEdgeDistanceForLabel, EdgeDistanceProperties } from 'src/ducks/graph/utils';
import { Attributes } from 'graphology-types';
import { getControlPointsFromGroupSize } from './edge.self';

const PADDING_SCALAR = 5;

const getXPadding = (inverseSqrtZoomRatio: number) => {
    return PADDING_SCALAR * inverseSqrtZoomRatio;
};

const drawBackground = (
    context: CanvasRenderingContext2D,
    edgeData: Attributes,
    settings: Settings,
    fadeAlphaFromZoom: number,
    textLength: number
) => {
    const inverseSqrtZoomRatio = edgeData.inverseSqrtZoomRatio || 1;
    if (edgeData.selected) {
        context.fillStyle = HIGHLIGHTED_LABEL_BACKGROUND_COLOR;
        context.globalAlpha = fadeAlphaFromZoom;
    } else {
        context.fillStyle = '#FFF';
        context.globalAlpha = fadeAlphaFromZoom * 0.8;
    }

    const { deltaX, deltaY, width, height } = getBackgroundBoundInfo(
        inverseSqrtZoomRatio,
        textLength,
        edgeData,
        settings.edgeLabelSize
    );

    context.fillRect(deltaX, deltaY, width, height);
};

export const getBackgroundBoundInfo = (
    inverseSqrtZoomRatio: number,
    textLength: number,
    edgeData: Attributes,
    edgeLabelSize: number
) => {
    const xPadding = getXPadding(inverseSqrtZoomRatio);
    const deltaX = -textLength / 2 - xPadding;
    const deltaY = (edgeData.size / 2) * inverseSqrtZoomRatio - edgeLabelSize * inverseSqrtZoomRatio;
    const width = textLength + 2 * xPadding;
    const height = edgeLabelSize * inverseSqrtZoomRatio * 1.4;

    return { deltaX: deltaX, deltaY: deltaY, width: width, height: height };
};

const drawText = (
    context: CanvasRenderingContext2D,
    edgeData: Attributes,
    fadeAlphaFromZoom: number,
    textLength: number
) => {
    const label = edgeData.label;
    if (!label) return;

    // Text should always be completely opaque, before factoring in fade from zoom level
    context.globalAlpha = fadeAlphaFromZoom;
    context.fillStyle = edgeData.selected ? HIGHLIGHTED_LABEL_FONT_COLOR : edgeData.color || '#000';

    context.fillText(label, -textLength / 2, (edgeData.size / 2) * (edgeData.inverseSqrtZoomRatio || 1));
};

const setContextFont = (context: CanvasRenderingContext2D, edgeData: Attributes, settings: Settings) => {
    const font = settings.edgeLabelFont;
    const weight = settings.edgeLabelWeight;
    const size = settings.edgeLabelSize * (edgeData.inverseSqrtZoomRatio || 1);

    context.font = `${weight} ${size}px ${font}`;
};

const getCurvedEdgeStartingPoint = (
    sourceCoords: Coordinates,
    targetCoords: Coordinates,
    control: Coordinates
): Coordinates => {
    return bezier.getCoordinatesAlongQuadraticBezier(sourceCoords, targetCoords, control, 0.5);
};

export const getSelfEdgeStartingPoint = (data: Attributes, sourceCoords: Coordinates, radius: number): Coordinates => {
    const control1 = { x: sourceCoords.x, y: sourceCoords.y };
    const { control2, control3 } = getControlPointsFromGroupSize(
        data.groupPosition,
        radius * 3,
        sourceCoords,
        false,
        true
    );
    const control4 = control1;

    return bezier.getCoordinatesAlongCubicBezier(control1, control2, control3, control4, 0.5);
};

const getStartingPoint = (
    edgeData: Attributes,
    sourceData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    targetData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    edgeDistance: EdgeDistanceProperties
): Coordinates => {
    const inverseSqrtZoomRatio = edgeData.inverseSqrtZoomRatio || 1;

    if (edgeData.controlInViewport) {
        const sourceCoords = { x: sourceData.x, y: sourceData.y };
        const targetCoords = { x: targetData.x, y: targetData.y };
        return getCurvedEdgeStartingPoint(sourceCoords, targetCoords, edgeData.controlInViewport);
    } else if (edgeData.type === 'self') {
        const sourceCoords = { x: sourceData.x, y: sourceData.y };
        const radius = bezier.getLineLength(
            { x: 0, y: 0 },
            {
                x: sourceData.size * inverseSqrtZoomRatio * inverseSqrtZoomRatio,
                y: sourceData.size * inverseSqrtZoomRatio * inverseSqrtZoomRatio,
            }
        );

        return getSelfEdgeStartingPoint(edgeData, sourceCoords, radius);
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
    const inverseSqrtZoomRatio = edgeData.inverseSqrtZoomRatio || 1;
    setContextFont(context, edgeData, settings);

    const edgeDistance = calculateEdgeDistanceForLabel(sourceData, targetData);
    const textLength = getEdgeLabelTextLength(context, label, edgeDistance.distance);
    if (!textLength) return;

    const startingPoint = getStartingPoint(edgeData, sourceData, targetData, edgeDistance);

    context.save();
    context.translate(startingPoint.x, startingPoint.y);

    const fadeAlphaFromZoom = calculateLabelOpacity(inverseSqrtZoomRatio);

    drawBackground(context, edgeData, settings, fadeAlphaFromZoom, textLength);
    drawText(context, edgeData, fadeAlphaFromZoom, textLength);

    context.restore();
}
