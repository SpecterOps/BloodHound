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
import { EdgeDirection } from 'src/utils';
import { GROUP_SPREAD } from 'src/ducks/graph/utils';

// Collection of helper functions for working with 2D Coordinates
export const bezier = {
    getNormals(start: Coordinates, end: Coordinates): Coordinates {
        const dx = end.x - start.x;
        const dy = end.y - start.y;

        const dist = 1 / bezier.getLineLength(start, end);

        return {
            x: dx * dist,
            y: dy * dist,
        };
    },

    // Gets a point along the perpindicular line that intersects the target line segment at a certain offset
    getOffsetVertex: (start: Coordinates, end: Coordinates, intersect: Coordinates, offset: number): Coordinates => {
        const normal = bezier.getNormals(start, end);

        return {
            x: intersect.x + offset * normal.y,
            y: intersect.y - offset * normal.x,
        };
    },

    // Use parametric formula to calculate point along quadratic bezier curve based on 0 < t < 1
    getCoordinatesAlongCurve: (start: Coordinates, end: Coordinates, control: Coordinates, t: number): Coordinates => {
        return {
            x: (1 - t) * (1 - t) * start.x + 2 * (1 - t) * t * control.x + t * t * end.x,
            y: (1 - t) * (1 - t) * start.y + 2 * (1 - t) * t * control.y + t * t * end.y,
        };
    },

    getMidpoint: (start: Coordinates, end: Coordinates): Coordinates => {
        return {
            x: (start.x + end.x) / 2,
            y: (start.y + end.y) / 2,
        };
    },

    getLineLength: (start: Coordinates, end: Coordinates): number => {
        return Math.sqrt((end.x - start.x) * (end.x - start.x) + (end.y - start.y) * (end.y - start.y));
    },

    calculateCurveHeight: (
        size: number | undefined,
        position: number | undefined,
        direction?: EdgeDirection
    ): number => {
        if (position === undefined || size === undefined) return 0;

        const isEvenSized = size % 2 === 0;
        const isEvenPosition = position % 2 === 0;

        // First edge of odd-totaled group should be flat
        if (!isEvenSized && position === 0) return 0;

        // Values in odd-numbered groups are offset because the first edge is a line
        if (!isEvenSized) position -= 1;

        let height = Math.floor(position / 2) * GROUP_SPREAD + GROUP_SPREAD;

        if (isEvenSized) height -= GROUP_SPREAD / 2;

        // Alternate direction of curve by flipping the sign
        if (isEvenPosition) height *= -1;

        // Handle groups with bidirectional edges by flipping the curve direction for edges pointing "backwards"
        if (direction) height *= direction;

        return height;
    },

    getControlAtMidpoint: (
        height: number,
        sourceCoordinates: Coordinates,
        targetCoordinates: Coordinates
    ): Coordinates => {
        const midpoint = bezier.getMidpoint(sourceCoordinates, targetCoordinates);
        return bezier.getOffsetVertex(sourceCoordinates, targetCoordinates, midpoint, height);
    },
};
