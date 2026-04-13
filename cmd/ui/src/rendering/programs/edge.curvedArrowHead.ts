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
import { AbstractEdgeProgram } from 'sigma/rendering/webgl/programs/common/edge';
import { RenderParams } from 'sigma/rendering/webgl/programs/common/program';
import { NodeDisplayData } from 'sigma/types';
import { floatColor } from 'sigma/utils';
import { CurvedEdgeDisplayData } from 'src/rendering/programs/edge.curvedArrow';
import { fragmentShaderSource } from 'src/rendering/shaders/edge.arrowHead.frag';
import { vertexShaderSource } from 'src/rendering/shaders/edge.arrowHead.vert';
import { bezier } from 'src/rendering/utils/bezier';
import { getNodeRadius } from 'src/rendering/utils/utils';

const POINTS = 3,
    ATTRIBUTES = 9,
    STRIDE = POINTS * ATTRIBUTES;

export default class CurvedEdgeArrowHeadProgram extends AbstractEdgeProgram {
    // Locations
    positionLocation: GLint;
    colorLocation: GLint;
    normalLocation: GLint;
    radiusLocation: GLint;
    barycentricLocation: GLint;
    matrixLocation: WebGLUniformLocation;
    sqrtZoomRatioLocation: WebGLUniformLocation;
    correctionRatioLocation: WebGLUniformLocation;

    constructor(gl: WebGLRenderingContext) {
        super(gl, vertexShaderSource, fragmentShaderSource, POINTS, ATTRIBUTES);

        // Locations
        this.positionLocation = gl.getAttribLocation(this.program, 'a_position');
        this.colorLocation = gl.getAttribLocation(this.program, 'a_color');
        this.normalLocation = gl.getAttribLocation(this.program, 'a_normal');
        this.radiusLocation = gl.getAttribLocation(this.program, 'a_radius');
        this.barycentricLocation = gl.getAttribLocation(this.program, 'a_barycentric');

        // Uniform locations
        const matrixLocation = gl.getUniformLocation(this.program, 'u_matrix');
        if (matrixLocation === null) throw new Error('EdgeArrowHeadProgram: error while getting matrixLocation');
        this.matrixLocation = matrixLocation;

        const sqrtZoomRatioLocation = gl.getUniformLocation(this.program, 'u_sqrtZoomRatio');
        if (sqrtZoomRatioLocation === null)
            throw new Error('EdgeArrowHeadProgram: error while getting sqrtZoomRatioLocation');
        this.sqrtZoomRatioLocation = sqrtZoomRatioLocation;

        const correctionRatioLocation = gl.getUniformLocation(this.program, 'u_correctionRatio');
        if (correctionRatioLocation === null)
            throw new Error('EdgeArrowHeadProgram: error while getting correctionRatioLocation');
        this.correctionRatioLocation = correctionRatioLocation;

        this.bind();
    }

    bind(): void {
        const gl = this.gl;

        // Bindings
        gl.enableVertexAttribArray(this.positionLocation);
        gl.enableVertexAttribArray(this.normalLocation);
        gl.enableVertexAttribArray(this.radiusLocation);
        gl.enableVertexAttribArray(this.colorLocation);
        gl.enableVertexAttribArray(this.barycentricLocation);

        gl.vertexAttribPointer(
            this.positionLocation,
            2,
            gl.FLOAT,
            false,
            ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT,
            0
        );
        gl.vertexAttribPointer(this.normalLocation, 2, gl.FLOAT, false, ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT, 8);
        gl.vertexAttribPointer(
            this.radiusLocation,
            1,
            gl.FLOAT,
            false,
            ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT,
            16
        );
        gl.vertexAttribPointer(
            this.colorLocation,
            4,
            gl.UNSIGNED_BYTE,
            true,
            ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT,
            20
        );

        // TODO: maybe we can optimize here by packing this in a bit mask
        gl.vertexAttribPointer(
            this.barycentricLocation,
            3,
            gl.FLOAT,
            false,
            ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT,
            24
        );
    }

    computeIndices(): void {
        // nothing to do
    }

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

    process(
        sourceData: NodeDisplayData,
        targetData: NodeDisplayData,
        data: CurvedEdgeDisplayData,
        hidden: boolean,
        offset: number
    ): void {
        if (hidden) {
            for (let i = offset * STRIDE, l = i + STRIDE; i < l; i++) this.array[i] = 0;
            return;
        }

        const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;
        const thickness = data.size || 1;
        const radius = getNodeRadius(targetData.highlighted, inverseSqrtZoomRatio, targetData.size);
        const color = floatColor(data.color);

        // We are going to try and approximate the intersection here
        const height = bezier.calculateCurveHeight(data.groupSize, data.groupPosition, data.direction);
        const control = bezier.getControlAtMidpoint(height, sourceData, targetData);

        const sqrtZoomRatio = 1 / inverseSqrtZoomRatio;
        const correctionRatio = data.correctionRatio || 1;
        const graphSpaceRadius = targetData.size * 2.0 * correctionRatio * sqrtZoomRatio;
        const graphSpaceArrowHeadLength = thickness * 6 * correctionRatio * sqrtZoomRatio;

        let dx: number, dy: number;
        if (height !== 0) {
            let low = 0.5;
            let mid = 0.75;
            let high = 1;

            for (let i = 0; i < 10; i++) {
                mid = (low + high) / 2;
                const coords = bezier.getCoordinatesAlongQuadraticBezier(sourceData, targetData, control, mid);
                const distanceSquared = bezier.getDistanceSquared(coords, targetData);

                if (distanceSquared > graphSpaceRadius * graphSpaceRadius) {
                    low = mid;
                } else {
                    high = mid;
                }
            }

            const intersection = bezier.getCoordinatesAlongQuadraticBezier(sourceData, targetData, control, mid);

            let low2 = 0.5;
            let mid2 = mid + low2 * 0.5;
            let high2 = mid;

            for (let i = 0; i < 10; i++) {
                mid2 = (low2 + high2) / 2;
                const coords = bezier.getCoordinatesAlongQuadraticBezier(sourceData, targetData, control, mid2);
                const distanceSquared = bezier.getDistanceSquared(coords, intersection);

                if (distanceSquared > graphSpaceArrowHeadLength * graphSpaceArrowHeadLength) {
                    low2 = mid2;
                } else {
                    high2 = mid2;
                }
            }

            const endpoint = bezier.getCoordinatesAlongQuadraticBezier(sourceData, targetData, control, mid2);

            dx = intersection.x - endpoint.x;
            dy = intersection.y - endpoint.y;
        } else {
            dx = targetData.x - sourceData.x;
            dy = targetData.y - sourceData.y;
        }

        const len = 1 / Math.sqrt(dx * dx + dy * dy);
        const normal = { x: dx * len, y: dy * len };

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

    render(params: RenderParams): void {
        if (this.hasNothingToRender()) return;

        const gl = this.gl;

        const program = this.program;
        gl.useProgram(program);

        // Binding uniforms
        gl.uniformMatrix3fv(this.matrixLocation, false, params.matrix);
        gl.uniform1f(this.sqrtZoomRatioLocation, Math.sqrt(params.ratio));
        gl.uniform1f(this.correctionRatioLocation, params.correctionRatio);

        // Drawing:
        gl.drawArrays(gl.TRIANGLES, 0, this.array.length / ATTRIBUTES);
    }
}
