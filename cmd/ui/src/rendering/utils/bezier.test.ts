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

import { bezier } from './bezier';

describe('bezier', () => {
    const startpoint1 = { x: 0, y: 0 };
    const endpoint1 = { x: 100, y: 100 };
    const offset1 = 20;

    const startpoint2 = { x: 20, y: 30 };
    const endpoint2 = { x: 40, y: 100 };
    const offset2 = 10;

    it('should compute normal vectors for line segments', () => {
        let normal;

        normal = bezier.getNormals(startpoint1, endpoint1);
        expect(normal.x).toBeCloseTo(0.7071);
        expect(normal.y).toBeCloseTo(0.7071);

        normal = bezier.getNormals(startpoint2, endpoint2);
        expect(normal.x).toBeCloseTo(0.2747);
        expect(normal.y).toBeCloseTo(0.9615);
    });

    it('should compute offset vertices for bezier curves at a given intersection', () => {
        let midpoint, control;

        midpoint = bezier.getMidpoint(startpoint1, endpoint1);
        control = bezier.getOffsetVertex(startpoint1, endpoint1, midpoint, offset1);
        expect(control.x).toBeCloseTo(64.1421);
        expect(control.y).toBeCloseTo(35.8579);

        midpoint = bezier.getMidpoint(startpoint2, endpoint2);
        control = bezier.getOffsetVertex(startpoint2, endpoint2, midpoint, offset2);
        expect(control.x).toBeCloseTo(39.6152);
        expect(control.y).toBeCloseTo(62.2528);
    });

    it('should compute coordinates along a quadratic bezier curve for a given t value', () => {
        let midpoint, control, t, point;

        midpoint = bezier.getMidpoint(startpoint1, endpoint1);
        control = bezier.getOffsetVertex(startpoint1, endpoint1, midpoint, offset1);

        t = 0;
        point = bezier.getCoordinatesAlongCurve(startpoint1, endpoint1, control, t);
        expect(point.x).toBeCloseTo(startpoint1.x);
        expect(point.y).toBeCloseTo(startpoint1.y);

        t = 0.5;
        point = bezier.getCoordinatesAlongCurve(startpoint1, endpoint1, control, t);
        expect(point.x).toBeCloseTo(57.0711);
        expect(point.y).toBeCloseTo(42.9289);

        t = 1;
        point = bezier.getCoordinatesAlongCurve(startpoint1, endpoint1, control, t);
        expect(point.x).toBeCloseTo(endpoint1.x);
        expect(point.y).toBeCloseTo(endpoint1.y);

        midpoint = bezier.getMidpoint(startpoint2, endpoint2);
        control = bezier.getOffsetVertex(startpoint2, endpoint2, midpoint, offset2);

        t = 0;
        point = bezier.getCoordinatesAlongCurve(startpoint2, endpoint2, control, t);
        expect(point.x).toBeCloseTo(startpoint2.x);
        expect(point.y).toBeCloseTo(startpoint2.y);

        t = 0.5;
        point = bezier.getCoordinatesAlongCurve(startpoint2, endpoint2, control, t);
        expect(point.x).toBeCloseTo(34.8076);
        expect(point.y).toBeCloseTo(63.6264);

        t = 1;
        point = bezier.getCoordinatesAlongCurve(startpoint2, endpoint2, control, t);
        expect(point.x).toBeCloseTo(endpoint2.x);
        expect(point.y).toBeCloseTo(endpoint2.y);
    });

    it('should compute the midpoint of a line segment', () => {
        let midpoint;

        midpoint = bezier.getMidpoint(startpoint1, endpoint1);
        expect(midpoint.x).toBeCloseTo(50);
        expect(midpoint.y).toBeCloseTo(50);

        midpoint = bezier.getMidpoint(startpoint2, endpoint2);
        expect(midpoint.x).toBeCloseTo(30);
        expect(midpoint.y).toBeCloseTo(65);
    });

    it('should compute the length of a line segment', () => {
        let length;

        length = bezier.getLineLength(startpoint1, endpoint1);
        expect(length).toBeCloseTo(141.4214);

        length = bezier.getLineLength(startpoint2, endpoint2);
        expect(length).toBeCloseTo(72.8011);
    });
});
