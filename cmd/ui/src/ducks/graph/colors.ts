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

import clamp from 'lodash/clamp';

function hslToRgb(h: number, s: number, l: number) {
    let r, g, b;

    if (s === 0) {
        r = g = b = l; // achromatic
    } else {
        const hue2rgb = function hue2rgb(p: number, q: number, t: number) {
            if (t < 0) t += 1;
            if (t > 1) t -= 1;
            if (t < 1 / 6) return p + (q - p) * 6 * t;
            if (t < 1 / 2) return q;
            if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
            return p;
        };

        const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        const p = 2 * l - q;
        r = hue2rgb(p, q, h + 1 / 3);
        g = hue2rgb(p, q, h);
        b = hue2rgb(p, q, h - 1 / 3);
    }

    return [Math.round(r * 255), Math.round(g * 255), Math.round(b * 255)];
}

export const mapToRange = (val: number, source: [number, number], target: [number, number]) => {
    const factor = (val - source[0]) / (source[1] - source[0]);
    const distance = (target[1] - target[0]) * factor;
    return target[0] + distance;
};

enum Palette360 {
    YELLOW = 58,
    ORANGE = 24,
    RED = 0,
    PURPLE = 268,
}

const hue360to1 = (hue: number) => mapToRange(hue, [0, 360], [0, 1]);

// -------------------------------
const thresholds: [number, number, number, string][] = [
    [0.4, hue360to1(Palette360.YELLOW), 5, 'LOW'],
    [0.8, hue360to1(Palette360.ORANGE), 10, 'MODERATE'],
    [0.95, hue360to1(Palette360.RED), 15, 'HIGH'],
    [1, hue360to1(Palette360.PURPLE), 20, 'CRITICAL'],
];

const toRgb: (value: number, luminance?: number) => [number, number, number] = (value, luminance = 0.5) => {
    const [r, g, b] = hslToRgb(value, 1, luminance);
    return [r, g, b];
};

const toGrade: (value: number) => number = (value) => {
    const boundedValue = clamp(value, 0, 1);
    const lower = thresholds.findIndex((threshold) => threshold[0] > boundedValue);
    return lower === -1 ? (boundedValue === 1 ? thresholds.length - 1 : 0) : lower;
};

const toHue: (value: number) => number = (value) => {
    const boundedValue = clamp(value, 0, 1);
    return thresholds[toGrade(boundedValue)][1];
};

export const getGrade: (value: number, range?: [number, number]) => number = (value, range = [0, 1]) => {
    return toGrade(mapToRange(value, range, [0, 1]));
};

export const toColor: (value: number, range?: [number, number], luminance?: number) => string = (
    value,
    range = [0, 1],
    luminance
) => {
    const [r, g, b] = toRgb(toHue(mapToRange(value, range, [0, 1])), luminance);
    return `rgba(${r}, ${g}, ${b}, 1)`;
};

export const gradeToColor: (grade: number) => string = (grade) => {
    const [r, g, b] = toRgb(thresholds[grade][1]);
    return `rgba(${r}, ${g}, ${b}, 1)`;
};

export const toGraphLinkColor: (value: number, range?: [number, number], luminance?: number) => string = (
    value,
    range = [0, 1],
    luminance
) => {
    const grade: number = toGrade(mapToRange(value, range, [0, 1]));
    switch (grade) {
        case 0:
            return '#444';
        case 1:
            return '#faf263';
        case 3:
            return '#e84a49';
        default: {
            const [r, g, b] = toRgb(toHue(mapToRange(value, range, [0, 1])), luminance);
            return `rgba(${r}, ${g}, ${b}, 1)`;
        }
    }
};

export const percentToColor: (value: number, luminance?: number) => string = (value, luminance = 0.75) =>
    toColor(value, [0, 100], luminance);

export const percentToContrastColor: (value: number) => string = (value) => {
    if (value >= 80) {
        return '#fff';
    }
    return '#000';
};

export const percentToTitle: (value: number) => string = (value) => toTitle(value, [0, 100]);

export const toWidth: (value: number, range?: [number, number]) => number = (value, range = [0, 1]) => {
    const mapped: number = mapToRange(value, range, [0, 1]);
    return thresholds[toGrade(mapped)][2];
};

const toTitle: (value: number, range?: [number, number]) => string = (value, range = [0, 1]) => {
    const mapped: number = mapToRange(value, range, [0, 1]);
    return thresholds[toGrade(mapped)][3];
};
