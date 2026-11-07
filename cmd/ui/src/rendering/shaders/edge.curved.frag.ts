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

export const fragmentShaderSource = /* glsl */ `
precision mediump float;

varying vec4 v_color;
varying vec2 v_normal;
varying float v_thickness;
varying highp float v_distance;
varying float v_dashed;

const float feather = 0.001;
const vec4 transparent = vec4(0.0, 0.0, 0.0, 0.0);

void main(void) {
  // Eight-pixel dashes and six-pixel gaps at the default zoom; arrowheads stay solid.
  if (v_dashed > 0.5 && mod(v_distance, 14.0) >= 8.0) discard;

  float dist = length(v_normal) * v_thickness;

  float t = smoothstep(
    v_thickness - feather,
    v_thickness,
    dist
  );

  gl_FragColor = mix(v_color, transparent, t);
}
`;
