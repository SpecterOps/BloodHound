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

import getNodeCombinedProgram from './node.combined';
import getNodeGlyphsProgram, { GlyphLocation } from './node.glyphs';

it.each([
    ['ordinary nodes', getNodeCombinedProgram, 3 * 11],
    ['glyph nodes', () => getNodeGlyphsProgram(6), 3 * 13 * 7],
] as const)('clears every hidden triangle for %s without touching the next node', (_name, factory, stride) => {
    const Program = factory();
    const program = Object.create(Program.prototype) as InstanceType<typeof Program>;
    const array = new Float32Array(stride * 2).fill(123);
    Object.assign(program, { array });
    const data = {
        x: 10,
        y: 20,
        size: 25,
        color: '#ffffff',
        label: 'leaf',
        hidden: false,
        forceLabel: false,
        highlighted: false,
        zIndex: 0,
        type: 'glyphs',
        isDimmed: false,
        glyphs: [{ location: GlyphLocation.TOP_RIGHT, image: '', backgroundColor: '#000000', color: '#ffffff' }],
    };
    program.process(data, true, 0);
    expect([...array.slice(0, stride)]).toEqual(new Array(stride).fill(0));
    expect([...array.slice(stride)]).toEqual(new Array(stride).fill(123));
});
