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
 * This is pulled directly from Sigma, we are just using different shaders
 * @module
 */
import { AbstractEdgeProgram } from 'sigma/rendering/webgl/programs/common/edge';
import { RenderParams } from 'sigma/rendering/webgl/programs/common/program';
import { NodeDisplayData } from 'sigma/types';
import { floatColor } from 'sigma/utils';
import { CurvedEdgeDisplayData } from 'src/rendering/programs/edge.curvedArrow';
import { fragmentShaderSource } from 'src/rendering/shaders/edge.arrowHead.frag';
import { vertexShaderSource } from 'src/rendering/shaders/edge.arrowHead.vert';
import { getNodeRadius } from 'src/rendering/utils/utils';

const POINTS = 3,
    ATTRIBUTES = 9,
    STRIDE = POINTS * ATTRIBUTES;

export default class EdgeArrowHeadProgram extends AbstractEdgeProgram {
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
        array[i++] = x2;
        array[i++] = y2;
        array[i++] = -n1;
        array[i++] = -n2;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 1;
        array[i++] = 0;
        array[i++] = 0;

        // Second point
        array[i++] = x2;
        array[i++] = y2;
        array[i++] = -n1;
        array[i++] = -n2;
        array[i++] = radius;
        array[i++] = color;
        array[i++] = 0;
        array[i++] = 1;
        array[i++] = 0;

        // Third point
        array[i++] = x2;
        array[i++] = y2;
        array[i++] = -n1;
        array[i++] = -n2;
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
