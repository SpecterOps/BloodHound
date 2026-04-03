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

import { AbstractEdgeProgram } from 'sigma/rendering/webgl/programs/common/edge';
import { RenderParams } from 'sigma/rendering/webgl/programs/common/program';
import vertexShaderSource from 'sigma/rendering/webgl/shaders/edge.clamped.vert.glsl';
import fragmentShaderSource from 'sigma/rendering/webgl/shaders/edge.frag.glsl';
import { EdgeDisplayData, NodeDisplayData } from 'sigma/types';
import { canUse32BitsIndices, floatColor } from 'sigma/utils';
import { getNodeRadius } from 'src/rendering/utils/utils';

const POINTS = 4,
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
        data: EdgeDisplayData & { inverseSqrtZoomRatio: number },
        hidden: boolean,
        offset: number
    ): void {
        if (hidden) {
            for (let i = offset * STRIDE, l = i + STRIDE; i < l; i++) this.array[i] = 0;
            return;
        }

        const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;

        const thickness = data.size || 1,
            x1 = sourceData.x,
            y1 = sourceData.y,
            x2 = targetData.x,
            y2 = targetData.y,
            color = floatColor(data.color);
        const radius = getNodeRadius(targetData.highlighted, inverseSqrtZoomRatio, targetData.size);

        // Computing normals
        const dx = x2 - x1,
            dy = y2 - y1;

        let len = dx * dx + dy * dy,
            n1 = 0,
            n2 = 0;

        if (len) {
            len = 1 / Math.sqrt(len);

            n1 = -dy * len * thickness;
            n2 = dx * len * thickness;
        }

        let i = POINTS * ATTRIBUTES * offset;

        const array = this.array;

        // First point
        array[i++] = x1;
        array[i++] = y1;
        array[i++] = n1;
        array[i++] = n2;
        array[i++] = color;
        array[i++] = 0;

        // First point flipped
        array[i++] = x1;
        array[i++] = y1;
        array[i++] = -n1;
        array[i++] = -n2;
        array[i++] = color;
        array[i++] = 0;

        // Second point
        array[i++] = x2;
        array[i++] = y2;
        array[i++] = n1;
        array[i++] = n2;
        array[i++] = color;
        array[i++] = radius;

        // Second point flipped
        array[i++] = x2;
        array[i++] = y2;
        array[i++] = -n1;
        array[i++] = -n2;
        array[i++] = color;
        array[i] = -radius;
    }

    computeIndices(): void {
        const l = this.array.length / ATTRIBUTES;
        const size = l + l / 2;
        const indices = new this.IndicesArray(size);

        for (let i = 0, c = 0; i < l; i += 4) {
            indices[c++] = i;
            indices[c++] = i + 1;
            indices[c++] = i + 2;
            indices[c++] = i + 2;
            indices[c++] = i + 1;
            indices[c++] = i + 3;
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
