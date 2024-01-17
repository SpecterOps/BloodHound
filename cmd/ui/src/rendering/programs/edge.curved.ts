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
 * Sigma.js WebGL Renderer Edge Program
 * =====================================
 *
 * Program rendering edges as thick lines but with a twist: the end of edge
 * does not sit in the middle of target node but instead stays by some margin.
 *
 * This is useful when combined with arrows to draw directed edges.
 * @module
 */
import { AbstractEdgeProgram } from 'sigma/rendering/webgl/programs/common/edge';
import { RenderParams } from 'sigma/rendering/webgl/programs/common/program';
import { NodeDisplayData } from 'sigma/types';
import { canUse32BitsIndices, floatColor } from 'sigma/utils';
import { CurvedEdgeDisplayData } from 'src/rendering/programs/edge.curvedArrow';
import { fragmentShaderSource } from 'src/rendering/shaders/edge.curved.frag';
import { vertexShaderSource } from 'src/rendering/shaders/edge.curved.vert';
import { bezier } from 'src/rendering/utils/bezier';

const RESOLUTION = 0.02,
    POINTS = 2 / RESOLUTION + 2,
    ATTRIBUTES = 6,
    STRIDE = POINTS * ATTRIBUTES;

export default class EdgeClampedProgram extends AbstractEdgeProgram {
    IndicesArray: Uint32ArrayConstructor | Uint16ArrayConstructor;
    indicesArray: Uint32Array | Uint16Array;
    indicesBuffer: WebGLBuffer;
    indicesType: GLenum;
    positionLocation: GLint;
    colorLocation: GLint;
    normalLocation: GLint;
    radiusLocation: GLint;
    matrixLocation: WebGLUniformLocation;
    sqrtZoomRatioLocation: WebGLUniformLocation;
    correctionRatioLocation: WebGLUniformLocation;
    canUse32BitsIndices: boolean;

    constructor(gl: WebGLRenderingContext) {
        super(gl, vertexShaderSource, fragmentShaderSource, POINTS, ATTRIBUTES);

        // Initializing indices buffer
        const indicesBuffer = gl.createBuffer();
        if (indicesBuffer === null) throw new Error('EdgeClampedProgram: error while getting resolutionLocation');
        this.indicesBuffer = indicesBuffer;

        // Locations:
        this.positionLocation = gl.getAttribLocation(this.program, 'a_position');
        this.colorLocation = gl.getAttribLocation(this.program, 'a_color');
        this.normalLocation = gl.getAttribLocation(this.program, 'a_normal');
        this.radiusLocation = gl.getAttribLocation(this.program, 'a_radius');

        // Uniform locations
        const matrixLocation = gl.getUniformLocation(this.program, 'u_matrix');
        if (matrixLocation === null) throw new Error('EdgeClampedProgram: error while getting matrixLocation');
        this.matrixLocation = matrixLocation;

        const sqrtZoomRatioLocation = gl.getUniformLocation(this.program, 'u_sqrtZoomRatio');
        if (sqrtZoomRatioLocation === null)
            throw new Error('EdgeClampedProgram: error while getting cameraRatioLocation');
        this.sqrtZoomRatioLocation = sqrtZoomRatioLocation;

        const correctionRatioLocation = gl.getUniformLocation(this.program, 'u_correctionRatio');
        if (correctionRatioLocation === null)
            throw new Error('EdgeClampedProgram: error while getting viewportRatioLocation');
        this.correctionRatioLocation = correctionRatioLocation;

        // Enabling the OES_element_index_uint extension
        // NOTE: on older GPUs, this means that really large graphs won't
        // have all their edges rendered. But it seems that the
        // `OES_element_index_uint` is quite everywhere so we'll handle
        // the potential issue if it really arises.
        // NOTE: when using webgl2, the extension is enabled by default
        this.canUse32BitsIndices = canUse32BitsIndices(gl);
        this.IndicesArray = this.canUse32BitsIndices ? Uint32Array : Uint16Array;
        this.indicesArray = new this.IndicesArray();
        this.indicesType = this.canUse32BitsIndices ? gl.UNSIGNED_INT : gl.UNSIGNED_SHORT;

        this.bind();
    }

    bind(): void {
        const gl = this.gl;

        gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.indicesBuffer);

        // Bindings
        gl.enableVertexAttribArray(this.positionLocation);
        gl.enableVertexAttribArray(this.normalLocation);
        gl.enableVertexAttribArray(this.colorLocation);
        gl.enableVertexAttribArray(this.radiusLocation);

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
            this.colorLocation,
            4,
            gl.UNSIGNED_BYTE,
            true,
            ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT,
            16
        );
        gl.vertexAttribPointer(
            this.radiusLocation,
            1,
            gl.FLOAT,
            false,
            ATTRIBUTES * Float32Array.BYTES_PER_ELEMENT,
            20
        );
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

        const start = { x: sourceData.x, y: sourceData.y };
        const end = { x: targetData.x, y: targetData.y };
        const thickness = data.size || 1;

        // 1. Calculate a control point for this edge

        let { control } = data;

        if (!control) {
            const height = bezier.calculateCurveHeight(data.groupSize, data.groupPosition, data.direction);
            control = bezier.getControlAtMidpoint(height, start, end);
        }

        // 2. Get collection of points at some constant resolution distributed on quadratic curve based on
        // starting point, ending point, previously calculated control point

        const points = [];

        for (let t = 0; t <= 1; t += RESOLUTION) {
            const pointOnCurve = bezier.getCoordinatesAlongCurve(start, end, control, t);
            points.push(pointOnCurve);
        }

        // 3. Loop through each line segment, calculate normals so we can render each segment as two triangles, then
        // add data to this.array

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

    computeIndices(): void {
        // Calculate index array here that allows us to share vertices between triangles and save on shader calls
        const l = this.array.length / ATTRIBUTES;
        const size = l * 3;
        const indices = new this.IndicesArray(size);

        for (let i = 0, c = 0; i + 3 < l; i += 2) {
            indices[c++] = i;
            indices[c++] = i + 1;
            indices[c++] = i + 2;
            indices[c++] = i + 2;
            indices[c++] = i + 1;
            indices[c++] = i + 3;

            // Disconnect if it is the last line segment in the curve
            const isLastSegment = (c / 6) % (RESOLUTION - 1) === 0;
            if (isLastSegment) i += 2;
        }

        this.indicesArray = indices;
    }

    bufferData(): void {
        super.bufferData();

        // Indices data
        const gl = this.gl;
        gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, this.indicesArray, gl.STATIC_DRAW);
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
        gl.drawElements(gl.TRIANGLES, this.indicesArray.length, this.indicesType, 0);
    }
}
