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

/**
 * Sigma.js WebGL Renderer Arrow Program
 * ======================================
 *
 * Program rendering direction arrows as a simple triangle.
 * @module
 */
import { Attributes } from 'graphology-types';
import { Coordinates, NodeDisplayData } from 'sigma/types';
import { floatColor } from 'sigma/utils';
import EdgeArrowHeadProgram from 'src/rendering/programs/edge.arrowHead';
import { BezierCoordinate, bezier } from 'src/rendering/utils/bezier';
import { getCurveCircleIntersection, getNodeRadius } from 'src/rendering/utils/utils';
import { EdgeDirection } from 'src/utils';
import { defaultEdgeColor } from 'src/views/Explore/utils';

const color = floatColor(defaultEdgeColor);

const POINTS = 3,
    ATTRIBUTES = 9,
    STRIDE = POINTS * ATTRIBUTES;

export default class CurvedEdgeArrowHeadProgram extends EdgeArrowHeadProgram {
    // If the arrow sits right along the line between the node and control point, it never quite lines up correctly.
    // This allows us to add a standard adjustment value to the control point height that works for most curve lengths,
    // then handle the special case of very short curve lengths.
    calculateAdjustmentFactor(distanceBetweenNodes: number): number {
        const startingValue = 0.007;

        if (distanceBetweenNodes >= 0.1) {
            return startingValue;
        }

        return startingValue + (0.1 - distanceBetweenNodes) * 0.15;
    }

    processFast(
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

        let start, end: Coordinates;
        if (data.direction === EdgeDirection.BACKWARDS) {
            end = { x: sourceData.x, y: sourceData.y };
            start = { x: targetData.x, y: targetData.y };
        } else {
            start = { x: sourceData.x, y: sourceData.y };
            end = { x: targetData.x, y: targetData.y };
        }

        const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;
        const thickness = data.size || 1;
        const radius = getNodeRadius(targetData.highlighted, inverseSqrtZoomRatio, targetData.size);

        // We are going to try and approximate the intersection here
        const height = bezier.calculateCurveHeight(data.groupSize, data.groupPosition);
        let adjustedHeight = 0;

        if (height !== 0) {
            const distanceBetweenNodes = bezier.getLineLength(start, end);
            const adjustmentFactor = this.calculateAdjustmentFactor(distanceBetweenNodes);
            adjustedHeight = Math.abs(height) - adjustmentFactor;
        }

        if (height < 0) {
            adjustedHeight *= -1;
        }

        const control = bezier.getControlAtMidpoint(adjustedHeight, start, end);

        const normal = bezier.getNormals(control, targetData);

        const vOffset = {
            x: normal.x * thickness,
            y: -normal.y * thickness,
        };

        let i = POINTS * ATTRIBUTES * offset;

        const array = this.array;

        // First point
        array[i++] = targetData.x;
        array[i++] = targetData.y;
        array[i++] = -vOffset.y;
        array[i++] = -vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 1;
        array[i++] = 0;
        array[i++] = 0;

        // Second point
        array[i++] = targetData.x;
        array[i++] = targetData.y;
        array[i++] = -vOffset.y;
        array[i++] = -vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 1;
        array[i++] = 0;

        // Third point
        array[i++] = targetData.x;
        array[i++] = targetData.y;
        array[i++] = -vOffset.y;
        array[i++] = -vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 0;
        array[i] = 1;
    }

    processFine(
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

        let start, end: Coordinates;
        if (data.direction === EdgeDirection.BACKWARDS) {
            end = { x: sourceData.x, y: sourceData.y };
            start = { x: targetData.x, y: targetData.y };
        } else {
            start = { x: sourceData.x, y: sourceData.y };
            end = { x: targetData.x, y: targetData.y };
        }

        const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;
        const thickness = data.size || 1;

        const height = bezier.calculateCurveHeight(data.groupSize, data.groupPosition);
        const control = bezier.getControlAtMidpoint(height, start, end);
        const radius = data.framedGraphNodeRadius;

        const LUT = bezier.getLUT([start, end, control], 50);
        const tValues = getCurveCircleIntersection(LUT, targetData, radius * inverseSqrtZoomRatio);
        let intersection: BezierCoordinate;

        if (tValues.length) {
            intersection = LUT[tValues[0]];
        } else {
            const tValuesB = getCurveCircleIntersection(LUT, sourceData, radius * inverseSqrtZoomRatio);
            if (!tValuesB.length) {
                console.warn('curve circle intersection not found for drawing arrowhead');
                return;
            }
            const inverseT = 1 - LUT[tValuesB[0]].t;
            intersection = bezier.getCoordinatesAlongQuadraticBezier(start, end, control, inverseT);
        }

        const normal = bezier.getNormals(intersection, targetData);
        const vOffset = {
            x: -normal.x * thickness,
            y: normal.y * thickness,
        };

        let i = STRIDE * offset;
        const array = this.array;

        // First point
        array[i++] = intersection.x;
        array[i++] = intersection.y;
        array[i++] = vOffset.y;
        array[i++] = vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 1;
        array[i++] = 0;
        array[i++] = 0;

        // Second point
        array[i++] = intersection.x;
        array[i++] = intersection.y;
        array[i++] = vOffset.y;
        array[i++] = vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 1;
        array[i++] = 0;

        // Third point
        array[i++] = intersection.x;
        array[i++] = intersection.y;
        array[i++] = vOffset.y;
        array[i++] = vOffset.x;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 0;
        array[i] = 1;
    }

    process(
        sourceData: NodeDisplayData,
        targetData: NodeDisplayData,
        data: Attributes,
        hidden: boolean,
        offset: number
    ): void {
        if (data.needsPerformance) this.processFast(sourceData, targetData, data, hidden, offset);
        else this.processFine(sourceData, targetData, data, hidden, offset);
    }
}
