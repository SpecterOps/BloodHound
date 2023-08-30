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

export const vertexShaderSource = /* glsl */ `
attribute vec2 a_position;
attribute vec2 a_normal;
attribute float a_radius;
attribute vec4 a_color;
attribute vec3 a_barycentric;

uniform mat3 u_matrix;
uniform float u_sqrtZoomRatio;
uniform float u_correctionRatio;

varying vec4 v_color;

const float minThickness = 1.7;
const float bias = 255.0 / 254.0;
const float arrowHeadWidthLengthRatio = 0.75;
const float arrowHeadLengthThicknessRatio = 3.0;

void main() {
  float normalLength = length(a_normal);
  vec2 unitNormal = a_normal / normalLength;

  // These first computations are taken from edge.vert.glsl and
  // edge.clamped.vert.glsl. Please read it to get better comments on what's
  // happening:
  float pixelsThickness = max(normalLength, minThickness * u_sqrtZoomRatio);
  float webGLThickness = pixelsThickness * u_correctionRatio;
  float adaptedWebGLThickness = webGLThickness * u_sqrtZoomRatio;
  float adaptedWebGLNodeRadius = a_radius * 2.0 * u_correctionRatio * u_sqrtZoomRatio;
  float adaptedWebGLArrowHeadLength = adaptedWebGLThickness * 2.0 * arrowHeadLengthThicknessRatio;
  float adaptedWebGLArrowHeadHalfWidth = adaptedWebGLArrowHeadLength * arrowHeadWidthLengthRatio / 2.0;

  float da = a_barycentric.x;
  float db = a_barycentric.y;
  float dc = a_barycentric.z;

  vec2 delta = vec2(
      da * (adaptedWebGLNodeRadius * unitNormal.y)
    + db * ((adaptedWebGLNodeRadius + adaptedWebGLArrowHeadLength) * unitNormal.y + adaptedWebGLArrowHeadHalfWidth * unitNormal.x)
    + dc * ((adaptedWebGLNodeRadius + adaptedWebGLArrowHeadLength) * unitNormal.y - adaptedWebGLArrowHeadHalfWidth * unitNormal.x),

      da * (-adaptedWebGLNodeRadius * unitNormal.x)
    + db * (-(adaptedWebGLNodeRadius + adaptedWebGLArrowHeadLength) * unitNormal.x + adaptedWebGLArrowHeadHalfWidth * unitNormal.y)
    + dc * (-(adaptedWebGLNodeRadius + adaptedWebGLArrowHeadLength) * unitNormal.x - adaptedWebGLArrowHeadHalfWidth * unitNormal.y)
  );

  vec2 position = (u_matrix * vec3(a_position + delta, 1)).xy;

  gl_Position = vec4(position, 0, 1);

  // Extract the color:
  v_color = a_color;
  v_color.a *= bias;
}`;
