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
import { Coordinates, EdgeDisplayData, NodeDisplayData, PartialButFor } from 'sigma/types';
import { bezier } from 'src/rendering/utils/bezier';
import {
    HIGHLIGHTED_LABEL_BACKGROUND_COLOR,
    HIGHLIGHTED_LABEL_FONT_COLOR,
    calculateLabelOpacity,
} from 'src/rendering/utils/utils';
import { getEdgeLabelTextLength, calculateEdgeDistanceForLabel } from 'src/ducks/graph/utils';

export default function drawEdgeLabel(
    context: CanvasRenderingContext2D,
    edgeData: PartialButFor<
        EdgeDisplayData & {
            groupPosition: number;
            groupSize: number;
            direction: number;
            control: Coordinates;
            controlInViewport: Coordinates;
            inverseSqrtZoomRatio: number;
            selected: boolean;
        },
        'label' | 'color' | 'size'
    >,
    sourceData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    targetData: PartialButFor<NodeDisplayData, 'x' | 'y' | 'size'>,
    settings: Settings
): void {
    const label = edgeData.label;
    if (!label) return;
    const font = settings.edgeLabelFont;
    const weight = settings.edgeLabelWeight;
    const color = edgeData.selected ? HIGHLIGHTED_LABEL_FONT_COLOR : edgeData.color || '#000';

    const inverseSqrtZoomRatio = edgeData.inverseSqrtZoomRatio || 1;
    const size = settings.edgeLabelSize * inverseSqrtZoomRatio;

    const edgeDistance = calculateEdgeDistanceForLabel(sourceData, targetData);
    if (!edgeDistance.distance) return;

    const textLength = getEdgeLabelTextLength(context, label, edgeDistance.distance);
    if (!textLength) return;

    let point = { x: edgeDistance.cx, y: edgeDistance.cy };

    if (edgeData.controlInViewport) {
        const sourceCoords = { x: sourceData.x, y: sourceData.y };
        const targetCoords = { x: targetData.x, y: targetData.y };
        point = bezier.getCoordinatesAlongCurve(sourceCoords, targetCoords, edgeData.controlInViewport, 0.5);
    }

    // Translate canvas to that point instead of the line between the nodes' center
    context.save();
    context.translate(point.x, point.y);
    const fadeAlphaFromZoom = calculateLabelOpacity(inverseSqrtZoomRatio);

    // Render label background
    if (edgeData.selected) {
        context.fillStyle = HIGHLIGHTED_LABEL_BACKGROUND_COLOR;
        context.globalAlpha = fadeAlphaFromZoom;
    } else {
        context.fillStyle = '#FFF';
        context.globalAlpha = fadeAlphaFromZoom * 0.8;
    }

    const xFill = (-textLength / 2) * inverseSqrtZoomRatio;
    const xPadding = 4 * inverseSqrtZoomRatio;

    const startX = xFill - xPadding;
    const startY = (edgeData.size / 2) * inverseSqrtZoomRatio - size;
    const fillWidth = textLength * 1.4 * inverseSqrtZoomRatio + 2 * xPadding;
    const fillHeight = size * 1.4;
    context.fillRect(startX, startY, fillWidth, fillHeight);

    // Text should always be completely opaque, before factoring in fade from zoom level
    context.globalAlpha = fadeAlphaFromZoom;

    // Render text
    context.fillStyle = color;
    context.font = `${weight} ${size}px ${font}`;
    context.fillText(label, xFill, (edgeData.size / 2) * inverseSqrtZoomRatio);

    context.restore();
}
