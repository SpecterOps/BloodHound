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
varying float v_border;
varying vec4 v_borderColor;
varying float v_softborder;
varying vec4 v_texture;
varying vec2 v_diffVector;
varying float v_radius;
varying float v_angle;
varying vec2 v_position;



uniform sampler2D u_atlas;

const float radius = 0.5;
const vec4 transparent = vec4(0.0, 0.0, 0.0, 0.0);

mat2 rotate2d(float _angle){
    return mat2(cos(_angle),-sin(_angle),
                sin(_angle),cos(_angle));
}

void main(void) {
  vec4 transparent = vec4(0.0, 0.0, 0.0, 0.0);
  vec4 color;

  if (v_texture.z > 0.0) {
    vec4 texel = texture2D(u_atlas, v_texture.xy, -1.0);
    color = vec4(mix(v_color, texel, texel.a).rgb, max(texel.a, v_color.a));
  } else {
    color = v_color;
  }

  float dist = length(v_diffVector) - v_radius;
  vec4 border_color = v_borderColor;

  if (dist > v_softborder)
    // outside border
    gl_FragColor = transparent;
  else if (dist > 0.0)
    // outside border antialias
    gl_FragColor = mix(border_color, transparent, dist / v_softborder);
  else if (dist > v_radius - v_border + v_softborder)
    // inner border
    gl_FragColor = mix(color, border_color, border_color.a);
  else if (dist > v_radius - v_border)
    // inner border antialias
    gl_FragColor = mix(color, border_color, (dist - v_radius + v_border) / v_softborder);
  else
    gl_FragColor = color;
}`;
