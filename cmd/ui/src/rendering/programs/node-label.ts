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

import { truncateText } from 'bh-shared-ui';
import { Settings } from 'sigma/settings';
import {
    EDGE_MIDDLE_ALIGN_OFFSET,
    EDGE_TYPES,
    GraphItemData,
    LABEL_PADDING,
    LabelBoundsParams,
    calculateLabelOpacity,
    getLabelBoundsFromContext,
    getNodeLabelBoundsBelowFromContext,
} from '../utils/utils';

const SUBLABEL_FONT_RATIO = 0.85;

export default function drawLabel(context: CanvasRenderingContext2D, data: GraphItemData, settings: Settings): void {
    if (!data.label || !settings.labelColor.color) return;

    const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;

    const size = settings.labelSize * inverseSqrtZoomRatio,
        font = settings.labelFont,
        weight = settings.labelWeight;

    const fillBackground = data.highlighted ? data.highlightedBackground : data.backgroundColor;
    const fillText = data.highlighted ? data.highlightedText : settings.labelColor.color;

    context.globalAlpha = calculateLabelOpacity(inverseSqrtZoomRatio);

    // Edge labels: use original right-of-node positioning with center-alignment offset
    if (EDGE_TYPES.includes(data.type ?? '')) {
        context.font = `${weight} ${size}px ${font}`;

        const labelParams: LabelBoundsParams = {
            inverseSqrtZoomRatio,
            label: data.label,
            position: data,
            size: data.size ?? 0,
        };

        const labelbounds = getLabelBoundsFromContext(context, labelParams);
        labelbounds[0] = labelbounds[0] - labelbounds[2] / 2 - EDGE_MIDDLE_ALIGN_OFFSET;

        context.fillStyle = fillBackground;
        context.fillRect(...labelbounds);

        context.fillStyle = fillText;
        context.fillText(data.label, labelbounds[0] + LABEL_PADDING, labelbounds[1] + labelbounds[3] - LABEL_PADDING);

        context.globalAlpha = 1;
        return;
    }

    // Node labels: truncated primary label + truncated sublabel, centered below the node
    // When the node is highlighted (selected), show the full label instead of truncating

    const labelTextRendered = (labelText: string) => {
        return data.highlighted ? labelText : truncateText(labelText) ?? labelText;
    };

    const primaryLabel = labelTextRendered(data.label);
    const nodeSize = data.size ?? 0;

    context.font = `bold ${size}px ${font}`;

    const primaryParams: LabelBoundsParams = {
        inverseSqrtZoomRatio,
        label: primaryLabel,
        position: data,
        size: nodeSize,
    };

    const primaryBounds = getNodeLabelBoundsBelowFromContext(context, primaryParams);

    context.fillStyle = fillBackground;
    context.fillRect(...primaryBounds);

    context.fillStyle = fillText;
    context.fillText(
        primaryLabel,
        primaryBounds[0] + LABEL_PADDING,
        primaryBounds[1] + primaryBounds[3] - LABEL_PADDING
    );

    if (data.source && data.kind) {
        const nodeSource = labelTextRendered(data.source);
        const nodeKind = labelTextRendered(data.kind);
        const sublabelText = `${nodeSource} | ${nodeKind}`;
        const sublabelSize = size * SUBLABEL_FONT_RATIO;
        context.font = `${weight} ${sublabelSize}px ${font}`;

        const sublabelParams: LabelBoundsParams = {
            inverseSqrtZoomRatio,
            label: sublabelText,
            position: data,
            size: nodeSize,
        };

        const sublabelBounds = getNodeLabelBoundsBelowFromContext(context, sublabelParams, primaryBounds[3]);

        context.fillStyle = fillBackground;
        context.fillRect(...sublabelBounds);

        context.fillStyle = fillText;
        context.fillText(
            sublabelText,
            sublabelBounds[0] + LABEL_PADDING,
            sublabelBounds[1] + sublabelBounds[3] - LABEL_PADDING
        );
    }

    context.globalAlpha = 1;
}
