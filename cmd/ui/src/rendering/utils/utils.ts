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
import { Coordinates } from 'sigma/types';
import { bezier } from './bezier';

export const HIGHLIGHTED_LABEL_BACKGROUND_COLOR = '#FF6D66';
export const HIGHLIGHTED_LABEL_FONT_COLOR = '#FFF';

export const STARTING_ZOOM_FADE_RATIO = 0.5;
export const ENDING_ZOOM_FADE_RATIO = 0.3;

export const calculateLabelOpacity = (inverseSqrtZoomRatio: number): number => {
    if (inverseSqrtZoomRatio >= STARTING_ZOOM_FADE_RATIO) {
        return 1;
    }

    if (inverseSqrtZoomRatio < STARTING_ZOOM_FADE_RATIO && inverseSqrtZoomRatio > ENDING_ZOOM_FADE_RATIO) {
        return (inverseSqrtZoomRatio - ENDING_ZOOM_FADE_RATIO) / (STARTING_ZOOM_FADE_RATIO - ENDING_ZOOM_FADE_RATIO);
    }

    return 0;
};

export const getNodeRadius = (highlighted: boolean, inverseSqrtZoomRatio: number, size?: number): number => {
    const PADDING = 2;
    let radius = 1;

    if (!size) return radius;

    if (highlighted) radius = size * inverseSqrtZoomRatio + (PADDING + 2) * inverseSqrtZoomRatio;
    else radius = size * inverseSqrtZoomRatio;

    return radius;
};

//This was sourced from https://pomax.github.io/bezierinfo/#circleintersection where the steps are explained in detail
export const getCurveCircleIntersection = (
    LUT: Coordinates[],
    circleCenter: Coordinates,
    radius: number,
    deltaEpsilon = 0.01
) => {
    const values = [];
    const maxIterations = 25;
    let start = 0,
        count = 0;

    while (++count < maxIterations) {
        if (!LUT[start - 2] || !LUT[start - 1]) {
            start = start + 1;
            continue;
        }
        const closest = findClosest(
            circleCenter.x,
            circleCenter.y,
            LUT.slice(start),
            bezier.getLineLength(LUT[start - 2], circleCenter),
            bezier.getLineLength(LUT[start - 1], circleCenter),
            radius,
            deltaEpsilon
        );
        const i = start + closest;

        if (i < start) break;
        if (i > 0 && i === start) break;
        values.push(i);
        start = i + 2; // We need to add 2, because findClosest already tells us that i+1 cannot be a viable candidate.
    }
    return values;
};

const findClosest = (
    x: number,
    y: number,
    LUT: Coordinates[],
    pd2: number,
    pd1: number,
    radius: number,
    distanceEpsilon: number
): number => {
    let distance = Infinity,
        prevDistance2 = pd2 || distance,
        prevDistance1 = pd1 || distance,
        i = -1;

    for (let index = 0; index < LUT.length; index++) {
        const pointOnCurve = LUT[index];
        const pointDistance = Math.abs(
            bezier.getLineLength({ x: x, y: y }, { x: pointOnCurve.x, y: pointOnCurve.y }) - radius
        );

        // Realistically, there's only going to be an intersection if
        // the distance to the circle center is already approximately
        // the circle's radius.
        if (prevDistance1 < distanceEpsilon && prevDistance2 > prevDistance1 && prevDistance1 < pointDistance) {
            i = index - 1;
            break;
        }

        if (pointDistance < distance) {
            distance = pointDistance;
        }

        prevDistance2 = prevDistance1;
        prevDistance1 = pointDistance;
    }

    return i;
};
