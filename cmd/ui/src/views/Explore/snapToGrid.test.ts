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

import { getOccupiedGridPoints, snapPositionsToGrid, snapPositionToGrid } from './snapToGrid';

describe('snapPositionsToGrid', () => {
    it('rounds positive and negative coordinates to the nearest grid point', () => {
        const positions = {
            positive: { x: 149, y: 150 },
            negative: { x: -149, y: -150 },
            origin: { x: -1, y: 1 },
        };

        expect(snapPositionsToGrid(positions, new Set(), 100)).toEqual({
            negative: { x: -100, y: -200 },
            origin: { x: 0, y: 0 },
            positive: { x: 100, y: 200 },
        });
        expect(positions.positive).toEqual({ x: 149, y: 150 });
    });

    it('uses deterministic neighboring points when nodes collide', () => {
        const positions = {
            alpha: { x: 10, y: 10 },
            bravo: { x: 20, y: 20 },
            charlie: { x: 30, y: 30 },
        };
        const reverseOrder = Object.fromEntries(Object.entries(positions).reverse());

        const snapped = snapPositionsToGrid(positions);

        expect(snapped).toEqual(snapPositionsToGrid(reverseOrder));
        expect(new Set(Object.values(snapped).map(({ x, y }) => `${x}:${y}`))).toHaveLength(3);
    });

    it('keeps fixed nodes in place and moves a dragged node to a free point', () => {
        const positions = {
            fixed: { x: 0, y: 0 },
            dragged: { x: 20, y: 20 },
        };

        expect(snapPositionsToGrid(positions, new Set(['fixed']), 100)).toEqual({
            fixed: { x: 0, y: 0 },
            dragged: { x: 0, y: -100 },
        });
    });

    it('allocates a dense collision group without duplicate cells', () => {
        const positions = Object.fromEntries(
            Array.from({ length: 5_000 }, (_, index) => [`node-${index.toString().padStart(4, '0')}`, { x: 1, y: 1 }])
        );

        const snapped = snapPositionsToGrid(positions);

        expect(new Set(Object.values(snapped).map(({ x, y }) => `${x}:${y}`))).toHaveLength(5_000);
    });

    it('normalizes non-finite coordinates without creating duplicate cells', () => {
        const snapped = snapPositionsToGrid({
            nan: { x: Number.NaN, y: Number.NaN },
            positiveInfinity: { x: Number.POSITIVE_INFINITY, y: Number.POSITIVE_INFINITY },
            negativeInfinity: { x: Number.NEGATIVE_INFINITY, y: Number.NEGATIVE_INFINITY },
        });

        expect(Object.values(snapped).every(({ x, y }) => Number.isFinite(x) && Number.isFinite(y))).toBe(true);
        expect(new Set(Object.values(snapped).map(({ x, y }) => `${x}:${y}`))).toHaveLength(3);
    });
});

describe('snapPositionToGrid', () => {
    it('snaps a dragged node around occupied graph positions', () => {
        const positions = {
            fixed: { x: 0, y: 0 },
            dragged: { x: 100, y: 0 },
        };
        const occupiedGridPoints = getOccupiedGridPoints(positions, new Set(['dragged']), 100);

        expect(snapPositionToGrid({ x: 20, y: 20 }, occupiedGridPoints, 100)).toEqual({ x: 0, y: -100 });
    });
});
