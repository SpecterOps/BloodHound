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

import { SigmaNodeEventPayload } from 'sigma/sigma';
import { MouseCoords, NodeDisplayData, PartialButFor } from 'sigma/types';

/** Threshold of graph zoom before labels fade out */
export const STARTING_ZOOM_FADE_RATIO = 0.65;
export const ENDING_ZOOM_FADE_RATIO = 0.5;

/** Padding displayed around label for node or edge */
export const LABEL_PADDING = 3;

/** Padding displayed around a node */
export const NODE_PADDING = 2;

/**
 * While edge labels are drawn with a custom renderer, the mouse target for capturing clicks is
 * handled by a separate process, making its position slightly differ. This offset is applied
 * to align a centered edge label with the invisible mouse target. Selected by trial and error.
 */
export const EDGE_MIDDLE_ALIGN_OFFSET = 8;

/** Edge item `type` string values. Items of this `type` are edges and will be middle aligned. */
export const EDGE_TYPES = ['curved', 'arrow'];

/** Type for node data passed to custom render programs */
export type GraphItemData = PartialButFor<
    NodeDisplayData & { inverseSqrtZoomRatio: number },
    'x' | 'y' | 'size' | 'label' | 'color'
>;

/**
 * Calculate a labels opacity relative to the current graph's zoom (to mimick ReGraph)
 *
 * @param inverseSqrtZoomRatio modified form of current graph's zoom value
 * @returns label opacity for given zoom
 */
export const calculateLabelOpacity = (inverseSqrtZoomRatio: number): number => {
    if (inverseSqrtZoomRatio >= STARTING_ZOOM_FADE_RATIO) {
        return 1;
    }

    if (inverseSqrtZoomRatio < STARTING_ZOOM_FADE_RATIO && inverseSqrtZoomRatio > ENDING_ZOOM_FADE_RATIO) {
        return (inverseSqrtZoomRatio - ENDING_ZOOM_FADE_RATIO) / (STARTING_ZOOM_FADE_RATIO - ENDING_ZOOM_FADE_RATIO);
    }

    return 0;
};

export interface LabelBoundsParams {
    /** One over the current zoom value */
    inverseSqrtZoomRatio: number;

    /** Text drawn to a canvas as a label */
    label: string;

    /** [Optional] Starting position of label */
    position: {
        x: number;
        y: number;
    };

    /** Size of node label will apply to */
    size: number;
}

const DEFAULT_PARAMS: LabelBoundsParams = {
    inverseSqrtZoomRatio: 1,
    label: '',
    position: { x: 0, y: 0 },
    size: 1,
};

/**
 * Use a context to calculate the bounds for text drawn to a canvas
 *
 * @param context the canvas's drawing context, used to calculate the bounds
 * @param params settings which influence label's bounds
 * @returns tuple contain label bounds as [x, y, width, height]
 */
export const getLabelBoundsFromContext = (
    context: CanvasRenderingContext2D,
    params: LabelBoundsParams = DEFAULT_PARAMS
): [x: number, y: number, width: number, height: number] => {
    const labelBounds = context.measureText(params.label);
    const labelWidth = labelBounds.width;
    // Add the space above the text baseline plus the space below it
    const labelHeight = labelBounds.actualBoundingBoxAscent + labelBounds.actualBoundingBoxDescent;
    const labelOffsetX = (params.size + LABEL_PADDING / 2) * params.inverseSqrtZoomRatio;
    const labelOffsetY = labelHeight / 2;

    return [
        // Subtracting 1 pixel prevents a gap between node and label
        params.position.x + labelOffsetX - 1,
        params.position.y - labelOffsetY - LABEL_PADDING,
        labelWidth + 2 * LABEL_PADDING,
        labelHeight + 2 * LABEL_PADDING,
    ];
};

/**
 * Calculate the radius of a node by the given size
 *
 * @param highlighted whether node is highlighted or not (highlights add to radius)
 * @param inverseSqrtZoomRatio modified form of current graph's zoom value
 * @param size size of node
 * @returns calculated radius of a node with given size
 */
export const getNodeRadius = (highlighted: boolean, inverseSqrtZoomRatio: number, size?: number): number => {
    let radius = 1;

    if (!size) return radius;

    if (highlighted) radius = (size + NODE_PADDING + 2) * inverseSqrtZoomRatio;
    else radius = size * inverseSqrtZoomRatio;

    return radius;
};

/**
 * Reusable method to prevent defaults for mouse move, right click, and double click
 *
 * @param event Sigma or mouse event object used to cancel defaults
 */
export const preventAllDefaults = (event: SigmaNodeEventPayload | MouseCoords) => {
    event.preventSigmaDefault();

    // Prevent events for MouseCoords type
    if ('original' in event && event.original instanceof MouseEvent) {
        event.original.preventDefault();
        event.original.stopPropagation();
    }
};
