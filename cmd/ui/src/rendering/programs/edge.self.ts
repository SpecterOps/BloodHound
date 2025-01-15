// Copyright 2024 Specter Ops, Inc.
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

import CurvedEdgeProgram from './edge.curved';
import { Coordinates, NodeDisplayData } from 'sigma/types';
import { Attributes } from 'graphology-types';
import { floatColor } from 'sigma/utils';
import { bezier } from 'src/rendering/utils/bezier';

const RESOLUTION = 0.02,
    POINTS = 2 / RESOLUTION + 2,
    ATTRIBUTES = 6,
    STRIDE = POINTS * ATTRIBUTES;

export const getControlPointsFromGroupSize = (
    groupPosition: number,
    radius: number,
    center: Coordinates,
    invertY: boolean,
    invertX: boolean
) => {
    const position = groupPosition++;
    const step = Math.PI / 2;

    const theta2 = step * position - Math.PI / 2;
    const theta1 = theta2 + step;

    const x1Offset = radius * Math.cos(theta1);
    const y1Offset = radius * Math.sin(theta1);

    const x2Offset = radius * Math.cos(theta2);
    const y2Offset = radius * Math.sin(theta2);

    if (invertX && !invertY) {
        const control2 = { x: center.x + -1 * x1Offset, y: center.y + y1Offset };
        const control3 = { x: center.x + -1 * x2Offset, y: center.y + y2Offset };

        return { control2: control2, control3: control3 };
    }
    if (invertY && !invertX) {
        const control2 = { x: center.x + x1Offset, y: center.y + -1 * y1Offset };
        const control3 = { x: center.x + x2Offset, y: center.y + -1 * y2Offset };

        return { control2: control2, control3: control3 };
    } else if (invertY && invertX) {
        const control2 = { x: center.x + -1 * x1Offset, y: center.y + -1 * y1Offset };
        const control3 = { x: center.x + -1 * x2Offset, y: center.y + -1 * y2Offset };

        return { control2: control2, control3: control3 };
    } else {
        const control2 = { x: center.x + x1Offset, y: center.y + y1Offset };
        const control3 = { x: center.x + x2Offset, y: center.y + y2Offset };

        return { control2: control2, control3: control3 };
    }
};

export default class SelfEdgeProgram extends CurvedEdgeProgram {
    process(
        sourceData: NodeDisplayData,
        targetData: NodeDisplayData,
        data: Attributes,
        hidden: boolean,
        offset: number
    ): void {
        if (hidden) {
            for (let i = offset * STRIDE, l = i + STRIDE; i < l; i++) this.array[i] = 0;
            return;
        }

        const start = { x: sourceData.x, y: sourceData.y };
        const thickness = data.size || 1;

        const points = [];

        for (let t = 0; t <= 1; t += RESOLUTION) {
            const { control2, control3 } = getControlPointsFromGroupSize(
                data.groupPosition,
                data.framedGraphNodeRadius * 3,
                start,
                true,
                true
            );
            const pointOnCurve = bezier.getCoordinatesAlongCubicBezier(start, control2, control3, start, t);
            points.push(pointOnCurve);
        }

        let i = POINTS * ATTRIBUTES * offset;
        const array = this.array;
        const color = floatColor(data.color);

        for (let j = 0; j < points.length; j++) {
            // Handle special cases, since we do not need to calculate a miter join for the endcaps
            const isFirstPoint = j === 0;
            const isLastPoint = j === points.length - 1;

            let normal;

            if (isFirstPoint) {
                normal = bezier.getNormals(points[j], points[j + 1]);
            } else if (isLastPoint) {
                normal = bezier.getNormals(points[j - 1], points[j]);
            } else {
                // average normal vectors of the two adjoining line segments
                const firstNormal = bezier.getNormals(points[j - 1], points[j]);
                const secondNormal = bezier.getNormals(points[j], points[j + 1]);
                normal = bezier.getMidpoint(firstNormal, secondNormal);
            }

            const vOffset = {
                x: normal.x * thickness,
                y: -normal.y * thickness,
            };

            // First point
            array[i++] = points[j].x;
            array[i++] = points[j].y;
            array[i++] = vOffset.y;
            array[i++] = vOffset.x;
            array[i++] = color;
            array[i++] = 0;

            // First point flipped
            array[i++] = points[j].x;
            array[i++] = points[j].y;
            array[i++] = -vOffset.y;
            array[i++] = -vOffset.x;
            array[i++] = color;
            array[i++] = 0;
        }
    }
}
