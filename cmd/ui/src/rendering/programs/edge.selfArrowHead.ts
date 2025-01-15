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

import CurvedEdgeArrowHeadProgram from './edge.curvedArrowHead';
import { NodeDisplayData } from 'sigma/types';
import { floatColor } from 'sigma/utils';
import { Attributes } from 'graphology-types';
import { bezier } from 'src/rendering/utils/bezier';
import { getNodeRadius } from 'src/rendering/utils/utils';
import { getControlPointsFromGroupSize } from './edge.self';

const POINTS = 3,
    ATTRIBUTES = 9,
    INTERSECT_APPROXIMATION_T = 0.8975,
    STRIDE = POINTS * ATTRIBUTES;

export default class SelfEdgeArrowHeadProgram extends CurvedEdgeArrowHeadProgram {
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

        const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;
        const start = { x: sourceData.x, y: sourceData.y };
        const thickness = data.size || 1;
        const radius = getNodeRadius(targetData.highlighted, inverseSqrtZoomRatio, targetData.size);
        const color = floatColor(data.color);
        const { control2, control3 } = getControlPointsFromGroupSize(
            data.groupPosition,
            data.framedGraphNodeRadius * 3,
            start,
            false,
            false
        );

        const curveCircleIntersectionApproximation = bezier.getCoordinatesAlongCubicBezier(
            start,
            control2,
            control3,
            start,
            INTERSECT_APPROXIMATION_T
        );

        const normal = bezier.getNormals(start, curveCircleIntersectionApproximation);

        const vOffset = {
            x: normal.x * thickness * inverseSqrtZoomRatio,
            y: -normal.y * thickness * inverseSqrtZoomRatio,
        };

        let i = POINTS * ATTRIBUTES * offset;

        const array = this.array;

        // First point
        array[i++] = start.x;
        array[i++] = start.y;
        array[i++] = -vOffset.y;
        array[i++] = -vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 1;
        array[i++] = 0;
        array[i++] = 0;

        // Second point
        array[i++] = start.x;
        array[i++] = start.y;
        array[i++] = -vOffset.y;
        array[i++] = -vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 1;
        array[i++] = 0;

        // Third point
        array[i++] = start.x;
        array[i++] = start.y;
        array[i++] = -vOffset.y;
        array[i++] = -vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 0;
        array[i] = 1;
    }
}
