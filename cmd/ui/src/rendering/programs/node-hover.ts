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

import { Settings } from 'sigma/settings';
import { NodeDisplayData, PartialButFor } from 'sigma/types';
import drawLabel from 'src/rendering/programs/node-label';
import {
    HIGHLIGHTED_LABEL_BACKGROUND_COLOR,
    HIGHLIGHTED_LABEL_FONT_COLOR,
    calculateLabelOpacity,
    getNodeRadius,
} from 'src/rendering/utils/utils';

export default function drawHover(
    context: CanvasRenderingContext2D,
    data: PartialButFor<NodeDisplayData & { inverseSqrtZoomRatio: number }, 'x' | 'y' | 'size' | 'label' | 'color'>,
    settings: Settings
): void {
    const font = settings.labelFont,
        weight = settings.labelWeight;

    const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;
    const size = settings.labelSize * inverseSqrtZoomRatio;

    context.font = `${weight} ${size}px ${font}`;
    context.fillStyle = HIGHLIGHTED_LABEL_BACKGROUND_COLOR;

    const PADDING = 2;
    const radius = getNodeRadius(true, inverseSqrtZoomRatio, data.size);

    const nodeLabelExists = typeof data.label === 'string';

    context.globalAlpha = calculateLabelOpacity(inverseSqrtZoomRatio);

    if (nodeLabelExists) {
        // Determine size of label's bounding rectangle
        const textWidth = context.measureText(data.label || '').width,
            boxWidth = Math.round(textWidth + 5),
            boxHeight = Math.round(size + 2 * PADDING);

        const angleRadian = Math.asin(boxHeight / 2 / radius);
        const xDeltaCoord = Math.sqrt(Math.abs(Math.pow(radius, 2) - Math.pow(boxHeight / 2, 2)));

        // Draw a circle around the node with a label background along the positive x-axis
        context.beginPath();
        context.moveTo(data.x + xDeltaCoord, data.y + boxHeight / 2);
        context.lineTo(data.x + radius + boxWidth, data.y + boxHeight / 2);
        context.lineTo(data.x + radius + boxWidth, data.y - boxHeight / 2);
        context.lineTo(data.x + xDeltaCoord, data.y - boxHeight / 2);
        context.arc(data.x, data.y, radius, angleRadian, -angleRadian);
        context.closePath();
        context.fill();
    } else {
        // Draw a circle with no label background
        context.beginPath();
        context.arc(data.x, data.y, radius, 0, Math.PI * 2);
        context.closePath();
        context.fill();
    }

    context.globalAlpha = 1;

    // toggle the default text color before/after this drawLabel call so that the color is only
    // changed for hovered nodes
    settings.labelColor.color = HIGHLIGHTED_LABEL_FONT_COLOR;
    drawLabel(context, data, settings);
    settings.labelColor.color = '#000';
}
