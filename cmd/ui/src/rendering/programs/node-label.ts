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
    if (!settings.labelColor.color) return;

    const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;

    const size = settings.labelSize * inverseSqrtZoomRatio,
        font = settings.labelFont,
        weight = settings.labelWeight;

    const isEdge = EDGE_TYPES.includes(data.type ?? '');
    const showName = (isEdge || data.showNodeNames !== false) && !!data.label;
    const showKind = !isEdge && data.showNodeKinds !== false && !!data.nodeKind;

    // Nothing to draw for this node
    if (!showName && !showKind) return;

    context.globalAlpha = calculateLabelOpacity(inverseSqrtZoomRatio);

    const kindSize = size * 0.85;

    // Measure the label text to draw (name line)
    const nameText = showName ? data.label : undefined;
    // Use correct font size for measurement: full size for name, reduced for kind-only
    context.font = showName ? `${weight} ${size}px ${font}` : `${weight} ${kindSize}px ${font}`;

    const labelParams: LabelBoundsParams = {
        inverseSqrtZoomRatio,
        label: nameText ?? data.nodeKind ?? '',
        position: data,
        size: data.size ?? 0, // fallback prevents NaN
    };

    const labelbounds = getLabelBoundsFromContext(context, labelParams);

    // This method is reused to draw edge labels as well, however, edge labels
    // must be offset to middle to align with unrendered edge label mouse target
    if (isEdge) {
        labelbounds[0] = labelbounds[0] - labelbounds[2] / 2 - EDGE_MIDDLE_ALIGN_OFFSET;
    }

    // Expand background rect to include kind line if present
    let kindLineHeight = 0;

    if (showKind && data.nodeKind) {
        context.save();
        context.font = `${weight} ${kindSize}px ${font}`;
        const kindMetrics = context.measureText(data.nodeKind);
        kindLineHeight = kindMetrics.actualBoundingBoxAscent + kindMetrics.actualBoundingBoxDescent + LABEL_PADDING;
        // Widen background if kind text is wider than the label
        const kindWidth = kindMetrics.width + 2 * LABEL_PADDING;
        if (kindWidth > labelbounds[2]) {
            labelbounds[2] = kindWidth;
        }
        if (showName) {
            labelbounds[3] += kindLineHeight;
        }
        context.restore();
    }

    context.fillStyle = data.highlighted ? data.highlightedBackground : data.backgroundColor;
    context.fillRect(...labelbounds);

    context.fillStyle = data.highlighted ? data.highlightedText : settings.labelColor.color;

    // Draw node name
    if (showName) {
        context.font = `${weight} ${size}px ${font}`;
        context.fillText(data.label, labelbounds[0] + LABEL_PADDING, labelbounds[1] + labelbounds[3] - kindLineHeight - LABEL_PADDING);
    }

    // Draw node kind on a new line below the label (or as the only line)
    if (showKind && data.nodeKind) {
        context.font = `${weight} ${kindSize}px ${font}`;
        if (showName) {
            context.globalAlpha = context.globalAlpha * 0.7;
        }
        context.fillText(data.nodeKind, labelbounds[0] + LABEL_PADDING, labelbounds[1] + labelbounds[3] - LABEL_PADDING);
        if (showName) {
            context.globalAlpha = context.globalAlpha / 0.7;
        }
    }

    context.globalAlpha = 1;
}
