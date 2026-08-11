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

import type { Coordinates } from 'sigma/types';

export const SNAP_TO_GRID_SIZE = 100;

const roundToGrid = (value: number, gridSize: number) => {
    const scaledValue = value / gridSize;
    const roundedValue = scaledValue >= 0 ? Math.floor(scaledValue + 0.5) : Math.ceil(scaledValue - 0.5);
    const result = roundedValue * gridSize;

    return Object.is(result, -0) ? 0 : result;
};

const getGridKey = (x: number, y: number) => `${x}:${y}`;

function* getGridOffsets(): Generator<[number, number], never> {
    yield [0, 0];

    for (let radius = 1; ; radius++) {
        for (let x = -radius; x <= radius; x++) yield [x, -radius];
        for (let y = -radius + 1; y <= radius; y++) yield [radius, y];
        for (let x = radius - 1; x >= -radius; x--) yield [x, radius];
        for (let y = radius - 1; y > -radius; y--) yield [-radius, y];
    }
}

const findAvailableGridPosition = (
    position: Coordinates,
    occupiedGridPoints: Set<string>,
    gridSize: number,
    offsets: Generator<[number, number], never>
): Coordinates => {
    const targetX = roundToGrid(position.x, gridSize);
    const targetY = roundToGrid(position.y, gridSize);

    while (true) {
        const [offsetX, offsetY] = offsets.next().value;
        const x = targetX + offsetX * gridSize;
        const y = targetY + offsetY * gridSize;

        if (!occupiedGridPoints.has(getGridKey(x, y))) return { x, y };
    }
};

export const getOccupiedGridPoints = (
    positions: Record<string, Coordinates>,
    excludedIds: ReadonlySet<string> = new Set(),
    gridSize = SNAP_TO_GRID_SIZE
) =>
    new Set(
        Object.entries(positions)
            .filter(([id]) => !excludedIds.has(id))
            .map(([, position]) => getGridKey(roundToGrid(position.x, gridSize), roundToGrid(position.y, gridSize)))
    );

export const snapPositionToGrid = (
    position: Coordinates,
    occupiedGridPoints: Set<string>,
    gridSize = SNAP_TO_GRID_SIZE
) => findAvailableGridPosition(position, occupiedGridPoints, gridSize, getGridOffsets());

export const snapPositionsToGrid = (
    positions: Record<string, Coordinates>,
    fixedNodeIds: ReadonlySet<string> = new Set(),
    gridSize = SNAP_TO_GRID_SIZE
) => {
    const snappedPositions: Record<string, Coordinates> = {};
    const occupiedGridPoints = new Set<string>();
    const offsetSearches = new Map<string, Generator<[number, number], never>>();
    const positionIds = Object.keys(positions).sort();

    for (const id of positionIds) {
        if (!fixedNodeIds.has(id)) continue;

        const position = positions[id];
        snappedPositions[id] = { ...position };
        occupiedGridPoints.add(getGridKey(roundToGrid(position.x, gridSize), roundToGrid(position.y, gridSize)));
    }

    for (const id of positionIds) {
        if (fixedNodeIds.has(id)) continue;

        const position = positions[id];
        const targetKey = getGridKey(roundToGrid(position.x, gridSize), roundToGrid(position.y, gridSize));
        const offsets = offsetSearches.get(targetKey) ?? getGridOffsets();
        offsetSearches.set(targetKey, offsets);

        const snappedPosition = findAvailableGridPosition(position, occupiedGridPoints, gridSize, offsets);
        snappedPositions[id] = snappedPosition;
        occupiedGridPoints.add(getGridKey(snappedPosition.x, snappedPosition.y));
    }

    return snappedPositions;
};
