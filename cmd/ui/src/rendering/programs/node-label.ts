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
import {
    EDGE_MIDDLE_ALIGN_OFFSET,
    EDGE_TYPES,
    GraphItemData,
    LABEL_PADDING,
    LabelBoundsParams,
    calculateLabelOpacity,
    getLabelBoundsFromContext,
} from '../utils/utils';

export default function drawLabel(context: CanvasRenderingContext2D, data: GraphItemData, settings: Settings): void {
    if (!data.label || !settings.labelColor.color) return;

    const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;

    const size = settings.labelSize * inverseSqrtZoomRatio,
        font = settings.labelFont,
        weight = settings.labelWeight;

    context.globalAlpha = calculateLabelOpacity(inverseSqrtZoomRatio);
    context.font = `${weight} ${size}px ${font}`;

    const labelParams: LabelBoundsParams = {
        inverseSqrtZoomRatio,
        label: data.label,
        position: data,
        size: data.size ?? 0, // fallback prevents NaN
    };

    const labelbounds = getLabelBoundsFromContext(context, labelParams);

    // This method is reused to draw edge labels as well, however, edge labels
    // must be offset to middle to align with unrendered edge label mouse target
    if (EDGE_TYPES.includes(data.type ?? '')) {
        labelbounds[0] = labelbounds[0] - labelbounds[2] / 2 - EDGE_MIDDLE_ALIGN_OFFSET;
    }

    context.fillStyle = data.highlighted ? data.highlightedBackground : data.backgroundColor;
    context.fillRect(...labelbounds);

    context.fillStyle = data.highlighted ? data.highlightedText : settings.labelColor.color;
    context.fillText(data.label, labelbounds[0] + LABEL_PADDING, data.y + size / 3);

    context.globalAlpha = 1;
}
