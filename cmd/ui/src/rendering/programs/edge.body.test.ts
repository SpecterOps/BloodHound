// Copyright 2026 Specter Ops, Inc.
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

import { NodeDisplayData } from 'sigma/types';
import EdgeClampedProgram from './edge.clamped';
import CurvedEdgeProgram from './edge.curved';
import SelfEdgeProgram from './edge.self';

const source: NodeDisplayData = {
    x: 0,
    y: 0,
    size: 1,
    color: '#000000',
    label: '',
    hidden: false,
    highlighted: false,
    forceLabel: false,
    zIndex: 0,
    type: 'circle',
};
const target = { ...source, x: 1 };

// These tests exercise the real geometry writers with a mocked WebGL context.
describe.each([
    ['straight', EdgeClampedProgram],
    ['curved', CurvedEdgeProgram],
    ['self', SelfEdgeProgram],
] as const)('%s edge body', (_shape, Program) => {
    it.each([false, true])('preserves geometry and arc length anchored at the arrowhead with dashed=%s', (dashed) => {
        const gl = document.createElement('canvas').getContext('webgl')!;
        const program = new Program(gl);
        if (program instanceof CurvedEdgeProgram) program.correctionRatio = 0.0001;
        program.allocate(2);
        const stride = program.points * program.attributes;
        program.array.fill(123, 0, stride);
        program.process(
            source,
            target,
            {
                color: '#55595C',
                size: 2,
                label: '',
                hidden: false,
                forceLabel: false,
                zIndex: 0,
                type: 'arrow',
                dashed,
                inverseSqrtZoomRatio: 1,
                groupSize: 2,
                groupPosition: 0,
                direction: 1,
                framedGraphNodeRadius: 0.1,
            },
            false,
            1
        );

        expect(program.array.slice(0, stride).every((value) => value === 123)).toBe(true);
        let sourceDistance: number | undefined;
        let previousDistance = Infinity;
        let visibleVertices = 0;
        for (let index = stride; index < program.array.length; index += program.attributes) {
            if (program.array[index + 2] === 0 && program.array[index + 3] === 0) continue;
            expect(program.array[index + 6]).toBeLessThanOrEqual(previousDistance);
            previousDistance = program.array[index + 6];
            sourceDistance ??= previousDistance;
            expect(program.array[index + 7]).toBe(dashed ? 1 : 0);
            visibleVertices++;
        }
        expect(visibleVertices).toBeGreaterThanOrEqual(4);
        // Solid edges skip arc length measurement entirely.
        if (dashed) expect(sourceDistance).toBeGreaterThan(0);
        else expect(sourceDistance).toBe(0);
        expect(previousDistance).toBeCloseTo(0, 5);
        program.process(
            source,
            target,
            {
                color: '#55595C',
                size: 2,
                label: '',
                hidden: false,
                forceLabel: false,
                zIndex: 0,
                type: 'arrow',
                inverseSqrtZoomRatio: 1,
            },
            true,
            1
        );
        expect(program.array.slice(stride).every((value) => value === 0)).toBe(true);
    });
});

describe('edge body zoom', () => {
    it('keeps the curved edge dash count constant as arrowhead clamping changes', () => {
        const gl = document.createElement('canvas').getContext('webgl')!;
        const program = new CurvedEdgeProgram(gl);
        program.allocate(1);
        let referenceDistance: number | undefined;

        for (const ratio of [1, 1.5, 2, 3]) {
            program.correctionRatio = 0.001 / ratio;
            program.process(
                source,
                { ...target, size: 25 },
                {
                    color: '#55595C',
                    size: 4,
                    label: '',
                    hidden: false,
                    forceLabel: false,
                    zIndex: 0,
                    type: 'curved',
                    dashed: true,
                    groupSize: 2,
                    groupPosition: 0,
                    direction: 1,
                    inverseSqrtZoomRatio: 1 / Math.sqrt(ratio),
                },
                false,
                0
            );

            let endDistance = 0;
            for (let index = 6; index < program.array.length; index += program.attributes) {
                endDistance = Math.max(endDistance, program.array[index]);
            }
            expect(endDistance).toBeGreaterThan(0);
            referenceDistance ??= endDistance;
            expect(endDistance).toBeCloseTo(referenceDistance, 5);
        }
    });

    it('keeps the dash phase constant as the loop geometry follows the node radius', () => {
        const gl = document.createElement('canvas').getContext('webgl')!;
        const program = new SelfEdgeProgram(gl);
        program.allocate(1);
        let reference: Float32Array | undefined;

        for (const ratio of [1, 1.5, 2, 3]) {
            program.process(
                source,
                source,
                {
                    color: '#55595C',
                    size: 4,
                    dashed: true,
                    groupPosition: 0,
                    framedGraphNodeRadius: 0.1 / Math.sqrt(ratio),
                    inverseSqrtZoomRatio: 1 / Math.sqrt(ratio),
                },
                false,
                0
            );

            if (!reference) {
                reference = program.array.slice();
                continue;
            }

            // The shape shrinks, while every vertex retains its position in the dash pattern.
            const pointOffset = program.attributes * 2;
            expect(Math.hypot(program.array[pointOffset], program.array[pointOffset + 1])).toBeLessThan(
                Math.hypot(reference[pointOffset], reference[pointOffset + 1])
            );
            for (let index = 6; index < program.array.length; index += program.attributes) {
                expect(program.array[index]).toBeCloseTo(reference[index], 5);
            }
        }
    });
});
