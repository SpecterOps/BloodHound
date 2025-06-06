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
import { calculateLabelOpacity, getNodeRadius } from '../utils/utils';
import drawLabel from './node-label';

export default function drawHover(
    context: CanvasRenderingContext2D,
    data: PartialButFor<NodeDisplayData & { inverseSqrtZoomRatio: number }, 'x' | 'y' | 'size' | 'label' | 'color'>,
    settings: Settings
): void {
    // Early exit to prevent node hover highlight, matching BHE and keeping consistent with lack of edge hover highlight
    if (!data.highlighted) {
        return;
    }

    const font = settings.labelFont;
    const weight = settings.labelWeight;

    const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;
    const size = settings.labelSize * inverseSqrtZoomRatio;

    context.font = `${weight} ${size}px ${font}`;
    context.fillStyle = data.highlightedBackground;

    const radius = getNodeRadius(true, inverseSqrtZoomRatio, data.size);
    const nodeLabelExists = typeof data.label === 'string';

    context.globalAlpha = calculateLabelOpacity(inverseSqrtZoomRatio);

    // Draw highlight - a circle outline
    context.beginPath();
    context.arc(data.x, data.y, radius, 0, Math.PI * 2);
    context.closePath();
    context.fill();

    // Draw label and highlight
    if (nodeLabelExists && settings.renderLabels) {
        drawLabel(context, data, settings);
    }

    context.globalAlpha = 1;
}
